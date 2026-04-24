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
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/stream"
	stream_mocks "github.com/mendersoftware/mender-server/pkg/stream/mocks"
	"github.com/mendersoftware/mender-server/pkg/ws"
	"github.com/mendersoftware/mender-server/pkg/ws/menderclient"
	"github.com/mendersoftware/mender-server/pkg/ws/shell"

	"github.com/mendersoftware/mender-server/services/deviceconnect/app"
	app_mocks "github.com/mendersoftware/mender-server/services/deviceconnect/app/mocks"
	nats_mocks "github.com/mendersoftware/mender-server/services/deviceconnect/client/nats/mocks"
	"github.com/mendersoftware/mender-server/services/deviceconnect/model"
)

var natsPort int32 = 14420

type discardRecorder struct{}

func (discardRecorder) Record(context.Context, []byte) error { return nil }
func (discardRecorder) Close(context.Context) error          { return nil }

func GenerateJWT(id identity.Identity) string {
	JWT := base64.RawURLEncoding.EncodeToString(
		[]byte(`{"alg":"HS256","typ":"JWT"}`),
	)
	b, _ := json.Marshal(id)
	JWT = JWT + "." + base64.RawURLEncoding.EncodeToString(b)
	hash := hmac.New(sha256.New, []byte("hmac-sha256-secret"))
	JWT = JWT + "." + base64.RawURLEncoding.EncodeToString(
		hash.Sum([]byte(JWT)),
	)
	return JWT
}

func TestManagementGetDevice(t *testing.T) {
	testCases := []struct {
		Name     string
		DeviceID string
		Identity *identity.Identity

		GetDevice      *model.Device
		GetDeviceError error

		HTTPStatus int
		Body       *model.Device
	}{
		{
			Name:     "ok",
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

			HTTPStatus: 200,
			Body: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
			},
		},
		{
			Name:     "ko, missing auth",
			DeviceID: "1234567890",

			HTTPStatus: 401,
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

			HTTPStatus: 404,
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

			HTTPStatus: 400,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			app := &app_mocks.App{}

			router, _ := NewRouter(app, nil, nil)
			s := httptest.NewServer(router)
			defer s.Close()

			url := strings.Replace(APIURLManagementDevice, ":deviceId", tc.DeviceID, 1)
			req, err := http.NewRequest("GET", "http://localhost"+url, nil)
			if tc.Identity != nil {
				jwt := GenerateJWT(*tc.Identity)
				app.On("GetDevice",
					mock.MatchedBy(func(_ context.Context) bool {
						return true
					}),
					tc.Identity.Tenant,
					tc.DeviceID,
				).Return(tc.GetDevice, tc.GetDeviceError)
				req.Header.Set(headerAuthorization, "Bearer "+jwt)
			}
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tc.HTTPStatus, w.Code)

			if tc.HTTPStatus == http.StatusOK {
				var response *model.Device
				body := w.Body.Bytes()
				_ = json.Unmarshal(body, &response)
				assert.Equal(t, tc.Body, response)
			}

			app.AssertExpectations(t)
		})
	}
}

func TestManagementConnect(t *testing.T) {
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
		Subject: "00000000-0000-0000-0000-000000000000",
		Tenant:  "000000000000000000000000",
		IsUser:  true,
		Plan:    "professional",
	}
	var sessionID string
	app := &app_mocks.App{}
	app.On("PrepareUserSession",
		mock.MatchedBy(func(_ context.Context) bool {
			return true
		}),
		mock.MatchedBy(func(sess *model.Session) bool {
			sessionID = sess.ID
			return true
		}),
	).
		Run(func(args mock.Arguments) {
			app.On("LogUserSession",
				mock.MatchedBy(func(_ context.Context) bool {
					return true
				}),
				mock.MatchedBy(func(sess *model.Session) bool {
					return sess.ID == sessionID
				}),
				mock.AnythingOfType("string"),
			).Return(nil).
				On("FreeUserSession",
					mock.MatchedBy(func(_ context.Context) bool {
						return true
					}),
					sessionID,
					mock.AnythingOfType("[]string"),
				).Return(nil).
				On("GetControlRecorder",
					sessionID,
				).Return(discardRecorder{}).
				On("GetRecorder",
					sessionID,
				).Return(discardRecorder{})
		}).
		Return(nil)

	natsClient := nats_mocks.NewClient(t)
	router, _ := NewRouter(app, natsClient, nil)
	s := httptest.NewServer(router)
	defer s.Close()

	streamRecv := make(chan []byte, 10)
	streamConn := setupConn(t, streamRecv, Identity.Tenant+":foobar", Identity.Subject)

	var (
		conn *websocket.Conn
		err  error
	)
	connRecvCh := make(chan struct {
		Type  int
		Data  []byte
		Error error
	})
	if !t.Run("dial and connect", func(t *testing.T) {
		natsClient.On("Connect", contextMatcher, mock.MatchedBy(func(s string) bool {
			return strings.HasPrefix(s, Identity.Tenant)
		}), "1234567890").
			Return(streamConn, nil)

		url := "ws" + strings.TrimPrefix(s.URL, "http")
		url = url + strings.Replace(
			APIURLManagementDeviceConnect, ":deviceId",
			"1234567890", 1,
		)
		headers := http.Header{}
		headers.Set(
			headerAuthorization,
			"Bearer "+GenerateJWT(Identity),
		)
		conn, _, err = websocket.DefaultDialer.Dial(url, headers)
		require.NoError(t, err)
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
	}) {
		t.FailNow()
	}

	t.Run("websocket pong", func(t *testing.T) {
		pingReceived := make(chan struct{}, 1)
		pongReceived := make(chan struct{}, 1)
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
		err = websocketPing(conn)
		assert.NoError(t, err)
		select {
		case <-pongReceived:
		case <-time.After(pongWait * 2):
			assert.Fail(t, "did not receive pong within pongWait")
		}
	})

	assertSend := func(t *testing.T, msg ws.ProtoMsg) {
		wsBytes, _ := msgpack.Marshal(msg)
		msg.Header.SessionID = sessionID
		msg.Header.Properties = map[string]any{
			"user_id": Identity.Subject,
		}
		sendBytes, _ := msgpack.Marshal(msg)
		wait := make(chan struct{})
		streamConn.On("Send", contextMatcher, sendBytes).
			Run(func(args mock.Arguments) { close(wait) }).
			Return(nil).
			Once()
		err := conn.WriteMessage(websocket.BinaryMessage, wsBytes)
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
				Proto:   ws.ProtoTypeShell,
				MsgType: shell.MessageTypeSpawnShell,
			},
		})
		// terminal data
		assertSend(t, ws.ProtoMsg{
			Header: ws.ProtoHdr{
				Proto:   ws.ProtoTypeShell,
				MsgType: shell.MessageTypeShellCommand,
			},
			Body: []byte("echo YAY!"),
		})
		// stop the terminal
		assertSend(t, ws.ProtoMsg{
			Header: ws.ProtoHdr{
				Proto:   ws.ProtoTypeShell,
				MsgType: shell.MessageTypeStopShell,
			},
		})
	})

	t.Run("recv from device", func(t *testing.T) {
		// test receiving a message "from management"
		msg := ws.ProtoMsg{
			Header: ws.ProtoHdr{
				Proto:     ws.ProtoTypeShell,
				MsgType:   "cmd",
				SessionID: sessionID,
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
		closed := make(chan struct{})
		streamConn.On("Close", contextMatcher).
			Run(func(args mock.Arguments) { close(closed) }).
			Return(nil)
		err := conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "test"),
			time.Now().Add(time.Second*10),
		)
		assert.NoError(t, err)
		select {
		case <-closed:
		case <-time.After(time.Second):
			t.Error("error waiting for session to close")
		}
		assert.NoError(t, conn.Close())
	})
}

func TestManagementPlayback(t *testing.T) {
	testCases := []struct {
		Name            string
		SessionID       string
		Identity        *identity.Identity
		SleepIntervalMs string
		NoUpgrade       bool
	}{
		{
			Name:      "ok",
			SessionID: "session_id",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
				Plan:    "professional",
			},
		},
		{
			Name:      "ok with sleep interval",
			SessionID: "session_id",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
				Plan:    "professional",
			},
			SleepIntervalMs: "200",
		},
		{
			Name:      "internal error no upgrade",
			SessionID: "session_id",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
				Plan:    "professional",
			},
			SleepIntervalMs: "200",
			NoUpgrade:       true,
		},
		{
			Name:      "bad request no auth",
			SessionID: "session_id",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			app := &app_mocks.App{}
			defer app.AssertExpectations(t)
			natsClient := nats_mocks.NewClient(t)
			router, _ := NewRouter(app, natsClient, nil)

			headers := http.Header{}
			if tc.Identity != nil {
				headers.Set(headerAuthorization, "Bearer "+GenerateJWT(*tc.Identity))
			}

			if tc.Identity != nil && !tc.NoUpgrade {
				app.On("GetSessionRecording",
					mock.MatchedBy(func(_ context.Context) bool {
						return true
					}),
					tc.SessionID,
					mock.AnythingOfType("*app.PipeWriter"),
				).Return(nil)
			}

			s := httptest.NewServer(router)
			defer s.Close()

			if tc.NoUpgrade {
				url := s.URL + strings.Replace(
					APIURLManagementPlayback, ":sessionId",
					tc.SessionID, 1,
				)
				req, err := http.NewRequest(http.MethodGet, url, nil)
				req.Header.Set(headerAuthorization, "Bearer "+GenerateJWT(*tc.Identity))
				assert.NotNil(t, req)
				assert.NoError(t, err)
				tr := &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}

				client := &http.Client{
					Transport: tr,
				}
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				rsp, err := client.Do(req.WithContext(ctx))
				assert.Equal(t, rsp.StatusCode, http.StatusBadRequest)
				return
			}

			url := "ws" + strings.TrimPrefix(s.URL, "http")
			url = url + strings.Replace(
				APIURLManagementPlayback, ":sessionId",
				tc.SessionID, 1,
			)
			if len(tc.SleepIntervalMs) > 0 {
				url += "?" + PlaybackSleepIntervalMsField + "=" + tc.SleepIntervalMs
			}
			conn, _, err := websocket.DefaultDialer.Dial(url, headers)
			if tc.Identity == nil {
				assert.EqualError(t, err, "websocket: bad handshake")
				return
			} else {
				assert.NoError(t, err)
			}

			pingReceived := make(chan struct{}, 1)
			conn.SetPingHandler(func(message string) error {
				pingReceived <- struct{}{}
				return conn.WriteControl(
					websocket.PongMessage,
					[]byte{},
					time.Now().Add(writeWait),
				)
			})
			pongReceived := make(chan struct{}, 1)
			conn.SetPongHandler(func(message string) error {
				pongReceived <- struct{}{}
				return nil
			})

			// close the websocket
			conn.Close()

			// wait 100ms to let the websocket fully shutdown on the server
			time.Sleep(100 * time.Millisecond)
		})
	}
}

func TestManagementConnectFailures(t *testing.T) {
	testCases := []struct {
		Name                       string
		DeviceID                   string
		SessionID                  string
		OnConnect                  func(t *testing.T) (stream.Conn, error)
		PrepareUserSessionErr      error
		Authorization              string
		Identity                   identity.Identity
		RemoteTerminalAllowedError error
		RemoteTerminalAllowed      bool
		HTTPStatus                 int
		HTTPError                  error
	}{
		{
			Name:      "ko, unable to upgrade",
			SessionID: "1",
			Identity: identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Authorization: "Bearer " + GenerateJWT(identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			}),
			OnConnect: func(t *testing.T) (stream.Conn, error) {
				conn := stream_mocks.NewConn(t)
				conn.On("Close", contextMatcher).Return(nil)
				return conn, nil
			},
			HTTPStatus: http.StatusBadRequest,
		},
		{
			Name:                  "ko, session preparation failure",
			SessionID:             "1",
			PrepareUserSessionErr: errors.New("Error"),
			Identity: identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Authorization: "Bearer " + GenerateJWT(identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			}),
			OnConnect: func(t *testing.T) (stream.Conn, error) {
				conn := stream_mocks.NewConn(t)
				conn.On("Close", contextMatcher).Return(nil)
				return conn, nil
			},
			HTTPStatus: http.StatusInternalServerError,
		},
		{
			Name:                  "ko, device not found",
			SessionID:             "1",
			PrepareUserSessionErr: app.ErrDeviceNotFound,
			Identity: identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Authorization: "Bearer " + GenerateJWT(identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			}),
			OnConnect: func(t *testing.T) (stream.Conn, error) {
				return nil, stream.ErrConnectionRefused
			},
			HTTPStatus: http.StatusConflict,
		},
		{
			Name:       "ko, missing authorization header",
			HTTPStatus: http.StatusUnauthorized,
			HTTPError:  errors.New("Authorization not present in header"),
		},
		{
			Name: "ko, malformed authorization header",
			Identity: identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},
			Authorization: "malformed",
			HTTPStatus:    http.StatusUnauthorized,
			HTTPError:     errors.New("malformed Authorization header"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			app := &app_mocks.App{}
			if tc.SessionID != "" {
				app.On("PrepareUserSession",
					mock.MatchedBy(func(_ context.Context) bool {
						return true
					}),
					mock.MatchedBy(func(sess *model.Session) bool {
						sess.ID = tc.SessionID
						return true
					}),
				).Return(tc.PrepareUserSessionErr)
				if tc.PrepareUserSessionErr == nil {
					app.On("FreeUserSession",
						mock.MatchedBy(func(_ context.Context) bool {
							return true
						}),
						tc.SessionID,
						mock.AnythingOfType("[]string"),
					).Return(nil)
				}
			}

			natsClient := nats_mocks.NewClient(t)
			if tc.OnConnect != nil {
				conn, err := tc.OnConnect(t)
				natsClient.On("Connect",
					contextMatcher,
					mock.MatchedBy(func(s string) bool {
						return strings.HasPrefix(s, tc.Identity.Tenant)
					}),
					tc.DeviceID).
					Return(conn, err).
					Once()
			}
			router, _ := NewRouter(app, natsClient, nil)
			url := strings.Replace(APIURLManagementDeviceConnect, ":deviceId", tc.DeviceID, 1)
			req, err := http.NewRequest("GET", "http://localhost"+url, nil)
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

func TestManagementSessionLimit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	identity := identity.Identity{
		Subject: "00000000-0000-0000-0000-000000000000",
		Tenant:  "000000000000000000000000",
		IsUser:  true,
		Plan:    "professional",
	}
	devid := "1"

	headers := http.Header{}
	headers.Set(headerAuthorization, "Bearer "+GenerateJWT(identity))

	var sessionID string
	mapp := app_mocks.NewApp(t)
	mapp.On("PrepareUserSession",
		mock.MatchedBy(func(_ context.Context) bool {
			return true
		}),
		mock.MatchedBy(func(sess *model.Session) bool {
			sessionID = sess.ID
			return true
		}),
	).
		Run(func(args mock.Arguments) {
			mapp.On("FreeUserSession",
				mock.MatchedBy(func(_ context.Context) bool {
					return true
				}),
				sessionID,
				mock.AnythingOfType("[]string"),
			).Return(nil).
				On("GetControlRecorder",
					sessionID,
				).Return(discardRecorder{}).
				On("GetRecorder",
					sessionID,
				).Return(discardRecorder{})
		}).
		Return(nil)

	natsClient := nats_mocks.NewClient(t)
	router, _ := NewRouter(mapp, natsClient, nil)
	s := httptest.NewServer(router)
	defer s.Close()

	streamRecv := make(chan []byte, 10)
	streamConn := setupConn(t, streamRecv, identity.Tenant+":foobar", identity.Subject)
	streamConn.On("Close", contextMatcher).Return(nil).Maybe()
	natsClient.On("Connect", mock.Anything, mock.Anything, mock.Anything).
		Return(streamConn, nil)

	url := "ws" + strings.TrimPrefix(s.URL, "http")
	url = url + strings.Replace(
		APIURLManagementDeviceConnect, ":deviceId",
		devid, 1,
	)
	conn, _, err := websocket.DefaultDialer.Dial(url, headers)
	require.NoError(t, err)

	// MessageSizeLimit = 8 * 1024 * 1024
	// chunk spamming every 1msec should saturate session in ~1sec
	chunkSize := 8 * 1024
	buf := make([]byte, chunkSize)

	msg := ws.ProtoMsg{
		Header: ws.ProtoHdr{
			Proto:     ws.ProtoTypeShell,
			MsgType:   shell.MessageTypeShellCommand,
			SessionID: sessionID,
			Properties: map[string]interface{}{
				"status": shell.NormalMessage,
			},
		},
		Body: buf,
	}
	b, _ := msgpack.Marshal(msg)

	// device just spams us with chunks, as if we e.g. started 'top'
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				streamRecv <- b
			}
		}
	}()

	// receive shell data on user end
	// expect limit will be exceed at the app limit (error shell message)
	go func() {
		readBytes := 0
	Loop:
		for {
			select {
			case <-ctx.Done():
				return
			default:
				_, data, err := conn.ReadMessage()

				if readBytes < app.MessageSizeLimit {
					require.NoError(t, err)
					var rMsg ws.ProtoMsg
					err = msgpack.Unmarshal(data, &rMsg)
					assert.NoError(t, err)
					assert.Equal(t, ws.ProtoTypeShell, rMsg.Header.Proto)
					assert.Equal(t, shell.MessageTypeShellCommand, rMsg.Header.MsgType)
					assert.Equal(t, sessionID, rMsg.Header.SessionID)
					assert.Equal(t, int8(shell.NormalMessage), rMsg.Header.Properties["status"])

					readBytes += len(rMsg.Body)
				} else {
					assert.Error(t, err)
					break Loop
				}
			}
		}
	}()

	done := make(chan struct{})
	streamConn.On("Send", contextMatcher, mock.MatchedBy(func(b []byte) bool {
		var rMsg ws.ProtoMsg
		err = msgpack.Unmarshal(b, &rMsg)
		assert.NoError(t, err)
		assert.Equal(t, ws.ProtoTypeShell, rMsg.Header.Proto)
		assert.Equal(t, shell.MessageTypeStopShell, rMsg.Header.MsgType)
		assert.Equal(t, sessionID, rMsg.Header.SessionID)
		assert.Equal(t, identity.Subject, rMsg.Header.Properties["user_id"])
		assert.Equal(t, int8(shell.ErrorMessage), rMsg.Header.Properties["status"])
		assert.Equal(t, "session byte limit exceeded", string(rMsg.Body))
		return true
	})).
		Run(func(args mock.Arguments) { close(done) }).
		Return(nil)
	select {
	case <-done:
	case <-time.After(time.Second * 5):
		assert.Fail(t,
			"api did not forward shell_stop to nats device subject",
		)
		break
	}
}

func TestManagementCheckUpdate(t *testing.T) {
	testCases := []struct {
		Name     string
		DeviceID string
		Identity *identity.Identity

		OnConnect func(*testing.T, context.Context, string, string) (stream.Conn, error)

		HTTPStatus int
	}{
		{
			Name:     "ok",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},

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
			Name:     "ko, missing auth",
			DeviceID: "1234567890",

			HTTPStatus: 401,
		},
		{
			Name:     "ko, other error",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},

			OnConnect: func(*testing.T, context.Context, string, string) (stream.Conn, error) {
				return nil, fmt.Errorf("other error")
			},

			HTTPStatus: 500,
		},
		{
			Name:     "ko, device not connected",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},

			OnConnect: func(t *testing.T, ctx context.Context, s1, s2 string) (stream.Conn, error) {
				return nil, stream.ErrConnectionRefused
			},

			HTTPStatus: http.StatusConflict,
		},
		{
			Name:     "ko, publish error",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},

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
					Return(fmt.Errorf("internal error"))
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

			url := strings.Replace(APIURLManagementDeviceSendInventory, ":deviceId", tc.DeviceID, 1)
			req, err := http.NewRequest("POST", "http://localhost"+url, nil)
			if tc.Identity != nil {
				jwt := GenerateJWT(*tc.Identity)
				req.Header.Set(headerAuthorization, "Bearer "+jwt)
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

func TestManagementSendInventory(t *testing.T) {
	testCases := []struct {
		Name     string
		DeviceID string
		Identity *identity.Identity

		OnConnect func(*testing.T, context.Context, string, string) (stream.Conn, error)

		HTTPStatus int
	}{
		{
			Name:     "ok",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},

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
			Name:     "ko, missing auth",
			DeviceID: "1234567890",

			HTTPStatus: 401,
		},
		{
			Name:     "ko, other error",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},

			OnConnect: func(*testing.T, context.Context, string, string) (stream.Conn, error) {
				return nil, fmt.Errorf("other error")
			},

			HTTPStatus: 500,
		},
		{
			Name:     "ko, device not connected",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},

			OnConnect: func(t *testing.T, ctx context.Context, s1, s2 string) (stream.Conn, error) {
				return nil, stream.ErrConnectionRefused
			},

			HTTPStatus: http.StatusConflict,
		},
		{
			Name:     "ko, publish error",
			DeviceID: "1234567890",
			Identity: &identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
			},

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
					Return(fmt.Errorf("internal error"))
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

			url := strings.Replace(APIURLManagementDeviceSendInventory, ":deviceId", tc.DeviceID, 1)
			req, err := http.NewRequest("POST", "http://localhost"+url, nil)
			if tc.Identity != nil {
				jwt := GenerateJWT(*tc.Identity)
				req.Header.Set(headerAuthorization, "Bearer "+jwt)
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
