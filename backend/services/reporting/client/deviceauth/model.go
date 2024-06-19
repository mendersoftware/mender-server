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

package deviceauth

import (
	"time"
)

// DeviceAuthDevice is a wrapper for device auth devices
type DeviceAuthDevice struct {
	ID              string                    `json:"id"`
	IdDataStruct    map[string]string         `bson:"id_data_struct,omitempty"`
	Status          string                    `json:"status"`
	CreatedTs       time.Time                 `json:"created_ts"`
	UpdatedTs       time.Time                 `json:"updated_ts"`
	LastCheckinDate time.Time                 `json:"check_in_time,omitempty"`
	AuthSets        []DeviceAuthAuthSet       `json:"auth_sets"`
	External        *DeviceAuthExternalDevice `json:"external,omitempty"`
	Revision        uint                      `json:"revision"`
}

// DeviceAuthExternalDevice stores the external ID for the device
type DeviceAuthExternalDevice struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
	Name     string `json:"name,omitempty"`
}

// DeviceAuthAuthSet stores a device auth set
type DeviceAuthAuthSet struct {
	Id           string                 `json:"id"`
	IdData       string                 `json:"id_data"`
	IdDataStruct map[string]interface{} `bson:"id_data_struct,omitempty"`
	IdDataSha256 []byte                 `bson:"id_data_sha256,omitempty"`
	PubKey       string                 `json:"pubkey"`
	Timestamp    *time.Time             `json:"ts"`
	Status       string                 `json:"status"`
}
