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
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"io"
	"strings"

	"github.com/pkg/errors"
)

const V1 byte = 0x1
const encryptionKeyLength = 32
const hmacSHA512SignatureSizeInBytes = 64

var (
	encryptionKey         = ""
	encryptionFallbackKey = ""

	ErrDecryptionFailed         = errors.New("unable to decrypt the data")
	ErrCipherTextTooShort       = errors.New("cipher text is too short")
	ErrEncryptionKeyWrongLength = errors.New(
		"AES encryption key has a wrong length, expected 32 bytes")

	// Encode base64 secret in either std or URL encoding ignoring padding.
	base64Repl = strings.NewReplacer("-", "+", "_", "/", "=", "")
)

func SetAESEncryptionKey(key string) error {
	if key, err := base64.RawStdEncoding.DecodeString(base64Repl.Replace(key)); err == nil {
		if len(key) > 0 && len(key) != encryptionKeyLength {
			return ErrEncryptionKeyWrongLength
		}
		encryptionKey = string(key)
	} else {
		return errors.Wrap(err, "failed to base64-decode the AES encryption key")
	}
	return nil
}

func SetAESEncryptionFallbackKey(key string) error {
	if key, err := base64.RawStdEncoding.DecodeString(base64Repl.Replace(key)); err == nil {
		if len(key) > 0 && len(key) != encryptionKeyLength {
			return ErrEncryptionKeyWrongLength
		}
		encryptionFallbackKey = string(key)
	} else {
		return errors.Wrap(err, "failed to base64-decode the AES encryption fallback key")
	}
	return nil
}

func AESEncrypt(data string) ([]byte, error) {
	return encrypt(data, encryptionKey)
}

func AESDecrypt(data []byte) (string, error) {
	out, err := decrypt(data, encryptionKey)
	if err == ErrDecryptionFailed {
		out, err = decrypt(data, encryptionFallbackKey)
	}
	return out, err
}

func encrypt(data string, key string) ([]byte, error) {
	if len(key) == 0 {
		return []byte(data), nil
	}
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	dataToEncrypt := []byte(data)
	cipherText := make([]byte, aes.BlockSize+len(dataToEncrypt))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], dataToEncrypt)
	// AES and SHA-1 (or SHA-256) are "sufficiently different" that there
	// should be no practical issue with using the same key for AES and HMAC/SHA-*
	// see https://crypto.stackexchange.com/questions/8081/using-the-same-secret-
	// key-for-encryption-and-authentication-in-a-encrypt-then-ma/8086#8086
	HMAC := hmac.New(sha512.New, []byte(key))
	if _, err := HMAC.Write(dataToEncrypt); err != nil {
		return nil, err
	}
	result := append([]byte{V1}, append(cipherText, HMAC.Sum(nil)...)...)
	return result, nil
}

func decrypt(data []byte, key string) (string, error) {
	if len(key) == 0 {
		return string(data), nil
	} else if len(data) < aes.BlockSize {
		return "", ErrCipherTextTooShort
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", errors.Wrap(err, ErrDecryptionFailed.Error())
	}
	if data[0] != V1 {
		return "", ErrDecryptionFailed
	}
	signature := data[len(data)-hmacSHA512SignatureSizeInBytes:]
	data = data[1 : len(data)-hmacSHA512SignatureSizeInBytes]
	iv := make([]byte, aes.BlockSize)
	copy(iv, data[:aes.BlockSize])
	cipherData := make([]byte, len(data)-aes.BlockSize)
	copy(cipherData, data[aes.BlockSize:])
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherData, cipherData)
	cipherText := string(cipherData)
	//
	HMAC := hmac.New(sha512.New, []byte(key))
	if _, err := HMAC.Write(cipherData); err != nil {
		return "", err
	}
	if string(signature) != string(HMAC.Sum(nil)) {
		return "", ErrDecryptionFailed
	}
	return cipherText, err
}
