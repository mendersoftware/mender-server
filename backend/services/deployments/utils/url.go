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
)

// RewriteProxyURL replaces the base part of origin (everything up to the path)
// with the proxy URL. The query paremeters from proxy are copied.
// If proxy is nil, then origin is returned as is.
func RewriteProxyURL(origin, proxy *url.URL) (*url.URL, error) {
	if proxy == nil {
		return origin, nil
	}
	new := proxy.JoinPath(origin.Path)
	// Join query parameters
	q := origin.Query()
	for key, value := range proxy.Query() {
		q[key] = value
	}
	new.RawQuery = q.Encode()
	new.Fragment = origin.Fragment
	return new, nil
}
