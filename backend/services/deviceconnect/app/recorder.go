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
	"bufio"
	"compress/gzip"
	"context"
	"fmt"

	"github.com/mendersoftware/mender-server/services/deviceconnect/store"
)

type Recorder interface {
	Record(ctx context.Context, data []byte) error
	Close(ctx context.Context) error
}

type recorder struct {
	ctx        context.Context
	recordMu   chan struct{}
	gzipWriter *gzip.Writer
	bufWriter  *bufio.Writer
}

const (
	//It is used as a length of a memory region in bytes that is used to buffer
	//the session recording. 4455 comes from the estimated typical terminal size in
	//bytes (height=135 width=33) multiplied by 4 bytes of of terminal codes to get
	//an estimate of a typical screen frame size in bytes. So we round to 4kB
	RecorderBufferSize = 4 * 4096

	//Assuming there will be a moderate number of control messages, so let's set ot 1kB
	ControlRecorderBufferSize = 1024
)

type writerFunc func([]byte) (int, error)

func (f writerFunc) Write(data []byte) (int, error) {
	return f(data)
}

func (rec *recorder) init(recordFun func(context.Context, []byte) error, bufSize int) {
	rec.recordMu = make(chan struct{}, 1)
	rec.recordMu <- struct{}{}
	rec.bufWriter = bufio.NewWriterSize(writerFunc(
		func(data []byte) (int, error) {
			var err error
			if len(data) > 0 {
				err = recordFun(rec.ctx, data)
				if err != nil {
					return 0, err
				}
			}
			return len(data), err
		},
	), bufSize)
	rec.gzipWriter = gzip.NewWriter(writerFunc(func(data []byte) (int, error) {
		// Chop data into batches of at most RecorderBufferSize
		N := 0
		for len(data) > RecorderBufferSize {
			nn, err := rec.bufWriter.Write(data[:RecorderBufferSize])
			N += nn
			if err != nil {
				return N, err
			}
			data = data[RecorderBufferSize:]
		}
		if len(data) > 0 {
			nn, err := rec.bufWriter.Write(data)
			N += nn
			return N, err
		}
		return N, nil
	}))
}

func NewRecorder(sessionID string, store store.DataStore) Recorder {
	rec := new(recorder)
	rec.init(func(ctx context.Context, b []byte) error {
		return store.InsertSessionRecording(ctx, sessionID, b)
	}, RecorderBufferSize)
	return rec
}

func NewControlRecorder(sessionID string, store store.DataStore) Recorder {
	rec := new(recorder)
	rec.init(func(ctx context.Context, b []byte) error {
		return store.InsertControlRecording(ctx, sessionID, b)
	}, ControlRecorderBufferSize)
	return rec
}

func (r *recorder) Record(ctx context.Context, data []byte) error {
	select {
	case _, open := <-r.recordMu:
		if !open {
			return fmt.Errorf("closed")
		}
	case <-ctx.Done():
		return ctx.Err()
	}
	defer func() { r.recordMu <- struct{}{} }()
	r.ctx = ctx
	_, err := r.gzipWriter.Write(data)
	return err
}

func (r *recorder) Close(ctx context.Context) error {
	select {
	case _, open := <-r.recordMu:
		if !open {
			return fmt.Errorf("closed")
		}
	case <-ctx.Done():
		return ctx.Err()
	}
	close(r.recordMu)
	r.ctx = ctx
	err := r.gzipWriter.Close()
	if err == nil {
		err = r.bufWriter.Flush()
	}
	return err
}
