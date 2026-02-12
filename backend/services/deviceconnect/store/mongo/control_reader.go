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
	"bytes"
	"compress/gzip"
	"context"
	"io"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/mendersoftware/mender-server/services/deviceconnect/app"
	"github.com/mendersoftware/mender-server/services/deviceconnect/model"
)

const (
	controlReadBufferSize = 4096
)

var (
	noopReader = &ControlMessageReader{}
)

type ControlMessageReader struct {
	currentOffset int
	output        []byte
	outputLength  int
	buffer        bytes.Buffer
	c             *mongo.Cursor
	gzipReader    *gzip.Reader
}

func NewControlMessageReader(ctx context.Context, c *mongo.Cursor) *ControlMessageReader {
	reader := &ControlMessageReader{
		currentOffset: 0,
		output:        make([]byte, controlReadBufferSize),
		buffer:        bytes.Buffer{},
		c:             c,
		outputLength:  0,
	}

	hasNext := c.Next(ctx)
	if !hasNext {
		return noopReader
	}

	var r model.ControlData
	err := c.Decode(&r)
	if err != nil {
		return noopReader
	}

	reader.buffer.Reset()
	reader.buffer.Write(r.Control)
	gzipReader, e := gzip.NewReader(&reader.buffer)
	if e != nil {
		return noopReader
	}

	n, e := gzipReader.Read(reader.output)
	reader.gzipReader = gzipReader
	reader.outputLength = n
	if e != nil && e != io.EOF {
		return noopReader
	}

	return reader
}

func (r *ControlMessageReader) Pop() *app.Control {
	if r.c == nil {
		return nil
	}

	if r.outputLength < 3 { // at least we have to have type: 1 byte, and two bytes of offset
		n, e := r.gzipReader.Read(r.output[r.outputLength:])
		if e != nil && e != io.EOF {
			return nil
		}
		r.outputLength += n
	}

	if r.outputLength < 3 { // at least we have to have type: 1 byte, and two bytes of offset
		return nil
	}
	m := &app.Control{}
	offset := r.currentOffset
	//now here we can start deserializing the control messages
	//output[:n] contains the uncompressed buffer
	// +---------+----------+---------+
	// | type: 1 | offset:4 | data: l |
	// +---------+----------+---------+
	// where l is type-dependent
	controlMessageBuffer := r.output[:r.outputLength]
	switch controlMessageBuffer[offset] {
	case app.DelayMessage:
		if offset+7 > r.outputLength {
			return nil
		}

		e := m.UnmarshalBinary(controlMessageBuffer[offset:])
		if e != nil {
			return nil
		}
		offset += 7
	case app.ResizeMessage:
		if offset+9 > r.outputLength {
			return nil
		}

		e := m.UnmarshalBinary(controlMessageBuffer[offset:])
		if e != nil {
			return nil
		}
		offset += 9
	default:
		return nil
	}

	r.currentOffset = offset
	if r.currentOffset >= r.outputLength {
		r.outputLength = 0
		r.currentOffset = 0
	}
	return m
}
