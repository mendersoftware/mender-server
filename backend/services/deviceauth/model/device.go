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

import (
	"encoding/json"
	"net/url"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/ratelimits"
)

const (
	DevStatusAccepted = "accepted"
	DevStatusRejected = "rejected"
	DevStatusPending  = "pending"
	DevStatusPreauth  = "preauthorized"
	DevStatusNoAuth   = "noauth"

	DevKeyId           = "_id"
	DevKeyIdData       = "id_data"
	DevKeyIdDataSha256 = "id_data_sha256"
	DevKeyStatus       = "status"
)

var (
	DevStatuses = []string{
		DevStatusPending,
		DevStatusRejected,
		DevStatusAccepted,
		DevStatusPreauth,
		DevStatusNoAuth,
	}
)

// note: fields with underscores need the 'bson' decorator
// otherwise the underscore will be removed upon write to mongo
type Device struct {
	Id              string                 `json:"id" bson:"_id,omitempty"`
	IdData          string                 `json:"-" bson:"id_data,omitempty"`
	IdDataStruct    map[string]interface{} `json:"identity_data" bson:"id_data_struct,omitempty"`
	IdDataSha256    []byte                 `json:"-" bson:"id_data_sha256,omitempty"`
	Status          string                 `json:"status" bson:",omitempty"`
	Decommissioning bool                   `json:"decommissioning" bson:",omitempty"`
	CreatedTs       time.Time              `json:"created_ts" bson:"created_ts,omitempty"`
	UpdatedTs       time.Time              `json:"updated_ts" bson:"updated_ts,omitempty"`
	AuthSets        []AuthSet              `json:"auth_sets" bson:"-"`
	CheckInTime     *time.Time             `json:"check_in_time,omitempty" bson:"check_in_time,omitempty"` // nolint:lll
	//ApiLimits override tenant-wide quota/burst config
	ApiLimits ratelimits.ApiLimits `json:"-" bson:"api_limits"`

	//object revision which we use when synchronizing status with inventory service
	Revision uint   `json:"-" bson:"revision"`
	TenantID string `json:"-" bson:"tenant_id"`
}

type DeviceUpdate struct {
	PubKey          string                 `json:"-" bson:",omitempty"`
	IdData          string                 `json:"id_data" bson:"id_data,omitempty"`
	IdDataStruct    map[string]interface{} `bson:"id_data_struct,omitempty"`
	IdDataSha256    []byte                 `bson:"id_data_sha256,omitempty"`
	Status          string                 `json:"-" bson:",omitempty"`
	Decommissioning *bool                  `json:"-" bson:",omitempty"`
	UpdatedTs       *time.Time             `json:"updated_ts" bson:"updated_ts,omitempty"`
	CheckInTime     *time.Time             `json:"-" bson:"check_in_time,omitempty"`
}

func NewDevice(id, id_data, pubkey string) *Device {
	now := time.Now()

	return &Device{
		Id:              id,
		IdData:          id_data,
		Status:          DevStatusNoAuth,
		Decommissioning: false,
		CreatedTs:       now,
		UpdatedTs:       now,
		Revision:        1,
	}
}

type DeviceFilter struct {
	Status []string `json:"status,omitempty"`
	IDs    []string `json:"id,omitempty"`
}

func (fltr DeviceFilter) Validate() error {
	for _, stat := range fltr.Status {
		if !govalidator.IsIn(stat, DevStatuses...) {
			return errors.Errorf(
				`filter status must be one of: `+
					`%s, %s, %s, %s or %s`,
				DevStatusAccepted, DevStatusPending,
				DevStatusRejected, DevStatusPreauth,
				DevStatusNoAuth,
			)
		}
	}
	return nil
}

func (fltr *DeviceFilter) UnmarshalJSON(b []byte) error {
	schema := struct {
		Status interface{} `json:"status"`
		IDs    interface{} `json:"id"`
	}{}

	err := json.Unmarshal(b, &schema)
	if err != nil {
		return err
	}
	switch t := schema.Status.(type) {
	case string:
		fltr.Status = []string{t}
	case []interface{}:
		fltr.Status = make([]string, 0, len(t))
		for _, elem := range t {
			if str, ok := elem.(string); ok {
				fltr.Status = append(fltr.Status, str)
			} else {
				return errors.New(
					"invalid JSON type for 'status': " +
						"must be string or []string",
				)
			}
		}
	case nil:
		break
	default:
		return errors.New(
			"invalid JSON type for 'status': " +
				"must be string or []string",
		)
	}
	switch t := schema.IDs.(type) {
	case string:
		fltr.IDs = []string{t}
	case []interface{}:
		fltr.IDs = make([]string, 0, len(t))
		for _, elem := range t {
			if str, ok := elem.(string); ok {
				fltr.IDs = append(fltr.IDs, str)
			} else {
				return errors.New(
					"invalid JSON type for 'id': " +
						"must be string or []string",
				)
			}
		}
	case nil:
		break
	default:
		return errors.New(
			"invalid JSON type for 'id': " +
				"must be string or []string",
		)
	}
	return nil
}

func (fltr *DeviceFilter) ParseForm(form url.Values) error {
	if stat, ok := form["status"]; ok {
		fltr.Status = stat
	}
	if IDs, ok := form["id"]; ok {
		fltr.IDs = IDs
	}
	return fltr.Validate()
}

type DeviceAttribute struct {
	Name        string      `json:"name" bson:",omitempty"`
	Description *string     `json:"description,omitempty" bson:",omitempty"`
	Value       interface{} `json:"value" bson:",omitempty"`
	Scope       string      `json:"scope" bson:",omitempty"`
}

type DeviceInventoryUpdate struct {
	Id       string `json:"id"`
	Revision uint   `json:"revision"`
}
