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
	"time"

	"github.com/gorilla/websocket"
)

// handleShutdown is the common handler for handling server shutdown event for
// both management and devices API. The server will initiate a close handshake
// and wait for the reader to receive the peer close frame signaled from async
// ws reader go routine through errChan.
func handleShutdown(ctx context.Context, conn WSConn, errChan <-chan error) error {
	deadline := time.Now().Add(writeWait)
	err := conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseGoingAway, "server shutting down"),
		deadline,
	)
	if err == nil {
		ctx, cancel := context.WithDeadline(ctx, deadline)
		defer cancel()
		select {
		case err = <-errChan:
		case <-ctx.Done():
			err = ctx.Err()
		}
	}
	return err
}
