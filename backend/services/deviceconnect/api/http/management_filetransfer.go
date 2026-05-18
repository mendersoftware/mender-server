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
	"slices"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"
	"github.com/mendersoftware/mender-server/pkg/stream"
	"github.com/mendersoftware/mender-server/pkg/ws"
	wsft "github.com/mendersoftware/mender-server/pkg/ws/filetransfer"

	"github.com/mendersoftware/mender-server/services/deviceconnect/model"
)

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

func (h ManagementController) publishFileTransferProtoMessage(
	ctx context.Context,
	conn stream.Conn,
	userID, sessionID, msgType string,
	body interface{},
	offset int64,
) error {
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

	err = conn.Send(ctx, data)
	if err != nil {
		return errors.Wrap(err, errFileTransferPublishing.Error())
	}
	return nil
}

func (h ManagementController) publishControlMessage(
	ctx context.Context, conn stream.Conn, sessionID, messageType string, body interface{},
) error {
	msg := &ws.ProtoMsg{
		Header: ws.ProtoHdr{
			SessionID: sessionID,
			Proto:     ws.ProtoTypeControl,
			MsgType:   messageType,
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
	err = conn.Send(ctx, data)
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
	if !c.Writer.Written() {
		var statusError *Error
		var statusCode int
		switch {
		case errors.As(err, &statusError):
			statusCode = statusError.statusCode
		case errors.Is(err, context.DeadlineExceeded):
			statusCode = http.StatusRequestTimeout
		default:
			rest.RenderInternalError(c, err)
		}
		rest.RenderError(c, statusCode, err)
	} else {
		l.Warn("response already written")
	}
}

func (h ManagementController) statFile(
	ctx context.Context,
	conn stream.Conn,
	path, userID, sessionID string) (*wsft.FileInfo, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	// stat the remote file
	req := wsft.StatFile{
		Path: &path,
	}
	if err := h.publishFileTransferProtoMessage(
		ctx, conn, userID, sessionID,
		wsft.MessageTypeStat, req, 0); err != nil {
		return nil, err
	}
	data, err := conn.Recv(ctx)
	if err != nil {
		return nil, err
	}
	var msg ws.ProtoMsg
	err = msgpack.Unmarshal(data, &msg)
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
	var fileInfo wsft.FileInfo
	err = msgpack.Unmarshal(msg.Body, &fileInfo)
	if err != nil {
		return nil, fmt.Errorf("malformed message body from device: %w", err)
	}
	return &fileInfo, nil
}

func (h ManagementController) downloadFileResponse(
	c *gin.Context, conn stream.Conn, userID, sessionID string, request *model.DownloadFileRequest,
) {
	ctx := c.Request.Context()
	// send a JSON-encoded error message in case of failure

	if err := h.filetransferHandshake(ctx, conn, sessionID); err != nil {
		h.handleResponseError(c, err)
		return
	}
	// Inform the device that we're closing the session
	//nolint:errcheck
	defer h.publishControlMessage(ctx, conn, sessionID, ws.MessageTypeClose, nil)

	fileInfo, err := h.statFile(
		ctx, conn, *request.Path,
		userID, sessionID,
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
		ctx, conn, c.Writer, *request.Path, userID, sessionID,
	)
	if err != nil {
		h.handleResponseError(c, err)
		log.FromContext(ctx).
			Errorf("error downloading file from device: %s", err.Error())
	}
}

type timerCtx struct {
	*time.Timer
	context.Context
}

func newTimerCtx(ctx context.Context, timeout time.Duration) (timerCtx, context.CancelFunc) {
	ctx, cancel := context.WithCancelCause(ctx)
	return timerCtx{
		Timer: time.AfterFunc(timeout, func() {
			cancel(context.DeadlineExceeded)
		}),
		Context: ctx,
	}, func() { cancel(context.Canceled) }
}

func (ctx timerCtx) Err() error {
	// Return Cause instead of Err to get the appropriate Done error.
	return context.Cause(ctx.Context)
}

func (h ManagementController) downloadFile(
	ctx context.Context,
	conn stream.Conn,
	dst io.Writer,
	path, userID, sessionID string,
) error {
	latestOffset := int64(0)
	bw := bufio.NewWriter(dst)
	numberOfChunks := 0
	req := wsft.GetFile{
		Path: &path,
	}
	if err := h.publishFileTransferProtoMessage(
		ctx, conn,
		userID, sessionID,
		wsft.MessageTypeGet,
		req, 0); err != nil {
		return err
	}
	timerCtx, cancel := newTimerCtx(ctx, fileTransferTimeout)
	defer cancel()
	ctx = timerCtx
	for {
		b, err := conn.Recv(ctx)
		if err != nil {
			return err
		}
		timerCtx.Reset(fileTransferTimeout)

		// process the message
		msg, msgBody, err := h.decodeFileTransferProtoMessage(b)
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
			if len(msg.Body) == 0 {
				if err := h.publishFileTransferProtoMessage(
					ctx, conn, userID, sessionID,
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
					ctx, conn, userID, sessionID,
					wsft.MessageTypeACK, nil,
					latestOffset); err != nil {
					return err
				}
				numberOfChunks = 0
			}

		case ws.MessageTypePing:
			if err := h.publishFileTransferProtoMessage(
				ctx, conn, userID, sessionID,
				ws.MessageTypePong, nil,
				-1); err != nil {
				return err
			}
		}
	}
}

func (h ManagementController) DownloadFile(c *gin.Context) {
	ctx := c.Request.Context()
	idata := identity.FromContext(ctx)
	if idata == nil || !idata.IsUser {
		rest.RenderError(c, http.StatusUnauthorized, ErrMissingUserAuthentication)
		return
	}

	path := c.Request.URL.Query().Get(paramDownloadPath)
	request := &model.DownloadFileRequest{
		Path: &path,
	}

	if err := request.Validate(); err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	tenantID := idata.Tenant
	deviceID := c.Param("deviceId")
	sessionID := uuid.NewString()

	srcAddr := fmt.Sprintf("%s:%s", tenantID, sessionID)
	ctxWithTimeout, cancel := context.WithTimeout(ctx, fileTransferTimeout)
	conn, err := h.nats.Connect(ctxWithTimeout, srcAddr, deviceID)
	cancel()
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			rest.RenderError(c, http.StatusRequestTimeout, err)
		case errors.Is(err, stream.ErrConnectionRefused):
			rest.RenderError(c, http.StatusConflict, fmt.Errorf("device disconnected"))
		default:
			rest.RenderInternalError(c, err)
		}
		return
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second*10)
		_ = conn.Close(ctx)
		cancel()
	}()
	if err := h.app.DownloadFile(ctx, idata.Subject, deviceID, *request.Path); err != nil {
		rest.RenderInternalError(c, err)
		return
	}

	h.downloadFileResponse(c, conn, idata.Subject, sessionID, request)
}

func (h ManagementController) uploadFileResponseHandleInboundMessages(
	ctx context.Context,
	conn stream.Conn,
	userID, sessionID string,
	errorChan chan<- error,
	latestAckOffsets chan int64,
) {
	l := log.FromContext(ctx)
	defer l.SimpleRecovery()
	var latestAckOffset int64
	for {
		data, err := conn.Recv(ctx)
		if err != nil {
			errorChan <- err
			return
		}
		msg, msgBody, err := h.decodeFileTransferProtoMessage(data)
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
				ctx, conn, userID, sessionID,
				ws.MessageTypePong, nil,
				-1); err != nil {
				errorChan <- err
			}
		}
	}
}

// filetransferHandshake initiates a handshake and checks that the device
// is willing to accept file transfer requests.
func (h ManagementController) filetransferHandshake(
	ctx context.Context, conn stream.Conn, sessionID string,
) error {
	ctx, cancel := context.WithTimeout(ctx, fileTransferTimeout)
	defer cancel()
	if err := h.publishControlMessage(
		ctx, conn, sessionID, ws.MessageTypeOpen, ws.Open{
			Versions: []int{ws.ProtocolVersion},
		}); err != nil {
		return errFileTransferPublishing
	}
	data, err := conn.Recv(ctx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return errFileTransferTimeout
		}
		return err
	}
	var msg ws.ProtoMsg
	err = msgpack.Unmarshal(data, &msg)
	if err != nil {
		return err
	}

	if msg.Header.MsgType == ws.MessageTypeError {
		erro := new(ws.Error)
		//nolint:errcheck
		msgpack.Unmarshal(data, erro)
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
		return err
	}

	if slices.Contains(accept.Protocols, ws.ProtoTypeFileTransfer) {
		return nil
	}
	// Let's try to be polite and close the session before returning
	//nolint:errcheck
	h.publishControlMessage(ctx, conn, sessionID, ws.MessageTypeClose, nil)
	return errFileTransferDisabled
}

func (h ManagementController) uploadFileResponse(
	c *gin.Context,
	conn stream.Conn,
	request *model.UploadFileRequest,
	idata identity.Identity,
	sessionID string,
) {
	userID := idata.Subject
	ctx := c.Request.Context()
	if err := h.filetransferHandshake(ctx, conn, sessionID); err != nil {
		h.handleResponseError(c, err)
		return
	}

	// Inform the device that we're closing the session
	//nolint:errcheck
	defer h.publishControlMessage(ctx, conn, sessionID, ws.MessageTypeClose, nil)

	ctxWithTimeout, cancel := context.WithTimeout(ctx, fileTransferTimeout)
	defer cancel()

	// initialize the file transfer
	req := wsft.UploadRequest{
		SrcPath: request.SrcPath,
		Path:    request.Path,
		UID:     request.UID,
		GID:     request.GID,
		Mode:    request.Mode,
	}
	if err := h.publishFileTransferProtoMessage(
		ctxWithTimeout, conn,
		userID, sessionID,
		wsft.MessageTypePut, req, 0,
	); err != nil {
		h.handleResponseError(c, err)
		return
	}

	var (
		msg     *ws.ProtoMsg
		msgBody any
	)
	// receive the message from the device
	data, err := conn.Recv(ctxWithTimeout)
	if err == nil {
		msg, msgBody, err = h.decodeFileTransferProtoMessage(data)
	}
	if err != nil {
		h.handleResponseError(c, err)
		return
	}

	// process incoming messages from the device by type
	switch msg.Header.MsgType {

	// error message, stop here
	case wsft.MessageTypeError:
		errorMsg := msgBody.(*ws.Error)
		errorStatusCode := http.StatusBadRequest
		if errorMsg.Code > 0 {
			errorStatusCode = errorMsg.Code
		}
		rest.RenderError(c, errorStatusCode, NewError(errors.New(errorMsg.Error), errorStatusCode))
		return

	// you can continue the upload
	case wsft.MessageTypeACK:
	}

	// receive the ack message from the device
	latestAckOffsets := make(chan int64, 1)
	errorChan := make(chan error)
	go h.uploadFileResponseHandleInboundMessages(
		ctx, conn, userID, sessionID, errorChan, latestAckOffsets,
	)

	err = h.uploadFileResponseWriter(
		c, conn, userID, sessionID, request, errorChan, latestAckOffsets,
	)
	if err != nil {
		h.handleResponseError(c, err)
		return
	}
	c.Status(http.StatusCreated)
}

func (h ManagementController) uploadFileResponseWriter(ctx context.Context,
	conn stream.Conn,
	userID, sessionID string,
	request *model.UploadFileRequest,
	errorChan <-chan error, latestAckOffsets <-chan int64,
) error {
	var (
		offset          int64
		latestAckOffset int64
		err             error
		n               int
	)
	timeout := time.NewTimer(fileTransferTimeout)
	data := make([]byte, fileTransferBufferSize)
	for {
		n, err = request.File.Read(data)
		if n > 0 {
			// send the chunk
			errSend := h.publishFileTransferProtoMessage(
				ctx, conn, userID, sessionID,
				wsft.MessageTypeChunk, data[:n], offset,
			)
			if errSend != nil {
				return errSend
			}
			// update the offset
			offset += int64(n)
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = h.publishFileTransferProtoMessage(
					ctx, conn, userID, sessionID,
					wsft.MessageTypeChunk, nil, offset,
				)
			}
			break
		}

		// wait for acks, in case the ack sliding window is over
		if offset > latestAckOffset+int64(fileTransferBufferSize*ackSlidingWindowRecv) {
			timeout.Reset(fileTransferTimeout)
			select {
			case err := <-errorChan:
				return err
			case latestAckOffset = <-latestAckOffsets:
			case <-timeout.C:
				return err
			}
		} else {
			// in case of error, report it
			select {
			case err := <-errorChan:
				return err
			default:
			}
		}
	}
	if err != nil {
		return err
	}

	for offset > latestAckOffset {
		timeout.Reset(fileTransferTimeout)
		select {
		case latestAckOffset = <-latestAckOffsets:
		case err := <-errorChan:
			return err
		case <-timeout.C:
			return errFileTransferTimeout
		}
	}
	return nil
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
	ctx := c.Request.Context()
	idata := identity.FromContext(ctx)
	if idata == nil || !idata.IsUser {
		rest.RenderError(c, http.StatusUnauthorized, ErrMissingUserAuthentication)
		return
	}
	request, err := h.parseUploadFileRequest(c)
	if err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}
	if err := request.Validate(); err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}

	defer request.File.Close()

	tenantID := idata.Tenant
	deviceID := c.Param("deviceId")
	sessionID := uuid.NewString()

	srcAddr := fmt.Sprintf("%s:%s", tenantID, sessionID)
	ctxWithTimeout, cancel := context.WithTimeout(ctx, fileTransferTimeout)
	conn, err := h.nats.Connect(ctxWithTimeout, srcAddr, deviceID)
	cancel()
	if err != nil {
		switch {
		case errors.Is(err, stream.ErrConnectionRefused):
			rest.RenderError(c, http.StatusConflict, fmt.Errorf("device disconnected"))
		case errors.Is(err, context.DeadlineExceeded):
			rest.RenderError(c, http.StatusRequestTimeout, err)
		default:
			rest.RenderInternalError(c, err)
		}
		return
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second*10)
		_ = conn.Close(ctx)
		cancel()
	}()

	if err := h.app.UploadFile(ctx, idata.Subject, deviceID,
		*request.Path); err != nil {
		l.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errors.Wrap(err, "bad request").Error(),
		})
		return
	}

	h.uploadFileResponse(c, conn, request, *idata, sessionID)
}
