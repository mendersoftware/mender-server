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

package storage

import (
	"crypto/x509"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv(t *testing.T) {
	res := getEnv(storageBackendCertEnvName)
	assert.Equal(t, os.Getenv(storageBackendCertEnvName), res)
}

const (
	cert = `-----BEGIN CERTIFICATE-----
MIICIjCCAccCFA1IUl+HvFgbHdfsszblPyqMc62cMAoGCCqGSM49BAMCMIGOMQ4w
DAYDVQQDDAVNeSBDQTEYMBYGA1UECgwPTXkgT3JnYW5pemF0aW9uMRAwDgYDVQQL
DAdNeSBVbml0MSUwIwYJKoZIhvcNAQkBFhZteXVzZXJuYW1lQGV4YW1wbGUuY29t
MQswCQYDVQQGEwJOTzENMAsGA1UEBwwET3NsbzENMAsGA1UECAwET3NsbzAeFw0y
MTA2MjAwOTI5NDhaFw0zMTA2MTgwOTI5NDhaMIGWMRYwFAYDVQQDDA1teS1zZXJ2
ZXIuY29tMRgwFgYDVQQKDA9NeSBPcmdhbml6YXRpb24xEDAOBgNVBAsMB015IFVu
aXQxJTAjBgkqhkiG9w0BCQEWFm15dXNlcm5hbWVAZXhhbXBsZS5jb20xCzAJBgNV
BAYTAk5PMQ0wCwYDVQQHDARPc2xvMQ0wCwYDVQQIDARPc2xvMFkwEwYHKoZIzj0C
AQYIKoZIzj0DAQcDQgAEC7ebufzn6gPbsXfnGfLultQOJKkP+9o5UITLeGwX2ENJ
cCpC1gH6uNyBuM3kWzZcXW4of8uTSyF9zM384SlGLzAKBggqhkjOPQQDAgNJADBG
AiEAvPewBMFu0zLeFbIlI/O7qyvSHQY3p4Wll9XQasWdE8sCIQCneHR4lo4cwknI
G8WhQ/MyAyqLjNEgi5d0K3cIb0Xc6g==
-----END CERTIFICATE-----`
)

func TestGetRootCAs(t *testing.T) {
	rootCAs, _ := x509.SystemCertPool()
	systemCerts := len(rootCAs.Subjects())

	testCases := map[string]struct {
		certPEM string
		certCN  string
	}{
		"no custom cert": {
			certPEM: "",
		},
		"custom cert": {
			certPEM: cert,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			oldGetEnv := getEnv
			defer func() {
				getEnv = oldGetEnv
			}()
			getEnv = func(key string) string {
				return tc.certPEM
			}
			deploymentsRootCAs := GetRootCAs()
			assert.NotNil(t, deploymentsRootCAs)

			if tc.certPEM == "" {
				assert.Equal(t, systemCerts, len(deploymentsRootCAs.Subjects()))
			} else {
				assert.Less(t, systemCerts, len(deploymentsRootCAs.Subjects()))
			}
		})
	}
}
