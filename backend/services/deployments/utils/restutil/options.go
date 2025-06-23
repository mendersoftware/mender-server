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
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/mendersoftware/mender-server/services/inventory/utils"
)

const (
	HttpHeaderAllow string = "Allow"
)

type CreateOptionsHandler func(methods ...string) gin.HandlerFunc

type OptionsHandler struct {
	// Shared  reads, need locking of any write mathod is introduced.
	methods map[string]bool
}

func supportsMethod(method string, methods []string) bool {
	return utils.ContainsString(method, methods)
}

// NewOptionsHandler creates http handler object that will server OPTIONS method requests,
// Accepts a list of http methods.
// Adds information that it serves OPTIONS method automatically.
func NewOptionsHandler(methods ...string) gin.HandlerFunc {
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
func (o *OptionsHandler) handle(c *gin.Context) {
	for method := range o.methods {
		c.Writer.Header().Add(HttpHeaderAllow, method)
	}
}

// AutogenOptionsRoutes automatically add OPTIONS method support for each defined route.
func AutogenOptionsRoutes(createHandler CreateOptionsHandler, router *gin.Engine) {

	routes := router.Routes()
	methodGroups := make(map[string][]string, len(routes))

	for _, route := range routes {
		if strings.HasPrefix(route.Path, "/api/internal") {
			continue
		}
		methods, ok := methodGroups[route.Path]
		if !ok {
			methods = make([]string, 0)
		}

		methodGroups[route.Path] = append(methods, route.Method)
	}

	for route, methods := range methodGroups {
		// skip if there's a handler for OPTIONS already
		if !supportsMethod(http.MethodOptions, methods) {
			router.OPTIONS(route, createHandler(methods...))
		}
	}
}
