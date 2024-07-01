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
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

var (
	tUUID = reflect.TypeOf(uuid.UUID{})
)

func init() {
	bson.DefaultRegistry = bson.NewRegistryBuilder().
		RegisterCodec(tUUID, UUIDCodec{}).
		Build()
}

type UUIDCodec struct{}

func (codec UUIDCodec) RegisterCodec(rb *bsoncodec.RegistryBuilder) *bsoncodec.RegistryBuilder {
	return rb.RegisterCodec(tUUID, codec)
}

func (UUIDCodec) EncodeValue(
	ec bsoncodec.EncodeContext,
	w bsonrw.ValueWriter,
	val reflect.Value,
) error {
	if !val.IsValid() || val.Type() != tUUID {
		return bsoncodec.ValueEncoderError{
			Name:     "UUIDCodec",
			Types:    []reflect.Type{tUUID},
			Received: val,
		}
	}
	uid := val.Interface().(uuid.UUID)
	return w.WriteBinaryWithSubtype(uid[:], bsontype.BinaryUUID)
}

func (UUIDCodec) DecodeValue(
	ec bsoncodec.DecodeContext,
	r bsonrw.ValueReader,
	val reflect.Value,
) error {
	if !val.CanSet() || val.Type() != tUUID {
		return bsoncodec.ValueDecoderError{
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
	case bsontype.Binary:
		data, subtype, err = r.ReadBinary()
		switch subtype {
		case bsontype.BinaryGeneric:
			if len(data) != 16 {
				return fmt.Errorf(
					"cannot decode %v as a UUID: "+
						"incorrect length: %d",
					data, len(data),
				)
			}

			fallthrough
		case bsontype.BinaryUUID, bsontype.BinaryUUIDOld:
			copy(uid[:], data)

		default:
			err = fmt.Errorf(
				"cannot decode %v as a UUID: "+
					"incorrect subtype 0x%02x",
				data, subtype,
			)
		}

	case bsontype.Undefined:
		err = r.ReadUndefined()

	case bsontype.Null:
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
