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
package identity

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mendersoftware/mender-server/pkg/addons"
	"github.com/pkg/errors"
)

type Identity struct {
	Subject  string         `json:"sub" valid:"required"`
	Tenant   string         `json:"mender.tenant,omitempty"`
	IsUser   bool           `json:"mender.user,omitempty"`
	IsDevice bool           `json:"mender.device,omitempty"`
	Plan     string         `json:"mender.plan,omitempty"`
	Addons   []addons.Addon `json:"mender.addons,omitempty"`
	Trial    bool           `json:"mender.trial"`
}

// ExtractJWTFromHeader inspect the Authorization header for a Bearer token and
// if not present looks for a "JWT" cookie.
func ExtractJWTFromHeader(r *http.Request) (jwt string, err error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		jwtCookie, err := r.Cookie("JWT")
		if err != nil {
			return "", errors.New("Authorization not present in header")
		}
		jwt = jwtCookie.Value
	} else {
		auths := strings.Split(auth, " ")

		if len(auths) != 2 {
			return "", errors.Errorf("malformed Authorization header")
		}

		if !strings.EqualFold(auths[0], "Bearer") {
			return "", errors.Errorf("unknown Authorization method %s", auths[0])
		}
		jwt = auths[1]
	}
	return jwt, nil
}

// Generate identity information from given JWT by extracting subject and tenant claims.
// Note that this function does not perform any form of token signature
// verification.
func ExtractIdentity(token string) (id Identity, err error) {
	var (
		claims []byte
		jwt    []string
	)
	jwt = strings.Split(token, ".")
	if len(jwt) != 3 {
		return id, errors.New("identity: incorrect token format")
	}
	claims, err = base64.RawURLEncoding.DecodeString(jwt[1])
	if err != nil {
		return id, errors.Wrap(err,
			"identity: failed to decode base64 JWT claims")
	}
	err = json.Unmarshal(claims, &id)
	if err != nil {
		return id, errors.Wrap(err,
			"identity: failed to decode JSON JWT claims")
	}
	return id, id.Validate()
}

func (id Identity) Validate() error {
	if id.Subject == "" {
		return errors.New("identity: claim \"sub\" is required")
	}
	return nil
}
