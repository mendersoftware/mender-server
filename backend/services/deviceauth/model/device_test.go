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

package model

import (
	"encoding/json"
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeviceFilterParseForm(t *testing.T) {

	testCases := []struct {
		Name string

		Form url.Values

		Result DeviceFilter
		Error  error
	}{{
		Name: "ok",

		Form: url.Values{
			"status": []string{"pending"},
			"id": []string{
				"66c3a801-c5da-4f23-9ab7-7489127ad473",
				"66c3a801-c5da-4f23-9ab7-7489127ad474",
			},
		},
		Result: DeviceFilter{
			Status: []string{"pending"},
			IDs: []string{
				"66c3a801-c5da-4f23-9ab7-7489127ad473",
				"66c3a801-c5da-4f23-9ab7-7489127ad474",
			},
		},
	}, {
		Name: "error, invalid status",

		Form: url.Values{
			"status": []string{"occupied"},
		},
		Error: errors.New("filter status must be one of: " +
			"accepted, pending, rejected, preauthorized or noauth"),
	}}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var fltr DeviceFilter

			err := fltr.ParseForm(tc.Form)
			if tc.Error != nil {
				assert.EqualError(t, err, tc.Error.Error())
			} else {
				assert.Equal(t, tc.Result, fltr)
			}
		})
	}
}

func TestDeviceFilterUnmarshalJSON(t *testing.T) {
	testCases := []struct {
		Name string

		JSON []byte

		Expected DeviceFilter
		Error    error
	}{{
		Name: "ok, empty",
		JSON: []byte{'{', '}'},

		Expected: DeviceFilter{},
	}, {
		Name: "ok, slices",
		JSON: func() []byte {
			b, _ := json.Marshal(map[string]interface{}{
				"status": []string{"accepted", "pending"},
				"id":     []string{"1", "2", "3"},
			})
			return b
		}(),
		Expected: DeviceFilter{
			Status: []string{"accepted", "pending"},
			IDs:    []string{"1", "2", "3"},
		},
	}, {
		Name: "ok, strings",
		JSON: func() []byte {
			b, _ := json.Marshal(map[string]string{
				"status": "accepted",
				"id":     "123456",
			})
			return b
		}(),
		Expected: DeviceFilter{
			Status: []string{"accepted"},
			IDs:    []string{"123456"},
		},
	}, {
		Name: "error, unmarshal error",
		JSON: []byte("Lorem ipsum"),

		Error: errors.New("invalid character 'L' looking for beginning of value"),
	}, {
		Name: "error, invalid status type",
		JSON: func() []byte {
			b, _ := json.Marshal(map[string]interface{}{
				"status": 404,
			})
			return b
		}(),
		Error: errors.New(`invalid JSON type for 'status': ` +
			`must be string or \[\]string`),
	}, {
		Name: "error, invalid status array type",
		JSON: func() []byte {
			b, _ := json.Marshal(map[string]interface{}{
				"status": []int{404, 200},
			})
			return b
		}(),
		Error: errors.New(`invalid JSON type for 'status': ` +
			`must be string or \[\]string`),
	}, {
		Name: "error, invalid id type",
		JSON: func() []byte {
			b, _ := json.Marshal(map[string]interface{}{
				"id": 404,
			})
			return b
		}(),
		Error: errors.New(`invalid JSON type for 'id': ` +
			`must be string or \[\]string`),
	}, {
		Name: "error, invalid id array type",
		JSON: func() []byte {
			b, _ := json.Marshal(map[string]interface{}{
				"id": []int{404, 200},
			})
			return b
		}(),
		Error: errors.New(`invalid JSON type for 'id': ` +
			`must be string or \[\]string`),
	}}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			var fltr DeviceFilter
			err := fltr.UnmarshalJSON(tc.JSON)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t,
						tc.Error.Error(),
						err.Error(),
					)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.Expected, fltr)
			}
		})
	}
}
