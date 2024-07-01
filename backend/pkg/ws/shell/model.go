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

package shell

type MenderShellMessageStatus int

const (
	MessageTypePingShell    = "ping"
	MessageTypePongShell    = "pong"
	MessageTypeResizeShell  = "resize"
	MessageTypeShellCommand = "shell"
	MessageTypeSpawnShell   = "new"
	MessageTypeStopShell    = "stop"
)

const (
	NormalMessage MenderShellMessageStatus = iota + 1
	ErrorMessage
	ControlMessage
)
