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

package http

import (
	"net/http"
)

func allowAllOrigins(r *http.Request) bool { return true }

func SetAcceptedOrigins(origins []string) {
	switch len(origins) {
	case 0:
		wsUpgrader.CheckOrigin = allowAllOrigins
	case 1:
		origin := origins[0]
		wsUpgrader.CheckOrigin = func(r *http.Request) bool {
			if actual, ok := r.Header[HdrKeyOrigin]; ok {
				return len(actual) > 0 && origin == actual[0]
			}
			// Origin header not present
			return true
		}
	default:
		// Compile a hashmap of valid origins for fast lookup
		originSet := make(map[string]struct{}, len(origins))
		for _, origin := range origins {
			originSet[origin] = struct{}{}
		}
		wsUpgrader.CheckOrigin = func(r *http.Request) bool {
			if actual, ok := r.Header[HdrKeyOrigin]; ok {
				if len(actual) == 0 {
					return false
				}
				_, allowed := originSet[actual[0]]
				return allowed
			}
			// Origin header not present
			return true
		}
	}
}
