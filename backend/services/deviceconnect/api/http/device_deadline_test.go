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
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	http_mocks "github.com/mendersoftware/mender-server/services/deviceconnect/api/http/mocks"
)

func TestHandleDeviceMessagesReadDeadlineError(t *testing.T) {
	errDeadline := errors.New("set read deadline failed")

	conn := http_mocks.NewWSConn(t)
	conn.On("ReadMessage").
		Return(websocket.BinaryMessage, []byte{0x00}, nil).Once()
	conn.On("SetReadDeadline", mock.AnythingOfType("time.Time")).
		Return(errDeadline).Once()

	errChan := make(chan error, 1)
	var h DeviceController
	h.handleDeviceMessages(context.Background(), conn, new(streamMap), errChan)

	select {
	case err := <-errChan:
		assert.ErrorIs(t, err, errDeadline)
	default:
		t.Fatal("expected SetReadDeadline error to be sent on errChan")
	}
}

func TestConnectWSWriterInitialReadDeadlineError(t *testing.T) {
	errDeadline := errors.New("set read deadline failed")

	conn := http_mocks.NewWSConn(t)
	conn.On("SetReadDeadline", mock.AnythingOfType("time.Time")).
		Return(errDeadline).Once()
	conn.On("WriteControl",
		mock.AnythingOfType("int"),
		mock.Anything,
		mock.AnythingOfType("time.Time"),
	).Return(nil).Once()
	conn.On("Close").Return(nil).Once()

	var h DeviceController
	err := h.connectWSWriter(context.Background(), conn, nil)

	assert.ErrorIs(t, err, errDeadline)
}

func TestPingHandlerReadDeadlineError(t *testing.T) {
	errDeadline := errors.New("set read deadline failed")

	conn := http_mocks.NewWSConn(t)
	conn.On("SetReadDeadline", mock.AnythingOfType("time.Time")).
		Return(errDeadline).Once()

	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	err := newPingHandler(conn, ticker, time.Hour)("ping")

	assert.ErrorIs(t, err, errDeadline)
	conn.AssertNotCalled(t, "WriteControl",
		mock.Anything, mock.Anything, mock.Anything)
}

func TestPingHandlerWritesPong(t *testing.T) {
	conn := http_mocks.NewWSConn(t)
	conn.On("SetReadDeadline", mock.AnythingOfType("time.Time")).
		Return(nil).Once()
	conn.On("WriteControl",
		websocket.PongMessage,
		[]byte("ping"),
		mock.AnythingOfType("time.Time"),
	).Return(nil).Once()

	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	assert.NoError(t, newPingHandler(conn, ticker, time.Hour)("ping"))
}

func TestPongHandlerReadDeadlineError(t *testing.T) {
	errDeadline := errors.New("set read deadline failed")

	conn := http_mocks.NewWSConn(t)
	conn.On("SetReadDeadline", mock.AnythingOfType("time.Time")).
		Return(errDeadline).Once()

	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	assert.ErrorIs(t, newPongHandler(conn, ticker, time.Hour)(""), errDeadline)
}
