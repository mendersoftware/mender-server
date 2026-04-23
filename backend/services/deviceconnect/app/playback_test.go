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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPipeWriter(t *testing.T) {
	r := NewPipeWriter()
	assert.NotNil(t, r)
	d := []byte("testing123")
	errCh := make(chan error, 1)
	go func() {
		_, err := r.Write(d)
		errCh <- err
	}()
	select {
	case actual := <-r.RecvChan():
		assert.Equal(t, d, actual, "did not receive the data written to PipeWriter")
	case err := <-errCh:
		assert.NoError(t, err, "unexpected error from PipeWriter.Write")
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for PipeWriter data")
	}

	assert.NoError(t, r.Close())

	errCh = make(chan error, 1)
	go func() {
		_, err := r.Write(d)
		errCh <- err
	}()
	select {
	case err := <-errCh:
		assert.ErrorIs(t, err, io.EOF)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for PipeWriter data")
	}
}
