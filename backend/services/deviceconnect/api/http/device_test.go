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
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/stream"
	stream_mocks "github.com/mendersoftware/mender-server/pkg/stream/mocks"
	"github.com/mendersoftware/mender-server/pkg/ws"
	"github.com/mendersoftware/mender-server/pkg/ws/shell"

	app_mocks "github.com/mendersoftware/mender-server/services/deviceconnect/app/mocks"
	nats_mocks "github.com/mendersoftware/mender-server/services/deviceconnect/client/nats/mocks"
	"github.com/mendersoftware/mender-server/services/deviceconnect/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var contextMatcher = mock.MatchedBy(func(context.Context) bool {
	return true
})

func setupMockListener(t *testing.T, connCh chan stream.Conn) *stream_mocks.Listener {
	listener := stream_mocks.NewListener(t)
	listener.On("Close", contextMatcher).
		Run(func(args mock.Arguments) { close(connCh) }).
		Return(nil).
		Maybe()

	go func() {
		listener.On("Accept", contextMatcher).
			Return(func(context.Context) (stream.Conn, error) {
				conn, open := <-connCh
				if !open {
					return nil, stream.ErrClosed
				}
				return conn, nil
			})
	}()
	return listener
}

func setupConn(t *testing.T, recvCh <-chan []byte, remoteAddr, localAddr string) *stream_mocks.Conn {
	conn := stream_mocks.NewConn(t)
	conn.On("RemoteAddr").Return(remoteAddr).Maybe()
	conn.On("LocalAddr").Return(localAddr).Maybe()
	conn.On("Recv", contextMatcher).
		Return(func(context.Context) ([]byte, error) {
			data, open := <-recvCh
			if !open {
				return nil, stream.ErrClosed
			}
			return data, nil
		}).
		Maybe()
	return conn
}

func TestDeviceConnect(t *testing.T) {
	// temporarily speed things up a bit
	prevPongWait := pongWait
	prevWriteWait := writeWait
	defer func() {
		pongWait = prevPongWait
		writeWait = prevWriteWait
	}()
	pongWait = time.Second
	writeWait = time.Second

	Identity := identity.Identity{
		Subject:  "00000000-0000-0000-0000-000000000000",
		Tenant:   "000000000000000000000000",
		IsDevice: true,
	}
	app := &app_mocks.App{}
	app.On("RegisterShutdownCancel",
		mock.AnythingOfType("context.CancelFunc"),
	).Return(uint32(1))
	app.On("UnregisterShutdownCancel",
		mock.AnythingOfType("uint32"),
	).Return()
	app.On("SetDeviceConnected",
		mock.MatchedBy(func(_ context.Context) bool {
			return true
		}),
		Identity.Tenant,
		Identity.Subject,
	).Return(int64(1), nil).Once()

	app.On("SetDeviceDisconnected",
		mock.MatchedBy(func(_ context.Context) bool {
			return true
		}),
		Identity.Tenant,
		Identity.Subject,
		int64(1),
	).Return(nil).Once()

	connCh := make(chan stream.Conn, 1)
	listener := setupMockListener(t, connCh)
	natsClient := nats_mocks.NewClient(t)
	natsClient.On("Listen", Identity.Subject).
		Return(listener, nil)

	router, _ := NewRouter(app, natsClient, nil)
	s := httptest.NewServer(router)
	defer s.Close()

	url := "ws" + strings.TrimPrefix(s.URL, "http")

	headers := http.Header{}
	headers.Set(
		headerAuthorization,
		"Bearer "+GenerateJWT(Identity),
	)

	streamRecv := make(chan []byte, 10)
	streamConn := setupConn(t, streamRecv, Identity.Tenant+":foobar", Identity.Subject)

	var (
		conn *websocket.Conn
		err  error
	)
	pingReceived := make(chan struct{}, 1)
	pongReceived := make(chan struct{}, 1)
	connRecvCh := make(chan struct {
		Type  int
		Data  []byte
		Error error
	})
	t.Run("dial and accept", func(t *testing.T) {
		conn, _, err = websocket.DefaultDialer.Dial(url+APIURLDevicesConnect, headers)
		conn.SetPingHandler(func(message string) error {
			select {
			case pingReceived <- struct{}{}:
			default:
			}
			return conn.WriteControl(
				websocket.PongMessage,
				[]byte{},
				time.Now().Add(writeWait),
			)
		})
		conn.SetPongHandler(func(message string) error {
			select {
			case pongReceived <- struct{}{}:
			default:
			}
			return nil
		})
		go func() {
			// Async read from websocket to keep ping/pong hooks active
			for {
				typ, data, err := conn.ReadMessage()
				connRecvCh <- struct {
					Type  int
					Data  []byte
					Error error
				}{
					Type:  typ,
					Data:  data,
					Error: err,
				}
				if err != nil {
					break
				}
			}
		}()
		require.NoError(t, err)
		connCh <- streamConn
	})

	t.Run("websocket ping", func(t *testing.T) {
		select {
		case <-pingReceived:
		case <-time.After(pongWait * 2):
			assert.Fail(t, "did not receive ping within pongWait")
		}
	})
	t.Run("websocket pong", func(t *testing.T) {
		err = websocketPing(conn)
		assert.NoError(t, err)
		select {
		case <-pongReceived:
		case <-time.After(pongWait * 2):
			assert.Fail(t, "did not receive pong within pongWait")
		}
	})

	assertSend := func(t *testing.T, msg ws.ProtoMsg) {
		b, _ := msgpack.Marshal(msg)
		wait := make(chan struct{})
		streamConn.On("Send", contextMatcher, b).
			Run(func(args mock.Arguments) { close(wait) }).
			Return(nil).
			Once()
		err := conn.WriteMessage(websocket.BinaryMessage, b)
		assert.NoError(t, err)
		select {
		case <-wait:
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for message from websocket")
		}
	}

	t.Run("send from device", func(t *testing.T) {
		// terminal data
		assertSend(t, ws.ProtoMsg{
			Header: ws.ProtoHdr{
				Proto:     ws.ProtoTypeShell,
				MsgType:   shell.MessageTypeShellCommand,
				SessionID: "foobar",
			},
			Body: []byte("[batman@hogwartz ~]$"),
		})
		// stop the terminal
		assertSend(t, ws.ProtoMsg{
			Header: ws.ProtoHdr{
				Proto:     ws.ProtoTypeShell,
				MsgType:   shell.MessageTypeStopShell,
				SessionID: "foobar",
			},
		})
	})

	t.Run("recv from management", func(t *testing.T) {
		// test receiving a message "from management"
		msg := ws.ProtoMsg{
			Header: ws.ProtoHdr{
				Proto:     ws.ProtoTypeShell,
				MsgType:   "cmd",
				SessionID: "foobar",
			},
		}
		b, _ := msgpack.Marshal(msg)
		streamRecv <- b
		wsMessage := <-connRecvCh
		require.NoError(t, wsMessage.Error)
		assert.Equal(t, websocket.BinaryMessage, wsMessage.Type)
		assert.Equal(t, b, wsMessage.Data)
	})

	t.Run("close websocket", func(t *testing.T) {
		err := conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "test"),
			time.Now().Add(time.Second*10),
		)
		assert.NoError(t, err)
		select {
		case <-connCh:
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for connection cleanup")
		}
	})
}

func TestDeviceConnectFailures(t *testing.T) {
	JWT := GenerateJWT(identity.Identity{
		Subject:  "00000000-0000-0000-0000-000000000000",
		Tenant:   "000000000000000000000000",
		IsDevice: true,
	})
	testCases := []struct {
		Name          string
		Authorization string
		WithNATS      bool
		HTTPStatus    int
		HTTPError     error
	}{
		{
			Name:          "ko, unable to upgrade",
			Authorization: "Bearer " + JWT,
			WithNATS:      true,
			HTTPStatus:    http.StatusBadRequest,
		},
		{
			Name:          "error, unable to subscribe",
			Authorization: "Bearer " + JWT,
			HTTPStatus:    http.StatusInternalServerError,
		},
		{
			Name: "error, user auth",
			Authorization: "Bearer " + GenerateJWT(identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			}),
			HTTPStatus: http.StatusBadRequest,
		},
		{
			Name:       "ko, missing authorization header",
			HTTPStatus: http.StatusUnauthorized,
			HTTPError:  errors.New("Authorization not present in header"),
		},
		{
			Name:          "ko, malformed authorization header",
			Authorization: "malformed",
			HTTPStatus:    http.StatusUnauthorized,
			HTTPError:     errors.New("malformed Authorization header"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var natsClient *nats_mocks.Client
			if tc.WithNATS {
				natsClient = nats_mocks.NewClient(t)
				connCh := make(chan stream.Conn)
				listener := stream_mocks.NewListener(t)
				listener.On("Close", contextMatcher).
					Run(func(args mock.Arguments) { close(connCh) }).
					Return(nil).
					Maybe()
				listener.On("Accept", contextMatcher).
					Return(func(context.Context) (stream.Conn, error) {
						<-connCh
						return nil, stream.ErrClosed
					}).
					Maybe()
				natsClient.On("Listen", mock.Anything).
					Return(listener, nil)
			}

			router, _ := NewRouter(nil, natsClient, nil)
			req, err := http.NewRequest("GET", "http://localhost"+APIURLDevicesConnect, nil)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			if tc.Authorization != "" {
				req.Header.Add("Authorization", tc.Authorization)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tc.HTTPStatus, w.Code)

			if tc.HTTPError != nil {
				var response map[string]string
				body := w.Body.Bytes()
				_ = json.Unmarshal(body, &response)
				value := response["error"]
				assert.Equal(t, tc.HTTPError.Error(), value)
			}
		})
	}
}

func TestProvisionDevice(t *testing.T) {
	testCases := []struct {
		Name               string
		TenantID           string
		DeviceID           string
		Device             string
		ProvisionDeviceErr error
		HTTPStatus         int
	}{
		{
			Name:       "ok",
			TenantID:   "1234",
			DeviceID:   "1234",
			Device:     `{"device_id": "1234"}`,
			HTTPStatus: http.StatusCreated,
		},
		{
			Name:       "ko, empty payload",
			TenantID:   "1234",
			Device:     ``,
			HTTPStatus: http.StatusBadRequest,
		},
		{
			Name:       "ko, bad payload",
			TenantID:   "1234",
			Device:     `...`,
			HTTPStatus: http.StatusBadRequest,
		},
		{
			Name:       "ko, empty device ID",
			TenantID:   "1234",
			Device:     `{"device_id": ""}`,
			HTTPStatus: http.StatusBadRequest,
		},
		{
			Name:               "ko, error",
			TenantID:           "1234",
			DeviceID:           "1234",
			Device:             `{"device_id": "1234"}`,
			ProvisionDeviceErr: errors.New("error"),
			HTTPStatus:         http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			deviceConnectApp := &app_mocks.App{}
			if tc.DeviceID != "" {
				deviceConnectApp.On("ProvisionDevice",
					mock.MatchedBy(func(_ context.Context) bool {
						return true
					}),
					tc.TenantID,
					&model.Device{ID: tc.DeviceID},
				).Return(tc.ProvisionDeviceErr)
			}

			router, _ := NewRouter(deviceConnectApp, nil, nil)

			url := strings.Replace(APIURLInternalDevices, ":tenantId", tc.TenantID, 1)
			req, err := http.NewRequest("POST", url, strings.NewReader(tc.Device))
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tc.HTTPStatus, w.Code)

			deviceConnectApp.AssertExpectations(t)
		})
	}
}

func TestDeleteDevice(t *testing.T) {
	testCases := []struct {
		Name               string
		TenantID           string
		DeviceID           string
		ProvisionDeviceErr error
		HTTPStatus         int
	}{
		{
			Name:       "ok",
			TenantID:   "1234",
			DeviceID:   "abcd",
			HTTPStatus: http.StatusAccepted,
		},
		{
			Name:               "ko, empty device id",
			TenantID:           "1234",
			ProvisionDeviceErr: errors.New("error"),
			HTTPStatus:         http.StatusNotFound,
		},
		{
			Name:               "ko, error",
			TenantID:           "1234",
			DeviceID:           "abcd",
			ProvisionDeviceErr: errors.New("error"),
			HTTPStatus:         http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			deviceConnectApp := &app_mocks.App{}
			if tc.DeviceID != "" {
				deviceConnectApp.On("DeleteDevice",
					mock.MatchedBy(func(_ context.Context) bool {
						return true
					}),
					tc.TenantID,
					tc.DeviceID,
				).Return(tc.ProvisionDeviceErr)
			}

			router, _ := NewRouter(deviceConnectApp, nil, nil)

			url := strings.Replace(APIURLInternalDevicesID, ":tenantId", tc.TenantID, 1)
			url = strings.Replace(url, ":deviceId", tc.DeviceID, 1)
			req, err := http.NewRequest("DELETE", url, nil)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tc.HTTPStatus, w.Code)

			deviceConnectApp.AssertExpectations(t)
		})
	}
}
