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

package ws

import "encoding"

const ProtocolVersion = 1

// ProtoType defines how the ProtoMsg should be interpreted.
type ProtoType uint16

const (
	// ProtoInvalid signifies an invalid (uninitialized) ProtoMsg.
	ProtoInvalid ProtoType = iota
	// ProtoTypeShell is used for communicating remote terminal session data.
	ProtoTypeShell
	// ProtoTypeFileTransfer is used for file transfer from/to the device.
	ProtoTypeFileTransfer
	// ProtoTypePortForward is used for port-forwarding connections to the device.
	ProtoTypePortForward
	// ProtoTypeMenderClient is used for communication with the Mender client.
	ProtoTypeMenderClient

	// ProtoTypeControl is a reserved proto type for session control messages.
	ProtoTypeControl ProtoType = 0xFFFF
)

const (
	// MessageTypes for session control messages (ProtoTypeControl).

	// MessageTypePing sends a ping. After receiving a ping, the receiver
	// MUST respond with a pong message or the session will time out.
	MessageTypePing = "ping"
	// MessageTypePong is sent in response to a MessageTypePing.
	MessageTypePong = "pong"
	// MessageTypeOpen allocates a new peer-to-peer deviceconnect session.
	// The other peer can either respond with MessageTypeAccept or
	// MessageTypeError
	MessageTypeOpen = "open"
	// MessageTypeAccept is a successful response to an open request.
	MessageTypeAccept = "accept"
	// MessageTypeClose is sent when the session MUST close. All
	// communication on the session stop after receiving this message.
	MessageTypeClose = "close"
	// MessageTypeError is sent on a general protocol violation/error.
	// An error message MUST contain an Error object. If the object's
	// "close" field is set this message also closes the session.
	MessageTypeError = "error"
)

// ProtoHdr provides the info about what the ProtoMsg contains and
// to which protocol the message should be routed.
type ProtoHdr struct {
	// Proto defines which protocol this message belongs
	// to (required).
	Proto ProtoType `msgpack:"proto"`
	// MsgType is an optional content type header describing
	// the protocol specific content type of the message.
	MsgType string `msgpack:"typ,omitempty"`
	// SessionID is used to identify one ProtoMsg stream for
	// multiplexing multiple ProtoMsg sessions over the same connection.
	SessionID string `msgpack:"sid,omitempty"`
	// Properties provide a map of optional prototype specific
	// properties (such as http headers or other meta-data).
	Properties map[string]interface{} `msgpack:"props,omitempty"`
}

// ProtoMsg is a wrapper to messages communicated on bidirectional interfaces
// such as websockets to wrap data from other application protocols.
type ProtoMsg struct {
	// Header contains a protocol specific header with a single
	// fixed ProtoType ("typ") field and optional hints for decoding
	// the payload.
	Header ProtoHdr `msgpack:"hdr"`
	// Body contains the raw protocol data. The data contained in Body
	// can be arbitrary and must be decoded according to the protocol
	// defined in the header.
	Body []byte `msgpack:"body,omitempty"`
}

func (m *ProtoMsg) Bind(b encoding.BinaryMarshaler) error {
	data, err := b.MarshalBinary()
	m.Body = data
	return err
}

// The Error struct is passed in the Body of MessageTypeError.
type Error struct {
	// The error description, as in "Permission denied while opening a file"
	Error string `msgpack:"err" json:"error"`
	// Close determines whether the session closed as a result of this error.
	Close bool `msgpack:"close" json:"close"`
	// Code is the error code associated with the error. Values in the
	// range 400-599 carry the same semantic meaning as the HTTP equivalent.
	// Please allocate new error codes starting from 1000 for custom error codes.
	Code int `msgpack:"code,omitempty" json:"code,omitempty"`
	// MessageProto is the protocol of the message that caused the error.
	MessageProto ProtoType `msgpack:"msgproto,omitempty" json:"message_protocol,omitempty"`
	// Type of message that raised the error
	MessageType string `msgpack:"msgtype,omitempty" json:"message_type,omitempty"`
	// Message id is passed in the MsgProto Properties, and in case it is available and
	// error occurs it is passed for reference in the Body of the error message
	MessageID string `msgpack:"msgid,omitempty" json:"message_id,omitempty"`
}

// ProtoMsg handshake semantics:
// 1)  The requester sends an "open" control message with all the protocol
//     versions it supports to the peer.
// 2a) On success, the peer will respond with a message of type Accept with an
//     agreed upon version together with all the ProtoTypes the peer is willing
//     to accept.
// 2b) On failure, the peer will respond with an Error message with a reasoning.*
//
//     *The client can also expect to get an Open message type in return. In
//     this case it MUST check the header's "status" property, if the property
//     exists and equal to 1 the client is version 0, and does not support
//     features added after Mender 2.6.

// Open is the schema used for initiating a ProtoMsg handshake.
type Open struct {
	// Versions is a list of versions the client is able to interpret.
	Versions []int `msgpack:"versions"`
}

// Accept is the schema for the message type "accept" for a successful response to
// a ProtoMsg handshake.
type Accept struct {
	// Version is the accepted version used for this session.
	Version int `msgpack:"version"`
	// Protocols is a list of protocols the peer is willing to accept.
	Protocols []ProtoType `msgpack:"protocols"`
}
