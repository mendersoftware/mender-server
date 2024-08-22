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
	"crypto/rsa"
	"encoding/json"
	"io"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/services/deviceauth/utils"
)

type PreAuthReq struct {
	Force     bool   `json:"force" bson:"force"`
	DeviceId  string `json:"device_id" valid:"required" bson:"device_id"`
	AuthSetId string `json:"auth_set_id" valid:"required" bson:"auth_set_id"`
	IdData    string `json:"id_data" valid:"required" bson:"id_data"`
	PubKey    string `json:"pubkey" valid:"required" bson:"pubkey"`
}

func ParsePreAuthReq(source io.Reader) (*PreAuthReq, error) {
	jd := json.NewDecoder(source)

	var req PreAuthReq

	if err := jd.Decode(&req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	return &req, nil
}

func (r *PreAuthReq) Validate() error {
	err := validation.ValidateStruct(r,
		validation.Field(&r.DeviceId, validation.Required),
		validation.Field(&r.AuthSetId, validation.Required),
		validation.Field(&r.IdData, validation.Required),
		validation.Field(&r.PubKey, validation.Required),
	)
	if err != nil {
		return err
	}

	if sorted, err := utils.JsonSort(r.IdData); err != nil {
		return err
	} else {
		r.IdData = sorted
	}

	//normalize key
	key, err := utils.ParsePubKey(r.PubKey)
	if err != nil {
		return err
	}

	keyStruct, ok := key.(*rsa.PublicKey)
	if !ok {
		return errors.New("cannot decode key as RSA public key")
	}

	serialized, err := utils.SerializePubKey(keyStruct)
	if err != nil {
		return err
	}

	r.PubKey = serialized

	return nil
}
