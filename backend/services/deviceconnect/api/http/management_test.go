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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/ws"
	"github.com/mendersoftware/mender-server/pkg/ws/shell"

	"github.com/mendersoftware/mender-server/services/deviceconnect/app"
	app_mocks "github.com/mendersoftware/mender-server/services/deviceconnect/app/mocks"
	nats_mocks "github.com/mendersoftware/mender-server/services/deviceconnect/client/nats/mocks"
	"github.com/mendersoftware/mender-server/services/deviceconnect/model"
)

var natsPort int32 = 14420

func NewNATSTestClient(t *testing.T) *nats.Conn {
	port := atomic.AddInt32(&natsPort, 1)
	opts := &server.Options{
		Port: int(port),
	}
	srv, err := server.NewServer(opts)
	if err != nil {
		panic(err)
	}
	go srv.Start()
	t.Cleanup(srv.Shutdown)

	// Spinlock until go routine is listening
	for i := 0; srv.Addr() == nil && i < 1000; i++ {
		time.Sleep(time.Millisecond)
	}
	if srv.Addr() == nil {
		panic("failed to setup NATS test server")
	}
	client, err := nats.Connect("nats://" + srv.Addr().String())
	if err != nil {
		panic(err)
	}
	return client
}

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
	prevPongWait := pongWait
	prevWriteWait := writeWait
	defer func() {
		pongWait = prevPongWait
		writeWait = prevWriteWait
	}()
	pongWait = time.Second
	writeWait = time.Second
	testCases := []struct {
		Name                       string
		DeviceID                   string
		SessionID                  string
		Identity                   identity.Identity
		RemoteTerminalAllowedError error
		RemoteTerminalAllowed      bool
	}{
		{
			Name:      "ok",
			DeviceID:  "1234567890",
			SessionID: "session_id",
			Identity: identity.Identity{
				Subject: "00000000-0000-0000-0000-000000000000",
				Tenant:  "000000000000000000000000",
				IsUser:  true,
				Plan:    "professional",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			app := &app_mocks.App{}
			defer app.AssertExpectations(t)
			natsClient := NewNATSTestClient(t)
			router, _ := NewRouter(app, natsClient, nil)

			headers := http.Header{}
			headers.Set(headerAuthorization, "Bearer "+GenerateJWT(tc.Identity))

			app.On("PrepareUserSession",
				mock.MatchedBy(func(_ context.Context) bool {
					return true
				}),
				mock.MatchedBy(func(sess *model.Session) bool {
					sess.ID = tc.SessionID
					return true
				}),
			).Return(nil)
			app.On("LogUserSession",
				mock.MatchedBy(func(_ context.Context) bool {
					return true
				}),
				mock.MatchedBy(func(sess *model.Session) bool {
					sess.ID = tc.SessionID
					return true
				}),
				mock.AnythingOfType("string"),
			).Return(nil)
			app.On("FreeUserSession",
				mock.MatchedBy(func(_ context.Context) bool {
					return true
				}),
				tc.SessionID,
				mock.AnythingOfType("[]string"),
			).Return(nil)
			app.On("GetControlRecorder",
				mock.MatchedBy(func(_ context.Context) bool {
					return true
				}),
				tc.SessionID,
			).Return(nil)
			app.On("GetRecorder",
				mock.MatchedBy(func(_ context.Context) bool {
					return true
				}),
				tc.SessionID,
			).Return(nil)

			s := httptest.NewServer(router)
			defer s.Close()

			url := "ws" + strings.TrimPrefix(s.URL, "http")
			url = url + strings.Replace(
				APIURLManagementDeviceConnect, ":deviceId",
				tc.DeviceID, 1,
			)
			conn, _, err := websocket.DefaultDialer.Dial(url, headers)
			assert.NoError(t, err)

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

			receivedMsg := make(chan []byte, 1)
			go func() {
				for {
					_, data, err := conn.ReadMessage()
					if err != nil {
						break
					}
					receivedMsg <- data
				}
			}()
			natsChan := make(chan *nats.Msg, 2)
			sub, _ := natsClient.ChanSubscribe(
				model.GetDeviceSubject(
					tc.Identity.Tenant,
					tc.DeviceID,
				), natsChan,
			)
			defer sub.Unsubscribe()
			msg := ws.ProtoMsg{
				Header: ws.ProtoHdr{
					Proto:   ws.ProtoTypeShell,
					MsgType: "hello",
				},
			}
			b, _ := msgpack.Marshal(msg)
			err = conn.WriteMessage(websocket.BinaryMessage, b)
			assert.NoError(t, err)
			select {
			case natsMsg := <-natsChan:
				var rMsg ws.ProtoMsg
				err = msgpack.Unmarshal(natsMsg.Data, &rMsg)
				if assert.NoError(t, err) {
					msg.Header.SessionID = tc.SessionID
					msg.Header.Properties = map[string]interface{}{
						"user_id": tc.Identity.Subject,
					}
					assert.Equal(t, msg, rMsg)
				}
			case <-time.After(time.Second * 5):
				assert.Fail(t,
					"api did not forward message to message bus",
				)
			}
			err = conn.WriteControl(
				websocket.PingMessage,
				[]byte("1"),
				time.Now().Add(time.Second),
			)
			assert.NoError(t, err)

			msg.Header.SessionID = tc.SessionID
			b, _ = msgpack.Marshal(msg)
			err = natsClient.Publish(
				model.GetSessionSubject(tc.Identity.Tenant, tc.SessionID),
				b,
			)
			assert.NoError(t, err)
			select {
			case p := <-receivedMsg:
				assert.Equal(t, b, p)
			case <-time.After(time.Second * 5):
				assert.Fail(t, "timed out waiting for message from device")
			}

			// check that ping and pong works as expected
			select {
			case <-pingReceived:
			case <-time.After(pongWait):
				assert.Fail(t, "did not receive ping within pongWait")
			}

			select {
			case <-pongReceived:
			case <-time.After(pongWait):
				assert.Fail(t, "did not receive pong within pongWait")
			}

			// start the remote terminal
			msg = ws.ProtoMsg{
				Header: ws.ProtoHdr{
					Proto:     ws.ProtoTypeShell,
					MsgType:   shell.MessageTypeSpawnShell,
					SessionID: "foobar",
				},
			}
			b, _ = msgpack.Marshal(msg)
			natsClient.Publish(model.GetDeviceSubject(
				tc.Identity.Tenant,
				tc.Identity.Subject),
				b,
			)
			_ = conn.WriteMessage(websocket.BinaryMessage, b)

			select {
			case msg := <-natsChan:
				var stopMsg ws.ProtoMsg
				err := msgpack.Unmarshal(msg.Data, &stopMsg)
				if assert.NoError(t, err) {
					assert.Equal(t,
						ws.ProtoTypeShell,
						stopMsg.Header.Proto,
					)
					assert.Equal(t,
						shell.MessageTypeSpawnShell,
						stopMsg.Header.MsgType,
					)
				}

			case <-time.After(time.Second * 5):
				assert.Fail(t,
					"timeout waiting for stop message on nats channel",
				)
			}

			// stop the remote terminal
			msg = ws.ProtoMsg{
				Header: ws.ProtoHdr{
					Proto:     ws.ProtoTypeShell,
					MsgType:   shell.MessageTypeStopShell,
					SessionID: "foobar",
				},
			}
			b, _ = msgpack.Marshal(msg)
			natsClient.Publish(model.GetDeviceSubject(
				tc.Identity.Tenant,
				tc.Identity.Subject),
				b,
			)
			_ = conn.WriteMessage(websocket.BinaryMessage, b)

			select {
			case msg := <-natsChan:
				var stopMsg ws.ProtoMsg
				err := msgpack.Unmarshal(msg.Data, &stopMsg)
				if assert.NoError(t, err) {
					assert.Equal(t,
						ws.ProtoTypeShell,
						stopMsg.Header.Proto,
					)
					assert.Equal(t,
						shell.MessageTypeStopShell,
						stopMsg.Header.MsgType,
					)
				}

			case <-time.After(time.Second * 5):
				assert.Fail(t,
					"timeout waiting for stop message on nats channel",
				)
			}

			// start the remote terminal again
			msg = ws.ProtoMsg{
				Header: ws.ProtoHdr{
					Proto:     ws.ProtoTypeShell,
					MsgType:   shell.MessageTypeSpawnShell,
					SessionID: "foobar",
				},
			}
			b, _ = msgpack.Marshal(msg)
			natsClient.Publish(model.GetDeviceSubject(
				tc.Identity.Tenant,
				tc.Identity.Subject),
				b,
			)
			_ = conn.WriteMessage(websocket.BinaryMessage, b)

			select {
			case msg := <-natsChan:
				var stopMsg ws.ProtoMsg
				err := msgpack.Unmarshal(msg.Data, &stopMsg)
				if assert.NoError(t, err) {
					assert.Equal(t,
						ws.ProtoTypeShell,
						stopMsg.Header.Proto,
					)
					assert.Equal(t,
						shell.MessageTypeSpawnShell,
						stopMsg.Header.MsgType,
					)
				}

			case <-time.After(time.Second * 5):
				assert.Fail(t,
					"timeout waiting for stop message on nats channel",
				)
			}

			// close the websocket
			conn.Close()

			select {
			case msg := <-natsChan:
				var stopMsg ws.ProtoMsg
				err := msgpack.Unmarshal(msg.Data, &stopMsg)
				if assert.NoError(t, err) {
					assert.Equal(t,
						ws.ProtoTypeShell,
						stopMsg.Header.Proto,
					)
					assert.Equal(t,
						shell.MessageTypeStopShell,
						stopMsg.Header.MsgType,
					)
				}

			case <-time.After(time.Second * 5):
				assert.Fail(t,
					"timeout waiting for stop message on nats channel",
				)
			}

			conn, _, err = websocket.DefaultDialer.Dial(url, headers)
			app.On("PrepareUserSession",
				mock.MatchedBy(func(_ context.Context) bool {
					return true
				}),
				mock.MatchedBy(func(sess *model.Session) bool {
					sess.ID = tc.SessionID
					return true
				}),
			).Return(nil)
			app.On("FreeUserSession",
				mock.MatchedBy(func(_ context.Context) bool {
					return true
				}),
				tc.SessionID,
				mock.AnythingOfType("[]string"),
			).Return(nil)
			err = conn.WriteMessage(websocket.BinaryMessage, []byte("bogus"))
			assert.NoError(t, err)
			_, _, err = conn.ReadMessage()
			assert.Error(t, err)
			assert.True(t, websocket.IsCloseError(err,
				websocket.CloseInternalServerErr),
			)
			conn.Close()

			app.AssertExpectations(t)
		})
	}
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
			natsClient := NewNATSTestClient(t)
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
					mock.AnythingOfType("*app.Playback"),
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
			HTTPStatus: http.StatusNotFound,
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

			natsClient := NewNATSTestClient(t)
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

	mapp := &app_mocks.App{}
	natsClient := NewNATSTestClient(t)
	router, _ := NewRouter(mapp, natsClient, nil)

	sid := "test_session_id"

	identity := identity.Identity{
		Subject: "00000000-0000-0000-0000-000000000000",
		Tenant:  "000000000000000000000000",
		IsUser:  true,
		Plan:    "professional",
	}
	devid := "1"

	headers := http.Header{}
	headers.Set(headerAuthorization, "Bearer "+GenerateJWT(identity))

	mapp.On("PrepareUserSession",
		mock.MatchedBy(func(_ context.Context) bool {
			return true
		}),
		mock.MatchedBy(func(sess *model.Session) bool {
			sess.ID = sid
			return true
		}),
	).Return(nil)
	mapp.On("FreeUserSession",
		mock.MatchedBy(func(_ context.Context) bool {
			return true
		}),
		sid,
		mock.AnythingOfType("[]string"),
	).Return(nil)
	mapp.On("GetRecorder",
		mock.MatchedBy(func(_ context.Context) bool {
			return true
		}),
		sid,
	).Return(ioutil.Discard)
	mapp.On("GetControlRecorder",
		mock.MatchedBy(func(_ context.Context) bool {
			return true
		}),
		sid,
	).Return(nil)

	s := httptest.NewServer(router)
	defer s.Close()

	natsChan := make(chan *nats.Msg)
	sub, _ := natsClient.ChanSubscribe(
		model.GetDeviceSubject(
			identity.Tenant,
			devid,
		), natsChan,
	)
	defer sub.Unsubscribe()

	url := "ws" + strings.TrimPrefix(s.URL, "http")
	url = url + strings.Replace(
		APIURLManagementDeviceConnect, ":deviceId",
		devid, 1,
	)
	conn, _, err := websocket.DefaultDialer.Dial(url, headers)
	assert.NoError(t, err)

	// MessageSizeLimit = 8 * 1024 * 1024
	// chunk spamming every 1msec should saturate session in ~1sec
	chunkSize := 8 * 1024
	buf := make([]byte, chunkSize)

	msg := ws.ProtoMsg{
		Header: ws.ProtoHdr{
			Proto:     ws.ProtoTypeShell,
			MsgType:   shell.MessageTypeShellCommand,
			SessionID: sid,
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
				err = natsClient.Publish(
					model.GetSessionSubject(identity.Tenant, sid),
					b,
				)
				assert.NoError(t, err)
				time.Sleep(1 * time.Millisecond)
			}
		}
	}()

	// receive shell data on user end
	// expect limit will be exceed at the app limit (error shell message)
	go func() {
		readBytes := 0
		for {
			select {
			case <-ctx.Done():
				return
			default:
				_, data, err := conn.ReadMessage()
				assert.NoError(t, err)

				var rMsg ws.ProtoMsg
				err = msgpack.Unmarshal(data, &rMsg)
				assert.NoError(t, err)

				if readBytes < app.MessageSizeLimit {
					assert.Equal(t, ws.ProtoTypeShell, rMsg.Header.Proto)
					assert.Equal(t, shell.MessageTypeShellCommand, rMsg.Header.MsgType)
					assert.Equal(t, sid, rMsg.Header.SessionID)
					assert.Equal(t, int8(shell.NormalMessage), rMsg.Header.Properties["status"])

					readBytes += len(rMsg.Body)
				} else {
					assert.Equal(t, ws.ProtoTypeShell, rMsg.Header.Proto)
					assert.Equal(t, shell.MessageTypeShellCommand, rMsg.Header.MsgType)
					assert.Equal(t, sid, rMsg.Header.SessionID)
					assert.Equal(t, int8(shell.ErrorMessage), rMsg.Header.Properties["status"])
					assert.Equal(t, "session byte limit exceeded", string(rMsg.Body))
				}
			}
		}
	}()

	select {
	case natsMsg := <-natsChan:
		var rMsg ws.ProtoMsg
		err = msgpack.Unmarshal(natsMsg.Data, &rMsg)
		assert.NoError(t, err)
		assert.Equal(t, ws.ProtoTypeShell, rMsg.Header.Proto)
		assert.Equal(t, shell.MessageTypeStopShell, rMsg.Header.MsgType)
		assert.Equal(t, sid, rMsg.Header.SessionID)
		assert.Equal(t, identity.Subject, rMsg.Header.Properties["user_id"])
		assert.Equal(t, int8(shell.ErrorMessage), rMsg.Header.Properties["status"])
		assert.Equal(t, "session byte limit exceeded", string(rMsg.Body))
		break
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

		GetDevice      *model.Device
		GetDeviceError error

		PublishErr error

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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
			},

			HTTPStatus: http.StatusAccepted,
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
		{
			Name:     "ko, device not connected",
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
			Name:     "ko, publish error",
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

			PublishErr: errors.New("error"),

			HTTPStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			app := &app_mocks.App{}
			defer app.AssertExpectations(t)

			natsClient := &nats_mocks.Client{}
			defer natsClient.AssertExpectations(t)

			router, _ := NewRouter(app, natsClient, nil)
			s := httptest.NewServer(router)
			defer s.Close()

			url := strings.Replace(APIURLManagementDeviceCheckUpdate, ":deviceId", tc.DeviceID, 1)
			req, err := http.NewRequest("POST", "http://localhost"+url, nil)
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

				if tc.GetDeviceError == nil && tc.GetDevice != nil &&
					tc.GetDevice.Status == model.DeviceStatusConnected {
					natsClient.On("Publish",
						mock.AnythingOfType("string"),
						mock.AnythingOfType("[]uint8"),
					).Return(tc.PublishErr)
				}
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

		GetDevice      *model.Device
		GetDeviceError error

		PublishErr error

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

			GetDevice: &model.Device{
				ID:     "1234567890",
				Status: model.DeviceStatusConnected,
			},

			HTTPStatus: http.StatusAccepted,
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
		{
			Name:     "ko, device not connected",
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
			Name:     "ko, publish error",
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

			PublishErr: errors.New("error"),

			HTTPStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			app := &app_mocks.App{}
			defer app.AssertExpectations(t)

			natsClient := &nats_mocks.Client{}
			defer natsClient.AssertExpectations(t)

			router, _ := NewRouter(app, natsClient, nil)
			s := httptest.NewServer(router)
			defer s.Close()

			url := strings.Replace(APIURLManagementDeviceSendInventory, ":deviceId", tc.DeviceID, 1)
			req, err := http.NewRequest("POST", "http://localhost"+url, nil)
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

				if tc.GetDeviceError == nil && tc.GetDevice != nil &&
					tc.GetDevice.Status == model.DeviceStatusConnected {
					natsClient.On("Publish",
						mock.AnythingOfType("string"),
						mock.AnythingOfType("[]uint8"),
					).Return(tc.PublishErr)
				}
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
