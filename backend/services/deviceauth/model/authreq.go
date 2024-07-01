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
	"crypto"
	"errors"

	"github.com/mendersoftware/mender-server/services/deviceauth/utils"
)

// note: fields with underscores need the 'bson' decorator
// otherwise the underscore will be removed upon write to mongo
type AuthReq struct {
	IdData      string `json:"id_data" bson:"id_data"`
	TenantToken string `json:"tenant_token" bson:"tenant_token"`
	PubKey      string `json:"pubkey"`

	//helpers, not serialized
	PubKeyStruct crypto.PublicKey `json:"-" bson:"-"`
}

func (r *AuthReq) Validate() error {
	if r.IdData == "" {
		return errors.New("id_data must be provided")
	}

	if r.PubKey == "" {
		return errors.New("pubkey must be provided")
	}

	// normalize pubkey by parsing+serializing the key string
	//in between, save it in a temp field because it will be useful outside of Validate()
	key, err := utils.ParsePubKey(r.PubKey)
	if err != nil {
		return err
	}

	r.PubKeyStruct = key

	serialized, err := utils.SerializePubKey(key)
	if err != nil {
		return err
	}

	r.PubKey = serialized

	if sorted, err := utils.JsonSort(r.IdData); err != nil {
		return err
	} else {
		r.IdData = sorted
	}

	// not checking tenant token for now - TODO
	return nil
}
