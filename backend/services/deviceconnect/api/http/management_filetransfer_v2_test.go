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
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vmihailenco/msgpack/v5"

	stream_mocks "github.com/mendersoftware/mender-server/pkg/stream/mocks"
	"github.com/mendersoftware/mender-server/pkg/ws"
	wsft "github.com/mendersoftware/mender-server/pkg/ws/filetransfer"
)

func TestDownloadFileV2DeviceErrorCode(t *testing.T) {
	testCases := []struct {
		Name string

		DeviceErrorCode int

		HTTPStatus int
	}{
		{
			Name: "ko, error from device with status code",

			DeviceErrorCode: http.StatusForbidden,

			HTTPStatus: http.StatusForbidden,
		},
		{
			Name: "ko, error from device without status code",

			HTTPStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			conn := stream_mocks.NewConn(t)
			conn.On("Send", contextMatcher, mock.Anything).
				Return(nil).
				Once().
				On("Recv", contextMatcher).
				Return(func(context.Context) ([]byte, error) {
					b, _ := msgpack.Marshal(ws.Error{Error: "access denied", Code: tc.DeviceErrorCode})
					return msgpack.Marshal(&ws.ProtoMsg{
						Header: ws.ProtoHdr{
							Proto:   ws.ProtoTypeFileTransferV2,
							MsgType: wsft.MessageTypeError,
						},
						Body: b,
					})
				}).
				Once()

			var h ManagementController
			err := h.downloadFileV2(
				context.Background(), conn, io.Discard, "/path", "userID", "sessionID",
			)

			var statusErr *Error
			if assert.ErrorAs(t, err, &statusErr) {
				assert.Equal(t, tc.HTTPStatus, statusErr.statusCode)
			}
		})
	}
}
