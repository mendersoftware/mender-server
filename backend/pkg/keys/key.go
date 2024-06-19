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
package keys

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"

	"github.com/pkg/errors"
)

const (
	ErrMsgPrivKeyReadFailed    = "failed to read server private key file"
	ErrMsgPrivKeyNotPEMEncoded = "server private key not PEM-encoded"
)

func LoadRSAPrivate(privKeyPath string) (*rsa.PrivateKey, error) {
	// read key from file
	pemData, err := ioutil.ReadFile(privKeyPath)
	if err != nil {
		return nil, errors.Wrap(err, ErrMsgPrivKeyReadFailed)
	}
	// decode pem key
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New(ErrMsgPrivKeyNotPEMEncoded)
	}
	// check if it is a RSA PRIVATE KEY
	if got, want := block.Type, "RSA PRIVATE KEY"; got != want {
		return nil, errors.Errorf(
			"unknown server private key type; got: %s, want: %s", got, want)
	}
	// return parsed key
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
