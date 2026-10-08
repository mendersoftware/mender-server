// Copyright 2026 Northern.tech AS
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
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/mendersoftware/mender-server/pkg/identity"
	stream_mocks "github.com/mendersoftware/mender-server/pkg/stream/mocks"
	"github.com/mendersoftware/mender-server/pkg/ws"
	wsft "github.com/mendersoftware/mender-server/pkg/ws/filetransfer"
	app_mocks "github.com/mendersoftware/mender-server/services/deviceconnect/app/mocks"
	nats_mocks "github.com/mendersoftware/mender-server/services/deviceconnect/client/nats/mocks"
)

// A small upload fits in a single chunk. The device can only report that it
// cannot write the file once it has seen the chunk, i.e. after the server has
// already sent the final (empty) chunk. The server must not report success
// before the device had a chance to answer.
func TestManagementUploadFileV2ErrorAfterFinalChunk(t *testing.T) {
	originalFileTransferTimeout := fileTransferTimeout
	t.Cleanup(func() { fileTransferTimeout = originalFileTransferTimeout })
	fileTransferTimeout = 2 * time.Second

	app := &app_mocks.App{}
	defer app.AssertExpectations(t)
	natsClient := &nats_mocks.Client{}
	conn := stream_mocks.NewConn(t)

	var sessionID string
	natsClient.On("Connect", contextMatcher, mock.MatchedBy(func(srcAddr string) bool {
		var ok bool
		_, sessionID, ok = strings.Cut(srcAddr, ":")
		return ok
	}), "000000000000000000000000:1234567890").
		Return(conn, nil).
		Once()

	finalChunkSent := make(chan struct{})
	conn.On("Close", contextMatcher).Return(nil).Maybe()
	conn.On("Send", contextMatcher, mock.Anything).
		Return(func(_ context.Context, data []byte) error {
			msg := ws.ProtoMsg{}
			if assert.NoError(t, msgpack.Unmarshal(data, &msg)) &&
				msg.Header.Proto == ws.ProtoTypeFileTransferV2 &&
				msg.Header.MsgType == wsft.MessageTypeChunk &&
				len(msg.Body) == 0 {
				close(finalChunkSent)
			}
			return nil
		})
	conn.On("Recv", contextMatcher).
		Return(func(context.Context) ([]byte, error) {
			b, _ := msgpack.Marshal(ws.Accept{
				Version:   ws.ProtocolVersion,
				Protocols: []ws.ProtoType{ws.ProtoTypeFileTransferV2},
			})
			return msgpack.Marshal(&ws.ProtoMsg{
				Header: ws.ProtoHdr{
					Proto:     ws.ProtoTypeControl,
					MsgType:   ws.MessageTypeAccept,
					SessionID: sessionID,
				},
				Body: b,
			})
		}).
		Once()
	conn.On("Recv", contextMatcher).
		Return(func(context.Context) ([]byte, error) {
			// the device answers only after the final chunk was sent
			<-finalChunkSent
			b, _ := msgpack.Marshal(ws.Error{
				Error: "failed to create target file",
				Code:  http.StatusBadRequest,
			})
			return msgpack.Marshal(&ws.ProtoMsg{
				Header: ws.ProtoHdr{
					Proto:     ws.ProtoTypeFileTransferV2,
					MsgType:   wsft.MessageTypeError,
					SessionID: sessionID,
				},
				Body: b,
			})
		}).
		Maybe()

	router, _ := NewRouter(app, natsClient, nil)

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	w.SetBoundary("boundary")
	_ = w.WriteField(fieldUploadPath, "/does/not/exist/dummy.txt")
	_ = w.WriteField(fieldUploadUID, "0")
	_ = w.WriteField(fieldUploadGID, "0")
	_ = w.WriteField(fieldUploadMode, "0600")
	fw, _ := w.CreateFormFile(fieldUploadFile, "dummy.txt")
	_, _ = fw.Write([]byte("dummy"))
	_ = w.Close()

	url := strings.Replace(APIURLManagementDeviceUpload, ":deviceId", "1234567890", 1)
	req, err := http.NewRequest(http.MethodPut, "http://localhost"+url, &buf)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	req.Header.Add("Content-Type", "multipart/form-data; boundary=\"boundary\"")
	req.Header.Set(headerAuthorization, "Bearer "+GenerateJWT(identity.Identity{
		Subject: "00000000-0000-0000-0000-000000000000",
		Tenant:  "000000000000000000000000",
		IsUser:  true,
	}))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
}
