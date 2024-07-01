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
package testing

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"testing"

	"github.com/pkg/errors"
)

func AuthReqSign(data []byte, privkey crypto.PrivateKey, t *testing.T) []byte {

	var (
		digest    []byte
		signature []byte
		err       error
	)

	switch privkey.(type) {
	case *rsa.PrivateKey, *ecdsa.PrivateKey:
		hash := sha256.New()
		if _, err = bytes.NewReader(data).WriteTo(hash); err != nil {
			t.Fatal(err)
		}
		digest = hash.Sum(nil)

	case ed25519.PrivateKey:
		digest = data

	default:
		panic("private key not understood")
	}
	switch pkey := privkey.(type) {
	case *rsa.PrivateKey:
		signature, err = rsa.SignPKCS1v15(
			rand.Reader, pkey, crypto.SHA256, digest,
		)
		if err != nil {
			panic(err)
		}
	case *ecdsa.PrivateKey:
		r, s, err := ecdsa.Sign(rand.Reader, pkey, digest)
		if err != nil {
			panic(err)
		}
		signature, err = asn1.Marshal([]*big.Int{r, s})
		if err != nil {
			panic(err)
		}
	case ed25519.PrivateKey:
		signature = ed25519.Sign(pkey, digest)
	}

	b64 := make([]byte, base64.StdEncoding.EncodedLen(len(signature)))
	base64.StdEncoding.Encode(b64, signature)

	return b64
}

func LoadPrivKey(path string) crypto.PrivateKey {
	PEMData, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return DecodePrivKey(PEMData)
}

func DecodePrivKey(priv []byte) crypto.PrivateKey {
	block, _ := pem.Decode(priv)

	if block == nil {
		panic("error decoding private key (empty PEM block)")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		var e error
		key, e = x509.ParsePKCS1PrivateKey(block.Bytes)
		if e != nil {
			panic(err)
		}
	}
	return key
}

func LoadPubKeyStr(path string) string {
	pem_data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return string(pem_data)
}

func DecodePubKey(pub []byte) crypto.PublicKey {
	block, _ := pem.Decode(pub)
	if block == nil {
		panic("error decoding private key (empty PEM block)")
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(errors.Wrapf(err, "PEM: %s", string(pub)))
	}
	return pubKey
}
