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
	"io"
	"sync"
)

const (
	DefaultPlaybackSleepIntervalMs = uint(100)
)

type PipeWriter struct {
	dataChan  chan []byte
	closed    chan struct{}
	closeOnce sync.Once
}

func NewPipeWriter() *PipeWriter {
	return &PipeWriter{
		dataChan: make(chan []byte),
		closed:   make(chan struct{}),
	}
}

func (r *PipeWriter) Write(d []byte) (int, error) {
	select {
	case <-r.closed:
		return 0, io.EOF
	case r.dataChan <- d:
		return len(d), nil
	}
}

func (r *PipeWriter) RecvChan() <-chan []byte {
	return r.dataChan
}

func (r *PipeWriter) Close() error {
	r.closeOnce.Do(func() {
		close(r.closed)
	})

	return nil
}
