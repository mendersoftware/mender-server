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
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/mendersoftware/mender-server/pkg/stream"
	"github.com/mendersoftware/mender-server/pkg/ws"
	"github.com/mendersoftware/mender-server/pkg/ws/menderclient"
	app_mocks "github.com/mendersoftware/mender-server/services/deviceconnect/app/mocks"
	nats_mocks "github.com/mendersoftware/mender-server/services/deviceconnect/client/nats/mocks"
)

func TestInternalCheckUpdate(t *testing.T) {
	testCases := []struct {
		Name     string
		TenantID string
		DeviceID string

		OnConnect func(*testing.T, context.Context, string, string) (stream.Conn, error)

		PublishErr error

		HTTPStatus int
	}{
		{
			Name:     "ok",
			DeviceID: "1234567890",

			OnConnect: func(t *testing.T,
				ctx context.Context,
				localAddr, remoteAddr string) (stream.Conn, error) {
				conn := setupConn(t, nil, "foo", "bar")
				conn.On("Send", ctx, mock.MatchedBy(func(b []byte) bool {
					var msg ws.ProtoMsg
					err := msgpack.Unmarshal(b, &msg)
					return assert.NoError(t, err) &&
						assert.Equal(t, ws.ProtoTypeMenderClient, msg.Header.Proto) &&
						assert.Equal(t, menderclient.MessageTypeMenderClientCheckUpdate, msg.Header.MsgType)
				})).
					Return(nil)
				conn.On("Close", ctx).
					Return(nil)
				return conn, nil
			},

			HTTPStatus: http.StatusAccepted,
		},
		{
			Name:     "ok, with tenantID",
			DeviceID: "1234567890",
			TenantID: "tenant_id",

			OnConnect: func(t *testing.T,
				ctx context.Context,
				localAddr, remoteAddr string) (stream.Conn, error) {
				conn := setupConn(t, nil, "foo", "bar")
				conn.On("Send", ctx, mock.MatchedBy(func(b []byte) bool {
					var msg ws.ProtoMsg
					err := msgpack.Unmarshal(b, &msg)
					return assert.NoError(t, err) &&
						assert.Equal(t, ws.ProtoTypeMenderClient, msg.Header.Proto) &&
						assert.Equal(t, menderclient.MessageTypeMenderClientCheckUpdate, msg.Header.MsgType)
				})).
					Return(nil)
				conn.On("Close", ctx).
					Return(nil)
				return conn, nil
			},

			HTTPStatus: http.StatusAccepted,
		},
		{
			Name:     "ko, not found",
			DeviceID: "1234567890",

			OnConnect: func(t *testing.T,
				ctx context.Context,
				localAddr, remoteAddr string) (stream.Conn, error) {
				return nil, stream.ErrConnectionRefused
			},

			HTTPStatus: 404,
		},
		{
			Name:     "ko, other error",
			DeviceID: "1234567890",

			OnConnect: func(t *testing.T,
				ctx context.Context,
				localAddr, remoteAddr string) (stream.Conn, error) {
				return nil, errors.New("error!")
			},

			HTTPStatus: 500,
		},
		{
			Name:     "ok, with tenantID",
			DeviceID: "1234567890",
			TenantID: "tenant_id",

			OnConnect: func(t *testing.T,
				ctx context.Context,
				localAddr, remoteAddr string) (stream.Conn, error) {
				conn := setupConn(t, nil, "foo", "bar")
				conn.On("Send", ctx, mock.MatchedBy(func(b []byte) bool {
					var msg ws.ProtoMsg
					err := msgpack.Unmarshal(b, &msg)
					return assert.NoError(t, err) &&
						assert.Equal(t, ws.ProtoTypeMenderClient, msg.Header.Proto) &&
						assert.Equal(t, menderclient.MessageTypeMenderClientCheckUpdate, msg.Header.MsgType)
				})).
					Return(errors.New("error!"))
				conn.On("Close", ctx).
					Return(nil)
				return conn, nil
			},

			HTTPStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			app := app_mocks.NewApp(t)
			natsClient := nats_mocks.NewClient(t)

			router, _ := NewRouter(app, natsClient, nil)
			s := httptest.NewServer(router)
			defer s.Close()

			url := strings.Replace(APIURLInternalDevicesIDCheckUpdate, ":tenantId", tc.TenantID, 1)
			url = strings.Replace(url, ":deviceId", tc.DeviceID, 1)
			req, err := http.NewRequest("POST", "http://localhost"+url, nil)

			if tc.OnConnect != nil {
				natsClient.On("Connect", contextMatcher, mock.Anything, mock.Anything).
					Return(func(
						ctx context.Context, localAddr, remoteAddr string,
					) (stream.Conn, error) {
						return tc.OnConnect(t, ctx, localAddr, remoteAddr)
					})
			}
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tc.HTTPStatus, w.Code)
		})
	}
}

func TestInternalSendInventory(t *testing.T) {
	testCases := []struct {
		Name     string
		TenantID string
		DeviceID string

		OnConnect func(*testing.T, context.Context, string, string) (stream.Conn, error)

		PublishErr error

		HTTPStatus int
	}{
		{
			Name:     "ok",
			DeviceID: "1234567890",

			OnConnect: func(t *testing.T,
				ctx context.Context,
				localAddr, remoteAddr string) (stream.Conn, error) {
				conn := setupConn(t, nil, "foo", "bar")
				conn.On("Send", ctx, mock.MatchedBy(func(b []byte) bool {
					var msg ws.ProtoMsg
					err := msgpack.Unmarshal(b, &msg)
					return assert.NoError(t, err) &&
						assert.Equal(t, ws.ProtoTypeMenderClient, msg.Header.Proto) &&
						assert.Equal(t, menderclient.MessageTypeMenderClientSendInventory, msg.Header.MsgType)
				})).
					Return(nil)
				conn.On("Close", ctx).
					Return(nil)
				return conn, nil
			},

			HTTPStatus: http.StatusAccepted,
		},
		{
			Name:     "ok, with tenantID",
			DeviceID: "1234567890",
			TenantID: "tenant_id",

			OnConnect: func(t *testing.T,
				ctx context.Context,
				localAddr, remoteAddr string) (stream.Conn, error) {
				conn := setupConn(t, nil, "foo", "bar")
				conn.On("Send", ctx, mock.MatchedBy(func(b []byte) bool {
					var msg ws.ProtoMsg
					err := msgpack.Unmarshal(b, &msg)
					return assert.NoError(t, err) &&
						assert.Equal(t, ws.ProtoTypeMenderClient, msg.Header.Proto) &&
						assert.Equal(t, menderclient.MessageTypeMenderClientSendInventory, msg.Header.MsgType)
				})).
					Return(nil)
				conn.On("Close", ctx).
					Return(nil)
				return conn, nil
			},

			HTTPStatus: http.StatusAccepted,
		},
		{
			Name:     "ko, not found",
			DeviceID: "1234567890",

			OnConnect: func(t *testing.T,
				ctx context.Context,
				localAddr, remoteAddr string) (stream.Conn, error) {
				return nil, stream.ErrConnectionRefused
			},

			HTTPStatus: 404,
		},
		{
			Name:     "ko, other error",
			DeviceID: "1234567890",

			OnConnect: func(t *testing.T,
				ctx context.Context,
				localAddr, remoteAddr string) (stream.Conn, error) {
				return nil, errors.New("error")
			},

			HTTPStatus: 500,
		},
		{
			Name:     "ok, with tenantID",
			DeviceID: "1234567890",
			TenantID: "tenant_id",

			OnConnect: func(t *testing.T,
				ctx context.Context,
				localAddr, remoteAddr string) (stream.Conn, error) {
				conn := setupConn(t, nil, "foo", "bar")
				conn.On("Send", ctx, mock.MatchedBy(func(b []byte) bool {
					var msg ws.ProtoMsg
					err := msgpack.Unmarshal(b, &msg)
					return assert.NoError(t, err) &&
						assert.Equal(t, ws.ProtoTypeMenderClient, msg.Header.Proto) &&
						assert.Equal(t, menderclient.MessageTypeMenderClientSendInventory, msg.Header.MsgType)
				})).
					Return(errors.New("error!"))
				conn.On("Close", ctx).
					Return(nil)
				return conn, nil
			},

			HTTPStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			app := app_mocks.NewApp(t)
			natsClient := nats_mocks.NewClient(t)

			router, _ := NewRouter(app, natsClient, nil)
			s := httptest.NewServer(router)
			defer s.Close()

			url := strings.Replace(APIURLInternalDevicesIDSendInventory, ":tenantId", tc.TenantID, 1)
			url = strings.Replace(url, ":deviceId", tc.DeviceID, 1)
			req, err := http.NewRequest("POST", "http://localhost"+url, nil)

			if tc.OnConnect != nil {
				natsClient.On("Connect",
					contextMatcher,
					mock.MatchedBy(func(s string) bool { return strings.HasPrefix(s, tc.TenantID+":cmd") }),
					tc.DeviceID,
				).
					Return(func(
						ctx context.Context, localAddr, remoteAddr string,
					) (stream.Conn, error) {
						return tc.OnConnect(t, ctx, localAddr, remoteAddr)
					})
			}

			if !assert.NoError(t, err) {
				t.FailNow()
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tc.HTTPStatus, w.Code)
		})
	}
}

func TestDeleteTenant(t *testing.T) {
	t.Parallel()
	const tenantID = "123456789012345678901234"

	testCases := []struct {
		Name    string
		Request *http.Request
		Error   error
		Status  int
	}{{
		Name: "ok",

		Request: func() *http.Request {
			repl := strings.NewReplacer(
				":tenantId", tenantID,
			)
			req, _ := http.NewRequest("DELETE",
				"http://localhost"+repl.Replace(APIURLInternalTenant),
				nil,
			)
			return req
		}(),

		Status: http.StatusNoContent,
	}, {
		Name: "error, internal server error",

		Request: func() *http.Request {
			repl := strings.NewReplacer(
				":tenantId", tenantID,
			)
			req, _ := http.NewRequest("DELETE",
				"http://localhost"+repl.Replace(APIURLInternalTenant),
				nil,
			)
			return req
		}(),

		Error:  errors.New("error"),
		Status: http.StatusInternalServerError,
	}}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			app := &app_mocks.App{}
			defer app.AssertExpectations(t)

			router, _ := NewRouter(app, nil, nil)
			s := httptest.NewServer(router)
			defer s.Close()

			app.On("DeleteTenant",
				mock.MatchedBy(func(_ context.Context) bool {
					return true
				}),
				tenantID,
			).Return(tc.Error)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, tc.Request)
			assert.Equal(t, tc.Status, w.Code)
		})
	}
}
