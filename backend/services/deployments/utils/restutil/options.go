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
package restutil

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
)

const (
	HttpHeaderAllow string = "Allow"
)

type CreateOptionsHandler func(methods ...string) rest.HandlerFunc

type OptionsHandler struct {
	// Shared  reads, need locking of any write mathod is introduced.
	methods map[string]bool
}

// NewOptionsHandler creates http handler object that will server OPTIONS method requests,
// Accepts a list of http methods.
// Adds information that it serves OPTIONS method automatically.
func NewOptionsHandler(methods ...string) rest.HandlerFunc {
	handler := &OptionsHandler{
		methods: make(map[string]bool, len(methods)+1),
	}

	for _, method := range methods {
		handler.methods[method] = true
	}

	if _, ok := handler.methods[http.MethodOptions]; !ok {
		handler.methods[http.MethodOptions] = true
	}

	return handler.handle
}

// Handle is a method for handling OPTIONS method requests.
// This method is called concurrently while serving requests and should not modify self.
func (o *OptionsHandler) handle(w rest.ResponseWriter, r *rest.Request) {
	for method := range o.methods {
		w.Header().Add(HttpHeaderAllow, method)
	}
}

// AutogenOptionsRoutes automatically add OPTIONS method support for each defined route.
func AutogenOptionsRoutes(createHandler CreateOptionsHandler, routes ...*rest.Route) []*rest.Route {

	methodGroups := make(map[string][]string, len(routes))

	for _, route := range routes {
		methods, ok := methodGroups[route.PathExp]
		if !ok {
			methods = make([]string, 0)
		}

		methodGroups[route.PathExp] = append(methods, route.HttpMethod)
	}

	options := make([]*rest.Route, 0, len(methodGroups))
	for route, methods := range methodGroups {
		options = append(options, rest.Options(route, createHandler(methods...)))
	}

	return append(routes, options...)
}
