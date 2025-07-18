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
	"github.com/nats-io/nats.go"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/ws"
	"github.com/mendersoftware/mender-server/pkg/ws/shell"
	app_mocks "github.com/mendersoftware/mender-server/services/deviceconnect/app/mocks"
	"github.com/mendersoftware/mender-server/services/deviceconnect/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

	natsClient := NewNATSTestClient(t)
	router, _ := NewRouter(app, natsClient, nil)
	s := httptest.NewServer(router)
	defer s.Close()

	url := "ws" + strings.TrimPrefix(s.URL, "http")

	headers := http.Header{}
	headers.Set(
		headerAuthorization,
		"Bearer "+GenerateJWT(Identity),
	)

	conn, _, err := websocket.DefaultDialer.Dial(url+APIURLDevicesConnect, headers)
	assert.NoError(t, err)

	pingReceived := make(chan struct{}, 10)
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

	dataReceived := make(chan []byte, 2)
	go func() {
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				break
			}
			dataReceived <- data
		}
	}()

	// test receiving a message "from management"
	msg := ws.ProtoMsg{
		Header: ws.ProtoHdr{
			Proto:     ws.ProtoTypeShell,
			MsgType:   "cmd",
			SessionID: "foobar",
		},
	}
	b, _ := msgpack.Marshal(msg)
	natsClient.Publish(model.GetDeviceSubject(
		Identity.Tenant,
		Identity.Subject),
		b,
	)
	select {
	case rMsg := <-dataReceived:
		assert.Equal(t, b, rMsg)

	case <-time.After(time.Second):
		assert.Fail(t,
			"timeout waiting for message from management",
		)
	}

	// test responding to message from management
	rChan := make(chan *nats.Msg, 1)
	_, err = natsClient.ChanSubscribe(
		model.GetSessionSubject(Identity.Tenant, "foobar"),
		rChan,
	)
	assert.NoError(t, err)
	err = conn.WriteMessage(websocket.BinaryMessage, b)
	assert.NoError(t, err)

	select {
	case <-time.After(time.Second):
		assert.Fail(t,
			"timeout waiting for message to propagate",
		)
	case rMsg := <-rChan:
		assert.Equal(t, b, rMsg.Data)
	}

	// check that ping and pong works as expected
	err = conn.WriteControl(
		websocket.PingMessage,
		[]byte("1"),
		time.Now().Add(time.Second),
	)
	assert.NoError(t, err)
	select {
	case <-pongReceived:
	case <-time.After(pongWait * 2):
		assert.Fail(t, "did not receive pong within pongWait")
	}

	// check that pings are received
	select {
	case <-pingReceived:
	case <-time.After(pongWait * 2):
		assert.Fail(t, "did not receive ping within pongWait")
	}

	// start a new terminal
	msg = ws.ProtoMsg{
		Header: ws.ProtoHdr{
			Proto:     ws.ProtoTypeShell,
			MsgType:   shell.MessageTypeSpawnShell,
			SessionID: "foobar",
		},
	}
	b, _ = msgpack.Marshal(msg)
	err = conn.WriteMessage(websocket.BinaryMessage, b)
	assert.NoError(t, err)

	select {
	case <-time.After(time.Second):
		assert.Fail(t,
			"timeout waiting for message to propagate",
		)
	case rMsg := <-rChan:
		assert.Equal(t, b, rMsg.Data)
	}

	// stop the terminal
	msg = ws.ProtoMsg{
		Header: ws.ProtoHdr{
			Proto:     ws.ProtoTypeShell,
			MsgType:   shell.MessageTypeStopShell,
			SessionID: "foobar",
		},
	}
	b, _ = msgpack.Marshal(msg)
	err = conn.WriteMessage(websocket.BinaryMessage, b)
	assert.NoError(t, err)

	select {
	case <-time.After(time.Second):
		assert.Fail(t,
			"timeout waiting for message to propagate",
		)
	case rMsg := <-rChan:
		assert.Equal(t, b, rMsg.Data)
	}

	// start again a new terminal
	msg = ws.ProtoMsg{
		Header: ws.ProtoHdr{
			Proto:     ws.ProtoTypeShell,
			MsgType:   shell.MessageTypeSpawnShell,
			SessionID: "foobar",
		},
	}
	b, _ = msgpack.Marshal(msg)
	err = conn.WriteMessage(websocket.BinaryMessage, b)
	assert.NoError(t, err)

	select {
	case <-time.After(time.Second):
		assert.Fail(t,
			"timeout waiting for message to propagate",
		)
	case rMsg := <-rChan:
		assert.Equal(t, b, rMsg.Data)
	}

	// send wrong message (session ID empty), this shuts down the connection
	msg = ws.ProtoMsg{
		Header: ws.ProtoHdr{
			Proto:     ws.ProtoTypeShell,
			MsgType:   shell.MessageTypeSpawnShell,
			SessionID: "",
		},
	}
	b, _ = msgpack.Marshal(msg)
	err = conn.WriteMessage(websocket.BinaryMessage, b)
	assert.NoError(t, err)

	// we receive the stop message
	select {
	case <-time.After(time.Second):
		assert.Fail(t,
			"timeout waiting for message to propagate",
		)
	case rMsg := <-rChan:
		msg := &ws.ProtoMsg{}
		_ = msgpack.Unmarshal(rMsg.Data, msg)
		assert.Equal(t, ws.ProtoTypeShell, msg.Header.Proto)
		assert.Equal(t, shell.MessageTypeStopShell, msg.Header.MsgType)
	}

	// close the websocket
	conn.Close()

	// Restart a connection to check error handling
	app.On("SetDeviceConnected",
		mock.MatchedBy(func(_ context.Context) bool {
			return true
		}),
		Identity.Tenant,
		Identity.Subject,
	).Return(int64(1), nil)

	app.On("SetDeviceDisconnected",
		mock.MatchedBy(func(_ context.Context) bool {
			return true
		}),
		Identity.Tenant,
		Identity.Subject,
		int64(1),
	).Return(nil)

	conn, _, err = websocket.DefaultDialer.Dial(url+APIURLDevicesConnect, headers)
	assert.NoError(t, err)
	err = conn.WriteMessage(websocket.BinaryMessage, []byte("bogus"))
	assert.NoError(t, err)
	_, _, err = conn.ReadMessage()
	assert.NotNil(t, err)
	assert.True(t, websocket.IsCloseError(err,
		websocket.CloseInternalServerErr), err.Error(),
	)
	conn.Close()

	app.AssertExpectations(t)
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
			var natsClient *nats.Conn
			if tc.WithNATS {
				natsClient = NewNATSTestClient(t)
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
