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

package portforward

const (
	// MessageTypePortForwardNew is the message type to start a port-forwarding connection
	// The body MUST contain a PortForwardNew object.
	MessageTypePortForwardNew = "new"
	// MessageTypePortForwardStop is the message type to stop a port-forwarding connection
	MessageTypePortForwardStop = "stop"
	// MessageTypePortForward is the message type for streaming data
	MessageTypePortForward = "forward"
	// MessageTypePortForwardAck is the message type for streaming data acknowledgement
	MessageTypePortForwardAck = "ack"
	// MessageTypeError is returned on internal or protocol errors. The
	// body MUST contain an Error object.
	MessageTypeError = "error"
)

// PortForwardProtocol stores the protocol
type PortForwardProtocol string

// Values for the PortForwardProtocol type
const (
	PortForwardProtocolTCP = "tcp"
	PortForwardProtocolUDP = "udp"
)

// PropertyConnectionID is the name of the property holding the connection ID
const PropertyConnectionID = "connection_id"

// PortForwardNew represents a new port forwarding request
type PortForwardNew struct {
	RemoteHost *string              `msgpack:"remote_host" json:"remote_host"`
	RemotePort *uint16              `msgpack:"remote_port" json:"remote_port"`
	Protocol   *PortForwardProtocol `msgpack:"protocol" json:"protocol"`
}

// Error struct is passed in the Body of MsgProto in case the message type is ErrorMessage
type Error struct {
	// The error description, as in "Permission denied while opening a file"
	Error *string `msgpack:"err" json:"error"`
	// Type of message that raised the error
	MessageType *string `msgpack:"msgtype,omitempty" json:"message_type,omitempty"`
}
