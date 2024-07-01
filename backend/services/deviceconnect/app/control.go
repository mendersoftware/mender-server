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
	"encoding/binary"
	"errors"
	"io"
)

const (
	ResizeMessage byte = iota + 1
	DelayMessage
)

var (
	ErrUnknownMessage = errors.New("message of unknown type")
)

type Control struct {
	Type           byte
	Offset         int
	DelayMs        uint16
	TerminalWidth  uint16
	TerminalHeight uint16
}

func (c Control) MarshalBinary() []byte {
	var b []byte

	switch c.Type {
	case ResizeMessage:
		b = make([]byte, 1+4+2+2)
		offset := 0
		b[offset] = c.Type
		offset++
		binary.LittleEndian.PutUint32(b[offset:], uint32(c.Offset))
		offset += 4
		binary.LittleEndian.PutUint16(b[offset:], c.TerminalWidth)
		offset += 2
		binary.LittleEndian.PutUint16(b[offset:], c.TerminalHeight)
	case DelayMessage:
		b = make([]byte, 1+4+2)
		offset := 0
		b[offset] = c.Type
		offset++
		binary.LittleEndian.PutUint32(b[offset:], uint32(c.Offset))
		offset += 4
		binary.LittleEndian.PutUint16(b[offset:], c.DelayMs)
	}
	return b
}

func (c *Control) UnmarshalBinary(controlMessageBuffer []byte) (err error) {
	var offset = 0

	switch controlMessageBuffer[0] {
	case DelayMessage:
		if len(controlMessageBuffer) < 7 {
			return io.ErrShortBuffer
		}
		offset++
		recordingOffset := binary.LittleEndian.Uint32(controlMessageBuffer[offset:])
		offset += 4
		delayMilliSeconds := binary.LittleEndian.Uint16(controlMessageBuffer[offset:])
		c.Type = DelayMessage
		c.Offset = int(recordingOffset)
		c.DelayMs = delayMilliSeconds
		return nil
	case ResizeMessage:
		if len(controlMessageBuffer) < 9 {
			return io.ErrShortBuffer
		}
		offset++
		recordingOffset := binary.LittleEndian.Uint32(controlMessageBuffer[offset:])
		offset += 4
		width := binary.LittleEndian.Uint16(controlMessageBuffer[offset:])
		offset += 2
		height := binary.LittleEndian.Uint16(controlMessageBuffer[offset:])
		c.Type = ResizeMessage
		c.Offset = int(recordingOffset)
		c.TerminalWidth = width
		c.TerminalHeight = height
		return nil
	}
	return ErrUnknownMessage
}
