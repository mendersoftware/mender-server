// Copyright 2021 Northern.tech AS
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
package http

import (
	"errors"
	"net/http"
	"strings"
)

// Action combines info about the requested resourd + http method.
type Action struct {
	Resource string
	Method   string
}

// ExtractResourceAction extracts resource action from the request url
func ExtractResourceAction(r *http.Request) (*Action, error) {
	action := Action{}

	// extract original uri
	uri := r.Header.Get("X-Forwarded-Uri")
	if uri == "" {
		uri = r.Header.Get("X-Forwarded-URI")
	}
	uriItems := strings.Split(uri, "/")

	if uri == "" || len(uriItems) < 4 {
		return nil, errors.New("can't parse service name from original uri " + uri)
	}

	action.Resource = strings.Join(uriItems[4:], ":")

	// extract original http method
	action.Method = r.Header.Get("X-Forwarded-Method")
	if action.Method == "" {
		action.Method = r.Header.Get("X-Forwarded-Method")
	}
	if action.Method == "" {
		return nil, errors.New("can't parse original request method")
	}

	return &action, nil
}

// extracts JWT from authorization header
func ExtractToken(req *http.Request) (string, error) {
	const authHeaderName = "Authorization"
	auth := req.Header.Get(authHeaderName)
	if auth != "" {
		auths := strings.Fields(auth)
		if !strings.EqualFold(auths[0], "Bearer") || len(auths) < 2 {
			return "", ErrInvalidAuthHeader
		}
		return auths[1], nil
	}
	cookie, err := req.Cookie("JWT")
	if err != nil {
		return "", ErrAuthzNoAuth
	}
	auth = cookie.Value
	if auth == "" {
		return "", ErrAuthzNoAuth
	}
	return auth, nil
}
