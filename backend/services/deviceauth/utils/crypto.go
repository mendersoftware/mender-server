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

package utils

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"math/big"

	"github.com/pkg/errors"
)

const (
	ErrMsgVerify = "verification failed"

	//PEM identifier of an RSA public key, needed for decoding
	//key content from a string
	PubKeyBlockType = "PUBLIC KEY"
)

// VerifyAuthReqSign verifies a SHA256 digested signature for a given public
// key. The current asymmetric crypto algorithms supported are RSA (PKCS 1.5
// signature), ED25519 and ECDSA (DSA is considered obsolete). The signature
// is fixed to the SHA256 hash algorithm.
func VerifyAuthReqSign(signature string, pubkey crypto.PublicKey, content []byte) error {
	var digest []byte = content

	decodedSig, err := base64.StdEncoding.DecodeString(string(signature))
	if err != nil {
		return errors.Wrap(err, ErrMsgVerify)
	}

	switch pubkey.(type) {
	case *rsa.PublicKey, *ecdsa.PublicKey:
		hash := sha256.Sum256(content)
		digest = hash[:]
	}

	switch key := pubkey.(type) {
	case *rsa.PublicKey:
		err = rsa.VerifyPKCS1v15(
			key, crypto.SHA256, digest, decodedSig,
		)
		if err != nil {
			return errors.Wrap(err, ErrMsgVerify)
		}

	case *ecdsa.PublicKey:
		var signInts struct {
			R *big.Int
			S *big.Int
		}
		_, err := asn1.Unmarshal(decodedSig, &signInts)
		if err != nil {
			return errors.Wrap(err, ErrMsgVerify)
		}
		valid := ecdsa.Verify(key, digest, signInts.R, signInts.S)
		if !valid {
			return errors.New(ErrMsgVerify)
		}

	case ed25519.PublicKey:
		valid := ed25519.Verify(key, digest, decodedSig)
		if !valid {
			return errors.New(ErrMsgVerify)
		}

	default:
		return errors.Wrap(errors.Errorf(
			"public key algorithm (%T) not supported", pubkey,
		), ErrMsgVerify)
	}

	return nil
}

// ParsePubKey
func ParsePubKey(pubkey string) (crypto.PublicKey, error) {
	block, _ := pem.Decode([]byte(pubkey))
	if block == nil || block.Type != PubKeyBlockType {
		return nil, errors.New("cannot decode public key")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "cannot decode public key")
	}

	return key, nil
}

func SerializePubKey(key crypto.PublicKey) (string, error) {

	switch key.(type) {
	case *rsa.PublicKey, *ecdsa.PublicKey, ed25519.PublicKey:
		break
	default:
		return "", errors.New("unrecognized public key type")
	}

	asn1, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return "", err
	}

	out := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1,
	})

	return string(out), nil
}
