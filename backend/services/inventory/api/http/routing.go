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
package http

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/mendersoftware/mender-server/services/inventory/utils"
)

type HttpOptionsGenerator func(methods []string) rest.HandlerFunc

func AllowHeaderOptionsGenerator(methods []string) rest.HandlerFunc {
	// return a dummy handler for now
	return func(w rest.ResponseWriter, r *rest.Request) {
		for _, m := range methods {
			w.Header().Add("Allow", m)
		}
	}
}

func supportsMethod(method string, methods []string) bool {
	return utils.ContainsString(method, methods)
}

// Automatically add OPTIONS method support for each defined route,
// only if there's no OPTIONS handler for that route yet
func AutogenOptionsRoutes(routes []*rest.Route, gen HttpOptionsGenerator) []*rest.Route {

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
		// skip if there's a handler for OPTIONS already
		if !supportsMethod(http.MethodOptions, methods) {
			options = append(options,
				rest.Options(route, gen(methods)))
		}
	}

	return append(routes, options...)
}
