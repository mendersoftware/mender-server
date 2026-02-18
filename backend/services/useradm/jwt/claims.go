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
package jwt

import (
	"encoding/json"
	"time"

	"github.com/mendersoftware/mender-server/pkg/mongo/v2/oid"
)

type Claims struct {
	// ID is the unique token UUID.
	ID oid.ObjectID `json:"jti,omitempty" bson:"_id,omitempty"`
	// Subject holds the UUID associated with the user's account.
	Subject oid.ObjectID `json:"sub,omitempty" bson:"sub,omitempty"`
	// ExpiresAt is the absolute time when the token expires.
	ExpiresAt *Time `json:"exp,omitempty" bson:"exp,omitempty"`
	// IssuedAt is the absolute time the token was created.
	IssuedAt Time `json:"iat,omitempty" bson:"iat,omitempty"`
	// Tenant holds the tenant ID claim
	Tenant string `json:"mender.tenant,omitempty" bson:"tenant,omitempty"`
	// User claims that this token is for the management API.
	User bool `json:"mender.user,omitempty" bson:"user,omitempty"`
	// Issuer contains the configured Issuer claim (defaults to "Mender")
	Issuer string `json:"iss,omitempty" bson:"iss,omitempty"`
	// Scope determines the API scope of the token (defaults to "mender.*")
	Scope     string `json:"scp,omitempty" bson:"scp,omitempty"`
	Audience  string `json:"aud,omitempty" bson:"aud,omitempty"`
	NotBefore Time   `json:"nbf,omitempty" bson:"nbf,omitempty"`
}

// Time is a simple wrapper of time.Time that marshals/unmarshals JSON
// to/from UNIX time.
type Time struct {
	time.Time
}

func (t Time) MarshalJSON() ([]byte, error) {
	timeUnix := t.Unix()
	return json.Marshal(timeUnix)
}

func (t *Time) UnmarshalJSON(b []byte) error {
	var timeUnix int64
	err := json.Unmarshal(b, &timeUnix)
	if err != nil {
		return err
	}
	t.Time = time.Unix(timeUnix, 0)
	return nil
}

// Valid checks if claims are valid. Returns error if validation fails.
// Note that for now we're only using iss, exp, sub, scp.
// Basic checks are done here, field correctness (e.g. issuer) - at the service level, where this
// info is available.
func (c *Claims) Valid() error {
	if c.Issuer == "" ||
		c.Subject.Type() == oid.TypeNil ||
		c.ID.Type() == oid.TypeNil ||
		c.Scope == "" {
		return ErrTokenInvalid
	}

	if c.ExpiresAt != nil {
		now := time.Now()
		if now.After(c.ExpiresAt.Time) {
			return ErrTokenExpired
		}
	}

	return nil
}
