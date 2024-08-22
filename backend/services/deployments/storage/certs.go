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
)

const storageBackendCertEnvName = "STORAGE_BACKEND_CERT"

var getEnv = func(key string) string {
	return os.Getenv(key)
}

func GetRootCAs() *x509.CertPool {
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	if storageBackendCert := getEnv(storageBackendCertEnvName); storageBackendCert != "" {
		rootCAs.AppendCertsFromPEM([]byte(storageBackendCert))
	}
	return rootCAs
}
