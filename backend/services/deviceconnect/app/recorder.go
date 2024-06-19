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

package app

import (
	"bytes"
	"compress/gzip"
	"context"

	"github.com/mendersoftware/mender-server/services/deviceconnect/store"
)

type Recorder struct {
	sessionID  string
	store      store.DataStore
	ctx        context.Context
	gzipWriter *gzip.Writer
	gzipBuffer bytes.Buffer
}

const (
	//It is used as a length of a memory region in bytes that is used to buffer
	//the session recording. 4455 comes from the estimated typical terminal size in
	//bytes (height=135 width=33) multiplied by 4 bytes of of terminal codes to get
	//an estimate of a typical screen frame size in bytes. So we round to 4kB
	RecorderBufferSize = 4 * 4096
)

func NewRecorder(ctx context.Context, sessionID string, store store.DataStore) *Recorder {
	buffer := bytes.Buffer{}
	return &Recorder{
		ctx:        ctx,
		sessionID:  sessionID,
		store:      store,
		gzipBuffer: buffer,
		gzipWriter: gzip.NewWriter(&buffer),
	}
}

func (r *Recorder) Write(d []byte) (n int, err error) {
	if len(r.sessionID) < 1 {
		return 0, nil
	}

	r.gzipBuffer.Reset()
	r.gzipWriter.Reset(&r.gzipBuffer)
	_, err = r.gzipWriter.Write(d)
	if err != nil {
		return -1, err
	}

	r.gzipWriter.Flush()
	r.gzipWriter.Close()
	output := r.gzipBuffer.Bytes()
	n, err = r.gzipBuffer.Read(output)
	if err != nil {
		return -1, err
	}

	err = r.store.InsertSessionRecording(r.ctx, r.sessionID, output[:n])

	if err != nil {
		return -1, err
	} else {
		return len(d), nil
	}
}
