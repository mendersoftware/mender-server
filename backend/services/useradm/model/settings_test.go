// Copyright 2022 Northern.tech AS
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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestSettingsMarshalJSON(t *testing.T) {
	testCases := map[string]struct {
		settings Settings
		out      []byte
		err      error
	}{
		"ok": {
			settings: Settings{
				ID:   "id",
				ETag: "etag",
				Values: map[string]interface{}{
					"key": "value",
				},
			},
			out: []byte(`{"key":"value"}`),
		},
	}

	for i, tc := range testCases {
		t.Run(i, func(t *testing.T) {
			out, err := json.Marshal(tc.settings)
			if tc.err != nil {
				assert.Equal(t, tc.err, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.out, out)
			}
		})
	}
}

func TestSettingsUnmarshalJSON(t *testing.T) {
	testCases := map[string]struct {
		in       []byte
		settings Settings
		err      error
	}{
		"ok": {
			in: []byte(`{"key":"value"}`),
			settings: Settings{
				Values: map[string]interface{}{
					"key": "value",
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(i, func(t *testing.T) {
			s := Settings{}
			err := json.Unmarshal(tc.in, &s)
			if tc.err != nil {
				assert.Equal(t, tc.err, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.settings, s)
			}
		})
	}
}

func TestSettingsMarshalBSON(t *testing.T) {
	settings := Settings{
		ID:   "id",
		ETag: "etag",
		Values: map[string]interface{}{
			"key":       "value",
			"tenant_id": "this is reserved",
		},
	}

	b, err := bson.Marshal(settings)
	assert.NoError(t, err)

	unmarshalled := Settings{}
	err = bson.Unmarshal(b, &unmarshalled)
	assert.NoError(t, err)

	expected := Settings{
		ID:   "id",
		ETag: "etag",
		Values: map[string]interface{}{
			"key": "value",
		},
	}
	assert.Equal(t, expected, unmarshalled)
}

func TestSettingsValidate(t *testing.T) {
	tooManyValues := map[string]interface{}{}
	for i := 0; i < maxSettings+1; i++ {
		tooManyValues[fmt.Sprintf("key%d", i)] = "value"
	}

	b := make([]rune, maxLength4096+1)
	for i := range b {
		b[i] = rune(65) // "A"
	}
	tooLongValue := string(b)

	testCases := map[string]struct {
		settings Settings
		err      error
	}{
		"ok": {
			settings: Settings{
				ID:   "id",
				ETag: "etag",
				Values: map[string]interface{}{
					"key": "value",
				},
			},
		},
		"ko, too many settings": {
			settings: Settings{
				ID:     "id",
				ETag:   "etag",
				Values: tooManyValues,
			},
			err: fmt.Errorf("Values: the length must be no more than %d.", maxSettings),
		},
		"ko, too long key": {
			settings: Settings{
				ID:   "id",
				ETag: "etag",
				Values: map[string]interface{}{
					tooLongValue: "value",
				},
			},
			err: fmt.Errorf("Values: the length must be no more than %d.", maxLength128),
		},
		"ko, too long value": {
			settings: Settings{
				ID:   "id",
				ETag: "etag",
				Values: map[string]interface{}{
					"key": tooLongValue,
				},
			},
			err: fmt.Errorf("Values: (key: the length must be no more than %d.).", maxLength4096),
		},
	}

	for i, tc := range testCases {
		t.Run(i, func(t *testing.T) {
			err := tc.settings.Validate()
			if tc.err != nil {
				assert.EqualError(t, tc.err, err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
