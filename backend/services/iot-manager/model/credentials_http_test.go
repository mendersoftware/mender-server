// Copyright 2022 Northern.tech AS
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
	"net"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"
)

func TestHTTPCredentials(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	const (
		TestCredentials  = `{"url":"http://mender.io","secret":"deadbeef12345678"}`
		TestInvalidCreds = `{"url":"http://localhost","secret":"nothex"}`
	)
	var creds HTTPCredentials
	err := json.Unmarshal([]byte(TestCredentials), &creds)
	if !assert.NoError(t, err) {
		return
	}
	creds.validateAddr = true

	err = creds.Validate()
	assert.NoError(t, err)

	creds.URL = "http://localhost"
	err = creds.Validate()
	var errs validation.Errors
	if assert.ErrorAs(t, err, &errs) {
		var netErr net.InvalidAddrError
		for _, err := range errs {
			assert.ErrorAs(t, err, &netErr)
		}
	}

	creds.URL = "https://%%%"
	err = creds.Validate()
	assert.Error(t, err)

	creds.URL = "http://"
	err = creds.Validate()
	if assert.ErrorAs(t, err, &errs) {
		var dnsError *net.DNSError
		for _, err := range errs {
			assert.ErrorAs(t, err, &dnsError)
		}
	}
	// Unmarshal invalid hex string
	err = json.Unmarshal([]byte(TestInvalidCreds), &creds)
	assert.Error(t, err)
}

func TestHexSecret(t *testing.T) {
	t.Parallel()

	hexSecret := HexSecret("shhh... this is a secret")
	secret, _ := hexSecret.MarshalText()
	assert.Equal(t, []byte("<omitted>"), secret)

	b, err := hexSecret.MarshalBSON()
	assert.NoError(t, err)
	var actualSecret HexSecret
	err = actualSecret.UnmarshalBSON(b)
	assert.NoError(t, err)
	assert.Equal(t, hexSecret, actualSecret)
}
