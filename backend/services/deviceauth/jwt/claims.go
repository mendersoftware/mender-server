// Copyright 2021 Northern.tech AS
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

package jwt

import (
	"encoding/json"
	"time"

	"github.com/mendersoftware/mender-server/pkg/addons"
	"github.com/mendersoftware/mender-server/pkg/mongo/oid"
)

type Claims struct {
	// ID is the unique jwt ID, also device AuthSet UUID. (Required)
	ID oid.ObjectID `json:"jti,omitempty" bson:"_id"`
	// Subject claim holds the device ID. (Required)
	Subject  oid.ObjectID `json:"sub,omitempty" bson:"sub"`
	Audience string       `json:"aud,omitempty" bson:"aud,omitempty"`
	Scope    string       `json:"scp,omitempty" bson:"scp,omitempty"`
	// Issuer holds the configurable issuer claim.
	Issuer string `json:"iss,omitempty" bson:"iss,omitempty"`
	// Tenant claim holds the tenant id this device belongs to.
	Tenant string `json:"mender.tenant,omitempty" bson:"mender.tenant,omitempty"`
	// ExpiresAt is the timestamp when the token becomes invalid. (Required)
	ExpiresAt Time `json:"exp,omitempty" bson:"exp"`
	IssuedAt  Time `json:"iat,omitempty" bson:"iat,omitempty"`
	NotBefore Time `json:"nbf,omitempty" bson:"nbf,omitempty"`
	// Device claim states that this token belongs to a device
	Device bool `json:"mender.device,omitempty" bson:"mender.device,omitempty"`
	// Plan holds the tenant's feature plan claim.
	Plan string `json:"mender.plan,omitempty"`
	// Trial claim holds a boolean which is true if the tenant is in trial mode
	Trial bool `json:"mender.trial" bson:"trial"`
	// Addons contains the settings for addons enabled for the tenant.
	Addons []addons.Addon `json:"mender.addons,omitempty"`
}

type Time struct {
	time.Time
}

func (t Time) MarshalJSON() ([]byte, error) {
	unixTime := t.Unix()
	return json.Marshal(unixTime)
}

func (t *Time) UnmarshalJSON(b []byte) error {
	var unixTime int64
	err := json.Unmarshal(b, &unixTime)
	t.Time = time.Unix(unixTime, 0)
	return err
}

// Valid checks if claims are valid. Returns error if validation fails.
// Note that for now we're only using iss, exp, sub, scp.
// Basic checks are done here, field correctness (e.g. issuer) - at the service
// level, where this info is available.
func (c *Claims) Valid() error {
	if c.Issuer == "" ||
		c.ID.String() == "" ||
		c.Subject.String() == "" {
		return ErrTokenInvalid
	}

	if c.ExpiresAt.Before(time.Now()) {
		return ErrTokenExpired
	}

	return nil
}
