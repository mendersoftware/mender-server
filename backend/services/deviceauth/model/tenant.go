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

	"github.com/pkg/errors"
)

type NewTenant struct {
	TenantId string `json:"tenant_id"`
}

func ParseNewTenant(source io.Reader) (*NewTenant, error) {
	jd := json.NewDecoder(source)

	var t NewTenant
	if err := jd.Decode(&t); err != nil {
		return nil, err
	}

	if t.TenantId == "" {
		return nil, errors.New("tenant_id must be provided")
	}

	return &t, nil
}
