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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAttributes2Map(t *testing.T) {
	configurationMap := map[string]interface{}{
		"hostname": "some0",
	}
	attributes := []Attribute{
		{
			Key:   "hostname",
			Value: "some0",
		},
	}
	assert.Equal(t, attributes2Map(attributes), configurationMap)

	configurationMap["hostname"] = "some0other"
	assert.NotEqual(t, attributes2Map(attributes), configurationMap)
}

func TestMap2Attributes(t *testing.T) {
	configurationMap := map[string]interface{}{
		"hostname": "some0",
	}
	attributes := Attributes{
		{
			Key:   "hostname",
			Value: "some0",
		},
	}
	assert.Equal(t, map2Attributes(configurationMap), attributes)

	configurationMap["hostname"] = "some0other"
	assert.NotEqual(t, map2Attributes(configurationMap), attributes)
}
