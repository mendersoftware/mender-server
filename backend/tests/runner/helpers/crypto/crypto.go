// Copyright 2026 Northern.tech AS
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
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

func GenerateRSAKeypair() (privPEM, pubPEM string, err error) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return "", "", err
	}
	privBytes := x509.MarshalPKCS1PrivateKey(key)
	privBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes}
	privPEM = string(pem.EncodeToMemory(privBlock))

	pubBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return "", "", err
	}
	pubBlock := &pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}
	pubPEM = string(pem.EncodeToMemory(pubBlock))

	return privPEM, pubPEM, nil
}

func GenerateECKeypair(curve elliptic.Curve) (privPEM, pubPEM string, err error) {
	key, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return "", "", err
	}
	privBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return "", "", err
	}
	privBlock := &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes}
	privPEM = string(pem.EncodeToMemory(privBlock))

	pubBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return "", "", err
	}
	pubBlock := &pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}
	pubPEM = string(pem.EncodeToMemory(pubBlock))

	return privPEM, pubPEM, nil
}

func GenerateEd25519Keypair() (privPEM, pubPEM string, err error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return "", "", err
	}
	privBlock := &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}
	privPEM = string(pem.EncodeToMemory(privBlock))

	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "", "", err
	}
	pubBlock := &pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}
	pubPEM = string(pem.EncodeToMemory(pubBlock))

	return privPEM, pubPEM, nil
}

func SignAuthRequest(data []byte, privKeyPEM string) (string, error) {
	block, _ := pem.Decode([]byte(privKeyPEM))
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block")
	}

	var sig []byte

	switch block.Type {
	case "RSA PRIVATE KEY":
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return "", err
		}
		hash := sha256.Sum256(data)
		sig, err = rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash[:])
		if err != nil {
			return "", err
		}

	case "EC PRIVATE KEY":
		key, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return "", err
		}
		hash := sha256.Sum256(data)
		sig, err = ecdsa.SignASN1(rand.Reader, key, hash[:])
		if err != nil {
			return "", err
		}

	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return "", err
		}
		switch k := key.(type) {
		case ed25519.PrivateKey:
			sig = ed25519.Sign(k, data)
		case *rsa.PrivateKey:
			hash := sha256.Sum256(data)
			sig, err = rsa.SignPKCS1v15(rand.Reader, k, crypto.SHA256, hash[:])
			if err != nil {
				return "", err
			}
		case *ecdsa.PrivateKey:
			hash := sha256.Sum256(data)
			sig, err = ecdsa.SignASN1(rand.Reader, k, hash[:])
			if err != nil {
				return "", err
			}
		default:
			return "", fmt.Errorf("unsupported PKCS8 key type: %T", key)
		}

	default:
		return "", fmt.Errorf("unsupported PEM block type: %s", block.Type)
	}

	return base64.StdEncoding.EncodeToString(sig), nil
}
