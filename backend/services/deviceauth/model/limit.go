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
package model

const (
	LimitMaxDeviceCount = "max_devices"
)

var (
	ValidLimits = []string{LimitMaxDeviceCount}
)

type Limit struct {
	Id       string `json:"-" bson:"_id,omitempty"`
	Name     string `bson:"name"`
	Value    uint64 `json:"value" bson:"value"`
	TenantID string `json:"-" bson:"tenant_id"`
}

func (l Limit) IsLess(what uint64) bool {
	return what < l.Value
}

func IsValidLimit(name string) bool {
	for _, n := range ValidLimits {
		if name == n {
			return true
		}
	}
	return false
}
