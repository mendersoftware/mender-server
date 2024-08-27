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
package main

import (
	"testing"

	"github.com/mendersoftware/mender-server/pkg/log"

	"github.com/stretchr/testify/assert"
)

func TestAddPrivateKeys(t *testing.T) {
	l := log.New(log.Ctx{})
	handlers, err := addPrivateKeys(l, "user/testdata", "private\\.id\\.([0-9]*)\\.pem")
	assert.NoError(t, err)
	assert.Equal(t, 10, len(handlers)) // there are 10 keys matching the pattern
	assert.Contains(t, handlers, 1024)
	assert.Contains(t, handlers, 13102)
	assert.Contains(t, handlers, 14211)
	assert.Contains(t, handlers, 20433)
	assert.Contains(t, handlers, 2048)
	assert.Contains(t, handlers, 21172)
	assert.Contains(t, handlers, 22899)
	assert.Contains(t, handlers, 5539)
	assert.Contains(t, handlers, 826)
	assert.Contains(t, handlers, 9478)
}
