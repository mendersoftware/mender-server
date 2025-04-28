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
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/mendersoftware/mender-server/services/deviceauth/utils"
)

type HttpOptionsGenerator func(methods []string) gin.HandlerFunc

func AllowHeaderOptionsGenerator(methods []string) gin.HandlerFunc {
	// return a dummy handler for now
	return func(c *gin.Context) {
		for _, m := range methods {
			c.Writer.Header().Add("Allow", m)
		}
	}
}

func supportsMethod(method string, methods []string) bool {
	return utils.ContainsString(method, methods)
}

// Automatically add OPTIONS method support for each defined route,
// only if there's no OPTIONS handler for that route yet
func AutogenOptionsRoutes(router *gin.Engine, gen HttpOptionsGenerator) {

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
			router.OPTIONS(route, gen(methods))
		}
	}

}

func init() {
	if mode := os.Getenv(gin.EnvGinMode); mode != "" {
		gin.SetMode(mode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	gin.DisableConsoleColor()
}
