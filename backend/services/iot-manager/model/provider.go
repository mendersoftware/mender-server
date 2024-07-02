// Copyright 2023 Northern.tech AS
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
	"crypto"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Provider string

const (
	ProviderEmpty   Provider = ""
	ProviderIoTHub  Provider = "iot-hub"
	ProviderIoTCore Provider = "iot-core"
	ProviderWebhook Provider = "webhook"
)

var validateProvider = validation.In(ProviderIoTHub, ProviderIoTCore, ProviderWebhook)

func (p Provider) Validate() error {
	return validateProvider.Validate(p)
}

// All crypto.PublicKey from standard library implements this interface:
type publicKeyExt interface {
	crypto.PublicKey
	Equal(crypto.PublicKey) bool
}

type PublicKey struct {
	PublicKey publicKeyExt
}

const (
	blockTypePublicKey      = "PUBLIC KEY"
	blockTypePKCS1PublicKey = "RSA PUBLIC KEY"
)

func (p *PublicKey) UnmarshalJSON(b []byte) error {
	var keyString string
	err := json.Unmarshal(b, &keyString)
	if err != nil {
		return err
	}
	return p.UnmarshalText([]byte(keyString))
}

func (p *PublicKey) UnmarshalText(b []byte) error {
	block, _ := pem.Decode(b)
	var (
		pub crypto.PublicKey
		err error
	)
	if block == nil {
		return fmt.Errorf("invalid public key format")
	}
	switch block.Type {
	case blockTypePublicKey:
		pub, err = x509.ParsePKIXPublicKey(block.Bytes)
	case blockTypePKCS1PublicKey:
		pub, err = x509.ParsePKCS1PublicKey(block.Bytes)
	default:
		err = fmt.Errorf("invalid PEM block type %q", block.Type)
	}
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}
	var ok bool
	p.PublicKey, ok = pub.(publicKeyExt)
	if !ok {
		return errors.New("public key type not supported")
	}
	return nil
}

type PreauthRequest struct {
	DeviceID     string                 `json:"external_id"`
	IdentityData map[string]interface{} `json:"id_data"`
	PublicKey    PublicKey              `json:"pubkey"`
}
