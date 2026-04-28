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
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/stream"
	stream_mocks "github.com/mendersoftware/mender-server/pkg/stream/mocks"
	"github.com/mendersoftware/mender-server/pkg/ws"

	wsft "github.com/mendersoftware/mender-server/pkg/ws/filetransfer"
	"github.com/mendersoftware/mender-server/pkg/ws/shell"
	app_mocks "github.com/mendersoftware/mender-server/services/deviceconnect/app/mocks"
	nats_mocks "github.com/mendersoftware/mender-server/services/deviceconnect/client/nats/mocks"
)

func string2pointer(v string) *string {
	return &v
}

func uint322pointer(v uint32) *uint32 {
	return &v
}

func int642pointer(v int64) *int64 {
	return &v
}

func TestManagementDownloadFile(t *testing.T) {
	originalFileTransferTimeout := fileTransferTimeout
	originalAckSlidingWindowSend := ackSlidingWindowSend
	t.Cleanup(func() {
		fileTransferTimeout = originalFileTransferTimeout
		ackSlidingWindowSend = originalAckSlidingWindowSend
	})
	fileTransferTimeout = 2 * time.Second
	ackSlidingWindowSend = 1

	testCases := []struct {
		Name     string
		DeviceID string
		Path     string
		Identity *identity.Identity

		DeviceFunc         func(t *testing.T, nc *nats_mocks.Client)
		AppDownloadFile    bool
		AppDownloadFileErr error

		HTTPStatus int
		HTTPBody   []byte
	}{
		{
			Name:     "ok, successful download, single chunk",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Path: "/absolute/path",

			DeviceFunc: func(t *testing.T, client *nats_mocks.Client) {
				conn := stream_mocks.NewConn(t)
				var sessionID string
				client.On("Connect", contextMatcher, mock.MatchedBy(func(srcAddr string) bool {
					var (
						tenantID string
						ok       bool
					)
					tenantID, sessionID, ok = strings.Cut(srcAddr, ":")
					return assert.Truef(t, ok, "unexpected srcAddr format: %s", srcAddr) &&
						assert.Equal(t, "000000000000000000000000", tenantID)
				}), "1234567890").
					Return(conn, nil).
					Once()
				conn.On("Close", contextMatcher).
					Return(nil).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						// accept the open request
						b, _ := msgpack.Marshal(ws.Accept{
							Version:   ws.ProtocolVersion,
							Protocols: []ws.ProtoType{ws.ProtoTypeFileTransfer},
						})
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeControl,
								MsgType:   ws.MessageTypeAccept,
								SessionID: sessionID,
							},
							Body: b,
						}
						return msgpack.Marshal(msg)
					}).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						// file info response
						body := wsft.FileInfo{
							Path: string2pointer("/absolute/path"),
							UID:  uint322pointer(0),
							GID:  uint322pointer(0),
							Mode: uint322pointer(777),
							Size: int642pointer(10),
						}
						bodyData, _ := msgpack.Marshal(body)
						return msgpack.Marshal(&ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeFileInfo,
								SessionID: sessionID,
							},
							Body: bodyData,
						})
					}).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						// first chunk
						return msgpack.Marshal(&ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeChunk,
								SessionID: sessionID,
								Properties: map[string]interface{}{
									PropertyOffset: int64(0),
								},
							},
							Body: []byte("12345"),
						})
					}).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						// final chunk
						return msgpack.Marshal(&ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeChunk,
								SessionID: sessionID,
								Properties: map[string]interface{}{
									PropertyOffset: int64(5),
								},
							},
							Body: nil,
						})
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							assert.Equal(t, ws.MessageTypeOpen, msg.Header.MsgType)
						}
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							assert.Equal(t, wsft.MessageTypeStat, msg.Header.MsgType)
						}
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							assert.Equal(t, wsft.MessageTypeGet, msg.Header.MsgType)
						}
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							assert.Equal(t, wsft.MessageTypeACK, msg.Header.MsgType)
						}
						return nil
					}).
					Twice().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							assert.Equal(t, ws.MessageTypeClose, msg.Header.MsgType)
						}
						return nil
					}).
					Once()
			},
			AppDownloadFile: true,

			HTTPStatus: http.StatusOK,
			HTTPBody:   []byte("12345"),
		},
		{
			Name:     "ok, successful download, two chunks",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Path: "/absolute/path",

			DeviceFunc: func(t *testing.T, client *nats_mocks.Client) {
				conn := stream_mocks.NewConn(t)
				var sessionID string
				client.On("Connect", contextMatcher, mock.MatchedBy(func(srcAddr string) bool {
					var (
						tenantID string
						ok       bool
					)
					tenantID, sessionID, ok = strings.Cut(srcAddr, ":")
					return assert.Truef(t, ok, "unexpected srcAddr format: %s", srcAddr) &&
						assert.Equal(t, "000000000000000000000000", tenantID)
				}), "1234567890").
					Return(conn, nil).
					Once()
				conn.On("Close", contextMatcher).
					Return(nil).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						// accept the open request
						b, _ := msgpack.Marshal(ws.Accept{
							Version:   ws.ProtocolVersion,
							Protocols: []ws.ProtoType{ws.ProtoTypeFileTransfer},
						})
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeControl,
								MsgType:   ws.MessageTypeAccept,
								SessionID: sessionID,
							},
							Body: b,
						}
						return msgpack.Marshal(msg)
					}).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						// file info response
						body := wsft.FileInfo{
							Path: string2pointer("/absolute/path"),
							UID:  uint322pointer(0),
							GID:  uint322pointer(0),
							Mode: uint322pointer(777),
							Size: int642pointer(10),
						}
						bodyData, _ := msgpack.Marshal(body)
						return msgpack.Marshal(&ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeFileInfo,
								SessionID: sessionID,
							},
							Body: bodyData,
						})
					}).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						// first chunk
						return msgpack.Marshal(&ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeChunk,
								SessionID: sessionID,
								Properties: map[string]interface{}{
									PropertyOffset: int64(0),
								},
							},
							Body: []byte("12345"),
						})
					}).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						// second chunk
						return msgpack.Marshal(&ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeChunk,
								SessionID: sessionID,
								Properties: map[string]interface{}{
									PropertyOffset: int64(5),
								},
							},
							Body: []byte("67890"),
						})
					}).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						// final chunk
						return msgpack.Marshal(&ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeChunk,
								SessionID: sessionID,
								Properties: map[string]interface{}{
									PropertyOffset: int64(10),
								},
							},
							Body: nil,
						})
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							assert.Equal(t, ws.MessageTypeOpen, msg.Header.MsgType)
						}
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							assert.Equal(t, wsft.MessageTypeStat, msg.Header.MsgType)
						}
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							assert.Equal(t, wsft.MessageTypeGet, msg.Header.MsgType)
						}
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							assert.Equal(t, wsft.MessageTypeACK, msg.Header.MsgType)
						}
						return nil
					}).
					Times(3).
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							assert.Equal(t, ws.MessageTypeClose, msg.Header.MsgType)
						}
						return nil
					}).
					Once()
			},
			AppDownloadFile: true,

			HTTPStatus: http.StatusOK,
			HTTPBody:   []byte("1234567890"),
		},
		{
			Name:     "ko, file not found",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Path: "/absolute/path",

			DeviceFunc: func(t *testing.T, client *nats_mocks.Client) {
				conn := stream_mocks.NewConn(t)
				var sessionID string
				client.On("Connect", contextMatcher, mock.MatchedBy(func(srcAddr string) bool {
					var (
						tenantID string
						ok       bool
					)
					tenantID, sessionID, ok = strings.Cut(srcAddr, ":")
					return assert.Truef(t, ok, "unexpected srcAddr format: %s", srcAddr) &&
						assert.Equal(t, "000000000000000000000000", tenantID)
				}), "1234567890").
					Return(conn, nil).
					Once()
				conn.On("Close", contextMatcher).
					Return(nil).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						// accept the open request
						b, _ := msgpack.Marshal(ws.Accept{
							Version:   ws.ProtocolVersion,
							Protocols: []ws.ProtoType{ws.ProtoTypeFileTransfer},
						})
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeControl,
								MsgType:   ws.MessageTypeAccept,
								SessionID: sessionID,
							},
							Body: b,
						}
						return msgpack.Marshal(msg)
					}).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						// file info response
						body := wsft.Error{
							Error:       string2pointer("file not found"),
							MessageType: string2pointer(wsft.MessageTypeStat),
						}
						bodyData, _ := msgpack.Marshal(body)
						return msgpack.Marshal(&ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeError,
								SessionID: sessionID,
							},
							Body: bodyData,
						})
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							assert.Equal(t, ws.MessageTypeOpen, msg.Header.MsgType)
						}
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							assert.Equal(t, wsft.MessageTypeStat, msg.Header.MsgType)
						}
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							assert.Equal(t, ws.MessageTypeClose, msg.Header.MsgType)
						}
						return nil
					}).
					Once()
			},
			AppDownloadFile: true,

			HTTPStatus: http.StatusBadRequest,
		},
		{
			Name:     "ko, not a regular file",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Path: "/absolute/path",

			DeviceFunc: func(t *testing.T, client *nats_mocks.Client) {
				conn := stream_mocks.NewConn(t)
				var sessionID string
				client.On("Connect", contextMatcher, mock.MatchedBy(func(srcAddr string) bool {
					var (
						tenantID string
						ok       bool
					)
					tenantID, sessionID, ok = strings.Cut(srcAddr, ":")
					return assert.Truef(t, ok, "unexpected srcAddr format: %s", srcAddr) &&
						assert.Equal(t, "000000000000000000000000", tenantID)
				}), "1234567890").
					Return(conn, nil).
					Once()
				conn.On("Close", contextMatcher).
					Return(nil).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						// accept the open request
						b, _ := msgpack.Marshal(ws.Accept{
							Version:   ws.ProtocolVersion,
							Protocols: []ws.ProtoType{ws.ProtoTypeFileTransfer},
						})
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeControl,
								MsgType:   ws.MessageTypeAccept,
								SessionID: sessionID,
							},
							Body: b,
						}
						return msgpack.Marshal(msg)
					}).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						// file info response
						body := wsft.FileInfo{
							Path: string2pointer("/absolute/path"),
							UID:  uint322pointer(0),
							GID:  uint322pointer(0),
							Mode: uint322pointer(777 | uint32(os.ModeDir)),
							Size: int642pointer(10),
						}
						bodyData, _ := msgpack.Marshal(body)
						return msgpack.Marshal(&ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeFileInfo,
								SessionID: sessionID,
							},
							Body: bodyData,
						})
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							assert.Equal(t, ws.MessageTypeOpen, msg.Header.MsgType)
						}
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							assert.Equal(t, wsft.MessageTypeStat, msg.Header.MsgType)
						}
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							assert.Equal(t, ws.MessageTypeClose, msg.Header.MsgType)
						}
						return nil
					}).
					Once()
			},
			AppDownloadFile: true,

			HTTPStatus: http.StatusBadRequest,
		},
		{
			Name:     "ko, error between chunks",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Path: "/absolute/path",

			DeviceFunc: func(t *testing.T, client *nats_mocks.Client) {
				conn := stream_mocks.NewConn(t)
				var sessionID string
				client.On("Connect", contextMatcher, mock.MatchedBy(func(srcAddr string) bool {
					var (
						tenantID string
						ok       bool
					)
					tenantID, sessionID, ok = strings.Cut(srcAddr, ":")
					return assert.Truef(t, ok, "unexpected srcAddr format: %s", srcAddr) &&
						assert.Equal(t, "000000000000000000000000", tenantID)
				}), "1234567890").
					Return(conn, nil).
					Once()
				conn.On("Close", contextMatcher).
					Return(nil).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						// accept the open request
						b, _ := msgpack.Marshal(ws.Accept{
							Version:   ws.ProtocolVersion,
							Protocols: []ws.ProtoType{ws.ProtoTypeFileTransfer},
						})
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeControl,
								MsgType:   ws.MessageTypeAccept,
								SessionID: sessionID,
							},
							Body: b,
						}
						return msgpack.Marshal(msg)
					}).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						// file info response
						body := wsft.FileInfo{
							Path: string2pointer("/absolute/path"),
							UID:  uint322pointer(0),
							GID:  uint322pointer(0),
							Mode: uint322pointer(777),
							Size: int642pointer(10),
						}
						bodyData, _ := msgpack.Marshal(body)
						return msgpack.Marshal(&ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeFileInfo,
								SessionID: sessionID,
							},
							Body: bodyData,
						})
					}).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						// first chunk
						return msgpack.Marshal(&ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeChunk,
								SessionID: sessionID,
								Properties: map[string]interface{}{
									PropertyOffset: int64(0),
								},
							},
							Body: []byte("12345"),
						})
					}).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						// final chunk
						errBody := wsft.Error{
							Error:       string2pointer("generic error"),
							MessageType: string2pointer(wsft.MessageTypeStat),
						}
						bodyData, _ := msgpack.Marshal(errBody)
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeError,
								SessionID: sessionID,
							},
							Body: bodyData,
						}
						return msgpack.Marshal(msg)
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							assert.Equal(t, ws.MessageTypeOpen, msg.Header.MsgType)
						}
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							assert.Equal(t, wsft.MessageTypeStat, msg.Header.MsgType)
						}
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							assert.Equal(t, wsft.MessageTypeGet, msg.Header.MsgType)
						}
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							assert.Equal(t, wsft.MessageTypeACK, msg.Header.MsgType)
						}
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							assert.Equal(t, ws.MessageTypeClose, msg.Header.MsgType)
						}
						return nil
					}).
					Once()
			},
			AppDownloadFile: true,

			HTTPStatus: http.StatusInternalServerError,
		},
		{
			Name:     "ko, request timeout",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Path: "/absolute/path",

			DeviceFunc: func(t *testing.T, client *nats_mocks.Client) {
				conn := stream_mocks.NewConn(t)
				var sessionID string
				client.On("Connect", contextMatcher, mock.MatchedBy(func(srcAddr string) bool {
					var (
						tenantID string
						ok       bool
					)
					tenantID, sessionID, ok = strings.Cut(srcAddr, ":")
					return assert.Truef(t, ok, "unexpected srcAddr format: %s", srcAddr) &&
						assert.Equal(t, "000000000000000000000000", tenantID)
				}), "1234567890").
					Return(conn, nil).
					Once()
				conn.On("Close", contextMatcher).
					Return(nil).
					Once().
					On("Recv", contextMatcher).
					Return(func(ctx context.Context) ([]byte, error) {
						select {
						case <-ctx.Done():
							return nil, ctx.Err()
						case <-time.After(fileTransferTimeout * 2):
							t.Fatalf("file transfer test did not time out as expected!")
							return nil, fmt.Errorf("test failure")
						}
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							assert.Equal(t, ws.MessageTypeOpen, msg.Header.MsgType)
						}
						return nil
					}).
					Once()
			},
			AppDownloadFile: true,

			HTTPStatus: http.StatusRequestTimeout,
		},
		{
			Name:     "ko, request timeout between chunks",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Path: "/absolute/path",

			DeviceFunc: func(t *testing.T, client *nats_mocks.Client) {
				conn := stream_mocks.NewConn(t)
				var sessionID string
				client.On("Connect", contextMatcher, mock.MatchedBy(func(srcAddr string) bool {
					var (
						tenantID string
						ok       bool
					)
					tenantID, sessionID, ok = strings.Cut(srcAddr, ":")
					return assert.Truef(t, ok, "unexpected srcAddr format: %s", srcAddr) &&
						assert.Equal(t, "000000000000000000000000", tenantID)
				}), "1234567890").
					Return(conn, nil).
					Once()
				conn.On("Close", contextMatcher).
					Return(nil).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						// accept the open request
						b, _ := msgpack.Marshal(ws.Accept{
							Version:   ws.ProtocolVersion,
							Protocols: []ws.ProtoType{ws.ProtoTypeFileTransfer},
						})
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeControl,
								MsgType:   ws.MessageTypeAccept,
								SessionID: sessionID,
							},
							Body: b,
						}
						return msgpack.Marshal(msg)
					}).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						// file info response
						body := wsft.FileInfo{
							Path: string2pointer("/absolute/path"),
							UID:  uint322pointer(0),
							GID:  uint322pointer(0),
							Mode: uint322pointer(777),
							Size: int642pointer(10),
						}
						bodyData, _ := msgpack.Marshal(body)
						return msgpack.Marshal(&ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeFileInfo,
								SessionID: sessionID,
							},
							Body: bodyData,
						})
					}).
					Once().
					On("Recv", contextMatcher).
					Return(func(ctx context.Context) ([]byte, error) {
						select {
						case <-ctx.Done():
							return nil, ctx.Err()
						case <-time.After(fileTransferTimeout * 2):
							t.Fatalf("file transfer test did not time out as expected!")
							return nil, fmt.Errorf("test failure")
						}
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							assert.Equal(t, ws.MessageTypeOpen, msg.Header.MsgType)
						}
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							assert.Equal(t, wsft.MessageTypeStat, msg.Header.MsgType)
						}
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							assert.Equal(t, wsft.MessageTypeGet, msg.Header.MsgType)
						}
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							assert.Equal(t, ws.MessageTypeClose, msg.Header.MsgType)
						}
						return nil
					}).
					Once()
			},
			AppDownloadFile: true,

			HTTPStatus: http.StatusRequestTimeout,
		},
		{
			Name:     "ko, wrong offset in chunks",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Path: "/absolute/path",

			DeviceFunc: func(t *testing.T, client *nats_mocks.Client) {
				conn := stream_mocks.NewConn(t)
				var sessionID string
				client.On("Connect", contextMatcher, mock.MatchedBy(func(srcAddr string) bool {
					var (
						tenantID string
						ok       bool
					)
					tenantID, sessionID, ok = strings.Cut(srcAddr, ":")
					return assert.Truef(t, ok, "unexpected srcAddr format: %s", srcAddr) &&
						assert.Equal(t, "000000000000000000000000", tenantID)
				}), "1234567890").
					Return(conn, nil).
					Once()
				conn.On("Close", contextMatcher).
					Return(nil).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						// accept the open request
						b, _ := msgpack.Marshal(ws.Accept{
							Version:   ws.ProtocolVersion,
							Protocols: []ws.ProtoType{ws.ProtoTypeFileTransfer},
						})
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeControl,
								MsgType:   ws.MessageTypeAccept,
								SessionID: sessionID,
							},
							Body: b,
						}
						return msgpack.Marshal(msg)
					}).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						// file info response
						body := wsft.FileInfo{
							Path: string2pointer("/absolute/path"),
							UID:  uint322pointer(0),
							GID:  uint322pointer(0),
							Mode: uint322pointer(777),
							Size: int642pointer(10),
						}
						bodyData, _ := msgpack.Marshal(body)
						return msgpack.Marshal(&ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeFileInfo,
								SessionID: sessionID,
							},
							Body: bodyData,
						})
					}).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						// first chunk
						return msgpack.Marshal(&ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeChunk,
								SessionID: sessionID,
								Properties: map[string]interface{}{
									PropertyOffset: int64(0xdeadbeef),
								},
							},
							Body: []byte("12345"),
						})
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							assert.Equal(t, ws.MessageTypeOpen, msg.Header.MsgType)
						}
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							assert.Equal(t, wsft.MessageTypeStat, msg.Header.MsgType)
						}
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							assert.Equal(t, wsft.MessageTypeGet, msg.Header.MsgType)
						}
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							assert.Equal(t, ws.MessageTypeClose, msg.Header.MsgType)
						}
						return nil
					}).
					Once()
			},
			AppDownloadFile: true,

			HTTPStatus: http.StatusInternalServerError,
		},
		{
			Name:     "error, device does not support filetransfer",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Path: "/absolute/path",

			DeviceFunc: func(t *testing.T, client *nats_mocks.Client) {
				conn := stream_mocks.NewConn(t)
				var sessionID string
				client.On("Connect", contextMatcher, mock.MatchedBy(func(srcAddr string) bool {
					var (
						tenantID string
						ok       bool
					)
					tenantID, sessionID, ok = strings.Cut(srcAddr, ":")
					return assert.Truef(t, ok, "unexpected srcAddr format: %s", srcAddr) &&
						assert.Equal(t, "000000000000000000000000", tenantID)
				}), "1234567890").
					Return(conn, nil).
					Once()
				conn.On("Close", contextMatcher).
					Return(nil).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeShell,
								MsgType:   shell.MessageTypeStopShell,
								SessionID: sessionID,
								Properties: map[string]interface{}{
									"status": shell.ErrorMessage,
								},
							},
							Body: []byte("mender-connect v1.0 simulation"),
						}
						return msgpack.Marshal(msg)
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							assert.Equal(t, ws.MessageTypeOpen, msg.Header.MsgType)
						}
						return nil
					}).
					Once()
			},
			AppDownloadFile: true,

			HTTPStatus: http.StatusBadGateway,
		},
		{
			Name:     "ko, failed to submit audit log",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Path: "/absolute/path",

			AppDownloadFile:    true,
			AppDownloadFileErr: errors.New("generic error"),
			DeviceFunc: func(t *testing.T, client *nats_mocks.Client) {
				conn := stream_mocks.NewConn(t)
				client.On("Connect", contextMatcher, mock.MatchedBy(func(srcAddr string) bool {
					var (
						tenantID string
						ok       bool
					)
					tenantID, _, ok = strings.Cut(srcAddr, ":")
					return assert.Truef(t, ok, "unexpected srcAddr format: %s", srcAddr) &&
						assert.Equal(t, "000000000000000000000000", tenantID)
				}), "1234567890").
					Return(conn, nil).
					Once()
				conn.On("Close", contextMatcher).
					Return(nil).
					Once()
			},

			HTTPStatus: http.StatusInternalServerError,
		},
		{
			Name:     "ko, bad request, relative path",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Path: "relative/path",

			HTTPStatus: http.StatusBadRequest,
		},
		{
			Name:     "ko, malformed request",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Path: "",

			HTTPStatus: http.StatusBadRequest,
		},
		{
			Name:     "ko, missing request body",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},

			HTTPStatus: http.StatusBadRequest,
		},
		{
			Name:     "ko, missing auth",
			DeviceID: "1234567890",

			HTTPStatus: http.StatusUnauthorized,
		},
		{
			Name: "ko, wrong auth",
			Identity: &identity.Identity{
				Subject:  "00000000-0000-0000-0000-000000000000",
				Tenant:   "000000000000000000000000",
				IsDevice: true,
			},

			DeviceID: "1234567890",

			HTTPStatus: http.StatusUnauthorized,
		},
		{
			Name:     "ko, not connected",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Path: "/absolute/path",

			DeviceFunc: func(t *testing.T, client *nats_mocks.Client) {
				client.On("Connect", contextMatcher, mock.MatchedBy(func(srcAddr string) bool {
					var (
						tenantID string
						ok       bool
					)
					tenantID, _, ok = strings.Cut(srcAddr, ":")
					return assert.Truef(t, ok, "unexpected srcAddr format: %s", srcAddr) &&
						assert.Equal(t, "000000000000000000000000", tenantID)
				}), "1234567890").
					Return(nil, stream.ErrConnectionRefused).
					Once()
			},

			HTTPStatus: http.StatusConflict,
		},
		{
			Name:     "ko, other error",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},

			HTTPStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			app := app_mocks.NewApp(t)

			if tc.AppDownloadFile {
				app.On("DownloadFile",
					mock.MatchedBy(func(_ context.Context) bool {
						return true
					}),
					tc.Identity.Subject,
					tc.DeviceID,
					mock.AnythingOfType("string"),
				).Return(tc.AppDownloadFileErr)
			}

			natsClient := nats_mocks.NewClient(t)
			if tc.DeviceFunc != nil {
				tc.DeviceFunc(t, natsClient)
			}

			router, _ := NewRouter(app, natsClient, nil)
			s := httptest.NewServer(router)
			defer s.Close()

			path := url.QueryEscape(tc.Path)
			url := strings.Replace(APIURLManagementDeviceDownload, ":deviceId", tc.DeviceID, 1)
			req, err := http.NewRequest(http.MethodGet, "http://localhost"+url+"?path="+path, nil)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			if tc.Identity != nil {
				jwt := GenerateJWT(*tc.Identity)
				req.Header.Set(headerAuthorization, "Bearer "+jwt)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tc.HTTPStatus, w.Code, w.Body.Bytes())
			if tc.HTTPStatus == http.StatusOK {
				assert.Equal(t, tc.HTTPBody, w.Body.Bytes())
			}
		})
	}
}

func TestManagementUploadFile(t *testing.T) {
	originalFileTransferTimeout := fileTransferTimeout
	originalAckSlidingWindowRecv := ackSlidingWindowRecv
	t.Cleanup(func() {
		fileTransferTimeout = originalFileTransferTimeout
		ackSlidingWindowRecv = originalAckSlidingWindowRecv
	})

	fileTransferTimeout = 2 * time.Second
	ackSlidingWindowRecv = 0

	testCases := []struct {
		Name     string
		DeviceID string
		Body     map[string][]string
		File     []byte
		Identity *identity.Identity

		DeviceFunc       func(*testing.T, *nats_mocks.Client)
		AppUploadFile    bool
		AppUploadFileErr error

		HTTPStatus int
	}{
		{
			Name:     "ok, upload",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Body: map[string][]string{
				fieldUploadPath: {"/absolute/path"},
				fieldUploadUID:  {"0"},
				fieldUploadGID:  {"0"},
				fieldUploadMode: {"0644"},
			},
			File: []byte("1234567890"),

			DeviceFunc: func(t *testing.T, client *nats_mocks.Client) {
				ackOffset := make(chan int64, 1)
				conn := stream_mocks.NewConn(t)
				var sessionID string
				client.On("Connect", contextMatcher, mock.MatchedBy(func(srcAddr string) bool {
					var (
						tenantID string
						ok       bool
					)
					tenantID, sessionID, ok = strings.Cut(srcAddr, ":")
					return assert.Truef(t, ok, "unexpected srcAddr format: %s", srcAddr) &&
						assert.Equal(t, "000000000000000000000000", tenantID)
				}), "1234567890").
					Return(conn, nil).
					Once()

				conn.On("Close", contextMatcher).
					Return(nil).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						// accept the open request
						b, _ := msgpack.Marshal(ws.Accept{
							Version:   ws.ProtocolVersion,
							Protocols: []ws.ProtoType{ws.ProtoTypeFileTransfer},
						})
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeControl,
								MsgType:   ws.MessageTypeAccept,
								SessionID: sessionID,
							},
							Body: b,
						}
						return msgpack.Marshal(msg)
					}).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						offset := <-ackOffset
						return msgpack.Marshal(&ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeACK,
								SessionID: sessionID,
								Properties: map[string]any{
									PropertyOffset: offset,
								},
							},
						})
					}).
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							assert.Equal(t, ws.MessageTypeOpen, msg.Header.MsgType)
						}
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							assert.Equal(t, wsft.MessageTypePut, msg.Header.MsgType)
						}
						ackOffset <- 0
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							assert.Equal(t, wsft.MessageTypeChunk, msg.Header.MsgType)
							ackOffset <- 10
						}
						return nil
					}).
					Times(2).
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							assert.Equal(t, ws.MessageTypeClose, msg.Header.MsgType)
						}
						return nil
					}).
					Once()
			},
			AppUploadFile: true,

			HTTPStatus: http.StatusCreated,
		},
		{
			Name:     "ko, missing ack from device after chunk",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Body: map[string][]string{
				fieldUploadPath: {"/absolute/path"},
				fieldUploadUID:  {"0"},
				fieldUploadGID:  {"0"},
				fieldUploadMode: {"0644"},
			},
			File: []byte("1234567890"),

			DeviceFunc: func(t *testing.T, client *nats_mocks.Client) {
				ackOffset := make(chan int64, 1)
				conn := stream_mocks.NewConn(t)
				var sessionID string
				client.On("Connect", contextMatcher, mock.MatchedBy(func(srcAddr string) bool {
					var (
						tenantID string
						ok       bool
					)
					tenantID, sessionID, ok = strings.Cut(srcAddr, ":")
					return assert.Truef(t, ok, "unexpected srcAddr format: %s", srcAddr) &&
						assert.Equal(t, "000000000000000000000000", tenantID)
				}), "1234567890").
					Return(conn, nil).
					Once()

				conn.On("Close", contextMatcher).
					Return(nil).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						// accept the open request
						b, _ := msgpack.Marshal(ws.Accept{
							Version:   ws.ProtocolVersion,
							Protocols: []ws.ProtoType{ws.ProtoTypeFileTransfer},
						})
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeControl,
								MsgType:   ws.MessageTypeAccept,
								SessionID: sessionID,
							},
							Body: b,
						}
						return msgpack.Marshal(msg)
					}).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						offset := <-ackOffset
						return msgpack.Marshal(&ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeACK,
								SessionID: sessionID,
								Properties: map[string]any{
									PropertyOffset: offset,
								},
							},
						})
					}).
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							assert.Equal(t, ws.MessageTypeOpen, msg.Header.MsgType)
						}
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							assert.Equal(t, wsft.MessageTypePut, msg.Header.MsgType)
						}
						ackOffset <- 0
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							assert.Equal(t, wsft.MessageTypeChunk, msg.Header.MsgType)
						}
						return nil
					}).
					Times(2).
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							assert.Equal(t, ws.MessageTypeClose, msg.Header.MsgType)
						}
						return nil
					}).
					Once()
			},
			AppUploadFile: true,

			HTTPStatus: http.StatusRequestTimeout,
		},
		{
			Name:     "error, filetransfer disabled",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Body: map[string][]string{
				fieldUploadPath: {"/absolute/path"},
				fieldUploadUID:  {"0"},
				fieldUploadGID:  {"0"},
				fieldUploadMode: {"0644"},
			},
			File: []byte("1234567890"),

			DeviceFunc: func(t *testing.T, client *nats_mocks.Client) {
				conn := stream_mocks.NewConn(t)
				var sessionID string
				client.On("Connect", contextMatcher, mock.MatchedBy(func(srcAddr string) bool {
					var (
						tenantID string
						ok       bool
					)
					tenantID, sessionID, ok = strings.Cut(srcAddr, ":")
					return assert.Truef(t, ok, "unexpected srcAddr format: %s", srcAddr) &&
						assert.Equal(t, "000000000000000000000000", tenantID)
				}), "1234567890").
					Return(conn, nil).
					Once()

				conn.On("Close", contextMatcher).
					Return(nil).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						b, _ := msgpack.Marshal(ws.Accept{
							Version: ws.ProtocolVersion,
							Protocols: []ws.ProtoType{
								ws.ProtoTypeShell,
								ws.ProtoTypeMenderClient,
							},
						})
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeControl,
								MsgType:   ws.MessageTypeAccept,
								SessionID: sessionID,
							},
							Body: b,
						}
						return msgpack.Marshal(msg)
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							assert.Equal(t, ws.MessageTypeOpen, msg.Header.MsgType)
						}
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							assert.Equal(t, ws.MessageTypeClose, msg.Header.MsgType)
						}
						return nil
					}).
					Once()
			},
			AppUploadFile: true,

			HTTPStatus: http.StatusBadGateway,
		},
		{
			Name:     "ko, error from device",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Body: map[string][]string{
				fieldUploadPath: {"/absolute/path"},
				fieldUploadUID:  {"0"},
				fieldUploadGID:  {"0"},
				fieldUploadMode: {"0644"},
			},
			File: []byte("1234567890"),

			DeviceFunc: func(t *testing.T, client *nats_mocks.Client) {
				conn := stream_mocks.NewConn(t)
				var sessionID string
				client.On("Connect", contextMatcher, mock.MatchedBy(func(srcAddr string) bool {
					var (
						tenantID string
						ok       bool
					)
					tenantID, sessionID, ok = strings.Cut(srcAddr, ":")
					return assert.Truef(t, ok, "unexpected srcAddr format: %s", srcAddr) &&
						assert.Equal(t, "000000000000000000000000", tenantID)
				}), "1234567890").
					Return(conn, nil).
					Once()

				conn.On("Close", contextMatcher).
					Return(nil).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						// accept the open request
						b, _ := msgpack.Marshal(ws.Accept{
							Version:   ws.ProtocolVersion,
							Protocols: []ws.ProtoType{ws.ProtoTypeFileTransfer},
						})
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeControl,
								MsgType:   ws.MessageTypeAccept,
								SessionID: sessionID,
							},
							Body: b,
						}
						return msgpack.Marshal(msg)
					}).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						body := wsft.Error{
							Error:       string2pointer("file not writeable"),
							MessageType: string2pointer(wsft.MessageTypePut),
						}
						bodyData, _ := msgpack.Marshal(body)
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								MsgType:   wsft.MessageTypeError,
								SessionID: sessionID,
							},
							Body: bodyData,
						}
						return msgpack.Marshal(msg)
					}).
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							assert.Equal(t, ws.MessageTypeOpen, msg.Header.MsgType)
						}
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							assert.Equal(t, wsft.MessageTypePut, msg.Header.MsgType)
						}
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							assert.Equal(t, ws.MessageTypeClose, msg.Header.MsgType)
						}
						return nil
					}).
					Once()
			},
			AppUploadFile: true,

			HTTPStatus: http.StatusBadRequest,
		},
		{
			Name:     "ko, error from device after the first chunk",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Body: map[string][]string{
				fieldUploadPath: {"/absolute/path"},
				fieldUploadUID:  {"0"},
				fieldUploadGID:  {"0"},
				fieldUploadMode: {"0644"},
			},
			File: []byte("1234567890"),

			DeviceFunc: func(t *testing.T, client *nats_mocks.Client) {
				conn := stream_mocks.NewConn(t)
				var sessionID string
				client.On("Connect", contextMatcher, mock.MatchedBy(func(srcAddr string) bool {
					var (
						tenantID string
						ok       bool
					)
					tenantID, sessionID, ok = strings.Cut(srcAddr, ":")
					return assert.Truef(t, ok, "unexpected srcAddr format: %s", srcAddr) &&
						assert.Equal(t, "000000000000000000000000", tenantID)
				}), "1234567890").
					Return(conn, nil).
					Once()

				conn.On("Close", contextMatcher).
					Return(nil).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						// accept the open request
						b, _ := msgpack.Marshal(ws.Accept{
							Version:   ws.ProtocolVersion,
							Protocols: []ws.ProtoType{ws.ProtoTypeFileTransfer},
						})
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeControl,
								MsgType:   ws.MessageTypeAccept,
								SessionID: sessionID,
							},
							Body: b,
						}
						return msgpack.Marshal(msg)
					}).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						return msgpack.Marshal(&ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeACK,
								SessionID: sessionID,
								Properties: map[string]any{
									PropertyOffset: int64(0),
								},
							},
						})
					}).
					Once().
					On("Recv", contextMatcher).
					Return(func(context.Context) ([]byte, error) {
						body := ws.Error{
							Error:       "failed to Write",
							MessageType: wsft.MessageTypePut,
							Code:        http.StatusBadRequest,
						}
						bodyData, _ := msgpack.Marshal(body)
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								MsgType:   wsft.MessageTypeError,
								SessionID: sessionID,
							},
							Body: bodyData,
						}
						return msgpack.Marshal(msg)
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							assert.Equal(t, ws.MessageTypeOpen, msg.Header.MsgType)
						}
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							assert.Equal(t, wsft.MessageTypePut, msg.Header.MsgType)
						}
						return nil
					}).
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							assert.Equal(t, wsft.MessageTypeChunk, msg.Header.MsgType)
						}
						return nil
					}).
					Times(2).
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							assert.Equal(t, ws.MessageTypeClose, msg.Header.MsgType)
						}
						return nil
					}).
					Once()
			},
			AppUploadFile: true,

			HTTPStatus: http.StatusBadRequest,
		},
		{
			Name:     "ko, timeout",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Body: map[string][]string{
				fieldUploadPath: {"/absolute/path"},
				fieldUploadUID:  {"0"},
				fieldUploadGID:  {"0"},
				fieldUploadMode: {"0644"},
			},
			File: []byte("1234567890"),

			DeviceFunc: func(t *testing.T, client *nats_mocks.Client) {
				client.On("Connect", contextMatcher, mock.MatchedBy(func(srcAddr string) bool {
					var (
						tenantID string
						ok       bool
					)
					tenantID, _, ok = strings.Cut(srcAddr, ":")
					return assert.Truef(t, ok, "unexpected srcAddr format: %s", srcAddr) &&
						assert.Equal(t, "000000000000000000000000", tenantID)
				}), "1234567890").
					Return(func(ctx context.Context, _, _ string) (stream.Conn, error) {
						select {
						case <-ctx.Done():
							return nil, ctx.Err()
						case <-time.After(2 * fileTransferTimeout):
							t.Fatalf("timed out waiting for timeout")
							return nil, fmt.Errorf("test failed!")
						}
					}).
					Once()
			},
			AppUploadFile: false,

			HTTPStatus: http.StatusRequestTimeout,
		},
		{
			Name:     "error, timeout in handshake",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Body: map[string][]string{
				fieldUploadPath: {"/absolute/path"},
				fieldUploadUID:  {"0"},
				fieldUploadGID:  {"0"},
				fieldUploadMode: {"0644"},
			},
			File: []byte("1234567890"),

			DeviceFunc: func(t *testing.T, client *nats_mocks.Client) {
				conn := stream_mocks.NewConn(t)
				var sessionID string
				client.On("Connect", contextMatcher, mock.MatchedBy(func(srcAddr string) bool {
					var (
						tenantID string
						ok       bool
					)
					tenantID, sessionID, ok = strings.Cut(srcAddr, ":")
					return assert.Truef(t, ok, "unexpected srcAddr format: %s", srcAddr) &&
						assert.Equal(t, "000000000000000000000000", tenantID)
				}), "1234567890").
					Return(conn, nil).
					Once()

				conn.On("Close", contextMatcher).
					Return(nil).
					Once().
					On("Recv", contextMatcher).
					Return(func(ctx context.Context) ([]byte, error) {
						select {
						case <-ctx.Done():
							return nil, ctx.Err()
						case <-time.After(2 * fileTransferTimeout):
							t.Fatal("test did not timeout as expected")
						}
						return nil, fmt.Errorf("test failed!")
					}).
					Once().
					Once().
					On("Send", contextMatcher, mock.Anything).
					Return(func(ctx context.Context, data []byte) error {
						msg := &ws.ProtoMsg{}
						if assert.NoError(t, msgpack.Unmarshal(data, msg)) {
							assert.Equal(t, sessionID, msg.Header.SessionID)
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							assert.Equal(t, ws.MessageTypeOpen, msg.Header.MsgType)
						}
						return nil
					}).
					Once()
			},
			AppUploadFile: true,

			HTTPStatus: http.StatusRequestTimeout,
		},
		{
			Name:     "ko, failed to submit audit log",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Body: map[string][]string{
				fieldUploadPath: {"/absolute/path"},
				fieldUploadUID:  {"0"},
				fieldUploadGID:  {"0"},
				fieldUploadMode: {"0644"},
			},
			File: []byte("1234567890"),

			AppUploadFile:    true,
			AppUploadFileErr: errors.New("generic error"),
			DeviceFunc: func(t *testing.T, client *nats_mocks.Client) {
				conn := stream_mocks.NewConn(t)
				client.On("Connect", contextMatcher, mock.MatchedBy(func(srcAddr string) bool {
					var (
						tenantID string
						ok       bool
					)
					tenantID, _, ok = strings.Cut(srcAddr, ":")
					return assert.Truef(t, ok, "unexpected srcAddr format: %s", srcAddr) &&
						assert.Equal(t, "000000000000000000000000", tenantID)
				}), "1234567890").
					Return(conn, nil).
					Once()

				conn.On("Close", contextMatcher).
					Return(nil).
					Once()
			},

			HTTPStatus: http.StatusInternalServerError,
		},
		{
			Name:     "ko, bad request, missing file",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Body: map[string][]string{
				fieldUploadPath: {"/absolute/path"},
				fieldUploadUID:  {"0"},
				fieldUploadGID:  {"0"},
				fieldUploadMode: {"0644"},
			},

			HTTPStatus: http.StatusBadRequest,
		},
		{
			Name:     "ko, bad request, relative path",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Body: map[string][]string{
				fieldUploadPath: {"relative/path"},
				fieldUploadUID:  {"0"},
				fieldUploadGID:  {"0"},
				fieldUploadMode: {"0644"},
			},
			File: []byte("1234567890"),

			HTTPStatus: http.StatusBadRequest,
		},
		{
			Name:     "ko, bad request, relative path",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Body: map[string][]string{
				fieldUploadPath: {"relative/path"},
				fieldUploadUID:  {"0"},
				fieldUploadGID:  {"0"},
				fieldUploadMode: {"0644"},
			},
			File: []byte("1234567890"),

			HTTPStatus: http.StatusBadRequest,
		},
		{
			Name:     "ko, malformed request",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},

			HTTPStatus: http.StatusBadRequest,
		},
		{
			Name:     "ko, missing request body",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},

			HTTPStatus: http.StatusBadRequest,
		},
		{
			Name:     "ko, missing auth",
			DeviceID: "1234567890",

			HTTPStatus: http.StatusUnauthorized,
		},
		{
			Name: "ko, wrong auth",
			Identity: &identity.Identity{
				Subject:  "00000000-0000-0000-0000-000000000000",
				Tenant:   "000000000000000000000000",
				IsDevice: true,
			},

			DeviceID: "1234567890",

			HTTPStatus: http.StatusUnauthorized,
		},
		{
			Name:     "ko, not connected",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Body: map[string][]string{
				fieldUploadPath: {"/absolute/path"},
				fieldUploadUID:  {"0"},
				fieldUploadGID:  {"0"},
				fieldUploadMode: {"0644"},
			},
			File: []byte("1234567890"),

			DeviceFunc: func(t *testing.T, client *nats_mocks.Client) {
				client.On("Connect", contextMatcher, mock.MatchedBy(func(srcAddr string) bool {
					var (
						tenantID string
						ok       bool
					)
					tenantID, _, ok = strings.Cut(srcAddr, ":")
					return assert.Truef(t, ok, "unexpected srcAddr format: %s", srcAddr) &&
						assert.Equal(t, "000000000000000000000000", tenantID)
				}), "1234567890").
					Return(nil, stream.ErrConnectionRefused).
					Once()
			},
			HTTPStatus: http.StatusConflict,
		},
		{
			Name:     "ko, other error",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Body: map[string][]string{
				fieldUploadPath: {"/absolute/path"},
				fieldUploadUID:  {"0"},
				fieldUploadGID:  {"0"},
				fieldUploadMode: {"0644"},
			},
			File: []byte("1234567890"),

			DeviceFunc: func(t *testing.T, client *nats_mocks.Client) {
				client.On("Connect", contextMatcher, mock.MatchedBy(func(srcAddr string) bool {
					var (
						tenantID string
						ok       bool
					)
					tenantID, _, ok = strings.Cut(srcAddr, ":")
					return assert.Truef(t, ok, "unexpected srcAddr format: %s", srcAddr) &&
						assert.Equal(t, "000000000000000000000000", tenantID)
				}), "1234567890").
					Return(nil, fmt.Errorf("generic error")).
					Once()
			},

			HTTPStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			app := &app_mocks.App{}
			defer app.AssertExpectations(t)

			if tc.AppUploadFile {
				app.On("UploadFile",
					mock.MatchedBy(func(_ context.Context) bool {
						return true
					}),
					tc.Identity.Subject,
					tc.DeviceID,
					mock.AnythingOfType("string"),
				).Return(tc.AppUploadFileErr)
			}

			natsClient := &nats_mocks.Client{}
			defer natsClient.AssertExpectations(t)

			if tc.DeviceFunc != nil {
				tc.DeviceFunc(t, natsClient)
			}

			router, _ := NewRouter(app, natsClient, nil)
			s := httptest.NewServer(router)
			defer s.Close()

			var body io.Reader
			if tc.Body != nil {
				var b bytes.Buffer
				w := multipart.NewWriter(&b)
				w.SetBoundary("boundary")
				for key, value := range tc.Body {
					for _, v := range value {
						w.WriteField(key, v)
					}
				}
				if tc.File != nil {
					fileWriter, _ := w.CreateFormFile(fieldUploadFile, "dummy.txt")
					fileWriter.Write(tc.File)
				}
				w.Close()
				data := make([]byte, 10240)
				n, _ := b.Read(data)
				body = bytes.NewReader(data[:n])
			}

			url := strings.Replace(APIURLManagementDeviceUpload, ":deviceId", tc.DeviceID, 1)
			req, err := http.NewRequest(http.MethodPut, "http://localhost"+url, body)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			if body != nil {
				req.Header.Add("Content-Type", "multipart/form-data; boundary=\"boundary\"")
			}

			if tc.Identity != nil {
				jwt := GenerateJWT(*tc.Identity)
				req.Header.Set(headerAuthorization, "Bearer "+jwt)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tc.HTTPStatus, w.Code, w.Body.String())
		})
	}
}
