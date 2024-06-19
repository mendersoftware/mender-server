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
	"sync"

	"github.com/mendersoftware/mender-server/services/deviceconnect/store"
)

type ControlRecorder struct {
	sessionID  string
	store      store.DataStore
	ctx        context.Context
	gzipWriter *gzip.Writer
	gzipBuffer bytes.Buffer
	mutex      sync.Mutex
}

const (
	//Assuming there will be a moderate number of control messages, so let's set ot 1kB
	ControlRecorderBufferSize = 1024
)

func NewControlRecorder(ctx context.Context,
	sessionID string,
	store store.DataStore) *ControlRecorder {
	buffer := bytes.Buffer{}
	return &ControlRecorder{
		ctx:        ctx,
		sessionID:  sessionID,
		store:      store,
		gzipBuffer: buffer,
		gzipWriter: gzip.NewWriter(&buffer),
		mutex:      sync.Mutex{},
	}
}

func (r *ControlRecorder) Write(d []byte) (n int, err error) {
	//safety precaution, there are two buffered writers using this writer
	r.mutex.Lock()
	defer r.mutex.Unlock()

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

	err = r.store.InsertControlRecording(r.ctx, r.sessionID, output[:n])

	if err != nil {
		return -1, err
	} else {
		return len(d), nil
	}
}
