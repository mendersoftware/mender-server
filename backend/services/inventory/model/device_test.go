// Copyright 2023 Northern.tech AS
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	    http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.
package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestDeviceAttributesUnmarshal(t *testing.T) {

	var da DeviceAttributes
	j := `[
{"name": "foo", "value": "bar"},
{"name": "baz", "value": "zen"}
]`
	err := json.Unmarshal([]byte(j), &da)
	assert.NoError(t, err)

	exp := DeviceAttributes{
		{
			Name:  "foo",
			Value: "bar",
		},
		{
			Name:  "baz",
			Value: "zen",
		},
	}

	assert.ObjectsAreEqualValues(exp, da)
}

func TestDeviceAttributesMarshal(t *testing.T) {

	da := DeviceAttributes{
		{
			Name:  "foo",
			Scope: "inventory",
			Value: "bar",
		},
		{
			Name:  "bar",
			Scope: "inventory",
			Value: []int{1, 2, 3},
		},
	}

	data, err := json.Marshal(&da)
	assert.NoError(t, err)

	exp := `[{"name":"foo","value":"bar","scope":"inventory"},{"name":"bar","value":[1,2,3],"scope":"inventory"}]`
	assert.JSONEq(t, string(data), exp)

	var uda DeviceAttributes
	json.Unmarshal(data, &da)

	assert.ObjectsAreEqualValues(uda, da)

	var daEmpty DeviceAttributes
	data, err = json.Marshal(&daEmpty)
	assert.NoError(t, err)
	assert.Equal(t, "[]", string(data))
}

func TestMarshalMarshalBSON(t *testing.T) {
	dev := Device{
		ID: "foo",
		Attributes: DeviceAttributes{{
			Name:  "a.b",
			Value: "foo",
			Scope: "bar",
		}, {
			Name:  "c$d",
			Value: "foo",
			Scope: "bar",
		}},
	}
	b, err := bson.Marshal(dev)
	if assert.NoError(t, err) {
		var tmp Device

		err := bson.Unmarshal(b, &tmp)
		assert.NoError(t, err)
		assert.EqualValues(t, dev, tmp)
	}
}

func TestMarshalUnmarshalBSON(t *testing.T) {
	str2Ptr := func(s string) *string {
		return &s
	}
	dev := Device{
		ID:    "foo",
		Group: "bar",
		Attributes: DeviceAttributes{{
			Name:        "str",
			Value:       "foo",
			Scope:       "bar",
			Description: str2Ptr("fooSan!"),
		}, {
			Name:  "float",
			Value: float64(123.0),
			Scope: "floaters",
		}},
	}
	// Expected added by bson.Marshal
	groupAttr := DeviceAttribute{
		Name:  AttrNameGroup,
		Value: "bar",
		Scope: AttrScopeSystem,
	}

	b, err := bson.Marshal(dev)
	if assert.NoError(t, err) {
		var tmp Device
		dev.Attributes = append(dev.Attributes, groupAttr)
		err := bson.Unmarshal(b, &tmp)
		assert.NoError(t, err)
		assert.EqualValues(t, dev, tmp)
	}
}

func TestValidateDeviceAttributes(t *testing.T) {
	testCases := []struct {
		Name string

		Attributes DeviceAttributes
		ErrMessage string
	}{
		{
			Name: "Strings and floats",

			Attributes: DeviceAttributes{{
				Name:  "foo",
				Value: "stringer",
				Scope: "stringers",
			}, {
				Name:  "bar",
				Value: float64(123.0),
				Scope: "floaters",
			}, {
				Name:  "baz",
				Value: "1234567.0",
				Scope: "nonFloaters",
			}},
		},
		{
			Name: "String arrays",

			Attributes: DeviceAttributes{{
				Name:  "foo",
				Value: []string{"bar", "baz"},
				Scope: "slicers",
			}, {
				Name:  "bar",
				Value: []interface{}{"123.4", "567.89"},
				Scope: "slicers",
			}},
		},
		{
			Name: "Float64 arrays",

			Attributes: DeviceAttributes{{
				Name:  "foo",
				Value: []float64{123.567, 456.789},
				Scope: "floatingSlicers",
			}, {
				Name:  "bar",
				Value: []interface{}{float64(1.0)},
				Scope: "floatingSlicer",
			}},
		},
		{
			Name:       "Empty attributes",
			Attributes: DeviceAttributes{},
		},
		{
			Name: "Empty slice",
			Attributes: DeviceAttributes{{
				Name:  "Empty slice is ok",
				Value: []interface{}{},
				Scope: "void",
			}},
		},
		{
			Name: "Attribute missing value",
			Attributes: DeviceAttributes{{
				Name:  "nil",
				Scope: "void",
			}},
			ErrMessage: "supported types are string, float64, " +
				"and arrays thereof",
		},
		{
			Name: "Illegal float",
			Attributes: DeviceAttributes{{
				Name:  "Wrong float",
				Value: float32(123),
				Scope: "totallyLegit",
			}},
			ErrMessage: "supported types are string, float64, " +
				"and arrays thereof",
		},
		{
			Name: "Illegal string type",
			Attributes: DeviceAttributes{{
				Name:  "foo",
				Value: []byte("foobar"),
				Scope: "prettyStringish",
			}},
			ErrMessage: "supported types are string, float64, " +
				"and arrays thereof",
		},
		{
			Name: "Mixed type slice",
			Attributes: DeviceAttributes{{
				Name:  "mixedSignals",
				Value: []interface{}{'c', 123, []byte("bleh")},
				Scope: "bagOfTypes",
			}},
			ErrMessage: "array values must be either " +
				"string or float64",
		},
		{
			Name: "Mixed slice of legal types",
			Attributes: DeviceAttributes{{
				Name:  "totallyLegit",
				Value: []interface{}{"stringer", 123.0},
				Scope: "bagOfTypes",
			}},
			ErrMessage: "array values must be of consistent type: " +
				"string or float64",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.Attributes.Validate()
			if tc.ErrMessage != "" {
				if assert.Error(t, err) {
					assert.Contains(
						t, err.Error(), tc.ErrMessage,
					)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}

}

func TestValidateGroupName(t *testing.T) {
	t.Parallel()
	group1 := GroupName(make([]byte, 1025))
	assert.EqualError(t, group1.Validate(), "Group name can at most have "+
		"1024 characters")
	group2 := GroupName("totally.legit")
	assert.EqualError(t, group2.Validate(), "Group name can only contain: "+
		"upper/lowercase alphanum, -(dash), _(underscore)")
	group3 := GroupName("")
	assert.EqualError(t, group3.Validate(), "Group name cannot be blank")
	group4 := GroupName("test")
	assert.NoError(t, group4.Validate())
}
