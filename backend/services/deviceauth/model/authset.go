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
	"time"

	"github.com/pkg/errors"
)

const (
	AuthSetKeyIdData       = "id_data"
	AuthSetKeyPubKey       = "pubkey"
	AuthSetKeyDeviceId     = "device_id"
	AuthSetKeyStatus       = "status"
	AuthSetKeyIdDataSha256 = "id_data_sha256"
)

type AuthSet struct {
	Id           string                 `json:"id" bson:"_id,omitempty"`
	IdData       string                 `json:"-" bson:"id_data,omitempty"`
	IdDataStruct map[string]interface{} `json:"identity_data" bson:"id_data_struct,omitempty"`
	IdDataSha256 []byte                 `json:"-" bson:"id_data_sha256,omitempty"`
	PubKey       string                 `json:"pubkey" bson:"pubkey,omitempty"`
	DeviceId     string                 `json:"-" bson:"device_id,omitempty"`
	Timestamp    *time.Time             `json:"ts" bson:"ts,omitempty"`
	Status       string                 `json:"status" bson:"status,omitempty"`
	TenantID     string                 `json:"-" bson:"tenant_id"`
}

type AuthSetUpdate struct {
	Id           string                 `bson:"id,omitempty"`
	IdData       string                 `bson:"id_data,omitempty"`
	IdDataStruct map[string]interface{} `bson:"id_data_struct,omitempty"`
	IdDataSha256 []byte                 `bson:"id_data_sha256,omitempty"`
	PubKey       string                 `bson:"pubkey,omitempty"`
	DeviceId     string                 `bson:"device_id,omitempty"`
	Timestamp    *time.Time             `bson:"ts,omitempty"`
	Status       string                 `bson:"status,omitempty"`
}

type DevAdmAuthSet struct {
	Id             string                 `json:"id" bson:"_id,omitempty"`
	DeviceIdentity string                 `json:"device_identity" bson:"id_data"`
	Key            string                 `json:"key" bson:"pubkey"`
	DeviceId       string                 `json:"device_id" bson:"device_id,omitempty"`
	RequestTime    *time.Time             `json:"request_time" bson:"request_time"`
	Status         string                 `json:"status" bson:"status"`
	Attributes     map[string]interface{} `json:"attributes" bson:"attributes"`
}

func NewDevAdmAuthSet(a AuthSet) (*DevAdmAuthSet, error) {
	as := &DevAdmAuthSet{
		Id:             a.Id,
		DeviceIdentity: a.IdData,
		Key:            a.PubKey,
		DeviceId:       a.DeviceId,
		RequestTime:    a.Timestamp,
		Status:         a.Status,
	}

	// we don't store decoded attributes, but we will
	// decode them on the fly for the time being
	err := json.Unmarshal([]byte(a.IdData), &as.Attributes)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to decode attributes for device %s, auth set %s",
			a.DeviceId,
			a.Id,
		)
	}

	return as, nil
}
