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
	"encoding/base64"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const (
	testEncryptionKey         = "passphrasewhichneedstobe32bytes!"
	testEncryptionFallbackKey = "anotherpassphrasewhichis32bytes!"
)

func TestSetEncryptionKeys(t *testing.T) {
	err := SetAESEncryptionKey("aa")
	assert.Error(t, err)
	assert.EqualError(t, err, ErrEncryptionKeyWrongLength.Error())

	err = SetAESEncryptionKey("%")
	assert.Error(t, err)
	assert.EqualError(t, err, "failed to base64-decode the AES encryption key: illegal base64 data at input byte 0")

	value := base64.RawStdEncoding.EncodeToString([]byte(testEncryptionKey))
	err = SetAESEncryptionKey(string(value))
	assert.NoError(t, err)
	assert.Equal(t, testEncryptionKey, encryptionKey)

	err = SetAESEncryptionFallbackKey("aa")
	assert.Error(t, err)
	assert.EqualError(t, err, ErrEncryptionKeyWrongLength.Error())

	err = SetAESEncryptionFallbackKey("%")
	assert.Error(t, err)
	assert.EqualError(t, err, "failed to base64-decode the AES encryption fallback key: illegal base64 data at input byte 0")

	value = base64.RawStdEncoding.EncodeToString([]byte(testEncryptionFallbackKey))
	err = SetAESEncryptionFallbackKey(value)
	assert.NoError(t, err)
	assert.Equal(t, testEncryptionFallbackKey, encryptionFallbackKey)
}

func TestEncryptDecrypt(t *testing.T) {
	testCases := []struct {
		Name        string
		Value       string
		Key         string
		FallbackKey string
		Err         error
	}{
		{
			Name: "ok, no keys",
		},
		{
			Name:  "ok, with encryption key",
			Value: "my data",
			Key:   testEncryptionKey,
		},
		{
			Name:        "ok, with encryption and fallback key",
			Value:       "my data",
			Key:         testEncryptionKey,
			FallbackKey: testEncryptionFallbackKey,
		},
		{
			Name: "ko, wrong key",
			Key:  "dummy",
			Err:  errors.New("crypto/aes: invalid key size 5"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			key := tc.Key
			if tc.FallbackKey != "" {
				key = tc.FallbackKey
			}
			encryptionKey = key
			out, err := AESEncrypt(tc.Value)
			if tc.Err != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.Err.Error())
				return
			}
			assert.NoError(t, err)

			encryptionKey = tc.Key
			encryptionFallbackKey = tc.FallbackKey
			decrypted, err := AESDecrypt(out)
			assert.NoError(t, err)
			assert.Equal(t, tc.Value, decrypted)
		})
	}
}

func TestDecryptErrCipherWrongKey(t *testing.T) {
	encryptionKey = testEncryptionKey
	out, _ := AESEncrypt("value")

	encryptionKey = "dummy"
	encryptionFallbackKey = ""
	_, err := AESDecrypt([]byte(out))
	assert.Error(t, err)
	assert.EqualError(t, err, "unable to decrypt the data: crypto/aes: invalid key size 5")
}

func TestDecryptErrCipherTextTooShort(t *testing.T) {
	encryptionKey = testEncryptionKey
	out, err := AESDecrypt([]byte("a"))
	assert.Equal(t, "", out)
	assert.EqualError(t, err, ErrCipherTextTooShort.Error())
}
