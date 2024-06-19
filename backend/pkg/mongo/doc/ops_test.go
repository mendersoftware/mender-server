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
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestUnwindMap(t *testing.T) {
	Array8 := []string{"1", "2", "3", "4", "5", "6", "7", "8"}

	cases := []struct {
		name string
		in   map[string]interface{}
		out  []bson.D
		err  error
	}{
		{
			name: "Success simple",
			in: map[string]interface{}{
				"device_type": "arm6",
				"chksum":      "123",
			},
			out: []bson.D{
				bson.D{
					bson.E{Key: "chksum", Value: "123"},
					bson.E{Key: "device_type", Value: "arm6"},
				},
			},
		},
		{
			name: "Success array",
			in: map[string]interface{}{
				"device_type": []string{"arm6", "arm7"},
				"chksum":      "123",
			},
			out: []bson.D{
				bson.D{
					bson.E{Key: "chksum", Value: "123"},
					bson.E{Key: "device_type", Value: "arm6"},
				},
				bson.D{
					bson.E{Key: "chksum", Value: "123"},
					bson.E{Key: "device_type", Value: "arm7"},
				},
			},
		},
		{
			name: "Success multi-type array",
			in: map[string]interface{}{
				"device_type": []string{"arm6", "arm7"},
				"foo":         []interface{}{"1", "2", "3"},
				"chksum":      "123",
			},
			out: []bson.D{
				bson.D{
					bson.E{Key: "chksum", Value: "123"},
					bson.E{Key: "device_type", Value: "arm6"},
					bson.E{Key: "foo", Value: "1"},
				},
				bson.D{
					bson.E{Key: "chksum", Value: "123"},
					bson.E{Key: "device_type", Value: "arm6"},
					bson.E{Key: "foo", Value: "2"},
				},
				bson.D{
					bson.E{Key: "chksum", Value: "123"},
					bson.E{Key: "device_type", Value: "arm6"},
					bson.E{Key: "foo", Value: "3"},
				},
				bson.D{
					bson.E{Key: "chksum", Value: "123"},
					bson.E{Key: "device_type", Value: "arm7"},
					bson.E{Key: "foo", Value: "1"},
				},
				bson.D{
					bson.E{Key: "chksum", Value: "123"},
					bson.E{Key: "device_type", Value: "arm7"},
					bson.E{Key: "foo", Value: "2"},
				},
				bson.D{
					bson.E{Key: "chksum", Value: "123"},
					bson.E{Key: "device_type", Value: "arm7"},
					bson.E{Key: "foo", Value: "3"},
				},
			},
		},
		{
			name: "Error permutations",
			// Input has 786432 permutations - above threshold
			in: map[string]interface{}{
				"device_type": []string{"foo", "bar", "baz"},
				"dep0":        Array8,
				"dep1":        Array8,
				"dep2":        Array8,
				"dep3":        Array8,
				"dep4":        Array8,
				"dep5":        Array8,
				"chksum":      "123",
			},
			err: ErrPermutations,
		},
		{
			name: "Error overflow",
			// Input has 8^22=2^66 -> should overflow to int(8) (once
			// or twice depending on 64 or 32 bit architecture).
			in: map[string]interface{}{
				"foo":  "bar",
				"dep0": Array8, "dep10": Array8,
				"dep1": Array8, "dep11": Array8,
				"dep2": Array8, "dep12": Array8,
				"dep3": Array8, "dep13": Array8,
				"dep4": Array8, "dep14": Array8,
				"dep5": Array8, "dep15": Array8,
				"dep6": Array8, "dep16": Array8,
				"dep7": Array8, "dep17": Array8,
				"dep8": Array8, "dep18": Array8,
				"dep9": Array8, "dep19": Array8,
				"dep20": Array8, "dep21": Array8,
			},
			err: ErrPermutations,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := UnwindMap(tc.in)
			if tc.err != nil {
				assert.NotNil(t, err)
				assert.Equal(t, tc.err, err)
			} else {
				assert.Nil(t, err)
				for i := range tc.out {
					assert.Contains(t, res, tc.out[i])
				}
			}
		})
	}
}
