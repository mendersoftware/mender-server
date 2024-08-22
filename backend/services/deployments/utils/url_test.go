// Copyright 2023 Northern.tech AS
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

package utils

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRewriteProxyURL(t *testing.T) {
	var (
		origin = &url.URL{
			Scheme:   "https",
			Host:     "localhost:1234",
			Path:     "/origin/path",
			Fragment: "thisisafragment",
		}
		proxy = &url.URL{
			Scheme:   "http",
			Host:     "localhost:8080",
			Path:     "proxy/prefix",
			Fragment: "irrelevant",
		}
	)

	res, err := RewriteProxyURL(origin, proxy)
	assert.NoError(t, err)
	assert.Equal(t, proxy.Scheme, res.Scheme)
	assert.Equal(t, proxy.Host, res.Host)
	expectedPath, _ := url.JoinPath(proxy.Path, origin.Path)
	assert.Equal(t, expectedPath, res.Path)
	assert.Equal(t, origin.Fragment, res.Fragment)

	res, err = RewriteProxyURL(origin, nil)
	assert.NoError(t, err)
	assert.Equal(t, origin, res)
}
