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

package testing

import (
	"encoding/json"
	"testing"

	"github.com/ant0ine/go-json-rest/rest/test"
	"github.com/stretchr/testify/assert"
)

type JSONResponseParams struct {
	OutputStatus     int
	OutputBodyObject interface{}
	OutputHeaders    map[string]string
}

func CheckRecordedResponse(t *testing.T, recorded *test.Recorded, params JSONResponseParams) {

	recorded.CodeIs(params.OutputStatus)
	recorded.ContentTypeIsJson()

	if params.OutputBodyObject != nil {
		assert.NotEmpty(t, recorded.Recorder.Body.String())

		expectedJSON, err := json.Marshal(params.OutputBodyObject)
		assert.NoError(t, err)
		assert.JSONEq(t, string(expectedJSON), recorded.Recorder.Body.String())
	} else {
		assert.Empty(t, recorded.Recorder.Body.String())
	}

	rsp := recorded.Recorder.Result()
	for name, value := range params.OutputHeaders {
		assert.Equal(t, value, rsp.Header.Get(name))
	}
	_ = rsp.Body.Close()
}
