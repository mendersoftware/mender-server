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
package http

import (
	"encoding/json"
	"io"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"

	"github.com/mendersoftware/mender-server/services/deviceauth/model"
	"github.com/mendersoftware/mender-server/services/deviceauth/utils"
)

type preAuthReq struct {
	Force  bool                   `json:"force" valid:"-"`
	IdData map[string]interface{} `json:"identity_data" valid:"-"`
	PubKey string                 `json:"pubkey" valid:"required"`
}

func parsePreAuthReq(source io.Reader) (*preAuthReq, error) {
	jd := json.NewDecoder(source)

	var req preAuthReq

	if err := jd.Decode(&req); err != nil {
		return nil, err
	}

	if err := req.validate(); err != nil {
		return nil, err
	}

	return &req, nil
}

func (r *preAuthReq) validate() error {
	err := validation.ValidateStruct(r,
		validation.Field(&r.IdData, validation.Required),
		validation.Field(&r.PubKey, validation.Required),
	)
	if err != nil {
		return err
	}

	_, err = json.Marshal(r.IdData)
	if err != nil {
		return err
	}

	//normalize key
	key, err := utils.ParsePubKey(r.PubKey)
	if err != nil {
		return err
	}

	serialized, err := utils.SerializePubKey(key)
	if err != nil {
		return err
	}

	r.PubKey = serialized

	return nil
}

func (r *preAuthReq) getDbModel() (*model.PreAuthReq, error) {
	enc, err := json.Marshal(r.IdData)
	if err != nil {
		return nil, err
	}

	dId := uuid.New()
	asId := uuid.New()

	return &model.PreAuthReq{
		Force:     r.Force,
		DeviceId:  dId.String(),
		AuthSetId: asId.String(),
		IdData:    string(enc),
		PubKey:    r.PubKey,
	}, nil
}
