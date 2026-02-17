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

package codec

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	tUUID = reflect.TypeOf(uuid.UUID{})
)

func NewRegistry() *bson.Registry {
	reg := bson.NewRegistry()
	// Add UUID encoder/decoder for github.com/google/uuid.UUID
	uuidCodec := UUIDCodec{}
	reg.RegisterTypeEncoder(tUUID, uuidCodec)
	reg.RegisterTypeDecoder(tUUID, uuidCodec)
	return reg
}

type UUIDCodec struct{}

func (UUIDCodec) EncodeValue(
	ec bson.EncodeContext,
	w bson.ValueWriter,
	val reflect.Value,
) error {
	if !val.IsValid() || val.Type() != tUUID {
		return bson.ValueEncoderError{
			Name:     "UUIDCodec",
			Types:    []reflect.Type{tUUID},
			Received: val,
		}
	}
	uid := val.Interface().(uuid.UUID)
	return w.WriteBinaryWithSubtype(uid[:], bson.TypeBinaryUUID)
}

func (UUIDCodec) DecodeValue(
	ec bson.DecodeContext,
	r bson.ValueReader,
	val reflect.Value,
) error {
	if !val.CanSet() || val.Type() != tUUID {
		return bson.ValueDecoderError{
			Name:     "UUIDCodec",
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
				return fmt.Errorf(
					"cannot decode %v as a UUID: "+
						"incorrect length: %d",
					data, len(data),
				)
			}

			fallthrough
		case bson.TypeBinaryUUID,
			bson.TypeBinaryUUIDOld:
			copy(uid[:], data)

		default:
			err = fmt.Errorf(
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
		err = fmt.Errorf("cannot decode %v as a UUID", rType)
	}

	if err != nil {
		return err
	}
	val.Set(reflect.ValueOf(uid))
	return nil
}
