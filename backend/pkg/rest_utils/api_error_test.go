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
package rest_utils

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseApiErrOk(t *testing.T) {
	body := `{"error": "some error message", "request_id":"12345"}`

	err := ParseApiError(bytes.NewBufferString(body))

	assert.True(t, IsApiError(err))
	assert.Equal(t, &ApiError{Err: "some error message", ReqId: "12345"}, err)
}

func TestParseApiErrInvalid(t *testing.T) {
	body := `asdf`

	err := ParseApiError(bytes.NewBufferString(body))

	assert.False(t, IsApiError(err))
}
