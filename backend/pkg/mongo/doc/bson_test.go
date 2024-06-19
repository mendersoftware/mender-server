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

package doc

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

type SampleBSONMarshalerStruct struct {
	Value bson.D
}

func (r *SampleBSONMarshalerStruct) MarshalBSON() ([]byte, error) {
	return bson.Marshal(r.Value)
}

func TestMarshallBSONOrDocumentFromStruct(t *testing.T) {
	testCases := []struct {
		Name string

		Input          interface{}
		AppendElements []bson.E
		Expected       bson.D
	}{
		{
			Name: "Simple success",

			Input: struct {
				Field1 string
				Field2 int
			}{
				Field1: "foo",
				Field2: 321,
			},
			Expected: bson.D{
				{Key: "field1", Value: "foo"},
				{Key: "field2", Value: 321},
			},
		},
		{
			Name: "bson.Marshaler interface",

			Input: &SampleBSONMarshalerStruct{
				Value: bson.D{
					{Key: "field", Value: "value"},
				},
			},
			Expected: bson.D{
				{Key: "field", Value: "value"},
			},
		},
		{
			Name: "Not a struct",

			Input:    "Panic attack!",
			Expected: nil,
		},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			doc := MarshallBSONOrDocumentFromStruct(tc.Input, tc.AppendElements...)
			assert.Equal(t, tc.Expected, doc)
		})
	}

}

func TestDocumentFromStruct(t *testing.T) {
	type InlineStruct struct {
		FirstField  string `bson:"first_field"`
		SecondField string `bson:"second_field"`
	}
	testCases := []struct {
		Name string

		Input          interface{}
		AppendElements []bson.E
		Expected       bson.D
	}{
		{
			Name: "Simple success",

			Input: struct {
				Field1 string
				Field2 int
			}{
				Field1: "foo",
				Field2: 321,
			},
			Expected: bson.D{
				{Key: "field1", Value: "foo"},
				{Key: "field2", Value: 321},
			},
		},
		{
			Name: "Bson tags and appends",

			Input: struct {
				Field1 string `bson:"foo"`
				Field2 string `bson:"bar,omitempty"`
			}{
				Field1: "baz",
			},
			AppendElements: []bson.E{
				{Key: "a1", Value: 123},
				{Key: "a2", Value: "foobarbaz"},
			},
			Expected: bson.D{
				{Key: "foo", Value: "baz"},
				{Key: "a1", Value: 123},
				{Key: "a2", Value: "foobarbaz"},
			},
		},
		{
			Name: "Bson tags, appends and inline struct",

			Input: struct {
				Field1       string `bson:"foo"`
				Field2       string `bson:"bar,omitempty"`
				InlineStruct `bson:"inline"`
			}{
				Field1: "baz",
				InlineStruct: InlineStruct{
					FirstField:  "f1",
					SecondField: "f2",
				},
			},
			AppendElements: []bson.E{
				{Key: "a1", Value: 123},
			},
			Expected: bson.D{
				{Key: "foo", Value: "baz"},
				{Key: "first_field", Value: "f1"},
				{Key: "second_field", Value: "f2"},
				{Key: "a1", Value: 123},
			},
		},
		{
			Name: "Not a struct",

			Input:    "Panic attack!",
			Expected: nil,
		},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			doc := DocumentFromStruct(tc.Input, tc.AppendElements...)
			assert.Equal(t, tc.Expected, doc)
		})
	}

}

func TestFlattenDocument(t *testing.T) {
	testCases := []struct {
		Name string

		Input  interface{}
		Output bson.D
		Error  error

		Options *FlattenOptions
	}{{
		Name: "OK, struct",

		Input: struct {
			String   string  `bson:",omitempty"`
			Float    float64 `bson:"floatie_mc_float_face"`
			IntSlice []int   `bson:"slice"`
			Struct   struct {
				NestedVal string `bson:"nested_val"`
			}
		}{
			String:   "foo",
			Float:    123.456,
			IntSlice: []int{1, 2, 3, 4, 5, 6},
			Struct: struct {
				NestedVal string `bson:"nested_val"`
			}{
				NestedVal: "test",
			},
		},
		Output: bson.D{
			{Key: "string", Value: "foo"},
			{Key: "floatie_mc_float_face", Value: 123.456},
			{Key: "slice", Value: []int{1, 2, 3, 4, 5, 6}},
			{Key: "struct.nested_val", Value: "test"},
		},
	}, {
		Name: "OK, struct, with transform",

		Input: &struct {
			String     string  `bson:",omitempty"`
			Float      float64 `bson:"floatie_mc_float_face"`
			IntSlice   []int   `bson:"slice"`
			Map        map[string]string
			unexported string
		}{
			Float:    123.456,
			IntSlice: []int{1, 2, 3, 4, 5, 6},
			Map: map[string]string{
				"nested_val": "test",
			},
			unexported: "should not show up in result",
		},

		Options: &FlattenOptions{
			Transform: func(
				key string, value interface{},
			) (string, interface{}) {
				rVal := reflect.ValueOf(value)
				switch rVal.Kind() {
				case reflect.Slice:
					return key, bson.M{"$in": value}
				default:
					return key, value
				}
			},
		},
		Output: bson.D{
			{Key: "floatie_mc_float_face", Value: 123.456},
			{Key: "slice", Value: bson.M{"$in": []int{1, 2, 3, 4, 5, 6}}},
			{Key: "map.nested_val", Value: "test"},
		},
	}, {
		Name: "OK, map-type",

		Input: map[string]interface{}{
			"key": "value",
			"map": bson.M{
				"int": 123,
				"str": "foo",
			},
		},

		Output: bson.D{
			{Key: "key", Value: "value"},
			{Key: "map.int", Value: 123},
			{Key: "map.str", Value: "foo"},
		},
	}, {
		Name: "OK, map-type, with transform",

		Input: map[string]interface{}{
			"key": "value",
			"struct": struct {
				Int int
				Str string
			}{
				Int: 123,
				Str: "foo",
			},
		},

		Output: bson.D{
			{Key: "key", Value: "value"},
			{Key: "struct.int", Value: "123"},
			{Key: "struct.str", Value: "foo"},
		},

		Options: NewFlattenOptions().SetTransform(func(
			key string, value interface{},
		) (string, interface{}) {
			return key, fmt.Sprintf("%v", value)
		}),
	}, {
		Name: "Error, invalid type",

		Input: "this is not allowed!",
		Error: errors.New("[programming error] invalid argument type string, " +
			"expected struct or map-like type"),
	}, {
		Name: "Error, invalid field type",

		Input: map[string]interface{}{
			"invalid": nil, // (type-less)
		},
		Error: errors.New(
			"reflect: call of reflect.Value.Interface on zero Value",
		),
	}}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			res, err := FlattenDocument(testCase.Input, testCase.Options)
			if testCase.Output != nil &&
				assert.NoError(t, err) &&
				assert.NotNil(t, res) {

				switch testCase.Input.(type) {
				case map[string]interface{}:
					for _, elem := range testCase.Output {
						assert.Contains(t, res, elem)
					}
				default:
					assert.Equal(t, testCase.Output, res)
				}
			} else {
				if assert.Error(t, err) {
					assert.Contains(t, err.Error(), testCase.Error.Error())
				}
			}
		})
	}
}
