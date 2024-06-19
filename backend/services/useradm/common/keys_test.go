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

package common

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyIdFromPath(t *testing.T) {
	var keyId int
	for i := 1; i < 1024; i++ {
		keyId = KeyIdFromPath("/etc/useradm/rsa/private.id."+strconv.Itoa(i)+".pem", "private\\.id\\.([0-9]*)\\.pem")
		assert.Equal(t, i, keyId)
	}
	for i := 1; i < 1024; i++ {
		keyId = KeyIdFromPath("/etc/useradm/rsa/private.id-"+strconv.Itoa(i)+".pem", "private\\.id\\.([0-9]*)\\.pem")
		assert.Equal(t, KeyIdZero, keyId)
	}
}
