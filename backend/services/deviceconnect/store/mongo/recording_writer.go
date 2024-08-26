// Copyright 2021 Northern.tech AS
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

package mongo

import (
	"io"

	"github.com/vmihailenco/msgpack/v5"

	"github.com/mendersoftware/mender-server/pkg/ws"
	"github.com/mendersoftware/mender-server/pkg/ws/shell"
)

type RecordingWriter struct {
	w         io.Writer
	sessionID string
}

func NewRecordingWriter(sessionID string, w io.Writer) *RecordingWriter {
	return &RecordingWriter{
		w:         w,
		sessionID: sessionID,
	}
}

func sendRecordingMessage(data []byte, sessionID string, w io.Writer) (int, error) {
	msg := ws.ProtoMsg{
		Header: ws.ProtoHdr{
			Proto:      ws.ProtoTypeShell,
			MsgType:    shell.MessageTypeShellCommand,
			SessionID:  sessionID,
			Properties: nil,
		},
		Body: data,
	}
	messagePacked, err := msgpack.Marshal(&msg)
	if err != nil {
		return 0, err
	} else {
		n, e := w.Write(messagePacked)
		if e != nil {
			return n, e
		} else {
			return len(data), nil
		}
	}
}

func (r *RecordingWriter) Write(d []byte) (n int, err error) {
	n, err = sendRecordingMessage(d, r.sessionID, r.w)
	if n == 0 || err != nil {
		return 0, io.ErrShortWrite
	}
	return n, nil
}
