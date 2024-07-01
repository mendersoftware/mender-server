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
package testing

import (
	"encoding/json"
	"mime"
	"strings"
	"testing"

	"github.com/ant0ine/go-json-rest/rest/test"
	"github.com/stretchr/testify/assert"
)

func CheckResponse(t *testing.T, want ResponseChecker, have *test.Recorded) {
	want.CheckStatus(t, have)
	want.CheckHeaders(t, have)
	want.CheckContentType(t, have)
	want.CheckBody(t, have)
}

// ResponseChecker is a generic response checker, regardless of content-type.
type ResponseChecker interface {
	CheckStatus(t *testing.T, recorded *test.Recorded)
	CheckHeaders(t *testing.T, recorded *test.Recorded)
	CheckContentType(t *testing.T, recorded *test.Recorded)
	CheckBody(t *testing.T, recorded *test.Recorded)
}

// BaseResponse is used for testing any response with a selected content type.
// Implements ResponseChecker, provides base methods for common tests.
type BaseResponse struct {
	Status      int
	ContentType string
	Headers     map[string]string
	Body        interface{}
}

func (b *BaseResponse) CheckStatus(t *testing.T, recorded *test.Recorded) {
	recorded.CodeIs(b.Status)
}

func (b *BaseResponse) CheckContentType(t *testing.T, recorded *test.Recorded) {
	mediaType, params, _ := mime.ParseMediaType(recorded.Recorder.HeaderMap.Get("Content-Type"))
	charset := params["charset"]

	if mediaType != b.ContentType {
		t.Errorf(
			"Content-Type media type: %s expected, got: %s",
			b.ContentType,
			mediaType,
		)
	}

	if charset != "" && strings.ToUpper(charset) != "UTF-8" {
		t.Errorf(
			"Content-Type charset: must be empty or UTF-8, got: %s",
			charset,
		)
	}
}

func (b *BaseResponse) CheckHeaders(t *testing.T, recorded *test.Recorded) {
	for name, value := range b.Headers {
		assert.Equal(t, value, recorded.Recorder.HeaderMap.Get(name))
	}
}

func (b *BaseResponse) CheckBody(t *testing.T, recorded *test.Recorded) {
	if b.Body != nil {
		recorded.BodyIs(b.Body.(string))
	}
}

// JSONResponse is used for testing 'application/json' responses.
// Embeds the BaseResponse (implements ResponseChecker), and overrides relevant methods.
type JSONResponse struct {
	BaseResponse
}

func NewJSONResponse(status int, headers map[string]string, body interface{}) *JSONResponse {
	return &JSONResponse{
		BaseResponse: BaseResponse{
			Status:      status,
			ContentType: "application/json",
			Headers:     headers,
			Body:        body,
		},
	}
}

func (j *JSONResponse) CheckBody(t *testing.T, recorded *test.Recorded) {
	if j.Body != nil {
		assert.NotEmpty(t, recorded.Recorder.Body.String())
		expected, err := json.Marshal(j.Body)
		assert.NoError(t, err)
		assert.JSONEq(t, string(expected), recorded.Recorder.Body.String())
	} else {
		assert.Empty(t, recorded.Recorder.Body.String())
	}
}
