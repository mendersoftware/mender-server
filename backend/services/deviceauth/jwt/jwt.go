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
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"

	"github.com/pkg/errors"
)

var (
	ErrTokenExpired = errors.New("jwt: token expired")
	ErrTokenInvalid = errors.New("jwt: token invalid")
)

const (
	pemHeaderPKCS1 = "RSA PRIVATE KEY"
	pemHeaderPKCS8 = "PRIVATE KEY"
)

// Handler jwt generator/verifier
//
//go:generate ../utils/mockgen.sh
type Handler interface {
	ToJWT(t *Token) (string, error)
	// FromJWT parses the token
	// returns:
	// ErrTokenInvalid when the token is invalid (malformed, missing required claims, etc.)
	FromJWT(string) (*Token, error)
	// Validate does basic validity checks (Claims.Valid()).
	// returns:
	// ErrTokenExpired when the token is valid but expired
	// ErrTokenInvalid when the token is invalid (malformed, missing required claims, etc.)
	Validate(string) error
}

func NewJWTHandler(privateKeyPath string) (Handler, error) {
	priv, err := os.ReadFile(privateKeyPath)
	block, _ := pem.Decode(priv)
	if block == nil {
		return nil, errors.Wrap(err, "failed to read private key")
	}
	switch block.Type {
	case pemHeaderPKCS1:
		privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read rsa private key")
		}
		return NewJWTHandlerRS256(privKey), nil
	case pemHeaderPKCS8:
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read private key")
		}
		switch v := key.(type) {
		case *rsa.PrivateKey:
			return NewJWTHandlerRS256(v), nil
		case ed25519.PrivateKey:
			return NewJWTHandlerEd25519(&v), nil
		}
	}
	return nil, errors.Errorf("unsupported server private key type")
}
