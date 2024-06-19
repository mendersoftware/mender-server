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

import (
	"time"
)

const (
	InventoryGroupScope         = "system"
	InventoryGroupAttributeName = "group"
)

type DeviceAttribute struct {
	Name        string      `json:"name" bson:",omitempty"`
	Description *string     `json:"description,omitempty" bson:",omitempty"`
	Value       interface{} `json:"value" bson:",omitempty"`
	Scope       string      `json:"scope" bson:",omitempty"`
}

// Inventory device wrapper
type InvDevice struct {
	//system-generated device ID
	ID string `json:"id" bson:"_id,omitempty"`

	//a map of attributes names and their values.
	Attributes []DeviceAttribute `json:"attributes,omitempty" bson:"attributes,omitempty"`

	//device's group name
	Group string `json:"-" bson:"group,omitempty"`

	CreatedTs time.Time `json:"-" bson:"created_ts,omitempty"`
	//Timestamp of the last attribute update.
	UpdatedTs time.Time `json:"updated_ts" bson:"updated_ts,omitempty"`
}

type DeviceIds struct {
	Devices []string `json:"devices,omitempty" valid:"required" bson:"-"`
}

type SearchParams struct {
	Page      int               `json:"page"`
	PerPage   int               `json:"per_page"`
	Filters   []FilterPredicate `json:"filters"`
	DeviceIDs []string          `json:"device_ids"`
}

type FilterPredicate struct {
	Scope     string      `json:"scope" bson:"scope"`
	Attribute string      `json:"attribute" bson:"attribute"`
	Type      string      `json:"type" bson:"type"`
	Value     interface{} `json:"value" bson:"value"`
}
