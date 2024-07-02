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

package mongo

import (
	"reflect"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
)

var (
	tUUID = reflect.TypeOf(uuid.UUID{})
)

func init() {
	// Add UUID encoder/decoder for github.com/google/uuid.UUID
	bson.DefaultRegistry = bson.NewRegistry()
	bson.DefaultRegistry.RegisterTypeEncoder(tUUID, bsoncodec.ValueEncoderFunc(uuidEncodeValue))
	bson.DefaultRegistry.RegisterTypeDecoder(tUUID, bsoncodec.ValueDecoderFunc(uuidDecodeValue))
}

func uuidEncodeValue(ec bsoncodec.EncodeContext, w bsonrw.ValueWriter, val reflect.Value) error {
	if !val.IsValid() || val.Type() != tUUID {
		return bsoncodec.ValueEncoderError{
			Name:     "UUIDEncodeValue",
			Types:    []reflect.Type{tUUID},
			Received: val,
		}
	}
	uid := val.Interface().(uuid.UUID)
	return w.WriteBinaryWithSubtype(uid[:], byte(bson.TypeBinaryUUID))
}

func uuidDecodeValue(ec bsoncodec.DecodeContext, r bsonrw.ValueReader, val reflect.Value) error {
	if !val.CanSet() || val.Type() != tUUID {
		return bsoncodec.ValueDecoderError{
			Name:     "UUIDDecodeValue",
			Types:    []reflect.Type{tUUID},
			Received: val,
		}
	}

	var (
		data    []byte
		err     error
		subtype byte
		uid     uuid.UUID = uuid.Nil
	)
	switch rType := r.Type(); rType {
	case bson.TypeBinary:
		data, subtype, err = r.ReadBinary()
		switch subtype {
		case bson.TypeBinaryGeneric:
			if len(data) != 16 {
				return errors.Errorf(
					"cannot decode %v as a UUID: "+
						"incorrect length: %d",
					data, len(data),
				)
			}

			fallthrough
		case bson.TypeBinaryUUID, bson.TypeBinaryUUIDOld:
			copy(uid[:], data)

		default:
			err = errors.Errorf(
				"cannot decode %v as a UUID: "+
					"incorrect subtype 0x%02x",
				data, subtype,
			)
		}

	case bson.TypeUndefined:
		err = r.ReadUndefined()

	case bson.TypeNull:
		err = r.ReadNull()

	default:
		err = errors.Errorf("cannot decode %v as a UUID", rType)
	}

	if err != nil {
		return err
	}
	val.Set(reflect.ValueOf(uid))
	return nil
}
