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
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMarshalBSON(t *testing.T) {
	encryptionKey = ""

	value := String("random value")
	marshalled, err := value.MarshalBSON()
	assert.NoError(t, err)
	assert.NotNil(t, marshalled)
	assert.Greater(t, len(marshalled), 0)

	anotherValue := String("")
	err = anotherValue.UnmarshalBSON(marshalled)
	assert.NoError(t, err)
	assert.Equal(t, value, anotherValue)

	encryptionKey = "dummy"
	_, err = value.MarshalBSON()
	assert.Error(t, err)
	assert.EqualError(t, err, "crypto/aes: invalid key size 5")
}

func TestMarshalBSONWithEncryptionKey(t *testing.T) {
	encryptionKey = testEncryptionKey

	value := String("random value")
	marshalled, err := value.MarshalBSON()
	assert.NoError(t, err)
	assert.NotNil(t, marshalled)
	assert.Greater(t, len(marshalled), 0)

	anotherValue := String("")
	err = anotherValue.UnmarshalBSON(marshalled)
	assert.NoError(t, err)
	assert.Equal(t, value, anotherValue)

	err = anotherValue.UnmarshalBSON([]byte("abc"))
	assert.Error(t, err)
	assert.EqualError(t, err, "EOF")
}

func TestUnmarshalBSONUnknownAlgorithm(t *testing.T) {
	data, _ := bson.Marshal(&stringBSON{
		Algorithm: "dummy",
		Data:      []byte{},
	})

	anotherValue := String("")
	err := anotherValue.UnmarshalBSON(data)
	assert.Error(t, err)
	assert.EqualError(t, err, ErrUnknownAlgorithm.Error())
}

func TestMarshalText(t *testing.T) {
	cryptoString := String("value")

	s, _ := cryptoString.MarshalText()
	assert.Equal(t, []byte(omitted), s)
	assert.NotContains(t, s, "value")
}
