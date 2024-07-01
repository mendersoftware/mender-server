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
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/ws"
	natsio "github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/mendersoftware/mender-server/services/deviceconnect/app"
	app_mocks "github.com/mendersoftware/mender-server/services/deviceconnect/app/mocks"
	nats_mocks "github.com/mendersoftware/mender-server/services/deviceconnect/client/nats/mocks"
	"github.com/mendersoftware/mender-server/services/deviceconnect/model"
	wsft "github.com/mendersoftware/mender-server/pkg/ws/filetransfer"
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
	originalNewFileTransferSessionID := newFileTransferSessionID
	originalFileTransferTimeout := fileTransferTimeout
	originalAckSlidingWindowSend := ackSlidingWindowSend
	defer func() {
		newFileTransferSessionID = originalNewFileTransferSessionID
		fileTransferTimeout = originalFileTransferTimeout
		ackSlidingWindowSend = originalAckSlidingWindowSend
	}()

	fileTransferTimeout = 2 * time.Second
	ackSlidingWindowSend = 1

	sessionID, _ := uuid.NewRandom()
	newFileTransferSessionID = func() (uuid.UUID, error) {
		return sessionID, nil
	}

	testCases := []struct {
		Name     string
		DeviceID string
		Path     string
		Identity *identity.Identity

		GetDevice          *model.Device
		GetDeviceError     error
		DeviceFunc         func(*nats_mocks.Client)
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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
			},
			DeviceFunc: func(client *nats_mocks.Client) {
				client.On("ChanSubscribe",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(chanMsg chan *natsio.Msg) bool {
						// accept the open request
						b, _ := msgpack.Marshal(ws.Accept{
							Version:   ws.ProtocolVersion,
							Protocols: []ws.ProtoType{ws.ProtoTypeFileTransfer},
						})
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeControl,
								MsgType:   ws.MessageTypeAccept,
								SessionID: sessionID.String(),
							},
							Body: b,
						}
						b, _ = msgpack.Marshal(msg)
						chanMsg <- &natsio.Msg{Data: b}

						// file info response
						body := wsft.FileInfo{
							Path: string2pointer("/absolute/path"),
							UID:  uint322pointer(0),
							GID:  uint322pointer(0),
							Mode: uint322pointer(777),
							Size: int642pointer(10),
						}
						bodyData, err := msgpack.Marshal(body)
						msg = &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeFileInfo,
								SessionID: sessionID.String(),
							},
							Body: bodyData,
						}

						data, err := msgpack.Marshal(msg)
						assert.NoError(t, err)
						chanMsg <- &natsio.Msg{Data: data}

						// first chunk
						msg = &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeChunk,
								SessionID: sessionID.String(),
								Properties: map[string]interface{}{
									PropertyOffset: int64(0),
								},
							},
							Body: []byte("12345"),
						}

						data, err = msgpack.Marshal(msg)
						assert.NoError(t, err)
						chanMsg <- &natsio.Msg{Data: data}

						// final chunk
						msg = &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeChunk,
								SessionID: sessionID.String(),
								Properties: map[string]interface{}{
									PropertyOffset: int64(5),
								},
							},
							Body: nil,
						}

						data, err = msgpack.Marshal(msg)
						assert.NoError(t, err)
						chanMsg <- &natsio.Msg{Data: data}

						return true
					}),
				).Return(&natsio.Subscription{}, nil)

				client.On("Publish",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(data []byte) bool {
						msg := &ws.ProtoMsg{}
						err := msgpack.Unmarshal(data, msg)
						assert.NoError(t, err)

						if msg.Header.MsgType == wsft.MessageTypeStat {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == wsft.MessageTypeGet {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == wsft.MessageTypeACK {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypePing {
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypeOpen {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						} else if msg.Header.MsgType == ws.MessageTypeClose {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						}

						return false
					}),
				).Return(nil)
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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
			},
			DeviceFunc: func(client *nats_mocks.Client) {
				client.On("ChanSubscribe",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(chanMsg chan *natsio.Msg) bool {
						// accept the open request
						b, _ := msgpack.Marshal(ws.Accept{
							Version:   ws.ProtocolVersion,
							Protocols: []ws.ProtoType{ws.ProtoTypeFileTransfer},
						})
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeControl,
								MsgType:   ws.MessageTypeAccept,
								SessionID: sessionID.String(),
							},
							Body: b,
						}
						b, _ = msgpack.Marshal(msg)
						chanMsg <- &natsio.Msg{Data: b}

						// file info response
						body := wsft.FileInfo{
							Path: string2pointer("/absolute/path"),
							Mode: uint322pointer(777),
							Size: int642pointer(10),
						}
						bodyData, err := msgpack.Marshal(body)
						msg = &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeFileInfo,
								SessionID: sessionID.String(),
							},
							Body: bodyData,
						}

						data, err := msgpack.Marshal(msg)
						assert.NoError(t, err)
						chanMsg <- &natsio.Msg{Data: data}

						// first chunk
						msg = &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeChunk,
								SessionID: sessionID.String(),
								Properties: map[string]interface{}{
									PropertyOffset: int64(0),
								},
							},
							Body: []byte("12345"),
						}

						data, err = msgpack.Marshal(msg)
						assert.NoError(t, err)
						chanMsg <- &natsio.Msg{Data: data}

						// second chunk
						msg = &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeChunk,
								SessionID: sessionID.String(),
								Properties: map[string]interface{}{
									PropertyOffset: int64(5),
								},
							},
							Body: []byte("67890"),
						}

						data, err = msgpack.Marshal(msg)
						assert.NoError(t, err)
						chanMsg <- &natsio.Msg{Data: data}

						// final chunk
						msg = &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeChunk,
								SessionID: sessionID.String(),
							},
							Body: nil,
						}

						data, err = msgpack.Marshal(msg)
						assert.NoError(t, err)
						chanMsg <- &natsio.Msg{Data: data}

						return true
					}),
				).Return(&natsio.Subscription{}, nil)

				client.On("Publish",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(data []byte) bool {
						msg := &ws.ProtoMsg{}
						err := msgpack.Unmarshal(data, msg)
						assert.NoError(t, err)

						if msg.Header.MsgType == wsft.MessageTypeStat {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == wsft.MessageTypeGet {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == wsft.MessageTypeACK {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypePing {
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypeOpen {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						} else if msg.Header.MsgType == ws.MessageTypeClose {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						}

						return false
					}),
				).Return(nil)
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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
			},
			DeviceFunc: func(client *nats_mocks.Client) {
				client.On("ChanSubscribe",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(chanMsg chan *natsio.Msg) bool {
						// accept the open request
						b, _ := msgpack.Marshal(ws.Accept{
							Version:   ws.ProtocolVersion,
							Protocols: []ws.ProtoType{ws.ProtoTypeFileTransfer},
						})
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeControl,
								MsgType:   ws.MessageTypeAccept,
								SessionID: sessionID.String(),
							},
							Body: b,
						}
						b, _ = msgpack.Marshal(msg)
						chanMsg <- &natsio.Msg{Data: b}

						// file info response
						body := wsft.Error{
							Error:       string2pointer("file not found"),
							MessageType: string2pointer(wsft.MessageTypeStat),
						}
						bodyData, err := msgpack.Marshal(body)
						msg = &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeError,
								SessionID: sessionID.String(),
							},
							Body: bodyData,
						}

						data, err := msgpack.Marshal(msg)
						assert.NoError(t, err)
						chanMsg <- &natsio.Msg{Data: data}

						return true
					}),
				).Return(&natsio.Subscription{}, nil)

				client.On("Publish",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(data []byte) bool {
						msg := &ws.ProtoMsg{}
						err := msgpack.Unmarshal(data, msg)
						assert.NoError(t, err)

						if msg.Header.MsgType == wsft.MessageTypeStat {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == wsft.MessageTypeGet {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == wsft.MessageTypeACK {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypePing {
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypeOpen {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						} else if msg.Header.MsgType == ws.MessageTypeClose {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						}

						return false
					}),
				).Return(nil)
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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
			},
			DeviceFunc: func(client *nats_mocks.Client) {
				client.On("ChanSubscribe",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(chanMsg chan *natsio.Msg) bool {
						// accept the open request
						b, _ := msgpack.Marshal(ws.Accept{
							Version:   ws.ProtocolVersion,
							Protocols: []ws.ProtoType{ws.ProtoTypeFileTransfer},
						})
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeControl,
								MsgType:   ws.MessageTypeAccept,
								SessionID: sessionID.String(),
							},
							Body: b,
						}
						b, _ = msgpack.Marshal(msg)
						chanMsg <- &natsio.Msg{Data: b}

						// file info response
						body := wsft.FileInfo{
							Path: string2pointer("/absolute/path"),
							UID:  uint322pointer(0),
							GID:  uint322pointer(0),
							Mode: uint322pointer(777 | uint32(os.ModeDir)),
							Size: int642pointer(10),
						}
						bodyData, err := msgpack.Marshal(body)
						msg = &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeFileInfo,
								SessionID: sessionID.String(),
							},
							Body: bodyData,
						}

						data, err := msgpack.Marshal(msg)
						assert.NoError(t, err)
						chanMsg <- &natsio.Msg{Data: data}

						return true
					}),
				).Return(&natsio.Subscription{}, nil)

				client.On("Publish",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(data []byte) bool {
						msg := &ws.ProtoMsg{}
						err := msgpack.Unmarshal(data, msg)
						assert.NoError(t, err)

						if msg.Header.MsgType == wsft.MessageTypeStat {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == wsft.MessageTypeGet {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == wsft.MessageTypeACK {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypePing {
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypeOpen {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						} else if msg.Header.MsgType == ws.MessageTypeClose {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						}

						return false
					}),
				).Return(nil)
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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
			},
			DeviceFunc: func(client *nats_mocks.Client) {
				client.On("ChanSubscribe",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(chanMsg chan *natsio.Msg) bool {
						// accept the open request
						b, _ := msgpack.Marshal(ws.Accept{
							Version:   ws.ProtocolVersion,
							Protocols: []ws.ProtoType{ws.ProtoTypeFileTransfer},
						})
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeControl,
								MsgType:   ws.MessageTypeAccept,
								SessionID: sessionID.String(),
							},
							Body: b,
						}
						b, _ = msgpack.Marshal(msg)
						chanMsg <- &natsio.Msg{Data: b}
						// file info response
						body := wsft.FileInfo{
							Path: string2pointer("/absolute/path"),
							Mode: uint322pointer(777),
							Size: int642pointer(10),
						}
						bodyData, err := msgpack.Marshal(body)
						msg = &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeFileInfo,
								SessionID: sessionID.String(),
							},
							Body: bodyData,
						}

						data, err := msgpack.Marshal(msg)
						assert.NoError(t, err)
						chanMsg <- &natsio.Msg{Data: data}

						// first chunk
						msg = &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeChunk,
								SessionID: sessionID.String(),
								Properties: map[string]interface{}{
									PropertyOffset: int64(0),
								},
							},
							Body: []byte("12345"),
						}

						data, err = msgpack.Marshal(msg)
						assert.NoError(t, err)
						chanMsg <- &natsio.Msg{Data: data}

						// error
						errBody := wsft.Error{
							Error:       string2pointer("file not found"),
							MessageType: string2pointer(wsft.MessageTypeStat),
						}
						bodyData, err = msgpack.Marshal(errBody)
						msg = &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeError,
								SessionID: sessionID.String(),
							},
							Body: bodyData,
						}

						data, err = msgpack.Marshal(msg)
						assert.NoError(t, err)
						chanMsg <- &natsio.Msg{Data: data}

						return true
					}),
				).Return(&natsio.Subscription{}, nil)

				client.On("Publish",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(data []byte) bool {
						msg := &ws.ProtoMsg{}
						err := msgpack.Unmarshal(data, msg)
						assert.NoError(t, err)

						if msg.Header.MsgType == wsft.MessageTypeStat {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == wsft.MessageTypeGet {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == wsft.MessageTypeACK {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypePing {
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypeOpen {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						} else if msg.Header.MsgType == ws.MessageTypeClose {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						}

						return false
					}),
				).Return(nil)
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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
			},
			DeviceFunc: func(client *nats_mocks.Client) {
				client.On("ChanSubscribe",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(chanMsg chan *natsio.Msg) bool {
						// accept the open request
						b, _ := msgpack.Marshal(ws.Accept{
							Version:   ws.ProtocolVersion,
							Protocols: []ws.ProtoType{ws.ProtoTypeFileTransfer},
						})
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeControl,
								MsgType:   ws.MessageTypeAccept,
								SessionID: sessionID.String(),
							},
							Body: b,
						}
						b, _ = msgpack.Marshal(msg)
						chanMsg <- &natsio.Msg{Data: b}
						// then timeout...
						return true
					}),
				).Return(&natsio.Subscription{}, nil)

				client.On("Publish",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(data []byte) bool {
						msg := &ws.ProtoMsg{}
						err := msgpack.Unmarshal(data, msg)
						assert.NoError(t, err)

						if msg.Header.MsgType == wsft.MessageTypeStat {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == wsft.MessageTypeGet {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == wsft.MessageTypeACK {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypePing {
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypeOpen {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						} else if msg.Header.MsgType == ws.MessageTypeClose {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						}

						return false
					}),
				).Return(nil)
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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
			},
			DeviceFunc: func(client *nats_mocks.Client) {
				client.On("ChanSubscribe",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(chanMsg chan *natsio.Msg) bool {
						// accept the open request
						b, _ := msgpack.Marshal(ws.Accept{
							Version:   ws.ProtocolVersion,
							Protocols: []ws.ProtoType{ws.ProtoTypeFileTransfer},
						})
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeControl,
								MsgType:   ws.MessageTypeAccept,
								SessionID: sessionID.String(),
							},
							Body: b,
						}
						b, _ = msgpack.Marshal(msg)
						chanMsg <- &natsio.Msg{Data: b}
						// file info response
						body := wsft.FileInfo{
							Path: string2pointer("/absolute/path"),
							Mode: uint322pointer(777),
							Size: int642pointer(10),
						}
						bodyData, err := msgpack.Marshal(body)
						msg = &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeFileInfo,
								SessionID: sessionID.String(),
							},
							Body: bodyData,
						}

						data, err := msgpack.Marshal(msg)
						assert.NoError(t, err)
						chanMsg <- &natsio.Msg{Data: data}

						// first chunk
						msg = &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeChunk,
								SessionID: sessionID.String(),
								Properties: map[string]interface{}{
									PropertyOffset: int64(0),
								},
							},
							Body: []byte("12345"),
						}

						data, err = msgpack.Marshal(msg)
						assert.NoError(t, err)
						chanMsg <- &natsio.Msg{Data: data}

						return true
					}),
				).Return(&natsio.Subscription{}, nil)

				client.On("Publish",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(data []byte) bool {
						msg := &ws.ProtoMsg{}
						err := msgpack.Unmarshal(data, msg)
						assert.NoError(t, err)

						if msg.Header.MsgType == wsft.MessageTypeStat {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == wsft.MessageTypeGet {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == wsft.MessageTypeACK {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypePing {
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypeOpen {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						} else if msg.Header.MsgType == ws.MessageTypeClose {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						}

						return false
					}),
				).Return(nil)
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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
			},
			DeviceFunc: func(client *nats_mocks.Client) {
				client.On("ChanSubscribe",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(chanMsg chan *natsio.Msg) bool {
						// accept the open request
						b, _ := msgpack.Marshal(ws.Accept{
							Version:   ws.ProtocolVersion,
							Protocols: []ws.ProtoType{ws.ProtoTypeFileTransfer},
						})
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeControl,
								MsgType:   ws.MessageTypeAccept,
								SessionID: sessionID.String(),
							},
							Body: b,
						}
						b, _ = msgpack.Marshal(msg)
						chanMsg <- &natsio.Msg{Data: b}
						// file info response
						body := wsft.FileInfo{
							Path: string2pointer("/absolute/path"),
							Mode: uint322pointer(777),
							Size: int642pointer(10),
						}
						bodyData, err := msgpack.Marshal(body)
						msg = &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeFileInfo,
								SessionID: sessionID.String(),
							},
							Body: bodyData,
						}

						data, err := msgpack.Marshal(msg)
						assert.NoError(t, err)
						chanMsg <- &natsio.Msg{Data: data}

						// first chunk
						msg = &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeChunk,
								SessionID: sessionID.String(),
								Properties: map[string]interface{}{
									PropertyOffset: int64(0),
								},
							},
							Body: []byte("12345"),
						}

						data, err = msgpack.Marshal(msg)
						assert.NoError(t, err)
						chanMsg <- &natsio.Msg{Data: data}

						// second chunk
						msg = &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeFileTransfer,
								MsgType:   wsft.MessageTypeChunk,
								SessionID: sessionID.String(),
								Properties: map[string]interface{}{
									PropertyOffset: int64(100),
								},
							},
							Body: []byte("67890"),
						}

						data, err = msgpack.Marshal(msg)
						assert.NoError(t, err)
						chanMsg <- &natsio.Msg{Data: data}

						return true
					}),
				).Return(&natsio.Subscription{}, nil)

				client.On("Publish",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(data []byte) bool {
						msg := &ws.ProtoMsg{}
						err := msgpack.Unmarshal(data, msg)
						assert.NoError(t, err)

						if msg.Header.MsgType == wsft.MessageTypeStat {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == wsft.MessageTypeGet {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == wsft.MessageTypeACK {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypePing {
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypeOpen {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						} else if msg.Header.MsgType == ws.MessageTypeClose {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						}

						return false
					}),
				).Return(nil)
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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
			},
			DeviceFunc: func(client *nats_mocks.Client) {
				client.On("ChanSubscribe",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(chanMsg chan *natsio.Msg) bool {
						// accept the open request
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeControl,
								MsgType:   ws.MessageTypeOpen,
								SessionID: sessionID.String(),
								Properties: map[string]interface{}{
									"status": 1,
								},
							},
							Body: []byte("mender-connect v1.0 simulation"),
						}
						b, _ := msgpack.Marshal(msg)
						chanMsg <- &natsio.Msg{Data: b}
						return true
					}),
				).Return(&natsio.Subscription{}, nil)

				client.On("Publish",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(data []byte) bool {
						msg := &ws.ProtoMsg{}
						err := msgpack.Unmarshal(data, msg)
						assert.NoError(t, err)

						if msg.Header.MsgType == wsft.MessageTypeStat {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == wsft.MessageTypeGet {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == wsft.MessageTypeACK {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypePing {
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypeOpen {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						} else if msg.Header.MsgType == ws.MessageTypeClose {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						}

						return false
					}),
				).Return(nil)
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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
			},
			AppDownloadFile:    true,
			AppDownloadFileErr: errors.New("generic error"),

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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
			},

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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusDisconnected,
			},
			HTTPStatus: http.StatusConflict,
		},
		{
			Name:     "ko, not found",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},

			GetDeviceError: app.ErrDeviceNotFound,

			HTTPStatus: http.StatusNotFound,
		},
		{
			Name:     "ko, other error",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},

			GetDeviceError: errors.New("error"),

			HTTPStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			app := &app_mocks.App{}
			defer app.AssertExpectations(t)

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

			natsClient := &nats_mocks.Client{}
			defer natsClient.AssertExpectations(t)

			if tc.DeviceFunc != nil {
				tc.DeviceFunc(natsClient)
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

				if tc.Identity.IsUser {
					app.On("GetDevice",
						mock.MatchedBy(func(_ context.Context) bool {
							return true
						}),
						tc.Identity.Tenant,
						tc.DeviceID,
					).Return(tc.GetDevice, tc.GetDeviceError)
				}
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
	originalNewFileTransferSessionID := newFileTransferSessionID
	originalFileTransferTimeout := fileTransferTimeout
	originalAckSlidingWindowRecv := ackSlidingWindowRecv
	defer func() {
		newFileTransferSessionID = originalNewFileTransferSessionID
		fileTransferTimeout = originalFileTransferTimeout
		ackSlidingWindowRecv = originalAckSlidingWindowRecv
	}()

	fileTransferTimeout = 2 * time.Second
	ackSlidingWindowRecv = 0

	sessionID, _ := uuid.NewRandom()
	newFileTransferSessionID = func() (uuid.UUID, error) {
		return sessionID, nil
	}

	testCases := []struct {
		Name     string
		DeviceID string
		Body     map[string][]string
		File     []byte
		Identity *identity.Identity

		GetDevice        *model.Device
		GetDeviceError   error
		DeviceFunc       func(*nats_mocks.Client)
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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
			},
			DeviceFunc: func(client *nats_mocks.Client) {
				client.On("ChanSubscribe",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(chanMsg chan *natsio.Msg) bool {
						// accept the open request
						b, _ := msgpack.Marshal(ws.Accept{
							Version:   ws.ProtocolVersion,
							Protocols: []ws.ProtoType{ws.ProtoTypeFileTransfer},
						})
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeControl,
								MsgType:   ws.MessageTypeAccept,
								SessionID: sessionID.String(),
							},
							Body: b,
						}
						b, _ = msgpack.Marshal(msg)
						chanMsg <- &natsio.Msg{Data: b}

						// ack the put operation
						msg = &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								MsgType:   wsft.MessageTypeACK,
								SessionID: sessionID.String(),
							},
						}

						data, _ := msgpack.Marshal(ws.ProtoMsg{
							Header: ws.ProtoHdr{
								MsgType:   wsft.MessageTypeACK,
								SessionID: sessionID.String(),
							},
						})
						chanMsg <- &natsio.Msg{Data: data}

						// ack the chunk
						data, _ = msgpack.Marshal(ws.ProtoMsg{
							Header: ws.ProtoHdr{
								MsgType:   wsft.MessageTypeACK,
								SessionID: sessionID.String(),
								Properties: map[string]interface{}{
									"offset": int64(10),
								},
							},
						})
						chanMsg <- &natsio.Msg{Data: data}

						return true
					}),
				).Return(&natsio.Subscription{}, nil)

				client.On("Publish",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(data []byte) bool {
						msg := &ws.ProtoMsg{}
						err := msgpack.Unmarshal(data, msg)
						assert.NoError(t, err)

						if msg.Header.MsgType == wsft.MessageTypePut {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == wsft.MessageTypeChunk {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypePing {
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypeOpen {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						} else if msg.Header.MsgType == ws.MessageTypeClose {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						}

						return false
					}),
				).Return(nil)
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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
			},
			DeviceFunc: func(client *nats_mocks.Client) {
				client.On("ChanSubscribe",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(chanMsg chan *natsio.Msg) bool {
						// accept the open request
						b, _ := msgpack.Marshal(ws.Accept{
							Version:   ws.ProtocolVersion,
							Protocols: []ws.ProtoType{ws.ProtoTypeFileTransfer},
						})
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeControl,
								MsgType:   ws.MessageTypeAccept,
								SessionID: sessionID.String(),
							},
							Body: b,
						}
						b, _ = msgpack.Marshal(msg)
						chanMsg <- &natsio.Msg{Data: b}
						// ack the put operation
						msg = &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								MsgType:   wsft.MessageTypeACK,
								SessionID: sessionID.String(),
							},
						}

						data, err := msgpack.Marshal(msg)
						assert.NoError(t, err)
						chanMsg <- &natsio.Msg{Data: data}

						return true
					}),
				).Return(&natsio.Subscription{}, nil)

				client.On("Publish",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(data []byte) bool {
						msg := &ws.ProtoMsg{}
						err := msgpack.Unmarshal(data, msg)
						assert.NoError(t, err)

						if msg.Header.MsgType == wsft.MessageTypePut {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == wsft.MessageTypeChunk {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypePing {
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypeOpen {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						} else if msg.Header.MsgType == ws.MessageTypeClose {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						}

						return false
					}),
				).Return(nil)
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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
			},
			DeviceFunc: func(client *nats_mocks.Client) {
				client.On("ChanSubscribe",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(chanMsg chan *natsio.Msg) bool {
						// accept the open request
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
								SessionID: sessionID.String(),
							},
							Body: b,
						}
						b, _ = msgpack.Marshal(msg)
						chanMsg <- &natsio.Msg{Data: b}
						return true
					}),
				).Return(&natsio.Subscription{}, nil)

				client.On("Publish",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(data []byte) bool {
						msg := &ws.ProtoMsg{}
						err := msgpack.Unmarshal(data, msg)
						assert.NoError(t, err)

						if msg.Header.MsgType == wsft.MessageTypePut {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == wsft.MessageTypeChunk {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypePing {
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypeOpen {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						} else if msg.Header.MsgType == ws.MessageTypeClose {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						}

						return false
					}),
				).Return(nil)
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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
			},
			DeviceFunc: func(client *nats_mocks.Client) {
				client.On("ChanSubscribe",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(chanMsg chan *natsio.Msg) bool {
						// accept the open request
						b, _ := msgpack.Marshal(ws.Accept{
							Version:   ws.ProtocolVersion,
							Protocols: []ws.ProtoType{ws.ProtoTypeFileTransfer},
						})
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeControl,
								MsgType:   ws.MessageTypeAccept,
								SessionID: sessionID.String(),
							},
							Body: b,
						}
						b, _ = msgpack.Marshal(msg)
						chanMsg <- &natsio.Msg{Data: b}
						// put response
						body := wsft.Error{
							Error:       string2pointer("file not writeable"),
							MessageType: string2pointer(wsft.MessageTypePut),
						}
						bodyData, err := msgpack.Marshal(body)
						msg = &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								MsgType:   wsft.MessageTypeError,
								SessionID: sessionID.String(),
							},
							Body: bodyData,
						}

						data, err := msgpack.Marshal(msg)
						assert.NoError(t, err)
						chanMsg <- &natsio.Msg{Data: data}

						return true
					}),
				).Return(&natsio.Subscription{}, nil)

				client.On("Publish",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(data []byte) bool {
						msg := &ws.ProtoMsg{}
						err := msgpack.Unmarshal(data, msg)
						assert.NoError(t, err)

						if msg.Header.MsgType == wsft.MessageTypePut {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == wsft.MessageTypeChunk {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypePing {
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypeOpen {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						} else if msg.Header.MsgType == ws.MessageTypeClose {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						}

						return false
					}),
				).Return(nil)
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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
			},
			DeviceFunc: func(client *nats_mocks.Client) {
				client.On("ChanSubscribe",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(chanMsg chan *natsio.Msg) bool {
						// accept the open request
						b, _ := msgpack.Marshal(ws.Accept{
							Version:   ws.ProtocolVersion,
							Protocols: []ws.ProtoType{ws.ProtoTypeFileTransfer},
						})
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeControl,
								MsgType:   ws.MessageTypeAccept,
								SessionID: sessionID.String(),
							},
							Body: b,
						}
						b, _ = msgpack.Marshal(msg)
						chanMsg <- &natsio.Msg{Data: b}
						// ack the put operation
						msg = &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								MsgType:   wsft.MessageTypeACK,
								SessionID: sessionID.String(),
							},
						}

						data, err := msgpack.Marshal(msg)
						assert.NoError(t, err)
						chanMsg <- &natsio.Msg{Data: data}

						// put response
						body := ws.Error{
							Error:       "failed to Write",
							MessageType: wsft.MessageTypePut,
							Code:        http.StatusBadRequest,
						}
						bodyData, err := msgpack.Marshal(body)
						msg = &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								MsgType:   wsft.MessageTypeError,
								SessionID: sessionID.String(),
							},
							Body: bodyData,
						}

						data, err = msgpack.Marshal(msg)
						assert.NoError(t, err)
						chanMsg <- &natsio.Msg{Data: data}

						return true
					}),
				).Return(&natsio.Subscription{}, nil)

				client.On("Publish",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(data []byte) bool {
						msg := &ws.ProtoMsg{}
						err := msgpack.Unmarshal(data, msg)
						assert.NoError(t, err)

						if msg.Header.MsgType == wsft.MessageTypePut {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == wsft.MessageTypeChunk {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypePing {
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypeOpen {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						} else if msg.Header.MsgType == ws.MessageTypeClose {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						}

						return false
					}),
				).Return(nil)
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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
			},
			DeviceFunc: func(client *nats_mocks.Client) {
				client.On("ChanSubscribe",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(chanMsg chan *natsio.Msg) bool {
						// accept the open request
						b, _ := msgpack.Marshal(ws.Accept{
							Version:   ws.ProtocolVersion,
							Protocols: []ws.ProtoType{ws.ProtoTypeFileTransfer},
						})
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeControl,
								MsgType:   ws.MessageTypeAccept,
								SessionID: sessionID.String(),
							},
							Body: b,
						}
						b, _ = msgpack.Marshal(msg)
						chanMsg <- &natsio.Msg{Data: b}

						// then timeout...
						return true
					}),
				).Return(&natsio.Subscription{}, nil)

				client.On("Publish",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(data []byte) bool {
						msg := &ws.ProtoMsg{}
						err := msgpack.Unmarshal(data, msg)
						assert.NoError(t, err)

						if msg.Header.MsgType == wsft.MessageTypePut {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == wsft.MessageTypeChunk {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypePing {
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypeOpen {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						} else if msg.Header.MsgType == ws.MessageTypeClose {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						}

						return false
					}),
				).Return(nil)
			},
			AppUploadFile: true,

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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
			},
			DeviceFunc: func(client *nats_mocks.Client) {
				client.On("ChanSubscribe",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(chanMsg chan *natsio.Msg) bool {
						return true
					}),
				).Return(&natsio.Subscription{}, nil)

				client.On("Publish",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(data []byte) bool {
						msg := &ws.ProtoMsg{}
						err := msgpack.Unmarshal(data, msg)
						assert.NoError(t, err)

						if msg.Header.MsgType == wsft.MessageTypePut {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == wsft.MessageTypeChunk {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypePing {
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypeOpen {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						} else if msg.Header.MsgType == ws.MessageTypeClose {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						}

						return false
					}),
				).Return(nil)
			},
			AppUploadFile: true,

			HTTPStatus: http.StatusRequestTimeout,
		},
		{
			Name:     "error, handshake client error",
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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
			},
			DeviceFunc: func(client *nats_mocks.Client) {
				client.On("ChanSubscribe",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(chanMsg chan *natsio.Msg) bool {
						b, _ := msgpack.Marshal(ws.Error{
							Error: "I don't understand...",
						})
						msg := &ws.ProtoMsg{
							Header: ws.ProtoHdr{
								Proto:     ws.ProtoTypeControl,
								MsgType:   ws.MessageTypeError,
								SessionID: sessionID.String(),
							},
							Body: b,
						}
						b, _ = msgpack.Marshal(msg)
						chanMsg <- &natsio.Msg{Data: b}
						return true
					}),
				).Return(&natsio.Subscription{}, nil)

				client.On("Publish",
					mock.AnythingOfType("string"),
					mock.MatchedBy(func(data []byte) bool {
						msg := &ws.ProtoMsg{}
						err := msgpack.Unmarshal(data, msg)
						assert.NoError(t, err)

						if msg.Header.MsgType == wsft.MessageTypePut {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == wsft.MessageTypeChunk {
							assert.Equal(t, ws.ProtoTypeFileTransfer, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypePing {
							assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
							return true
						} else if msg.Header.MsgType == ws.MessageTypeOpen {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						} else if msg.Header.MsgType == ws.MessageTypeError {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						} else if msg.Header.MsgType == ws.MessageTypeClose {
							return assert.Equal(t, ws.ProtoTypeControl, msg.Header.Proto)
						}

						return false
					}),
				).Return(nil)
			},
			AppUploadFile: true,

			HTTPStatus: http.StatusInternalServerError,
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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
			},
			AppUploadFile:    true,
			AppUploadFileErr: errors.New("generic error"),

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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
			},

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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusDisconnected,
			},
			HTTPStatus: http.StatusConflict,
		},
		{
			Name:     "ko, not found",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},

			GetDeviceError: app.ErrDeviceNotFound,

			HTTPStatus: http.StatusNotFound,
		},
		{
			Name:     "ko, other error",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},

			GetDeviceError: errors.New("error"),

			HTTPStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
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
				tc.DeviceFunc(natsClient)
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

				if tc.Identity.IsUser {
					app.On("GetDevice",
						mock.MatchedBy(func(_ context.Context) bool {
							return true
						}),
						tc.Identity.Tenant,
						tc.DeviceID,
					).Return(tc.GetDevice, tc.GetDeviceError)
				}
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tc.HTTPStatus, w.Code, w.Body.String())
		})
	}
}
