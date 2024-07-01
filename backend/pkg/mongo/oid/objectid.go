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

// Package oid contains wrappers for creating bson ObjectIDs that marshals
// correctly according to the bson specification: http://bsonspec.org/spec.html
package oid

import (
	"bytes"
	"encoding/binary"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Type int

const (
	TypeNil Type = iota
	TypeUUID
	TypeBSONID
	TypeString
)

// Size of a UUID in bytes.
const Size = 16

// ObjectID implements a bson compatible fusion of bson ObjectID, UUID and
// string types. Depending on the initialization, this type marshals to the
// correct bson type (and sub-type) for each type.
type ObjectID struct {
	id interface{}
}

// NewBSONID initialize a new bson-type ObjectID (used by default by mongo)
func NewBSONID() ObjectID {
	return ObjectID{id: primitive.NewObjectID()}
}

// NewUUIDv4 creates a new ObjectID initialized with a UUID v4 (random).
// In the rare event that the RNG returns an error, the null UUID is returned.
func NewUUIDv4() ObjectID {
	uid, _ := uuid.NewRandom()
	return ObjectID{id: uid}
}

// NewUUIDv5 returns a new version 5 uuid in the objectID namespace
func NewUUIDv5(id string) ObjectID {
	ret := uuid.NewSHA1(uuid.NameSpaceOID, []byte(id))
	return ObjectID{id: ret}
}

func fromString(id string) interface{} {
	var ret interface{}
	switch len(id) {
	case 24: // Hex-encoded bson-type objectID type.
		mgoID, e := primitive.ObjectIDFromHex(id)
		if e != nil {
			// Fall back on assigning the raw string if value does
			// not comply.
			ret = id
		} else {
			ret = mgoID
		}

	case 32, 36, 38, 41, 45: // All valid hex-encoded uuid formats.
		uid, e := uuid.Parse(id)
		if e != nil {
			// Fall back on using the string.
			ret = id
		} else {
			ret = uid
		}

	default:
		ret = id
	}
	return ret
}

// FromString tries to parse a hex-encoded mongo ObjectID/UUID string, and
// returns an error if the string is not a valid UUID.
func FromString(id string) ObjectID {
	return ObjectID{id: fromString(id)}
}

func marshalBSONUUID(u uuid.UUID) (bsontype.Type, []byte, error) {
	buf := make([]byte, Size+4+1)
	binary.LittleEndian.PutUint32(buf, Size)
	buf[4] = bsontype.BinaryUUID
	copy(buf[5:], u[:])
	return bsontype.Binary, buf, nil

}

// MarshalBSONValue provides the bson.ValueMarshaler interface.
func (oid ObjectID) MarshalBSONValue() (bsontype.Type, []byte, error) {
	switch objectID := oid.id.(type) {
	case uuid.UUID:
		return marshalBSONUUID(objectID)
	case primitive.ObjectID:
		return bson.MarshalValue(objectID)
	case string:
		return bson.MarshalValue(objectID)
	}
	return bsontype.Null, nil, errors.New(
		"unable to marshal ObjectID: not initialized",
	)
}

// UnmarshalBSONValue provides the bson.ValueUnmarshaler interace.
func (oid *ObjectID) UnmarshalBSONValue(t bsontype.Type, b []byte) error {
	var err error
	switch t {
	case bsontype.Binary: // Assume UUID type
		l := binary.LittleEndian.Uint32(b)
		if l != Size {
			return errors.Errorf("illegal uuid length: %d", l)
		}
		if b[4] != bsontype.BinaryUUID {
			return errors.Errorf(
				"illegal bson sub-type: %0x02X, expected: 0x%02X",
				b[4], bsontype.BinaryUUID,
			)
		}
		oid.id, err = uuid.FromBytes(b[5:])
	case bsontype.ObjectID:
		if len(b) != 12 {
			return errors.Errorf(
				"illegal objectID length: %d",
				len(b),
			)
		}
		bsonID := primitive.ObjectID{}
		copy(bsonID[:], b)
		oid.id = bsonID
		return nil
	case bsontype.String:
		l := binary.LittleEndian.Uint32(b)
		strLen := len(b) - 4
		if int(l) != strLen {
			return errors.Errorf(
				"illegal string length; buffer length: "+
					"%d != header length: %d",
				l, strLen,
			)
		}
		// Length include terminating zero-byte
		buf := make([]byte, l-1)
		copy(buf, b[4:])
		oid.id = string(buf)
		return nil
	default:
		return errors.Errorf(
			"illegal bson-type %s, expected ObjectID", t,
		)
	}
	return err
}

// MarshalJSON ensures the ObjectID marhsals correctly.
func (oid ObjectID) MarshalJSON() ([]byte, error) {
	// All supported types provides json.Marshaller interface.
	return json.Marshal(oid.id)
}

// UnmarshalJSON unmarshal string-type json to the appropriate ObjectID type.
func (oid *ObjectID) UnmarshalJSON(b []byte) error {
	b = bytes.Trim(b, `"`)
	oid.id = fromString(string(b))
	return nil
}

func (oid ObjectID) String() string {
	switch id := oid.id.(type) {
	case uuid.UUID:
		return id.String()
	case primitive.ObjectID:
		return id.Hex()
	case string:
		return id
	default:
		return ""
	}
}

func (oid ObjectID) Type() Type {
	switch oid.id.(type) {
	case uuid.UUID:
		return TypeUUID
	case primitive.ObjectID:
		return TypeBSONID
	case string:
		return TypeString
	default:
		return TypeNil
	}
}
