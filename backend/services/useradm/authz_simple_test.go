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
package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/pkg/mongo/oid"
	"github.com/mendersoftware/mender-server/services/useradm/jwt"
	"github.com/mendersoftware/mender-server/services/useradm/scope"
)

func TestSimpleAuthzAuthorize(t *testing.T) {
	testCases := map[string]struct {
		inResource string
		inAction   string
		inToken    *jwt.Token

		outErr string
	}{
		"ok - useradm resource": {
			inResource: "useradm:some:resource:id",
			inAction:   "POST",
			inToken: &jwt.Token{
				Claims: jwt.Claims{
					Subject: oid.NewUUIDv5("testsubject"),
					Issuer:  "mender",
					ExpiresAt: &jwt.Time{
						Time: time.Now().Add(time.Hour),
					},
					Scope: scope.All,
				},
			},
		},
		"ok - other service's resource": {
			inResource: "otherservice:some:resource:id",
			inAction:   "POST",
			inToken: &jwt.Token{
				Claims: jwt.Claims{
					Issuer: "mender",
					ExpiresAt: &jwt.Time{
						Time: time.Now().Add(time.Hour),
					},
					Subject: oid.NewUUIDv5("testsubject"),
					Scope:   scope.All,
				},
			},
		},
		"error: unknown/incompatible scope": {
			inResource: "useradm:some:resource:id",
			inAction:   "POST",
			inToken: &jwt.Token{
				Claims: jwt.Claims{
					Issuer: "mender",
					ExpiresAt: &jwt.Time{
						Time: time.Now().Add(time.Hour),
					},
					Subject: oid.NewUUIDv5("testsubject"),
					Scope:   "foobar",
				},
			},
			outErr: "unauthorized",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			authz := &SimpleAuthz{}

			err := authz.Authorize(context.TODO(),
				tc.inToken, tc.inResource, tc.inAction)

			if tc.outErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.outErr)
			}
		})
	}
}

func loadPrivKey(path string, t *testing.T) *rsa.PrivateKey {
	pem_data, err := ioutil.ReadFile(path)
	if err != nil {
		t.FailNow()
	}

	block, _ := pem.Decode(pem_data)

	if block == nil ||
		block.Type != "RSA PRIVATE KEY" {
		t.FailNow()
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.FailNow()
	}

	return key
}
