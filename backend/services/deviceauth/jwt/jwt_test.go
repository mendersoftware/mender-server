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
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewJWTHandler(t *testing.T) {
	testCases := map[string]struct {
		privateKeyPath string
		err            error
	}{
		"ok, pkcs1, rsa": {
			privateKeyPath: "./testdata/rsa.pem",
		},
		"ok, pkcs8, rsa": {
			privateKeyPath: "./testdata/rsa_pkcs8.pem",
		},
		"ok, pkcs8, ed25519": {
			privateKeyPath: "./testdata/ed25519.pem",
		},
		"ko": {
			privateKeyPath: "./testdata/doesnotexist.pem",
			err:            errors.New("failed to read private key: open ./testdata/doesnotexist.pem: no such file or directory"),
		},
		"unknown priate key type": {
			privateKeyPath: "./testdata/dsa.pem",
			err:            errors.New("unsupported server private key type"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			_, err := NewJWTHandler(tc.privateKeyPath)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
