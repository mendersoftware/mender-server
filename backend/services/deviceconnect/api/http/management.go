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
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"
	"github.com/mendersoftware/mender-server/pkg/stream"
	"github.com/mendersoftware/mender-server/pkg/ws"
	"github.com/mendersoftware/mender-server/pkg/ws/menderclient"
	"github.com/mendersoftware/mender-server/pkg/ws/shell"

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
		rest.RenderError(c, http.StatusBadRequest, ErrMissingUserAuthentication)
		return
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	tenantID := idata.Tenant
	userID := idata.Subject
	deviceID := c.Param("deviceId")

	session := &model.Session{
		ID:                 uuid.NewString(),
		TenantID:           tenantID,
		UserID:             userID,
		DeviceID:           deviceID,
		StartTS:            time.Now(),
		BytesRecordedMutex: &sync.Mutex{},
		Types:              []string{},
	}

	s, err := h.nats.Connect(ctx, session.TenantID+":"+session.ID, session.DeviceID)
	if err != nil {
		if errors.Is(err, stream.ErrConnectionRefused) {
			rest.RenderErrorWithMessage(c, http.StatusConflict, err,
				"failed to connect to device: device disconnected")
		} else {
			rest.RenderErrorWithMessage(c, http.StatusInternalServerError, err,
				"failed to connect to device: internal error",
			)
		}
		return
	}
	//nolint:errcheck
	defer s.Close(ctx)

	// Prepare the user session
	err = h.app.PrepareUserSession(ctx, session)
	if err == app.ErrDeviceNotFound || err == app.ErrDeviceNotConnected {
		rest.RenderError(c, http.StatusNotFound, err)
		return
	} else if _, ok := errors.Cause(err).(validation.Errors); ok {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	} else if err != nil {
		rest.RenderInternalError(c, err)
		return
	}
	defer func() {
		err := h.app.FreeUserSession(ctx, session.ID, session.Types)
		if err != nil {
			l.Warnf("failed to free session: %s", err.Error())
		}
	}()

	// upgrade get request to websocket protocol
	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		err = errors.Wrap(err, "unable to upgrade the request to websocket protocol")
		l.Error(err)
		// upgrader.Upgrade has already responded
		return
	}
	conn.SetReadLimit(int64(app.MessageSizeLimit))
	defer conn.Close()

	//nolint:errcheck
	h.ConnectServeWS(c, ctx, conn, session, s)
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

	sessionID := c.Param(PlaybackSessionIDField)
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
	defer conn.Close()

	pipeWriter := app.NewPipeWriter()
	defer pipeWriter.Close() //nolint:errcheck

	errChan := make(chan error, 1)
	go func() {
		defer l.SimpleRecovery(
			log.NewRecoveryOption().
				WithChannel(errChan))
		defer pipeWriter.Close()
		err = h.app.GetSessionRecording(ctx,
			sessionID,
			pipeWriter)
		if err != nil {
			err = errors.Wrap(err, "unable to get the session.")
		}
	}()
	go func() {
		defer l.SimpleRecovery(
			log.NewRecoveryOption().
				WithChannel(errChan))
		// We need to keep reading in order to keep ping/pong handlers functioning.
		for ; err == nil; _, _, err = conn.NextReader() {
		}
	}()

	dataChan := pipeWriter.RecvChan()
	ticker := time.NewTicker(time.Duration(sleepMilliseconds) * time.Millisecond)

	for {
		select {
		case data, open := <-dataChan:
			if !open {
				_ = conn.WriteControl(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
					time.Now().Add(time.Second*5))
				return
			}
			err = conn.WriteMessage(websocket.BinaryMessage, data)
			if err != nil {
				_ = c.Error(err)
				return
			}
			dataChan = nil // Will cause loop to wait for ticker.C next iteration

		case <-ticker.C:
			dataChan = pipeWriter.RecvChan()

		case <-ctx.Done():
			_ = c.Error(ctx.Err())
			return

		case err = <-errChan:
			if err != nil {
				_ = c.Error(err)
			}
			return

		}
	}
}

func writerFinalizer(conn *websocket.Conn, e *error, l *log.Logger) {
	err := *e
	if err != nil {
		var closeErr *websocket.CloseError
		// If err is a websocket.CloseError we assume that we have already
		// received a close frame, or a close frame has already been sent.
		if errors.As(err, &closeErr) {
			if closeErr.Code == websocket.CloseNormalClosure {
				return
			}
		} else {
			errClose := conn.WriteControl(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "internal error"),
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
	} else {
		err = conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			time.Now().Add(writeWait))
		if err != nil {
			l.Errorf("error sending websocket close frame: %s", err.Error())
		}
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
	s stream.Conn,
	errChan <-chan error,
	recorder app.Recorder,
	controlRecorder app.Recorder,
) (err error) {
	l := log.FromContext(ctx)
	defer l.SimpleRecovery(
		log.NewRecoveryOption().
			WithError(err))

	defer writerFinalizer(conn, &err, l)

	// handle the ping-pong connection health check
	conn.SetPingHandler(func(msg string) error {
		if err != nil {
			return err
		}
		return conn.WriteControl(
			websocket.PongMessage,
			[]byte(msg),
			time.Now().Add(writeWait),
		)
	})

	recordedBytes := 0
	controlBytes := 0

	lastKeystrokeAt := time.Now().UTC().UnixNano()
	for {
		var forwardedMsg []byte
		var data []byte

		data, err = s.Recv(ctx)
		if err != nil {
			return err
		}
		mr := &ws.ProtoMsg{}
		err = msgpack.Unmarshal(data, mr)
		if err != nil {
			return err
		}
		mr.Header.SessionID = session.ID
		forwardedMsg, _ = msgpack.Marshal(mr)

		if mr.Header.Proto == ws.ProtoTypeShell {
			switch mr.Header.MsgType {
			case shell.MessageTypeShellCommand:

				if recordedBytes >= app.MessageSizeLimit ||
					controlBytes >= app.MessageSizeLimit {

					_ = h.handleSessLimit(ctx, session, s)
					closeErr := &websocket.CloseError{
						Code: websocket.ClosePolicyViolation,
						Text: "session byte limit exhausted",
					}
					_ = conn.WriteControl(websocket.CloseMessage,
						websocket.FormatCloseMessage(closeErr.Code, closeErr.Text),
						time.Now().Add(time.Second*10),
					)
					err = closeErr
					return err

				} else {
					if err = recordSession(ctx,
						mr,
						recorder,
						controlRecorder,
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
			}
		}

		err = conn.WriteMessage(websocket.BinaryMessage, forwardedMsg)
		if err != nil {
			l.Error(err)
			break
		}
	}
	return err
}

func (h ManagementController) handleSessLimit(ctx context.Context,
	session *model.Session,
	s stream.Conn,
) []byte {
	l := log.FromContext(ctx)

	// possible error return message (ws->user)
	var retMsg []byte

	// attempt to clean up once
	sendLimitErrDevice(ctx, session, s)
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

	return retMsg
}

func recordSession(ctx context.Context,
	msg *ws.ProtoMsg,
	recorder app.Recorder,
	recorderCtrl app.Recorder,
	recBytes *int,
	ctrlBytes *int,
	lastKeystrokeAt *int64,
	session *model.Session) error {
	l := log.FromContext(ctx)

	e := recorder.Record(ctx, msg.Body)
	if e != nil {
		l.Errorf("session logging: "+
			"recorderBuffered.Write"+
			"(len=%d),%+v",
			len(msg.Body), e)
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
		msg := controlMsg.MarshalBinary()
		_ = recorderCtrl.Record(
			ctx,
			msg)
		l.Debugf("saving control delay message: %+v/%d",
			controlMsg, len(msg))
		(*ctrlBytes) += len(msg)
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
func sendLimitErrDevice(ctx context.Context, session *model.Session, s stream.Conn) {
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
	err = s.Send(ctx, data)
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
	c *gin.Context,
	ctx context.Context,
	conn *websocket.Conn,
	sess *model.Session,
	s stream.Conn,
) (err error) {
	l := log.FromContext(ctx)
	remoteTerminalRunning := false

	defer func() {
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
			errPublish := s.Send(ctx, data)
			if errPublish != nil {
				l.Warnf(
					"failed to propagate stop session "+
						"message to device: %s",
					errPublish.Error(),
				)
			}
		}
	}()

	controlRecorder := h.app.GetControlRecorder(sess.ID)
	sessionRecorder := h.app.GetRecorder(sess.ID)
	defer func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second*10)
		defer cancel()
		_ = sessionRecorder.Close(ctx)
		_ = controlRecorder.Close(ctx)
	}()

	errChan := make(chan error)
	//nolint:errcheck
	go h.connectServeWSProcessMessages(ctx, conn, s, sess, errChan,
		&remoteTerminalRunning, controlRecorder)

	err = h.websocketWriter(ctx,
		conn,
		sess,
		s,
		errChan,
		sessionRecorder,
		controlRecorder)
	if err != nil && !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		_ = c.Error(err)
	}
	return err
}

func (h ManagementController) handleReadErrors(
	ctx context.Context,
	errP *error,
	errChan chan<- error) {
	var err error
	if errP != nil {
		err = *errP
	}
	log.FromContext(ctx).SimpleRecovery(log.NewRecoveryOption().WithError(err))
	if err != nil && !websocket.IsUnexpectedCloseError(err) {
		select {
		case <-ctx.Done():
		case errChan <- err:
		}
	}
	close(errChan)
}

func (h ManagementController) connectServeWSProcessMessages(
	ctx context.Context,
	conn *websocket.Conn,
	s stream.Conn,
	sess *model.Session,
	errChan chan<- error,
	remoteTerminalRunning *bool,
	controlRecorder app.Recorder,
) {
	var err error
	l := log.FromContext(ctx)
	logTerminal := false
	logPortForward := false
	defer s.Close(ctx)
	defer h.handleReadErrors(ctx, &err, errChan)

	var data []byte
	controlBytes := 0
	ignoreControlMessages := false
	for {
		_, data, err = conn.ReadMessage()
		if err != nil {
			if _, ok := err.(*websocket.CloseError); ok {
				err = nil
			}
			return
		}
		m := &ws.ProtoMsg{}
		err = msgpack.Unmarshal(data, m)
		if err != nil {
			return
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
				if err = h.app.LogUserSession(ctx, sess,
					model.SessionTypeTerminal); err != nil {
					return
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

				controlBytes += sendResizeMessage(ctx, m, sess, controlRecorder)
			}
		case ws.ProtoTypePortForward:
			if !logPortForward {
				if err = h.app.LogUserSession(ctx, sess,
					model.SessionTypePortForward); err != nil {
					return
				}
				sess.Types = append(sess.Types, model.SessionTypePortForward)
				logPortForward = true
			}
		}

		err = s.Send(ctx, data)
		if err != nil {
			return
		}
	}
}

func sendResizeMessage(ctx context.Context, m *ws.ProtoMsg,
	sess *model.Session,
	controlRecorder app.Recorder) (n int) {
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

	_ = controlRecorder.Record(
		ctx,
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
