// Copyright 2023 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"
	"github.com/mendersoftware/mender-server/pkg/stream"
	"github.com/mendersoftware/mender-server/pkg/ws"

	"github.com/mendersoftware/mender-server/services/deviceconnect/app"
	"github.com/mendersoftware/mender-server/services/deviceconnect/client/nats"
	"github.com/mendersoftware/mender-server/services/deviceconnect/model"
)

var (
	// Time allowed to read the next pong message from the peer.
	pongWait = time.Minute
	// Seconds allowed to write a message to the peer.
	writeWait = time.Second * 10
)

// HTTP errors
var (
	ErrMissingAuthentication = errors.New(
		"missing or non-device identity in the authorization headers",
	)
)

// DeviceController container for end-points
type DeviceController struct {
	app  app.App
	nats nats.Client
}

// NewDeviceController returns a new DeviceController
func NewDeviceController(
	app app.App,
	natsClient nats.Client,
) *DeviceController {
	return &DeviceController{
		app:  app,
		nats: natsClient,
	}
}

// Provision responds to POST /tenants/:tenantId/devices
func (h DeviceController) Provision(c *gin.Context) {
	tenantID := c.Param("tenantId")

	rawData, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "bad request",
		})
		return
	}

	device := &model.Device{}
	if err = json.Unmarshal(rawData, device); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.Wrap(err, "invalid payload").Error(),
		})
		return
	} else if device.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "device_id is empty",
		})
		return
	}

	ctx := c.Request.Context()
	if err = h.app.ProvisionDevice(ctx, tenantID, device); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.Wrap(err, "error provisioning the device").Error(),
		})
		return
	}

	c.Writer.WriteHeader(http.StatusCreated)
}

// Delete responds to DELETE /tenants/:tenantId/devices/:deviceId
func (h DeviceController) Delete(c *gin.Context) {
	tenantID := c.Param("tenantId")
	deviceID := c.Param("deviceId")

	ctx := c.Request.Context()
	if err := h.app.DeleteDevice(ctx, tenantID, deviceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.Wrap(err, "error deleting the device").Error(),
		})
		return
	}

	c.Writer.WriteHeader(http.StatusAccepted)
}

// Connect starts a websocket connection with the device
func (h DeviceController) Connect(c *gin.Context) {
	ctx := c.Request.Context()
	l := log.FromContext(ctx)

	idata := identity.FromContext(ctx)
	if !idata.IsDevice {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": ErrMissingAuthentication.Error(),
		})
		return
	}

	listener, err := h.nats.Listen(idata.Subject)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to allocate internal device channel",
		})
		return
	}
	//nolint:errcheck
	defer listener.Close(ctx)

	upgrader := websocket.Upgrader{
		Subprotocols: []string{"protomsg/msgpack"},
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		Error: func(
			w http.ResponseWriter, r *http.Request, s int, e error) {
			rest.RenderError(c, s, e)
		},
	}

	// upgrade get request to websocket protocol
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		err = errors.Wrap(err,
			"failed to upgrade the request to "+
				"websocket protocol",
		)
		l.Error(err)
		return
	}
	conn.SetReadLimit(int64(app.MessageSizeLimit))

	// register the websocket for graceful shutdown
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()
	registerID := h.app.RegisterShutdownCancel(cancel)
	defer h.app.UnregisterShutdownCancel(registerID)

	var version int64
	version, err = h.app.SetDeviceConnected(ctx, idata.Tenant, idata.Subject)
	if err != nil {
		return
	}
	defer func() {
		// update the device status on websocket closing
		disconnectCtx, disconnectCancel := context.WithTimeout(
			context.WithoutCancel(ctx),
			time.Second*10,
		)
		defer disconnectCancel()
		eStatus := h.app.SetDeviceDisconnected(
			disconnectCtx, idata.Tenant,
			idata.Subject, version,
		)
		if eStatus != nil {
			l.Error(eStatus)
		}
	}()

	// websocketWriter is responsible for closing the websocket
	//nolint:errcheck
	err = h.connectWSWriter(ctxWithCancel, conn, listener)
	if err != nil && !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		_ = c.Error(err)
	}
}

type streamMap sync.Map

func (m *streamMap) Load(remoteAddr string) stream.Conn {
	sm := (*sync.Map)(m)
	conn, ok := sm.Load(remoteAddr)
	if ok {
		return conn.(stream.Conn)
	}
	return nil
}

func (m *streamMap) Store(conn stream.Conn) bool {
	sm := (*sync.Map)(m)
	_, loaded := sm.LoadOrStore(conn.RemoteAddr(), conn)
	return !loaded
}

func (m *streamMap) Delete(conn stream.Conn) {
	sm := (*sync.Map)(m)
	sm.Delete(conn.RemoteAddr())
}

func (h DeviceController) handleDeviceMessages(
	ctx context.Context,
	conn *websocket.Conn,
	sessions *streamMap,
	errChan chan<- error,
) {
	id := identity.FromContext(ctx)
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			errChan <- err
			return
		}
		var msg ws.ProtoMsg
		err = msgpack.Unmarshal(data, &msg)
		if err != nil {
			errChan <- err
			return
		}
		sess := sessions.Load(id.Tenant + ":" + msg.Header.SessionID)
		if conn == nil {
			// TODO: Handle invalid session
			continue
		}
		err = sess.Send(ctx, data)
		if err != nil {
			if errors.Is(err, stream.ErrClosed) {
				return
			}
			errChan <- err
			return
		}
	}
}

func (h DeviceController) handleManagementMessages(
	ctx context.Context,
	conn *websocket.Conn,
	listener stream.Listener,
	sessions *streamMap,
	ticker *time.Ticker,
	errChan chan<- error,
) {
	l := log.FromContext(ctx)
	defer close(errChan)
	for {
		s, err := listener.Accept(ctx)
		if err != nil {
			if errors.Is(err, stream.ErrClosed) {
				return
			}
			l.Errorf("error accepting connections: %s", err.Error())
			select {
			case errChan <- err:
			default:
			}
			return
		}
		sessions.Store(s)
		go func() {
			defer s.Close(ctx)
			defer sessions.Delete(s)
			sessionID := s.RemoteAddr()
			idx := strings.LastIndex(sessionID, ":")
			if idx >= 0 {
				sessionID = sessionID[idx+1:]
			}
			l := l.WithField("session_id", sessionID)
			for {
				data, err := s.Recv(ctx)
				if err != nil {
					if !errors.Is(err, io.EOF) {
						l.Errorf("fatal error on channel: %s", err.Error())
					} else {
						select {
						case errChan <- err:
						default:
						}
					}
					return
				}
				err = conn.WriteMessage(websocket.BinaryMessage, data)
				if err != nil {
					l.Errorf("fatal error writing to websocket: %s", err.Error())
					select {
					case errChan <- err:
					default:
					}
					return
				}
				ticker.Reset(pongWait / 2)
			}
		}()
	}
}

// websocketWriter is the go-routine responsible for the writing end of the
// websocket. The routine forwards messages posted on the NATS session subject
// and periodically pings the connection. If the connection times out or a
// protocol violation occurs, the routine closes the connection.
func (h DeviceController) connectWSWriter(
	ctx context.Context,
	conn *websocket.Conn,
	listener stream.Listener,
) (err error) {
	l := log.FromContext(ctx)
	defer func() {
		l.SimpleRecovery(
			log.NewRecoveryOption().
				WithError(err))
		writerFinalizer(conn, &err, l)
	}()

	// send periodic ping messages to keep the connection alive
	pingPeriod := pongWait / 2
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	conn.SetPongHandler(func(string) error {
		ticker.Reset(pingPeriod)
		return nil
	})

	conn.SetPingHandler(func(msg string) error {
		ticker.Reset(pingPeriod)
		return conn.WriteControl(
			websocket.PongMessage,
			[]byte(msg),
			time.Now().Add(writeWait),
		)
	})
	acceptErr := make(chan error, 1)
	errChan := make(chan error, 1)
	sessions := new(streamMap)
	go h.handleDeviceMessages(ctx, conn, sessions, errChan)
	go h.handleManagementMessages(ctx, conn, listener, sessions, ticker, acceptErr)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if pingErr := websocketPing(conn); pingErr != nil {
				err = errors.Wrap(pingErr, "failed to send a ping")
				return err
			}
		case err := <-errChan:
			return err
		case err := <-acceptErr:
			return err
		}
	}
}

func websocketPing(conn *websocket.Conn) error {
	pongWaitString := strconv.Itoa(int(pongWait.Seconds()))
	return conn.WriteControl(
		websocket.PingMessage,
		[]byte(pongWaitString),
		time.Now().Add(writeWait),
	)
}
