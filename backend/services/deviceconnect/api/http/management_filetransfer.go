// Copyright 2021 Northern.tech AS
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
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	natsio "github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/ws"
	wsft "github.com/mendersoftware/mender-server/pkg/ws/filetransfer"

	"github.com/mendersoftware/mender-server/services/deviceconnect/app"
	"github.com/mendersoftware/mender-server/services/deviceconnect/model"
)

type fileTransferParams struct {
	TenantID  string
	UserID    string
	SessionID string
	Device    *model.Device
}

const (
	hdrContentType            = "Content-Type"
	hdrContentDisposition     = "Content-Disposition"
	hdrMenderFileTransferPath = "X-MEN-File-Path"
	hdrMenderFileTransferUID  = "X-MEN-File-UID"
	hdrMenderFileTransferGID  = "X-MEN-File-GID"
	hdrMenderFileTransferMode = "X-MEN-File-Mode"
	hdrMenderFileTransferSize = "X-MEN-File-Size"
)

const (
	fieldUploadPath = "path"
	fieldUploadUID  = "uid"
	fieldUploadGID  = "gid"
	fieldUploadMode = "mode"
	fieldUploadFile = "file"

	PropertyOffset = "offset"

	paramDownloadPath = "path"
)

var fileTransferTimeout = 60 * time.Second
var fileTransferBufferSize = 4096
var ackSlidingWindowSend = 10
var ackSlidingWindowRecv = 20

type Error struct {
	error      error
	statusCode int
}

func NewError(err error, code int) error {
	return &Error{
		error:      err,
		statusCode: code,
	}
}

func (err *Error) Error() string {
	return err.error.Error()
}

func (err *Error) Unwrap() error {
	return err.error
}

var (
	errFileTransferMarshalling   = errors.New("failed to marshal the request")
	errFileTransferUnmarshalling = errors.New("failed to unmarshal the request")
	errFileTransferPublishing    = errors.New("failed to publish the message")
	errFileTransferSubscribing   = errors.New("failed to subscribe to the mesages")
	errFileTransferTimeout       = &Error{
		error:      errors.New("file transfer timed out"),
		statusCode: http.StatusRequestTimeout,
	}
	errFileTransferFailed = &Error{
		error:      errors.New("file transfer failed"),
		statusCode: http.StatusBadRequest,
	}
	errFileTransferNotImplemented = &Error{
		error:      errors.New("file transfer not implemented on device"),
		statusCode: http.StatusBadGateway,
	}
	errFileTransferDisabled = &Error{
		error:      errors.New("file transfer disabled on device"),
		statusCode: http.StatusBadGateway,
	}
)

var newFileTransferSessionID = func() (uuid.UUID, error) {
	return uuid.NewRandom()
}

func (h ManagementController) getFileTransferParams(c *gin.Context) (*fileTransferParams, int,
	error) {
	ctx := c.Request.Context()

	idata := identity.FromContext(ctx)
	if idata == nil || !idata.IsUser {
		return nil, http.StatusUnauthorized, ErrMissingUserAuthentication
	}
	tenantID := idata.Tenant
	deviceID := c.Param("deviceId")

	device, err := h.app.GetDevice(ctx, tenantID, deviceID)
	if err == app.ErrDeviceNotFound {
		return nil, http.StatusNotFound, err
	} else if err != nil {
		return nil, http.StatusBadRequest, err
	} else if device.Status != model.DeviceStatusConnected {
		return nil, http.StatusConflict, app.ErrDeviceNotConnected
	}

	if c.Request.Method != http.MethodGet && c.Request.Body == nil {
		return nil, http.StatusBadRequest, errors.New("missing request body")
	}

	sessionID, err := newFileTransferSessionID()
	if err != nil {
		return nil, http.StatusInternalServerError,
			errors.New("failed to generate session ID")
	}

	return &fileTransferParams{
		TenantID:  idata.Tenant,
		UserID:    idata.Subject,
		SessionID: sessionID.String(),
		Device:    device,
	}, 0, nil
}

func (h ManagementController) publishFileTransferProtoMessage(sessionID, userID, deviceTopic,
	msgType string, body interface{}, offset int64) error {
	var msgBody []byte
	if msgType == wsft.MessageTypeChunk && body != nil {
		msgBody = body.([]byte)
	} else if msgType == wsft.MessageTypeACK {
		msgBody = nil
	} else if body != nil {
		var err error
		msgBody, err = msgpack.Marshal(body)
		if err != nil {
			return errors.Wrap(err, errFileTransferMarshalling.Error())
		}
	}
	proto := ws.ProtoTypeFileTransfer
	if msgType == ws.MessageTypePing || msgType == ws.MessageTypePong {
		proto = ws.ProtoTypeControl
	}
	msg := &ws.ProtoMsg{
		Header: ws.ProtoHdr{
			Proto:     proto,
			MsgType:   msgType,
			SessionID: sessionID,
			Properties: map[string]interface{}{
				PropertyUserID: userID,
			},
		},
		Body: msgBody,
	}
	if msgType == wsft.MessageTypeChunk || msgType == wsft.MessageTypeACK {
		msg.Header.Properties[PropertyOffset] = offset
	}
	data, err := msgpack.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, errFileTransferMarshalling.Error())
	}

	err = h.nats.Publish(deviceTopic, data)
	if err != nil {
		return errors.Wrap(err, errFileTransferPublishing.Error())
	}
	return nil
}

func (h ManagementController) publishControlMessage(
	sessionID, deviceTopic, messageType string, body interface{},
) error {
	msg := &ws.ProtoMsg{
		Header: ws.ProtoHdr{
			Proto:     ws.ProtoTypeControl,
			MsgType:   messageType,
			SessionID: sessionID,
		},
	}

	if body != nil {
		if b, ok := body.([]byte); ok {
			msg.Body = b
		} else {
			b, err := msgpack.Marshal(body)
			if err != nil {
				return errors.Wrap(errFileTransferMarshalling, err.Error())
			}
			msg.Body = b
		}
	}

	data, err := msgpack.Marshal(msg)
	if err != nil {
		return errors.Wrap(errFileTransferMarshalling, err.Error())
	}
	err = h.nats.Publish(deviceTopic, data)
	if err != nil {
		return errors.Wrap(errFileTransferPublishing, err.Error())
	}
	return err
}

func (h ManagementController) decodeFileTransferProtoMessage(data []byte) (*ws.ProtoMsg,
	interface{}, error) {
	msg := &ws.ProtoMsg{}
	err := msgpack.Unmarshal(data, msg)
	if err != nil {
		return nil, nil, errors.Wrap(err, errFileTransferUnmarshalling.Error())
	}

	switch msg.Header.MsgType {
	case wsft.MessageTypeError:
		msgBody := &ws.Error{}
		err := msgpack.Unmarshal(msg.Body, msgBody)
		if err != nil {
			return nil, nil, errors.Wrap(err, errFileTransferUnmarshalling.Error())
		}
		return msg, msgBody, nil
	case wsft.MessageTypeFileInfo:
		msgBody := &wsft.FileInfo{}
		err := msgpack.Unmarshal(msg.Body, msgBody)
		if err != nil {
			return nil, nil, errors.Wrap(err, errFileTransferUnmarshalling.Error())
		}
		return msg, msgBody, nil
	case wsft.MessageTypeACK, wsft.MessageTypeChunk, ws.MessageTypePing, ws.MessageTypePong:
		return msg, nil, nil
	}

	return nil, nil, errors.Errorf("unexpected message type '%s'", msg.Header.MsgType)
}

func writeHeaders(c *gin.Context, fileInfo *wsft.FileInfo) {
	c.Writer.Header().Add(hdrContentType, "application/octet-stream")
	if fileInfo.Path != nil {
		filename := path.Base(*fileInfo.Path)
		c.Writer.Header().Add(hdrContentDisposition,
			"attachment; filename=\""+filename+"\"")
		c.Writer.Header().Add(hdrMenderFileTransferPath, *fileInfo.Path)
	}
	if fileInfo.UID != nil {
		c.Writer.Header().Add(hdrMenderFileTransferUID, fmt.Sprintf("%d", *fileInfo.UID))
	}
	if fileInfo.GID != nil {
		c.Writer.Header().Add(hdrMenderFileTransferGID, fmt.Sprintf("%d", *fileInfo.GID))
	}
	if fileInfo.Mode != nil {
		c.Writer.Header().Add(hdrMenderFileTransferMode, fmt.Sprintf("%o", *fileInfo.Mode))
	}
	if fileInfo.Size != nil {
		c.Writer.Header().Add(hdrMenderFileTransferSize, fmt.Sprintf("%d", *fileInfo.Size))
	}
	c.Writer.WriteHeader(http.StatusOK)
}
func (h ManagementController) handleResponseError(c *gin.Context, err error) {
	l := log.FromContext(c.Request.Context())
	l.Errorf("error handling request: %s", err.Error())
	if !c.Writer.Written() {
		var statusError *Error
		var errMsg string = err.Error()
		var statusCode int = http.StatusInternalServerError
		if errors.As(err, &statusError) {
			statusCode = statusError.statusCode
		}
		if statusCode >= 500 {
			errMsg = "internal error"
		}
		c.Writer.WriteHeader(statusCode)
		c.JSON(statusCode, gin.H{
			"error":      errMsg,
			"request_id": requestid.FromContext(c.Request.Context()),
		})
	} else {
		l.Warn("response already written")
	}
}

func chanTimeout(
	src <-chan *natsio.Msg,
	timeout time.Duration,
) <-chan *natsio.Msg {
	timer := time.NewTimer(timeout)
	dst := make(chan *natsio.Msg)
	go func() {
		for {
			select {
			case <-timer.C:
				close(dst)
				return
			case msg, ok := <-src:
				if !ok {
					close(dst)
					return
				}
				if !timer.Stop() {
					// Timer must be stopped and drained before calling Reset.
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(timeout)
				dst <- msg
			}
		}
	}()
	return dst
}

func (h ManagementController) statFile(
	ctx context.Context,
	sessChan <-chan *natsio.Msg,
	path, sessionID, userID, deviceTopic string) (*wsft.FileInfo, error) {
	// stat the remote file
	req := wsft.StatFile{
		Path: &path,
	}
	if err := h.publishFileTransferProtoMessage(sessionID,
		userID, deviceTopic, wsft.MessageTypeStat, req, 0); err != nil {
		return nil, err
	}
	var fileInfo *wsft.FileInfo
	select {
	case rsp, ok := <-sessChan:
		if !ok {
			return nil, errFileTransferTimeout
		}
		var msg ws.ProtoMsg
		err := msgpack.Unmarshal(rsp.Data, &msg)
		if err != nil {
			return nil, fmt.Errorf("malformed message from device: %w", err)
		}
		if msg.Header.MsgType == ws.MessageTypeError {
			var errMsg ws.Error
			_ = msgpack.Unmarshal(msg.Body, &errMsg)
			errCode := http.StatusBadRequest
			if errMsg.Code > 0 {
				errCode = errMsg.Code
			}
			rspErr := NewError(
				fmt.Errorf("error received from device: %s", errMsg.Error),
				errCode,
			)
			return nil, rspErr
		}
		if msg.Header.Proto != ws.ProtoTypeFileTransfer ||
			msg.Header.MsgType != wsft.MessageTypeFileInfo {
			return nil, fmt.Errorf("unexpected response from device %q", msg.Header.MsgType)
		}
		err = msgpack.Unmarshal(msg.Body, &fileInfo)
		if err != nil {
			return nil, fmt.Errorf("malformed message body from device: %w", err)
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return fileInfo, nil
}

func (h ManagementController) downloadFileResponse(c *gin.Context, params *fileTransferParams,
	request *model.DownloadFileRequest) {
	ctx := c.Request.Context()
	// send a JSON-encoded error message in case of failure

	// subscribe to messages from the device
	deviceTopic := model.GetDeviceSubject(params.TenantID, params.Device.ID)
	sessionTopic := model.GetSessionSubject(params.TenantID, params.SessionID)
	subChan := make(chan *natsio.Msg, channelSize)
	defer close(subChan)
	sub, err := h.nats.ChanSubscribe(sessionTopic, subChan)
	if err != nil {
		h.handleResponseError(c, errors.Wrap(err, errFileTransferSubscribing.Error()))
		return
	}
	//nolint:errcheck
	defer sub.Unsubscribe()

	msgChan := chanTimeout(subChan, fileTransferTimeout)

	if err = h.filetransferHandshake(msgChan, params.SessionID, deviceTopic); err != nil {
		h.handleResponseError(c, err)
		return
	}
	// Inform the device that we're closing the session
	//nolint:errcheck
	defer h.publishControlMessage(params.SessionID, deviceTopic, ws.MessageTypeClose, nil)

	fileInfo, err := h.statFile(
		ctx, msgChan, *request.Path,
		params.SessionID, params.UserID, deviceTopic,
	)
	if err != nil {
		h.handleResponseError(c, fmt.Errorf("failed to retrieve file info: %w", err))
		return
	}
	if fileInfo.Mode == nil || !os.FileMode(*fileInfo.Mode).IsRegular() {
		h.handleResponseError(
			c,
			NewError(fmt.Errorf("path is not a regular file"), http.StatusBadRequest),
		)
		return
	}
	writeHeaders(c, fileInfo)
	if c.Request.Method == http.MethodHead {
		return
	}
	err = h.downloadFile(
		ctx, msgChan, c.Writer, *request.Path,
		params.SessionID, params.UserID, deviceTopic,
	)
	if err != nil {
		if !c.Writer.Written() {
			h.handleResponseError(c, err)
		}
		log.FromContext(ctx).
			Errorf("error downloading file from device: %s", err.Error())
	}
}

func (h ManagementController) downloadFile(
	ctx context.Context,
	msgChan <-chan *natsio.Msg,
	dst io.Writer,
	path, sessionID, userID, deviceTopic string,
) error {
	latestOffset := int64(0)
	bw := bufio.NewWriter(dst)
	numberOfChunks := 0
	req := wsft.GetFile{
		Path: &path,
	}
	if err := h.publishFileTransferProtoMessage(
		sessionID,
		userID,
		deviceTopic,
		wsft.MessageTypeGet,
		req, 0); err != nil {
		return err
	}
	for {
		select {
		case wsMessage, ok := <-msgChan:
			if !ok {
				return errFileTransferTimeout
			}

			// process the message
			msg, msgBody, err := h.decodeFileTransferProtoMessage(wsMessage.Data)
			if err != nil {
				return err
			}

			// process incoming messages from the device by type
			switch msg.Header.MsgType {

			// error message, stop here
			case wsft.MessageTypeError:
				err := msgBody.(*ws.Error)
				errCode := http.StatusInternalServerError
				if err.Code > 0 {
					errCode = err.Code
				}
				return NewError(errors.New(err.Error), errCode)

			// file data chunk
			case wsft.MessageTypeChunk:
				if msg.Body == nil {
					if err := h.publishFileTransferProtoMessage(
						sessionID, userID, deviceTopic,
						wsft.MessageTypeACK, nil,
						latestOffset); err != nil {
						return err
					}
					return bw.Flush()
				}

				// verify the offset property
				propOffset, _ := msg.Header.Properties[PropertyOffset].(int64)
				if propOffset != latestOffset {
					return NewError(errors.Wrap(errFileTransferFailed,
						"wrong offset received"), http.StatusInternalServerError)
				}
				latestOffset += int64(len(msg.Body))

				_, err := bw.Write(msg.Body)
				if err != nil {
					return err
				}

				numberOfChunks++
				if numberOfChunks >= ackSlidingWindowSend {
					if err := h.publishFileTransferProtoMessage(
						sessionID, userID, deviceTopic,
						wsft.MessageTypeACK, nil,
						latestOffset); err != nil {
						return err
					}
					numberOfChunks = 0
				}

			case ws.MessageTypePing:
				if err := h.publishFileTransferProtoMessage(
					sessionID, userID, deviceTopic,
					ws.MessageTypePong, nil,
					-1); err != nil {
					return err
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (h ManagementController) DownloadFile(c *gin.Context) {
	l := log.FromContext(c.Request.Context())

	params, statusCode, err := h.getFileTransferParams(c)
	if err != nil {
		l.Error(err)
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	path := c.Request.URL.Query().Get(paramDownloadPath)
	request := &model.DownloadFileRequest{
		Path: &path,
	}

	if err := request.Validate(); err != nil {
		l.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.Wrap(err, "bad request").Error(),
		})
		return
	}

	ctx := c.Request.Context()
	if err := h.app.DownloadFile(ctx, params.UserID, params.Device.ID,
		*request.Path); err != nil {
		l.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.Wrap(err, "bad request").Error(),
		})
		return
	}

	h.downloadFileResponse(c, params, request)
}

func (h ManagementController) uploadFileResponseHandleInboundMessages(
	c *gin.Context, params *fileTransferParams,
	msgChan chan *natsio.Msg, errorChan chan error,
	latestAckOffsets chan int64,
) {
	var latestAckOffset int64
	deviceTopic := model.GetDeviceSubject(params.TenantID, params.Device.ID)
	for {
		select {
		case wsMessage := <-msgChan:
			msg, msgBody, err := h.decodeFileTransferProtoMessage(
				wsMessage.Data)
			if err != nil {
				errorChan <- err
				return
			}

			// process incoming messages from the device by type
			switch msg.Header.MsgType {

			// error message, stop here
			case wsft.MessageTypeError:
				errorMsg := msgBody.(*ws.Error)
				errCode := http.StatusBadRequest
				if errorMsg.Code > 0 {
					errCode = errorMsg.Code
				}
				errorChan <- NewError(errors.New(errorMsg.Error), errCode)
				return

			// you can continue the upload
			case wsft.MessageTypeACK:
				propValue := msg.Header.Properties[PropertyOffset]
				propOffset, _ := propValue.(int64)
				if propOffset > latestAckOffset {
					latestAckOffset = propOffset
					select {
					case latestAckOffsets <- latestAckOffset:
					case <-latestAckOffsets:
						// Replace ack offset with the latest one
						latestAckOffsets <- latestAckOffset
					}
				}

			// handle ping messages
			case ws.MessageTypePing:
				if err := h.publishFileTransferProtoMessage(
					params.SessionID, params.UserID, deviceTopic,
					ws.MessageTypePong, nil,
					-1); err != nil {
					errorChan <- err
				}
			}
		case <-c.Done():
			return
		}
	}
}

// filetransferHandshake initiates a handshake and checks that the device
// is willing to accept file transfer requests.
func (h ManagementController) filetransferHandshake(
	sessChan <-chan *natsio.Msg, sessionID, deviceTopic string,
) error {
	if err := h.publishControlMessage(
		sessionID, deviceTopic,
		ws.MessageTypeOpen, ws.Open{
			Versions: []int{ws.ProtocolVersion},
		}); err != nil {
		return errFileTransferPublishing
	}
	select {
	case natsMsg, ok := <-sessChan:
		if !ok {
			return errFileTransferTimeout
		}
		var msg ws.ProtoMsg
		err := msgpack.Unmarshal(natsMsg.Data, &msg)
		if err != nil {
			return errFileTransferUnmarshalling
		}

		if msg.Header.MsgType == ws.MessageTypeError {
			erro := new(ws.Error)
			//nolint:errcheck
			msgpack.Unmarshal(natsMsg.Data, erro)
			errCode := http.StatusInternalServerError
			if erro.Code > 0 {
				errCode = erro.Code
			}
			rspErr := NewError(
				fmt.Errorf("handshake error from client: %s", erro.Error),
				errCode,
			)
			return fmt.Errorf("handshake error from client: %w", rspErr)
		} else if msg.Header.MsgType != ws.MessageTypeAccept {
			return errFileTransferNotImplemented
		}
		accept := new(ws.Accept)
		err = msgpack.Unmarshal(msg.Body, accept)
		if err != nil {
			return errFileTransferUnmarshalling
		}

		for _, proto := range accept.Protocols {
			if proto == ws.ProtoTypeFileTransfer {
				return nil
			}
		}
		// Let's try to be polite and close the session before returning
		//nolint:errcheck
		h.publishControlMessage(sessionID, deviceTopic, ws.MessageTypeClose, nil)
		return errFileTransferDisabled

	case <-time.After(fileTransferTimeout):
		return errFileTransferTimeout
	}
}

func (h ManagementController) uploadFileResponse(c *gin.Context, params *fileTransferParams,
	request *model.UploadFileRequest) {
	l := log.FromContext(c.Request.Context())

	// send a JSON-encoded error message in case of failure
	var responseError error
	errorStatusCode := http.StatusInternalServerError
	defer func() {
		if responseError != nil {
			l.Error(responseError.Error())
			var statusError *Error
			if errors.As(responseError, &statusError) {
				errorStatusCode = statusError.statusCode
			}
			c.JSON(errorStatusCode, gin.H{
				"error": responseError.Error(),
			})
			return
		}
	}()

	// subscribe to messages from the device
	deviceTopic := model.GetDeviceSubject(params.TenantID, params.Device.ID)
	sessionTopic := model.GetSessionSubject(params.TenantID, params.SessionID)
	msgChan := make(chan *natsio.Msg, channelSize)
	sub, err := h.nats.ChanSubscribe(sessionTopic, msgChan)
	if err != nil {
		responseError = errors.Wrap(err, errFileTransferSubscribing.Error())
		return
	}

	//nolint:errcheck
	defer sub.Unsubscribe()

	if err = h.filetransferHandshake(msgChan, params.SessionID, deviceTopic); err != nil {
		switch err {
		case errFileTransferTimeout:
			errorStatusCode = http.StatusRequestTimeout
		case errFileTransferNotImplemented, errFileTransferDisabled:
			errorStatusCode = http.StatusBadGateway
		}
		responseError = err
		return
	}

	// Inform the device that we're closing the session
	//nolint:errcheck
	defer h.publishControlMessage(params.SessionID, deviceTopic, ws.MessageTypeClose, nil)

	// initialize the file transfer
	req := wsft.UploadRequest{
		SrcPath: request.SrcPath,
		Path:    request.Path,
		UID:     request.UID,
		GID:     request.GID,
		Mode:    request.Mode,
	}
	if err := h.publishFileTransferProtoMessage(params.SessionID,
		params.UserID, deviceTopic, wsft.MessageTypePut, req, 0); err != nil {
		responseError = err
		return
	}

	// receive the message from the device
	select {
	case wsMessage := <-msgChan:
		msg, msgBody, err := h.decodeFileTransferProtoMessage(wsMessage.Data)
		if err != nil {
			responseError = err
			return
		}

		// process incoming messages from the device by type
		switch msg.Header.MsgType {

		// error message, stop here
		case wsft.MessageTypeError:
			errorMsg := msgBody.(*ws.Error)
			errorStatusCode = http.StatusBadRequest
			if errorMsg.Code > 0 {
				errorStatusCode = errorMsg.Code
			}
			responseError = NewError(errors.New(errorMsg.Error), errorStatusCode)
			return

		// you can continue the upload
		case wsft.MessageTypeACK:
		}

	// no message after timeout expired, stop here
	case <-time.After(fileTransferTimeout):
		errorStatusCode = http.StatusRequestTimeout
		responseError = errFileTransferTimeout
		return
	}

	// receive the ack message from the device
	latestAckOffsets := make(chan int64, 1)
	errorChan := make(chan error)
	go h.uploadFileResponseHandleInboundMessages(
		c, params, msgChan, errorChan, latestAckOffsets,
	)

	h.uploadFileResponseWriter(
		c, params, request, errorChan, latestAckOffsets, &errorStatusCode, &responseError,
	)
}

func (h ManagementController) uploadFileResponseWriter(c *gin.Context,
	params *fileTransferParams, request *model.UploadFileRequest,
	errorChan chan error, latestAckOffsets <-chan int64,
	errorStatusCode *int, responseError *error) {
	var (
		offset          int64
		latestAckOffset int64
	)
	deviceTopic := model.GetDeviceSubject(params.TenantID, params.Device.ID)

	timeout := time.NewTimer(fileTransferTimeout)
	data := make([]byte, fileTransferBufferSize)
	for {
		n, err := request.File.Read(data)
		if err != nil && err != io.EOF {
			if err == io.ErrUnexpectedEOF {
				*errorStatusCode = http.StatusBadRequest
				*responseError = errors.New(
					"malformed request body: " +
						"did not find closing multipart boundary",
				)
			} else {
				*responseError = err
			}
			return
		} else if n == 0 {
			if err := h.publishFileTransferProtoMessage(params.SessionID,
				params.UserID, deviceTopic, wsft.MessageTypeChunk, nil,
				offset); err != nil {
				*responseError = err
				return
			}
			break
		}

		// send the chunk
		if err := h.publishFileTransferProtoMessage(params.SessionID,
			params.UserID, deviceTopic, wsft.MessageTypeChunk, data[0:n],
			offset); err != nil {
			*responseError = err
			return
		}

		// update the offset
		offset += int64(n)

		// wait for acks, in case the ack sliding window is over
		if offset > latestAckOffset+int64(fileTransferBufferSize*ackSlidingWindowRecv) {
			timeout.Reset(fileTransferTimeout)
			select {
			case err := <-errorChan:
				*errorStatusCode = http.StatusBadRequest
				*responseError = err
				return
			case latestAckOffset = <-latestAckOffsets:
			case <-timeout.C:
				*errorStatusCode = http.StatusRequestTimeout
				*responseError = errFileTransferTimeout
				return
			}
		} else {
			// in case of error, report it
			select {
			case err := <-errorChan:
				*errorStatusCode = http.StatusBadRequest
				*responseError = err
				return
			default:
			}
		}

	}

	for offset > latestAckOffset {
		timeout.Reset(fileTransferTimeout)
		select {
		case latestAckOffset = <-latestAckOffsets:
		case <-timeout.C:
			*errorStatusCode = http.StatusRequestTimeout
			*responseError = errFileTransferTimeout
			return
		}
	}

	c.Writer.WriteHeader(http.StatusCreated)
}

func (h ManagementController) parseUploadFileRequest(c *gin.Context) (*model.UploadFileRequest,
	error) {
	reader, err := c.Request.MultipartReader()
	if err != nil {
		return nil, err
	}

	request := &model.UploadFileRequest{}
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		var n int
		data := make([]byte, fileTransferBufferSize)
		partName := part.FormName()
		switch partName {
		case fieldUploadPath, fieldUploadUID, fieldUploadGID, fieldUploadMode:
			n, err = part.Read(data)
			var value string
			if err == nil || err == io.EOF {
				value = string(data[:n])
			}
			switch partName {
			case fieldUploadPath:
				request.Path = &value
			case fieldUploadUID:
				v, err := strconv.Atoi(string(data[:n]))
				if err != nil {
					return nil, err
				}
				nUID := uint32(v)
				request.UID = &nUID
			case fieldUploadGID:
				v, err := strconv.Atoi(string(data[:n]))
				if err != nil {
					return nil, err
				}
				nGID := uint32(v)
				request.GID = &nGID
			case fieldUploadMode:
				v, err := strconv.ParseUint(string(data[:n]), 8, 32)
				if err != nil {
					return nil, err
				}
				nMode := uint32(v)
				request.Mode = &nMode
			}
			part.Close()
		case fieldUploadFile:
			filename := part.FileName()
			request.SrcPath = &filename
			request.File = part
		}
		// file is the last part we can process, in order to avoid loading it in memory
		if request.File != nil {
			break
		}
	}

	return request, nil
}

func (h ManagementController) UploadFile(c *gin.Context) {
	l := log.FromContext(c.Request.Context())

	params, statusCode, err := h.getFileTransferParams(c)
	if err != nil {
		l.Error(err.Error())
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	request, err := h.parseUploadFileRequest(c)
	if err != nil {
		l.Error(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := request.Validate(); err != nil {
		l.Error(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.Wrap(err, "bad request").Error(),
		})
		return
	}

	defer request.File.Close()

	ctx := c.Request.Context()
	if err := h.app.UploadFile(ctx, params.UserID, params.Device.ID,
		*request.Path); err != nil {
		l.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.Wrap(err, "bad request").Error(),
		})
		return
	}

	h.uploadFileResponse(c, params, request)
}
