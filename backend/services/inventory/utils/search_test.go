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
package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/services/inventory/model"
)

func TestContainsString(t *testing.T) {
	if ContainsString("foo", []string{"foo", "bar", "baz"}) == false {
		t.Errorf("string not found")
	}
	if ContainsString("foo", []string{"bar", "baz"}) == true {
		t.Errorf("string found, expected not found")
	}
}

func TestTextToKeywords(t *testing.T) {
	testCases := map[string]struct {
		In     string
		Out    string
		Search bool
	}{
		"same string": {
			In:  "test",
			Out: "test",
		},
		"mixed case": {
			In:  "Test",
			Out: "Test",
		},
		"special characters": {
			In:  "te:st",
			Out: "te st",
		},
		"spaces are preserved": {
			In:  "te st",
			Out: "te st",
		},
		"dashes are removed": {
			In:  "te-st",
			Out: "te st",
		},
		"extra spaces are removed": {
			In:  " extra  spaces",
			Out: "extra spaces",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			out := TextToKeywords(tc.In)
			assert.Equal(t, tc.Out, out)
		})
	}
}

func TestGetTextField(t *testing.T) {
	testCases := map[string]struct {
		Device *model.Device
		Text   string
	}{
		"ok": {
			Device: &model.Device{
				ID:    "1",
				Group: "group",
				Attributes: model.DeviceAttributes{
					{
						Name:  "attribute",
						Scope: model.AttrScopeIdentity,
						Value: "value1",
					},
					{
						Name:  "attribute",
						Scope: model.AttrScopeInventory,
						Value: "value2",
					},
					{
						Name:  "attribute",
						Scope: model.AttrScopeTags,
						Value: "value3",
					},
					{
						Name:  "attribute",
						Scope: model.AttrScopeSystem,
						Value: "value4",
					},
				},
			},
			Text: "1 group value1 value2 value3",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			text := GetTextField(tc.Device)
			assert.Equal(t, tc.Text, text)
		})
	}
}
