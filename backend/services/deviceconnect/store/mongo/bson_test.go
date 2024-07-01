// Copyright 2021 Northern.tech AS
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
	"bytes"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

func TestUUIDEncodeDecode(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		Value      interface{}
		EncodError error
		DecodError error
	}{{
		Name: "ok, in a struct",
		Value: struct {
			uuid.UUID
		}{
			UUID: uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest")),
		},
	}, {
		Name: "ok, pointer in a struct",
		Value: struct {
			*uuid.UUID
		}{
			UUID: func() *uuid.UUID {
				uid := uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest"))
				return &uid
			}(),
		},
	}, {
		Name: "ok, in a struct",
		Value: struct {
			uuid.UUID `bson:",omitempty"`
		}{},
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			b, err := bson.Marshal(tc.Value)
			if tc.EncodError != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.EncodError.Error(), err.Error())
				}
				return
			}
			if !assert.NoError(t, err) {
				return
			}
			val := reflect.New(reflect.TypeOf(tc.Value))
			err = bson.Unmarshal(b, val.Interface())
			if tc.DecodError != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.DecodError.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, tc.Value, val.Elem().Interface())
			}
		})
	}
}

func TestUUIDEncodeValue(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		Value interface{}
		Error error
	}{{
		Name:  "ok",
		Value: uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest")),
	}, {
		Name:  "error, bad type",
		Value: "0c070528-236b-414b-b72b-42bfd10c3abc",
		Error: errors.New(
			"UUIDEncodeValue can only encode valid uuid.UUID, but got string",
		),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			w, err := bsonrw.NewBSONValueWriter(&buf)
			require.NoError(t, err)
			dw, err := w.WriteDocument()
			require.NoError(t, err)
			ew, err := dw.WriteDocumentElement("test")
			require.NoError(t, err)

			eCtx := bsoncodec.EncodeContext{Registry: bson.DefaultRegistry}
			err = uuidEncodeValue(eCtx, ew, reflect.ValueOf(tc.Value))
			dw.WriteDocumentEnd()
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				raw := bson.Raw(buf.Bytes())
				id, err := raw.LookupErr("test")
				if assert.NoError(t, err) {
					_, bin, ok := id.BinaryOK()
					if assert.True(t, ok, "document value not binary") {
						var uid uuid.UUID
						copy(uid[:], bin)
						assert.Equal(t, tc.Value, uid)
					}
				}
				return
			}

		})
	}
}

func TestUUIDDecodeValue(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		InputType bsontype.Type
		RawInput  []byte
		Value     interface{}
		Error     error
	}{{
		Name: "ok",

		InputType: bsontype.Binary,
		RawInput: []byte{
			16, 0, 0, 0, bsontype.BinaryUUID, '0', '1', '2', '3', '4',
			'5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F',
		},

		Value: uuid.UUID{
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
			'A', 'B', 'C', 'D', 'E', 'F',
		},
	}, {
		Name: "ok, old uuid subtype",

		InputType: bsontype.Binary,
		RawInput: []byte{
			16, 0, 0, 0, bsontype.BinaryUUIDOld, '0', '1', '2', '3', '4',
			'5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F',
		},

		Value: uuid.UUID{
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
			'A', 'B', 'C', 'D', 'E', 'F',
		},
	}, {
		Name: "ok, generic binary",

		InputType: bsontype.Binary,
		RawInput: []byte{
			16, 0, 0, 0, bsontype.BinaryGeneric, '0', '1', '2', '3', '4',
			'5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F',
		},

		Value: uuid.UUID{
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
			'A', 'B', 'C', 'D', 'E', 'F',
		},
	}, {
		Name: "error, invalid length",

		InputType: bsontype.Binary,
		RawInput: []byte{
			8, 0, 0, 0, bsontype.BinaryGeneric,
			'D', 'E', 'A', 'D', 'B', 'E', 'E', 'F',
		},

		Value: uuid.UUID{},
		Error: errors.New(
			`cannot decode \[68 69 65 68 66 69 69 70\] as a UUID: ` +
				`incorrect length: 8`,
		),
	}, {
		Name: "error, invalid length",

		InputType: bsontype.Binary,
		RawInput: []byte{
			8, 0, 0, 0, bsontype.BinaryUserDefined,
			'D', 'E', 'A', 'D', 'B', 'E', 'E', 'F',
		},

		Value: uuid.UUID{},
		Error: errors.Errorf(
			`cannot decode \[68 69 65 68 66 69 69 70\] as a UUID: `+
				`incorrect subtype 0x%02x`, bsontype.BinaryUserDefined,
		),
	}, {
		Name: "ok, undefined",

		InputType: bsontype.Undefined,
		RawInput: []byte{
			8, 0, 0, 0, bsontype.BinaryUserDefined,
			'D', 'E', 'A', 'D', 'B', 'E', 'E', 'F',
		},

		Value: uuid.UUID{},
	}, {
		Name: "ok, null",

		InputType: bsontype.Null,
		RawInput: []byte{
			8, 0, 0, 0, bsontype.BinaryUserDefined,
			'D', 'E', 'A', 'D', 'B', 'E', 'E', 'F',
		},

		Value: uuid.UUID{},
	}, {
		Name: "error, invalid bsontype",

		InputType: bsontype.Boolean,
		RawInput:  []byte{'1'},

		Value: uuid.UUID{},
		Error: errors.New(`cannot decode boolean as a UUID`),
	}, {
		Name: "error, bad encoder args",

		InputType: bsontype.Boolean,
		RawInput:  []byte{'1'},

		Value: "?",
		Error: errors.New(
			`UUIDDecodeValue can only decode valid and settable ` +
				`uuid\.UUID, but got string`),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			r := bsonrw.NewBSONValueReader(tc.InputType, tc.RawInput)
			dCtx := bsoncodec.DecodeContext{Registry: bson.DefaultRegistry}
			val := reflect.New(reflect.TypeOf(tc.Value))
			err := uuidDecodeValue(dCtx, r, val.Elem())
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.Value, val.Elem().Interface())
				return
			}

		})
	}
}
