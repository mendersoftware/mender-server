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

package http

import (
	"encoding/json"
	"mime"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/services/inventory/utils"
)

type JSONResponseParams struct {
	OutputStatus     int
	OutputBodyObject interface{}
	OutputHeaders    map[string][]string
}

func CheckRecordedResponse(t *testing.T, recorder *httptest.ResponseRecorder, params JSONResponseParams) {

	assert.Equal(t, params.OutputStatus, recorder.Code)
	if recorder.Body.Len() > 0 &&
		assert.Contains(t, recorder.HeaderMap, "Content-Type") {
		contentType, _, err := mime.ParseMediaType(recorder.Header().Get("Content-Type"))
		assert.NoError(t, err)
		assert.Equal(t, "application/json", contentType)
	}
	if params.OutputBodyObject != nil {
		assert.NotEmpty(t, recorder.Body.String())

		expectedJSON, err := json.Marshal(params.OutputBodyObject)
		assert.NoError(t, err)
		assert.JSONEq(t, string(expectedJSON), recorder.Body.String())
	} else {
		assert.Empty(t, recorder.Body.String())
	}

	for name, valueArr := range params.OutputHeaders {
		for _, value := range valueArr {
			assert.True(t, utils.ContainsString(value, recorder.HeaderMap[name]), "not found header with value: "+value)
		}
	}
}
