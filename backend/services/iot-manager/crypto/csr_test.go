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

package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPrivateKey(t *testing.T) {
	key, err := NewPrivateKey()
	assert.NoError(t, err)

	keyPem := string(PrivateKeyToPem(key))
	assert.Contains(t, keyPem, "-----BEGIN PRIVATE KEY-----")
	assert.Contains(t, keyPem, "-----END PRIVATE KEY-----")
}

func TestNewCertificateSigningRequest(t *testing.T) {
	key, err := NewPrivateKey()
	assert.NoError(t, err)

	csr, err := NewCertificateSigningRequest("dummy", key)
	assert.NoError(t, err)
	assert.Contains(t, string(csr), "-----BEGIN CERTIFICATE REQUEST-----")
	assert.Contains(t, string(csr), "-----END CERTIFICATE REQUEST-----")
}
