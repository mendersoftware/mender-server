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
	mu        chan struct{}
	dataChan  chan []byte
	closeOnce sync.Once
}

func NewPipeWriter() *PipeWriter {
	mu := make(chan struct{}, 1)
	mu <- struct{}{}
	return &PipeWriter{
		dataChan: make(chan []byte),
		mu:       mu,
	}
}

func (r *PipeWriter) Write(d []byte) (n int, err error) {
	if _, open := <-r.mu; !open {
		return 0, io.EOF
	}
	defer func() {
		if err != io.EOF {
			r.mu <- struct{}{}
		}
	}()
	select {
	case <-r.mu:
		err = io.EOF
		return 0, err
	case r.dataChan <- d:
		return len(d), nil
	}
}

func (r *PipeWriter) RecvChan() <-chan []byte {
	return r.dataChan
}

func (r *PipeWriter) Close() error {
	r.closeOnce.Do(func() {
		close(r.mu)
	})

	return nil
}
