// Copyright 2020 Northern.tech AS
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

package model

import "time"

// Values for the device status attribute
const (
	DeviceStatusDisconnected = "disconnected"
	DeviceStatusConnected    = "connected"
	DeviceStatusUnknown      = "unknown"
)

// Device represents a device and its attributes
type Device struct {
	ID        string    `json:"device_id" bson:"_id"`
	Status    string    `json:"status" bson:"status"`
	CreatedTs time.Time `json:"created_ts" bson:"created_ts,omitempty"`
	UpdatedTs time.Time `json:"updated_ts" bson:"updated_ts,omitempty"`
}
