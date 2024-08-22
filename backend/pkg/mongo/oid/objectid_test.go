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

package oid

import (
	"encoding/binary"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

type s struct {
	ID ObjectID
}

func TestUUID(t *testing.T) {
	// Test NewSHA1
	uuid5 := NewUUIDv5("foobar")
	assert.Equal(t, uuid5.Type(), TypeUUID)
	uuid5Eq := NewUUIDv5("foobar")
	assert.Equal(t, uuid5, uuid5Eq)

	// Test FromString
	uuidFromString := FromString(uuid5.String())
	assert.Equal(t, uuid5, uuidFromString)

	// Test NewRandom
	uRandom := NewUUIDv4()
	assert.NotEqual(t, uuid.Nil, uRandom.id)
	assert.NotEmpty(t, uRandom)

	// Test marshal/unmarshal bson
	bsonType, b, err := uRandom.MarshalBSONValue()
	assert.NoError(t, err)
	assert.Equal(t, bsontype.Binary, bsonType)

	// Illegal bsontype
	err = uRandom.UnmarshalBSONValue(bsontype.Undefined, b)
	assert.Error(t, err)
	// OK
	err = uRandom.UnmarshalBSONValue(bsonType, b)
	assert.NoError(t, err)
	// Invalid sub-type
	b[4] = bsontype.BinaryGeneric
	err = uRandom.UnmarshalBSONValue(bsonType, b)
	assert.Error(t, err)
	// Invalid length
	binary.LittleEndian.PutUint32(b, 10)
	err = uRandom.UnmarshalBSONValue(bsonType, b)
	assert.Error(t, err)

	// Test bson.Marshal / bson.Unmarshal with an embedded UUID
	s1 := s{ID: uRandom}
	var res s
	b, err = bson.Marshal(s1)
	assert.NoError(t, err)
	err = bson.Unmarshal(b, &res)
	assert.NoError(t, err)
	assert.Equal(t, s1, res)

	// Test Marshal / Unmarshal JSON
	b, err = json.Marshal(s1)
	assert.NoError(t, err)
	err = json.Unmarshal(b, &res)
	assert.NoError(t, err)
	assert.Equal(t, s1, res)
}

func TestBSONID(t *testing.T) {
	id := NewBSONID()
	assert.NotEmpty(t, id)
	assert.Equal(t, id.Type(), TypeBSONID)
	idEq := FromString(id.String())
	assert.Equal(t, id, idEq)

	// Test Marshal/Unmarshal BSON
	bsonType, b, err := id.MarshalBSONValue()
	assert.NoError(t, err)
	assert.Equal(t, bsontype.ObjectID, bsonType)
	assert.NotEmpty(t, b)

	// Unmarshal with wrong bsontype
	tmp := &ObjectID{}
	err = tmp.UnmarshalBSONValue(bsontype.Binary, b)
	assert.Error(t, err)
	// Unmarshal OK
	err = tmp.UnmarshalBSONValue(bsontype.ObjectID, b)
	assert.NoError(t, err)
	assert.Equal(t, id, *tmp)
	// Unmarshal malformed data
	err = tmp.UnmarshalBSONValue(bsontype.ObjectID, []byte{})
	assert.Error(t, err)

	// Marshal / Unmarshal JSON
	sct := s{ID: id}
	b, err = json.Marshal(sct)
	assert.NoError(t, err)
	var sctParsed s
	err = json.Unmarshal(b, &sctParsed)
	assert.NoError(t, err)
	assert.Equal(t, sct, sctParsed)
}

func TestString(t *testing.T) {
	str := ObjectID{id: "foobar"}
	assert.Equal(t, str.Type(), TypeString)
	assert.Equal(t, str.String(), "foobar")

	null := ObjectID{}
	assert.Equal(t, null.Type(), TypeNil)

	fromStrBSONIDLen := FromString("__ObjectIDLengthString__")
	assert.Equal(t, fromStrBSONIDLen.Type(), TypeString)
	fromStrUUIDLen := FromString("ThisStringHasSameLengthAsAUUIDString")
	assert.Equal(t, fromStrUUIDLen.Type(), TypeString)

	// Marshal / Unmarshal BSON
	tmp := &ObjectID{}
	bsonType, b, err := str.MarshalBSONValue()
	assert.NoError(t, err)
	assert.Equal(t, bsontype.String, bsonType)
	err = tmp.UnmarshalBSONValue(bsontype.Undefined, b)
	assert.Error(t, err)
	err = tmp.UnmarshalBSONValue(bsonType, b)
	assert.NoError(t, err)
	assert.Equal(t, str, *tmp)
	// corrupt length field
	b[0] = 0xFF
	err = tmp.UnmarshalBSONValue(bsonType, b)
	assert.Error(t, err)

	// Marshal / Unmarshal BSON as embedded field
	var sctParsed s
	sct := s{ID: str}
	b, err = bson.Marshal(sct)
	assert.NoError(t, err)
	err = bson.Unmarshal(b, &sctParsed)
	assert.NoError(t, err)
	assert.Equal(t, sct, sctParsed)

	// Marshal / Unmarshal JSON
	b, err = json.Marshal(sct)
	assert.NoError(t, err)
	err = json.Unmarshal(b, &sctParsed)
	assert.NoError(t, err)
	assert.Equal(t, sct, sctParsed)
}

func TestNull(t *testing.T) {
	null := ObjectID{}
	assert.Equal(t, null.String(), "")
	_, _, err := null.MarshalBSONValue()
	assert.Error(t, err)
}
