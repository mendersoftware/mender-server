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
	"encoding/binary"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	natsio "github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"
	"github.com/mendersoftware/mender-server/pkg/ws"
	"github.com/mendersoftware/mender-server/pkg/ws/shell"

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

	msgChan := make(chan *natsio.Msg, channelSize)
	sub, err := h.nats.ChanSubscribe(
		model.GetDeviceSubject(idata.Tenant, idata.Subject),
		msgChan,
	)
	if err != nil {
		l.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to allocate internal device channel",
		})
		return
	}
	//nolint:errcheck
	defer sub.Unsubscribe()

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

	errChan := make(chan error)
	defer close(errChan)

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
	registerID := h.app.RegisterShutdownCancel(cancel)
	defer h.app.UnregisterShutdownCancel(registerID)

	// websocketWriter is responsible for closing the websocket
	//nolint:errcheck
	go h.connectWSWriter(ctxWithCancel, conn, msgChan, errChan)
	err = h.ConnectServeWS(ctxWithCancel, conn)
	if err != nil {
		select {
		case errChan <- err:

		case <-time.After(time.Second):
			l.Warn("Failed to propagate error to client")
		}
	}
}

// websocketWriter is the go-routine responsible for the writing end of the
// websocket. The routine forwards messages posted on the NATS session subject
// and periodically pings the connection. If the connection times out or a
// protocol violation occurs, the routine closes the connection.
func (h DeviceController) connectWSWriter(
	ctx context.Context,
	conn *websocket.Conn,
	msgChan <-chan *natsio.Msg,
	errChan <-chan error,
) (err error) {
	l := log.FromContext(ctx)
	defer func() {
		if err != nil {
			if !websocket.IsUnexpectedCloseError(err) {
				// If the peer didn't send a close message we must.
				errMsg := err.Error()
				errBody := make([]byte, len(errMsg)+2)
				binary.BigEndian.PutUint16(errBody,
					websocket.CloseInternalServerErr,
				)
				copy(errBody[2:], errMsg)
				errClose := conn.WriteControl(
					websocket.CloseMessage,
					errBody,
					time.Now().Add(writeWait),
				)
				if errClose != nil {
					err = errors.Wrapf(err,
						"error sending websocket close frame: %s",
						errClose.Error(),
					)
				}
			}
			l.Errorf("websocket closed with error: %s", err.Error())
		}
		conn.Close()
	}()

	// send periodic ping messages to keep the connection alive
	pingPeriod := (pongWait * 9) / 10
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	conn.SetPongHandler(func(string) error {
		ticker.Reset(pingPeriod)
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	conn.SetPingHandler(func(msg string) error {
		ticker.Reset(pingPeriod)
		return conn.WriteControl(
			websocket.PongMessage,
			[]byte(msg),
			time.Now().Add(writeWait),
		)
	})
Loop:
	for {
		select {
		case msg := <-msgChan:
			err = conn.WriteMessage(websocket.BinaryMessage, msg.Data)
			if err != nil {
				l.Error(err)
				break Loop
			}
			ticker.Reset(pingPeriod)
		case <-ctx.Done():
			err = errors.New("connection closed by the server")
			break Loop
		case <-ticker.C:
			if pingErr := websocketPing(conn); pingErr != nil {
				err = errors.Wrap(pingErr, "failed to send a ping")
				break Loop
			}
		case err := <-errChan:
			return err
		}
	}
	return err
}

func (h DeviceController) ConnectServeWS(
	ctx context.Context,
	conn *websocket.Conn,
) (err error) {
	l := log.FromContext(ctx)
	id := identity.FromContext(ctx)
	sessMap := make(map[string]*model.ActiveSession)

	// update the device status on websocket opening
	var version int64
	version, err = h.app.SetDeviceConnected(ctx, id.Tenant, id.Subject)
	if err != nil {
		l.Error(err)
		return
	}
	defer func() {
		for sessionID, session := range sessMap {
			// TODO: notify the session NATS topic about the session
			//       being released.
			if session.RemoteTerminal {
				msg := ws.ProtoMsg{
					Header: ws.ProtoHdr{
						Proto:     ws.ProtoTypeShell,
						MsgType:   shell.MessageTypeStopShell,
						SessionID: sessionID,
						Properties: map[string]interface{}{
							"status": shell.ErrorMessage,
						},
					},
					Body: []byte("device disconnected"),
				}
				data, _ := msgpack.Marshal(msg)
				err = h.nats.Publish(
					model.GetSessionSubject(id.Tenant, sessionID),
					data,
				)
			}
		}
		// update the device status on websocket closing
		eStatus := h.app.SetDeviceDisconnected(
			ctx, id.Tenant,
			id.Subject, version,
		)
		if eStatus != nil {
			l.Error(eStatus)
		}
	}()

	var data []byte
	for {
		_, data, err = conn.ReadMessage()
		if err != nil {
			return err
		}
		m := &ws.ProtoMsg{}
		err = msgpack.Unmarshal(data, m)
		if err != nil {
			return err
		}

		sessMap[m.Header.SessionID] = &model.ActiveSession{}
		switch m.Header.Proto {
		case ws.ProtoTypeShell:
			if m.Header.SessionID == "" {
				return errors.New("api: message missing required session ID")
			} else if m.Header.MsgType == shell.MessageTypeSpawnShell {
				if session, ok := sessMap[m.Header.SessionID]; ok {
					session.RemoteTerminal = true
				}
			} else if m.Header.MsgType == shell.MessageTypeStopShell {
				delete(sessMap, m.Header.SessionID)
			}
		default:
			// TODO: Handle protocol violation
		}

		err = h.nats.Publish(
			model.GetSessionSubject(id.Tenant, m.Header.SessionID),
			data,
		)
		if err != nil {
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
