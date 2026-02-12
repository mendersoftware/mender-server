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
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func createV2Reader(t *testing.T, valType bson.Type, rawVal []byte) bson.ValueReader {
	// calculate total doc length: header(4) + type(1) + key(2) + value(len) + EOO(1)
	docLen := 4 + 1 + 2 + len(rawVal) + 1
	buf := make([]byte, docLen)

	binary.LittleEndian.PutUint32(buf, uint32(docLen))

	buf[4] = byte(valType)
	buf[5] = 'x'
	buf[6] = 0x00

	copy(buf[7:], rawVal)

	// write EOO
	buf[docLen-1] = 0x00

	vr := bson.NewDocumentReader(bytes.NewReader(buf))

	dr, err := vr.ReadDocument()
	if err != nil {
		t.Fatalf("failed to read document: %v", err)
	}

	name, valReader, err := dr.ReadElement()
	if err != nil {
		t.Fatalf("failed to read element: %v", err)
	}
	if name != "x" {
		t.Fatalf("expected key 'x', got '%s'", name)
	}

	return valReader
}

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
			UUID uuid.UUID
		}{
			UUID: uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest")),
		},
	}, {
		Name: "ok, pointer in a struct",
		Value: struct {
			UUID *uuid.UUID
		}{
			UUID: func() *uuid.UUID {
				uid := uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest"))
				return &uid
			}(),
		},
	}, {
		Name: "ok, in a struct",
		Value: struct {
			UUID uuid.UUID `bson:",omitempty"`
		}{},
	}, {
		Name: "ok, empty slice",
		Value: struct {
			UUIDS []uuid.UUID
		}{UUIDS: []uuid.UUID{}},
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
			"UUIDCodec can only encode valid uuid.UUID, but got string",
		),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			w := bson.NewDocumentWriter(&buf)
			dw, err := w.WriteDocument()
			require.NoError(t, err)
			ew, err := dw.WriteDocumentElement("test")
			require.NoError(t, err)

			eCtx := bson.EncodeContext{Registry: bson.NewMgoRegistry()}
			err = UUIDCodec{}.EncodeValue(eCtx, ew, reflect.ValueOf(tc.Value))
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

		InputType bson.Type
		RawInput  []byte
		Value     interface{}
		Error     error
	}{{
		Name: "ok",

		InputType: bson.TypeBinary,
		RawInput: []byte{
			16, 0, 0, 0, bson.TypeBinaryUUID, '0', '1', '2', '3', '4',
			'5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F',
		},

		Value: uuid.UUID{
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
			'A', 'B', 'C', 'D', 'E', 'F',
		},
	}, {
		Name: "ok, old uuid subtype",

		InputType: bson.TypeBinary,
		RawInput: []byte{
			16, 0, 0, 0, bson.TypeBinaryUUIDOld, '0', '1', '2', '3', '4',
			'5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F',
		},

		Value: uuid.UUID{
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
			'A', 'B', 'C', 'D', 'E', 'F',
		},
	}, {
		Name: "ok, generic binary",

		InputType: bson.TypeBinary,
		RawInput: []byte{
			16, 0, 0, 0, bson.TypeBinaryGeneric, '0', '1', '2', '3', '4',
			'5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F',
		},

		Value: uuid.UUID{
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
			'A', 'B', 'C', 'D', 'E', 'F',
		},
	}, {
		Name: "error, invalid length",

		InputType: bson.TypeBinary,
		RawInput: []byte{
			8, 0, 0, 0, bson.TypeBinaryGeneric,
			'D', 'E', 'A', 'D', 'B', 'E', 'E', 'F',
		},

		Value: uuid.UUID{},
		Error: errors.New(
			`cannot decode \[68 69 65 68 66 69 69 70\] as a UUID: ` +
				`incorrect length: 8`,
		),
	}, {
		Name: "error, invalid length",

		InputType: bson.TypeBinary,
		RawInput: []byte{
			8, 0, 0, 0, bson.TypeBinaryUserDefined,
			'D', 'E', 'A', 'D', 'B', 'E', 'E', 'F',
		},

		Value: uuid.UUID{},
		Error: fmt.Errorf(
			`cannot decode \[68 69 65 68 66 69 69 70\] as a UUID: `+
				`incorrect subtype 0x%02x`, bson.TypeBinaryUserDefined,
		),
	}, {
		Name: "ok, undefined",

		InputType: bson.TypeUndefined,
		RawInput: []byte{
			8, 0, 0, 0, bson.TypeBinaryUserDefined,
			'D', 'E', 'A', 'D', 'B', 'E', 'E', 'F',
		},

		Value: uuid.UUID{},
	}, {
		Name: "ok, null",

		InputType: bson.TypeNull,
		RawInput: []byte{
			8, 0, 0, 0, bson.TypeBinaryUserDefined,
			'D', 'E', 'A', 'D', 'B', 'E', 'E', 'F',
		},

		Value: uuid.UUID{},
	}, {
		Name: "error, invalid bson",

		InputType: bson.TypeBoolean,
		RawInput:  []byte{'1'},

		Value: uuid.UUID{},
		Error: errors.New(`cannot decode boolean as a UUID`),
	}, {
		Name: "error, bad encoder args",

		InputType: bson.TypeBoolean,
		RawInput:  []byte{'1'},

		Value: "?",
		Error: errors.New(
			`UUIDCodec can only decode valid and settable ` +
				`uuid\.UUID, but got string`),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			r := createV2Reader(t, tc.InputType, tc.RawInput)
			dCtx := bson.DecodeContext{Registry: bson.NewMgoRegistry()}
			val := reflect.New(reflect.TypeOf(tc.Value))
			err := UUIDCodec{}.DecodeValue(dCtx, r, val.Elem())
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
