// Copyright 2020 Northern.tech AS
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
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/mendersoftware/mender-server/pkg/identity"
)

const headerAuthorization = "Authorization"

// IdentityMiddleware is a gin middleware which extracts the identity from the JWT token
func IdentityMiddleware(c *gin.Context) {
	req := c.Request
	ctx := req.Context()

	var idata identity.Identity
	var err error

	if jwt := extractTokenFromRequest(req); jwt != "" {
		idata, err = identity.ExtractIdentity(jwt)
	}
	if err == nil {
		ctx = identity.WithContext(ctx, &idata)
		c.Request = req.WithContext(ctx)
	}
}

func extractTokenFromRequest(req *http.Request) string {
	jwt := req.URL.Query().Get("jwt")
	if jwt == "" {
		auth := strings.Split(req.Header.Get(headerAuthorization), " ")
		if len(auth) == 2 && auth[0] == "Bearer" {
			jwt = auth[1]
		}
	}
	return jwt
}
