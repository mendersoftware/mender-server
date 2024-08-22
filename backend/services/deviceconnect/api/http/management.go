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
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gorilla/websocket"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"
	"github.com/mendersoftware/mender-server/pkg/ws"
	"github.com/mendersoftware/mender-server/pkg/ws/menderclient"
	"github.com/mendersoftware/mender-server/pkg/ws/shell"
	natsio "github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/mendersoftware/mender-server/services/deviceconnect/app"
	"github.com/mendersoftware/mender-server/services/deviceconnect/client/nats"
	"github.com/mendersoftware/mender-server/services/deviceconnect/model"
)

// HTTP errors
var (
	ErrMissingUserAuthentication = errors.New(
		"missing or non-user identity in the authorization headers",
	)
	ErrMsgSessionLimit = "session byte limit exceeded"

	//The name of the field holding a number of milliseconds to sleep between
	//the consecutive writes of session recording data. Note that it does not have
	//anything to do with the sleep between the keystrokes send, lines printed,
	//or screen blinks, we are only aware of the stream of bytes.
	PlaybackSleepIntervalMsField = "sleep_ms"

	//The name of the field in the query parameter to GET that holds the id of a session
	PlaybackSessionIDField = "sessionId"

	//The threshold between the shell commands received (keystrokes) above which the
	//delay control message is saved (1.5 seconds)
	keyStrokeDelayRecordingThresholdNs = int64(1500 * 1000000)

	//The key stroke delay is recorded in two bytes, so this is the maximal
	//possible delay. We round down to this if the real delay is larger
	keyStrokeMaxDelayRecording = int64(65535 * 1000000)
)

const channelSize = 25 // TODO make configurable

const (
	PropertyUserID = "user_id"
)

var wsUpgrader = websocket.Upgrader{
	Subprotocols: []string{"protomsg/msgpack"},
	CheckOrigin:  allowAllOrigins,
	Error: func(
		w http.ResponseWriter, r *http.Request, s int, e error,
	) {
		w.WriteHeader(s)
		enc := json.NewEncoder(w)
		_ = enc.Encode(rest.Error{
			Err:       e.Error(),
			RequestID: requestid.FromContext(r.Context())},
		)
	},
}

// ManagementController container for end-points
type ManagementController struct {
	app  app.App
	nats nats.Client
}

// NewManagementController returns a new ManagementController
func NewManagementController(
	app app.App,
	nc nats.Client,
) *ManagementController {
	return &ManagementController{
		app:  app,
		nats: nc,
	}
}

// GetDevice returns a device
func (h ManagementController) GetDevice(c *gin.Context) {
	ctx := c.Request.Context()

	idata := identity.FromContext(ctx)
	if idata == nil || !idata.IsUser {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": ErrMissingUserAuthentication.Error(),
		})
		return
	}
	tenantID := idata.Tenant
	deviceID := c.Param("deviceId")

	device, err := h.app.GetDevice(ctx, tenantID, deviceID)
	if err == app.ErrDeviceNotFound {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	} else if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, device)
}

// Connect extracts identity from request, checks user permissions
// and calls ConnectDevice
func (h ManagementController) Connect(c *gin.Context) {
	ctx := c.Request.Context()
	l := log.FromContext(ctx)

	idata := identity.FromContext(ctx)
	if !idata.IsUser {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": ErrMissingUserAuthentication.Error(),
		})
		return
	}

	tenantID := idata.Tenant
	userID := idata.Subject
	deviceID := c.Param("deviceId")

	session := &model.Session{
		TenantID:           tenantID,
		UserID:             userID,
		DeviceID:           deviceID,
		StartTS:            time.Now(),
		BytesRecordedMutex: &sync.Mutex{},
		Types:              []string{},
	}

	// Prepare the user session
	err := h.app.PrepareUserSession(ctx, session)
	if err == app.ErrDeviceNotFound || err == app.ErrDeviceNotConnected {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	} else if _, ok := errors.Cause(err).(validation.Errors); ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	} else if err != nil {
		l.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer func() {
		err := h.app.FreeUserSession(ctx, session.ID, session.Types)
		if err != nil {
			l.Warnf("failed to free session: %s", err.Error())
		}
	}()

	deviceChan := make(chan *natsio.Msg, channelSize)
	sub, err := h.nats.ChanSubscribe(session.Subject(tenantID), deviceChan)
	if err != nil {
		l.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to establish internal device session",
		})
		return
	}
	//nolint:errcheck
	defer sub.Unsubscribe()

	// upgrade get request to websocket protocol
	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		err = errors.Wrap(err, "unable to upgrade the request to websocket protocol")
		l.Error(err)
		// upgrader.Upgrade has already responded
		return
	}
	conn.SetReadLimit(int64(app.MessageSizeLimit))

	//nolint:errcheck
	h.ConnectServeWS(ctx, conn, session, deviceChan)
}

func (h ManagementController) Playback(c *gin.Context) {
	ctx := c.Request.Context()
	l := log.FromContext(ctx)

	idata := identity.FromContext(ctx)
	if !idata.IsUser {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": ErrMissingUserAuthentication.Error(),
		})
		return
	}

	tenantID := idata.Tenant
	userID := idata.Subject
	sessionID := c.Param(PlaybackSessionIDField)
	session := &model.Session{
		TenantID:           tenantID,
		UserID:             userID,
		StartTS:            time.Now(),
		BytesRecordedMutex: &sync.Mutex{},
	}
	sleepInterval := c.Param(PlaybackSleepIntervalMsField)
	sleepMilliseconds := uint(app.DefaultPlaybackSleepIntervalMs)
	if len(sleepInterval) > 1 {
		n, err := strconv.ParseUint(sleepInterval, 10, 32)
		if err != nil {
			sleepMilliseconds = uint(n)
		}
	}

	l.Infof("Playing back the session session_id=%s", sessionID)

	// upgrade get request to websocket protocol
	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		err = errors.Wrap(err, "unable to upgrade the request to websocket protocol")
		l.Error(err)
		return
	}
	conn.SetReadLimit(int64(app.MessageSizeLimit))

	deviceChan := make(chan *natsio.Msg, channelSize)
	errChan := make(chan error, 1)

	//nolint:errcheck
	go h.websocketWriter(ctx,
		conn,
		session,
		deviceChan,
		errChan,
		bufio.NewWriterSize(ioutil.Discard, app.RecorderBufferSize),
		bufio.NewWriterSize(ioutil.Discard, app.RecorderBufferSize))

	go func() {
		err = h.app.GetSessionRecording(ctx,
			sessionID,
			app.NewPlayback(sessionID, deviceChan, sleepMilliseconds))
		if err != nil {
			err = errors.Wrap(err, "unable to get the session.")
			errChan <- err
			return
		}
	}()
	// We need to keep reading in order to keep ping/pong handlers functioning.
	for ; err == nil; _, _, err = conn.NextReader() {
	}
}

func websocketPing(conn *websocket.Conn) bool {
	pongWaitString := strconv.Itoa(int(pongWait.Seconds()))
	if err := conn.WriteControl(
		websocket.PingMessage,
		[]byte(pongWaitString),
		time.Now().Add(writeWait),
	); err != nil {
		return false
	}
	return true
}

func writerFinalizer(conn *websocket.Conn, e *error, l *log.Logger) {
	err := *e
	if err != nil {
		if !websocket.IsUnexpectedCloseError(errors.Cause(err)) {
			errMsg := err.Error()
			errBody := make([]byte, len(errMsg)+2)
			binary.BigEndian.PutUint16(errBody,
				websocket.CloseInternalServerErr)
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
}

// websocketWriter is the go-routine responsible for the writing end of the
// websocket. The routine forwards messages posted on the NATS session subject
// and periodically pings the connection. If the connection times out or a
// protocol violation occurs, the routine closes the connection.
func (h ManagementController) websocketWriter(
	ctx context.Context,
	conn *websocket.Conn,
	session *model.Session,
	deviceChan <-chan *natsio.Msg,
	errChan <-chan error,
	recorderBuffered *bufio.Writer,
	controlRecorderBuffered *bufio.Writer,
) (err error) {
	l := log.FromContext(ctx)
	defer writerFinalizer(conn, &err, l)

	// handle the ping-pong connection health check
	err = conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		l.Error(err)
		return err
	}

	pingPeriod := (pongWait * 9) / 10
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	conn.SetPongHandler(func(string) error {
		ticker.Reset(pingPeriod)
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	conn.SetPingHandler(func(msg string) error {
		ticker.Reset(pingPeriod)
		err := conn.SetReadDeadline(time.Now().Add(pongWait))
		if err != nil {
			return err
		}
		return conn.WriteControl(
			websocket.PongMessage,
			[]byte(msg),
			time.Now().Add(writeWait),
		)
	})

	defer recorderBuffered.Flush()
	defer controlRecorderBuffered.Flush()
	recordedBytes := 0
	controlBytes := 0

	sessOverLimit := false
	sessOverLimitHandled := false

	lastKeystrokeAt := time.Now().UTC().UnixNano()
Loop:
	for {
		var forwardedMsg []byte

		select {
		case msg := <-deviceChan:
			mr := &ws.ProtoMsg{}
			err = msgpack.Unmarshal(msg.Data, mr)
			if err != nil {
				return err
			}

			forwardedMsg = msg.Data

			if mr.Header.Proto == ws.ProtoTypeShell {
				switch mr.Header.MsgType {
				case shell.MessageTypeShellCommand:

					if recordedBytes >= app.MessageSizeLimit ||
						controlBytes >= app.MessageSizeLimit {
						sessOverLimit = true

						errMsg := h.handleSessLimit(ctx,
							session,
							&sessOverLimitHandled,
						)

						//override original message with shell error
						if errMsg != nil {
							forwardedMsg = errMsg
						}
					} else {
						if err = recordSession(ctx,
							mr,
							recorderBuffered,
							controlRecorderBuffered,
							&recordedBytes,
							&controlBytes,
							&lastKeystrokeAt,
							session,
						); err != nil {
							return err
						}
					}

				case shell.MessageTypeStopShell:
					l.Debugf("session logging: recorderBuffered.Flush()"+
						" at %d on stop shell", recordedBytes)
					recorderBuffered.Flush()
				}
			}

			if !sessOverLimit {
				err = conn.WriteMessage(websocket.BinaryMessage, forwardedMsg)
				if err != nil {
					l.Error(err)
					break Loop
				}
			}
		case <-ctx.Done():
			break Loop
		case <-ticker.C:
			if !websocketPing(conn) {
				err = errors.New("connection timeout")
				break Loop
			}
		case err := <-errChan:
			return err
		}
	}
	return err
}

func (h ManagementController) handleSessLimit(ctx context.Context,
	session *model.Session,
	handled *bool,
) []byte {
	l := log.FromContext(ctx)

	// possible error return message (ws->user)
	var retMsg []byte

	// attempt to clean up once
	if !(*handled) {
		sendLimitErrDevice(ctx, session, h.nats)
		userErrMsg, err := prepLimitErrUser(ctx, session)
		if err != nil {
			l.Errorf("session limit: " +
				"failed to notify user")
		}

		retMsg = userErrMsg

		err = h.app.FreeUserSession(ctx, session.ID, session.Types)
		if err != nil {
			l.Warnf("failed to free session"+
				"that went over limit: %s", err.Error())
		}

		*handled = true
	}

	return retMsg
}

func recordSession(ctx context.Context,
	msg *ws.ProtoMsg,
	recorder io.Writer,
	recorderCtrl io.Writer,
	recBytes *int,
	ctrlBytes *int,
	lastKeystrokeAt *int64,
	session *model.Session) error {
	l := log.FromContext(ctx)

	b, e := recorder.Write(msg.Body)
	if e != nil {
		l.Errorf("session logging: "+
			"recorderBuffered.Write"+
			"(len=%d)=%d,%+v",
			len(msg.Body), b, e)
	}
	timeNowUTC := time.Now().UTC().UnixNano()
	keystrokeDelay := timeNowUTC - (*lastKeystrokeAt)
	if keystrokeDelay >= keyStrokeDelayRecordingThresholdNs {
		if keystrokeDelay > keyStrokeMaxDelayRecording {
			keystrokeDelay = keyStrokeMaxDelayRecording
		}

		controlMsg := app.Control{
			Type:   app.DelayMessage,
			Offset: *recBytes,
			DelayMs: uint16(float64(keystrokeDelay) *
				0.000001),
			TerminalHeight: 0,
			TerminalWidth:  0,
		}
		n, _ := recorderCtrl.Write(
			controlMsg.MarshalBinary())
		l.Debugf("saving control delay message: %+v/%d",
			controlMsg, n)
		(*ctrlBytes) += n
	}

	(*lastKeystrokeAt) = timeNowUTC

	(*recBytes) += len(msg.Body)
	session.BytesRecordedMutex.Lock()
	session.BytesRecorded = *recBytes
	session.BytesRecordedMutex.Unlock()

	return nil
}

// prepLimitErrUser preps a session limit exceeded error for the user (shell cmd + err status)
func prepLimitErrUser(ctx context.Context, session *model.Session) ([]byte, error) {
	userErrMsg := ws.ProtoMsg{
		Header: ws.ProtoHdr{
			Proto:     ws.ProtoTypeShell,
			MsgType:   shell.MessageTypeShellCommand,
			SessionID: session.ID,
			Properties: map[string]interface{}{
				"status": shell.ErrorMessage,
			},
		},
		Body: []byte(ErrMsgSessionLimit),
	}

	return msgpack.Marshal(userErrMsg)
}

// sendLimitErrDevice preps and sends
// session limit exceeded error to device (stop shell + err status)
// this is best effort, log and swallow errors
func sendLimitErrDevice(ctx context.Context, session *model.Session, nats nats.Client) {
	l := log.FromContext(ctx)

	msg := ws.ProtoMsg{
		Header: ws.ProtoHdr{
			Proto:     ws.ProtoTypeShell,
			MsgType:   shell.MessageTypeStopShell,
			SessionID: session.ID,
			Properties: map[string]interface{}{
				"status":       shell.ErrorMessage,
				PropertyUserID: session.UserID,
			},
		},
		Body: []byte(ErrMsgSessionLimit),
	}
	data, err := msgpack.Marshal(msg)
	if err != nil {
		l.Errorf(
			"session limit: "+
				"failed to prep stop session"+
				"%s message to device: %s, error %v",
			session.ID,
			session.DeviceID,
			err,
		)
	}
	err = nats.Publish(model.GetDeviceSubject(
		session.TenantID, session.DeviceID),
		data,
	)
	if err != nil {
		l.Errorf(
			"session limit: failed to send stop session"+
				"%s message to device: %s, error %v",
			session.ID,
			session.DeviceID,
			err,
		)
	}
}

// ConnectServeWS starts a websocket connection with the device
// Currently this handler only properly handles a single terminal session.
func (h ManagementController) ConnectServeWS(
	ctx context.Context,
	conn *websocket.Conn,
	sess *model.Session,
	deviceChan chan *natsio.Msg,
) (err error) {
	l := log.FromContext(ctx)
	id := identity.FromContext(ctx)
	errChan := make(chan error, 1)
	remoteTerminalRunning := false

	defer func() {
		if err != nil {
			select {
			case errChan <- err:

			case <-time.After(time.Second):
				l.Warn("Failed to propagate error to client")
			}
		}
		if remoteTerminalRunning {
			msg := ws.ProtoMsg{
				Header: ws.ProtoHdr{
					Proto:     ws.ProtoTypeShell,
					MsgType:   shell.MessageTypeStopShell,
					SessionID: sess.ID,
					Properties: map[string]interface{}{
						"status":       shell.ErrorMessage,
						PropertyUserID: sess.UserID,
					},
				},
				Body: []byte("user disconnected"),
			}
			data, _ := msgpack.Marshal(msg)
			errPublish := h.nats.Publish(model.GetDeviceSubject(
				id.Tenant, sess.DeviceID),
				data,
			)
			if errPublish != nil {
				l.Warnf(
					"failed to propagate stop session "+
						"message to device: %s",
					errPublish.Error(),
				)
			}
		}
		close(errChan)
	}()

	controlRecorder := h.app.GetControlRecorder(ctx, sess.ID)
	controlRecorderBuffered := bufio.NewWriterSize(controlRecorder, app.RecorderBufferSize)
	defer controlRecorderBuffered.Flush()

	sessionRecorder := h.app.GetRecorder(ctx, sess.ID)
	sessionRecorderBuffered := bufio.NewWriterSize(sessionRecorder, app.RecorderBufferSize)
	defer sessionRecorderBuffered.Flush()

	// websocketWriter is responsible for closing the websocket
	//nolint:errcheck
	go h.websocketWriter(ctx,
		conn,
		sess,
		deviceChan,
		errChan,
		sessionRecorderBuffered,
		controlRecorderBuffered)

	return h.connectServeWSProcessMessages(ctx, conn, sess, deviceChan,
		&remoteTerminalRunning, controlRecorderBuffered)
}

func (h ManagementController) connectServeWSProcessMessages(
	ctx context.Context,
	conn *websocket.Conn,
	sess *model.Session,
	deviceChan chan *natsio.Msg,
	remoteTerminalRunning *bool,
	controlRecorderBuffered *bufio.Writer,
) (err error) {
	l := log.FromContext(ctx)
	id := identity.FromContext(ctx)
	logTerminal := false
	logPortForward := false

	var data []byte
	controlBytes := 0
	ignoreControlMessages := false
	for {
		_, data, err = conn.ReadMessage()
		if err != nil {
			if _, ok := err.(*websocket.CloseError); ok {
				return nil
			}
			return err
		}
		m := &ws.ProtoMsg{}
		err = msgpack.Unmarshal(data, m)
		if err != nil {
			return err
		}

		m.Header.SessionID = sess.ID
		if m.Header.Properties == nil {
			m.Header.Properties = make(map[string]interface{})
		}
		m.Header.Properties[PropertyUserID] = sess.UserID
		data, _ = msgpack.Marshal(m)
		switch m.Header.Proto {
		case ws.ProtoTypeShell:
			// send the audit log for remote terminal
			if !logTerminal {
				if err := h.app.LogUserSession(ctx, sess,
					model.SessionTypeTerminal); err != nil {
					return err
				}
				sess.Types = append(sess.Types, model.SessionTypeTerminal)
				logTerminal = true
			}
			// handle remote terminal-specific messages
			switch m.Header.MsgType {
			case shell.MessageTypeSpawnShell:
				*remoteTerminalRunning = true
			case shell.MessageTypeStopShell:
				*remoteTerminalRunning = false
			case shell.MessageTypeResizeShell:
				if ignoreControlMessages {
					continue
				}
				if controlBytes >= app.MessageSizeLimit {
					l.Infof("session_id=%s control data limit reached.",
						sess.ID)
					//see https://northerntech.atlassian.net/browse/MEN-4448
					ignoreControlMessages = true
					continue
				}

				controlBytes += sendResizeMessage(m, sess, controlRecorderBuffered)
			}
		case ws.ProtoTypePortForward:
			if !logPortForward {
				if err := h.app.LogUserSession(ctx, sess,
					model.SessionTypePortForward); err != nil {
					return err
				}
				sess.Types = append(sess.Types, model.SessionTypePortForward)
				logPortForward = true
			}
		}

		err = h.nats.Publish(model.GetDeviceSubject(id.Tenant, sess.DeviceID), data)
		if err != nil {
			return err
		}
	}
}

func sendResizeMessage(m *ws.ProtoMsg,
	sess *model.Session,
	controlRecorderBuffered *bufio.Writer) (n int) {
	if _, ok := m.Header.Properties[model.ResizeMessageTermHeightField]; ok {
		return 0
	}
	if _, ok := m.Header.Properties[model.ResizeMessageTermWidthField]; ok {
		return 0
	}

	var height uint16 = 0
	switch m.Header.Properties[model.ResizeMessageTermHeightField].(type) {
	case uint8:
		height = uint16(m.Header.Properties[model.ResizeMessageTermHeightField].(uint8))
	case int8:
		height = uint16(m.Header.Properties[model.ResizeMessageTermHeightField].(int8))
	}

	var width uint16 = 0
	switch m.Header.Properties[model.ResizeMessageTermWidthField].(type) {
	case uint8:
		width = uint16(m.Header.Properties[model.ResizeMessageTermWidthField].(uint8))
	case int8:
		width = uint16(m.Header.Properties[model.ResizeMessageTermWidthField].(int8))
	}

	sess.BytesRecordedMutex.Lock()
	controlMsg := app.Control{
		Type:           app.ResizeMessage,
		Offset:         sess.BytesRecorded,
		DelayMs:        0,
		TerminalHeight: height,
		TerminalWidth:  width,
	}
	sess.BytesRecordedMutex.Unlock()

	n, _ = controlRecorderBuffered.Write(
		controlMsg.MarshalBinary(),
	)
	return n
}

func (h ManagementController) CheckUpdate(c *gin.Context) {
	h.sendMenderCommand(c, menderclient.MessageTypeMenderClientCheckUpdate)
}

func (h ManagementController) SendInventory(c *gin.Context) {
	h.sendMenderCommand(c, menderclient.MessageTypeMenderClientSendInventory)
}

func (h ManagementController) sendMenderCommand(c *gin.Context, msgType string) {
	ctx := c.Request.Context()

	idata := identity.FromContext(ctx)
	if idata == nil || !idata.IsUser {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": ErrMissingUserAuthentication.Error(),
		})
		return
	}
	tenantID := idata.Tenant
	deviceID := c.Param("deviceId")

	device, err := h.app.GetDevice(ctx, tenantID, deviceID)
	if err == app.ErrDeviceNotFound {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	} else if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	} else if device.Status != model.DeviceStatusConnected {
		c.JSON(http.StatusConflict, gin.H{
			"error": app.ErrDeviceNotConnected,
		})
		return
	}

	msg := &ws.ProtoMsg{
		Header: ws.ProtoHdr{
			Proto:   ws.ProtoTypeMenderClient,
			MsgType: msgType,
			Properties: map[string]interface{}{
				PropertyUserID: idata.Subject,
			},
		},
	}
	data, _ := msgpack.Marshal(msg)

	err = h.nats.Publish(model.GetDeviceSubject(idata.Tenant, device.ID), data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	c.JSON(http.StatusAccepted, nil)
}
