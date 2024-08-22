// Copyright 2023 Northern.tech AS
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	    http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.
package inventory

import (
	"encoding/json"
	"time"
)

// 1:1 port of the inventory device
// for inventory api compat
const (
	AttrScopeInventory = "inventory"
	AttrScopeIdentity  = "identity"
	AttrScopeSystem    = "system"
)

type DeviceID string
type GroupName string
type DeviceAttributes []DeviceAttribute

type DeviceAttribute struct {
	Name        string      `json:"name" bson:",omitempty"`
	Description *string     `json:"description,omitempty" bson:",omitempty"`
	Value       interface{} `json:"value" bson:",omitempty"`
	Scope       string      `json:"scope" bson:",omitempty"`
}

// Device is a wrapper for inventory devices
type Device struct {
	// ID is the system-generated device ID
	ID DeviceID `json:"id" bson:"_id,omitempty"`

	// Attributes is a map of attributes names and their values.
	Attributes DeviceAttributes `json:"attributes,omitempty" bson:"attributes,omitempty"`

	// Group contains the device's group name
	Group GroupName `json:"-" bson:"group,omitempty"`

	// CreatedTs contains the timestamp of the creation time
	CreatedTs time.Time `json:"created_ts,omitempty" bson:"created_ts,omitempty"`

	// UpdatedTs contains the timestamp of the latest attribute update
	UpdatedTs time.Time `json:"updated_ts,omitempty" bson:"updated_ts,omitempty"`

	// LastCheckinDate contains the date of the latest device call to backend
	LastCheckinDate *time.Time `json:"check_in_time,omitempty" bson:"check_in_time,omitempty"`

	// Revision is the device object revision
	Revision uint `json:"-" bson:"revision,omitempty"`
}

func (d *DeviceAttributes) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, (*[]DeviceAttribute)(d))
	if err != nil {
		return err
	}
	for i := range *d {
		if (*d)[i].Scope == "" {
			(*d)[i].Scope = AttrScopeInventory
		}
	}

	return nil
}
