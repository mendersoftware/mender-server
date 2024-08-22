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

package filetransfer

import "time"

const (
	// MessageTypeGet requests a file from the device.
	MessageTypeGet = "get_file"
	// MessageTypePut requests a file upload to the device. The body MUST
	// contain a FileInfo object.
	MessageTypePut = "put_file"
	// MessageTypeACK messages MUST be sent in response to a
	// file_chunk or put_file message.
	MessageTypeACK = "ack"
	// MessageTypeStat requests file information from the device. The body
	// MUST contain a StatFile object.
	MessageTypeStat = "stat"
	// MessageTypeFileInfo is a response to a MessageTypeStat request.
	// The body MUST contain a FileInfo object.
	MessageTypeFileInfo = "file_info"
	// MessageTypeChunk is the message type for streaming file chunks. The
	// body contains a binary slice of the file, and optional "offset" property
	// can be passed in the header.
	MessageTypeChunk = "file_chunk"
	// MessageTypeError is returned on internal or protocol errors. The
	// body MUST contain an Error object.
	MessageTypeError = "error"
)

// The Error struct is passed in the Body of MsgProto in case the message type is ErrorMessage
type Error struct {
	// The error description, as in "Permission denied while opening a file"
	Error *string `msgpack:"err" json:"error"`
	// Type of message that raised the error
	MessageType *string `msgpack:"msgtype,omitempty" json:"message_type,omitempty"`
	// Message id is passed in the MsgProto Properties, and in case it is available and
	// error occurs it is passed for reference in the Body of the error message
	MessageID *string `msgpack:"msgid,omitempty" json:"message_id,omitempty"`
}

// Get file requests the flow of MessageTypeFileChunk messages to be started from the remote end
type GetFile struct {
	// The file path to the file we are requesting
	Path *string `msgpack:"path,omitempty" json:"path,omitempty"`
}

// Stat file requests the file stat structure from the remote end
type StatFile struct {
	// The file path to the file we are requesting
	Path *string `msgpack:"path" json:"path,omitempty"`
}

// FileInfo is the object returned from a StatFile request and is also used
// for "put_file" requests for specifying the target file.
type FileInfo struct {
	// The file path to the file we are sending status for
	Path *string `msgpack:"path" json:"path"`
	// The file size
	Size *int64 `msgpack:"size,omitempty" json:"size,omitempty"`
	// The file owner
	UID *uint32 `msgpack:"uid,omitempty" json:"uid,omitempty"`
	// The file group
	GID *uint32 `msgpack:"gid,omitempty" json:"gid,omitempty"`
	// Mode contains the file mode and permission bits.
	Mode *uint32 `msgpack:"mode,omitempty" json:"mode,omitempty"`
	// ModTime is the last modification time for the file.
	ModTime *time.Time `msgpack:"modtime,omitempty" json:"modification_time,omitempty"`
}

type UploadRequest struct {
	// SrcPath is the (optional) source filename which will be appended
	// to the target path if it points to a directory.
	SrcPath *string `msgpack:"src_path,omitempty" json:"src_path,omitempty"`
	// The file path to the file we are sending status for
	Path *string `msgpack:"path" json:"path"`
	// The file size
	Size *int64 `msgpack:"size,omitempty" json:"size,omitempty"`
	// The file owner
	UID *uint32 `msgpack:"uid,omitempty" json:"uid,omitempty"`
	// The file group
	GID *uint32 `msgpack:"gid,omitempty" json:"gid,omitempty"`
	// Mode contains the file mode and permission bits.
	Mode *uint32 `msgpack:"mode,omitempty" json:"mode,omitempty"`
	// ModTime is the last modification time for the file.
	ModTime *time.Time `msgpack:"modtime,omitempty" json:"modification_time,omitempty"`
}
