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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	SetTrustedHostnames([]string{"localhost", "totally.legit", "*.azure-devices.net", "*.iot.*.amazonaws.com"})
	lookupHostFunc = func(hostname string) error {
		return nil
	}
}

func TestParseConnectionString(t *testing.T) {
	cs, err := ParseConnectionString(
		"HostName=mender-test-hub.azure-devices.net;DeviceId=7b478313-de33-4735-bf00-0ebc31851faf;" +
			"SharedAccessKey=YWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWE=")
	assert.NoError(t, err)
	isZero := cs.IsZero()
	assert.False(t, isZero)
	err = cs.Validate()
	assert.NoError(t, err)

	cs, err = ParseConnectionString(
		"HostName=mender-test-hub.azure-devices.net;DeviceId=7b478313-de33-4735-bf00-0ebc31851faf;" +
			"SharedAccessKey=YWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWE=")
	assert.NoError(t, err)
	isZero = cs.IsZero()
	assert.False(t, isZero)
	err = cs.Validate()
	assert.NoError(t, err)

	_, err = ParseConnectionString("abc")
	assert.EqualError(t, err, "invalid connectionstring format")

	_, err = ParseConnectionString(
		"HostName=test-hub.bad-devices.net;DeviceId=7b478313-de33-4735-bf00-0ebc31851faf;" +
			"SharedAccessKey=YWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWE=")
	if assert.Error(t, err) {
		assert.Contains(t,
			err.Error(),
			"hostname does not refer to a trusted domain",
		)
	}
}

func TestConnectionStringAuthorization(t *testing.T) {
	cs, err := ParseConnectionString(
		"HostName=mender-test-hub.azure-devices.net;DeviceId=7b478313-de33-4735-bf00-0ebc31851faf;" +
			"SharedAccessKey=YWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWE=")
	assert.NoError(t, err)

	const layout = "Jan 2, 2006 at 3:04pm (MST)"
	tm, _ := time.Parse(layout, "Feb 4, 2014 at 6:05pm (PST)")
	token := cs.Authorization(tm)
	assert.Equal(t, "SharedAccessSignature sr=mender-test-hub.azure-devices.net&sig=PJYYfRCuL4bo5%2BCc%2Flj1L%2F4AShbEuisEcwGiK90fqYk%3D&se=1391537100", token)
}

func TestConnectionStringMarshalText(t *testing.T) {
	cs := &ConnectionString{}
	err := cs.UnmarshalText([]byte(
		"HostName=mender-test-hub.azure-devices.net;DeviceId=7b478313-de33-4735-bf00-0ebc31851faf;" +
			"SharedAccessKey=YWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWE="))
	assert.NoError(t, err)

	marshalled, err := cs.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, "HostName=mender-test-hub.azure-devices.net;DeviceId=7b478313-de33-4735-bf00-0ebc31851faf;SharedAccessKey=YWFh...<omitted>", string(marshalled))
}

func TestHostnameValidator(t *testing.T) {
	t.Parallel()
	valLocalhost := newHostnameValidator([]string{"localhost", ""})
	valEmpty := newHostnameValidator(nil)

	err := valEmpty.Validate("any.url.io")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "[PROG ERR(hostnameValidator)]")
	}

	err = valLocalhost.Validate(nil)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "[PROG ERR(hostnameValidator)]")
	}

	err = valLocalhost.Validate("")
	assert.ErrorIs(t, err, ErrHostnameTrust)
}
