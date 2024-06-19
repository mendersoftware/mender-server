// Copyright 2023 Northern.tech AS
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

package utils

import (
	"errors"
	"io"
)

var ErrStreamTooLarge = errors.New("read too many bytes")

type ReadCounter interface {
	io.Reader
	Count() int64
}

func CountReads(r io.Reader) ReadCounter {
	if rc, ok := r.(ReadCounter); ok {
		return rc
	}
	return &readCounter{
		R: r,
		N: 0,
	}
}

type readCounter struct {
	R io.Reader
	N int64
}

func (rc *readCounter) Read(b []byte) (int, error) {
	n, err := rc.R.Read(b)
	rc.N += int64(n)
	return n, err
}

func (rc readCounter) Count() int64 {
	return rc.N
}

type limitedReader struct {
	R       io.Reader // underlying reader
	N       int64     // bytes read
	atLeast int64     // expected number of bytes read before reaching EOF
	atMost  int64     // maximum number of bytes expected
}

// ReadExactly returns a reader that expects to read exactly size number of
// bytes. If io.EOF is returned before reading size bytes, the error is replaced
// by io.ErrUnexpectedEOF. Similarly if, the reader reads past size bytes, the
// reader replaces the error with ErrStreamTooLarge.
func ReadExactly(r io.Reader, size int64) ReadCounter {
	return &limitedReader{
		R:       r,
		atLeast: size,
		atMost:  size,
	}
}

// ReadAtMost, like ReadExactly, returns a reader that expects to read at most
// size bytes.
func ReadAtMost(r io.Reader, size int64) ReadCounter {
	return &limitedReader{
		R:      r,
		atMost: size,
	}
}

func (l *limitedReader) Read(p []byte) (n int, err error) {
	n, err = l.R.Read(p)
	l.N += int64(n)
	if l.N > l.atMost {
		err = ErrStreamTooLarge
	} else if err == io.EOF && l.N < l.atLeast {
		err = io.ErrUnexpectedEOF
	}
	return n, err
}

func (l *limitedReader) Count() int64 {
	return l.N
}
