// Copyright 2023 Northern.tech AS
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
	"encoding/json"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/mendersoftware/mender-server/pkg/mongo/oid"

	"github.com/mendersoftware/mender-server/services/useradm/jwt"
)

type TokenRequest struct {
	Name      *string `json:"name"`
	ExpiresIn int64   `json:"expires_in"`
}

const defaultTokenMaxExpiration = 31536000

func (tr TokenRequest) Validate(maxExpiration int) error {
	if maxExpiration <= 0 {
		maxExpiration = defaultTokenMaxExpiration
	}
	return validation.ValidateStruct(&tr,
		validation.Field(&tr.Name, validation.Required, lessThan4096),
		validation.Field(&tr.ExpiresIn, validation.Min(0), validation.Max(maxExpiration)))
}

type PersonalAccessToken struct {
	// system-generated user ID
	ID oid.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	// token name
	Name *string `json:"name,omitempty" on:"name,omitempty"`
	// timestamp of the last usage
	LastUsed *time.Time `json:"last_used,omitempty" bson:"last_used,omitempty"`
	// the absolute time when the token expires.
	ExpirationDate *jwt.Time `json:"expiration_date,omitempty" bson:"exp,omitempty"`
	// CreatedTs is the absolute time the token was created.
	CreatedTs jwt.Time `json:"created_ts,omitempty" bson:"iat,omitempty"`
}

type apiToken struct {
	// system-generated user ID
	ID oid.ObjectID `json:"id,omitempty"`
	// token name
	Name *string `json:"name,omitempty"`
	// timestamp of the last usage
	LastUsed *time.Time `json:"last_used,omitempty"`
	// the absolute time when the token expires
	ExpirationDate *time.Time `json:"expiration_date,omitempty"`
	// timestamp of the token creation
	CreatedTs *time.Time `json:"created_ts,omitempty"`
}

func newApiToken(t PersonalAccessToken) apiToken {
	var expiration *time.Time
	if t.ExpirationDate != nil {
		expiration = &t.ExpirationDate.Time
	}
	return apiToken{
		t.ID,
		t.Name,
		t.LastUsed,
		expiration,
		&t.CreatedTs.Time,
	}
}

func (t PersonalAccessToken) MarshalJSON() ([]byte, error) {
	return json.Marshal(newApiToken(t))
}
