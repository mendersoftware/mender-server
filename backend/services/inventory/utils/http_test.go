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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMsgQueryParmMissing(t *testing.T) {
	s := MsgQueryParmMissing("testparam")
	assert.Equal(t, "Missing required param testparam", s)
}

func TestMsgQueryParmOneOf(t *testing.T) {
	s := MsgQueryParmOneOf("testparam", []string{"foo", "bar"})
	assert.Equal(t, "Param testparam must be one of [foo bar]", s)
}

func mockRequest(url string, has_scheme bool) *http.Request {
	req, _ := http.NewRequest("GET", url, nil)

	if !has_scheme {
		req.URL.Scheme = ""
	}

	return req
}
func mockPageRequest(url, page, per_page string) *http.Request {
	req := mockRequest(url, true)
	reqUrl := req.URL
	q := reqUrl.Query()
	if page != "" {
		q.Set(PageName, page)
	}
	if per_page != "" {
		q.Set(PerPageName, per_page)
	}
	reqUrl.RawQuery = q.Encode()
	return req
}

func TestParseQueryParmUInt(t *testing.T) {
	url := "https://localhost:8080/resource?test=10"
	req := mockRequest(url, true)
	val, err := ParseQueryParmUInt(req, "test", true, 1, 10, 0)
	assert.Equal(t, uint64(10), val)
	assert.Nil(t, err)
}

func TestParseQueryParmUIntMissing(t *testing.T) {
	url := "https://localhost:8080/resource"
	req := mockRequest(url, true)
	_, err := ParseQueryParmUInt(req, "test", true, 1, 10, 0)
	assert.NotNil(t, err)
}

func TestParseQueryParmUIntBounds(t *testing.T) {
	url := "https://localhost:8080/resource?test=11"
	req := mockRequest(url, true)
	_, err := ParseQueryParmUInt(req, "test", true, 1, 10, 0)
	assert.NotNil(t, err)
}

func TestParseQueryParmUIntInvalid(t *testing.T) {
	url := "https://localhost:8080/resource?test=asdf"
	req := mockRequest(url, true)
	_, err := ParseQueryParmUInt(req, "test", true, 1, 10, 0)
	assert.NotNil(t, err)
}

func TestParseQueryParmUIntDefault(t *testing.T) {
	url := "https://localhost:8080/resource"
	req := mockRequest(url, true)
	val, err := ParseQueryParmUInt(req, "test", false, 1, 10, 10)
	assert.Nil(t, err)
	assert.Equal(t, uint64(10), val)
}

func TestParseQueryParmStr(t *testing.T) {
	url := "https://localhost:8080/resource?test=testval"
	req := mockRequest(url, true)
	val, err := ParseQueryParmStr(req, "test", false, []string{"testval"})
	assert.Nil(t, err)
	assert.Equal(t, "testval", val)
}

func TestParseQueryParmStrMissing(t *testing.T) {
	url := "https://localhost:8080/resource"
	req := mockRequest(url, true)
	_, err := ParseQueryParmStr(req, "test", true, []string{"testval"})
	assert.NotNil(t, err)
}

func TestParseQueryParmStrNotOneOf(t *testing.T) {
	url := "https://localhost:8080/resource?test=foo"
	req := mockRequest(url, true)
	_, err := ParseQueryParmStr(req, "test", true, []string{"testval"})
	assert.NotNil(t, err)
}
