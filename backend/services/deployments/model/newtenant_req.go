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
	"io"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/mendersoftware/mender-server/pkg/rules"
)

type NewTenantReq struct {
	TenantId string `json:"tenant_id"`
}

func (t NewTenantReq) Validate() error {
	return validation.ValidateStruct(&t,
		validation.Field(&t.TenantId, validation.Required, rules.TenantID),
	)
}

func ParseNewTenantReq(source io.Reader) (*NewTenantReq, error) {
	jd := json.NewDecoder(source)

	var r NewTenantReq
	if err := jd.Decode(&r); err != nil {
		return nil, err
	}

	if err := r.Validate(); err != nil {
		return nil, err
	}

	return &r, nil
}
