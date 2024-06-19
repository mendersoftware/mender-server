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
	neturl "net/url"
	"testing"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/stretchr/testify/assert"
)

func TestMsgQueryParmInvalid(t *testing.T) {
	s := MsgQueryParmInvalid("testparam")
	assert.Equal(t, "Can't parse param testparam", s)
}

func TestMsgQueryParmMissing(t *testing.T) {
	s := MsgQueryParmMissing("testparam")
	assert.Equal(t, "Missing required param testparam", s)
}

func TestMsgQueryParmLimit(t *testing.T) {
	s := MsgQueryParmLimit("testparam")
	assert.Equal(t, "Param testparam is out of bounds", s)
}

func TestMsgQueryParmOneOf(t *testing.T) {
	s := MsgQueryParmOneOf("testparam", []string{"foo", "bar"})
	assert.Equal(t, "Param testparam must be one of [foo bar]", s)
}

func mockRequest(url string, has_scheme bool) *rest.Request {
	req, _ := http.NewRequest("GET", url, nil)

	if !has_scheme {
		req.URL.Scheme = ""
	}

	return &rest.Request{Request: req, PathParams: nil, Env: nil}
}

func mockPageRequest(url, page, per_page string) *rest.Request {
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

func TestMakeLink(t *testing.T) {
	l := MakeLink("first", "resource", neturl.Values{}, 1, 10)
	assert.Equal(t, "<resource?page=1&per_page=10>; rel=\"first\"", l)
}

func TestMakePageLinkHdrs(t *testing.T) {
	url := "https://localhost:8080/base/url/resource?page=2&per_page=10"
	req := mockRequest(url, true)
	links := MakePageLinkHdrs(req, 2, 10, true)
	assert.Len(t, links, 3)
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

func TestParsePagination(t *testing.T) {
	url := "https://localhost:8080/resource"
	req := mockPageRequest(url, "1", "10")
	page, per_page, err := ParsePagination(req)
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), page)
	assert.Equal(t, uint64(10), per_page)
}

func TestParsePaginationInvalidPage(t *testing.T) {
	url := "https://localhost:8080/resource"
	req := mockPageRequest(url, "foo", "10")
	_, _, err := ParsePagination(req)
	assert.NotNil(t, err)
}

func TestParsePaginationInvalidPerPage(t *testing.T) {
	url := "https://localhost:8080/resource"
	req := mockPageRequest(url, "1", "bar")
	_, _, err := ParsePagination(req)
	assert.NotNil(t, err)
}

func TestParsePaginationDefault(t *testing.T) {
	url := "https://localhost:8080/resource"
	req := mockPageRequest(url, "", "")
	page, per_page, err := ParsePagination(req)
	assert.Nil(t, err)
	assert.Equal(t, uint64(PageDefault), page)
	assert.Equal(t, uint64(PerPageDefault), per_page)
}
