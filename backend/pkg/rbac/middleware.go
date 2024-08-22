// Copyright 2024 Northern.tech AS
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
package rbac

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/gin-gonic/gin"
)

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if scope := ExtractScopeFromHeader(c.Request); scope != nil {
			ctx := c.Request.Context()
			ctx = WithContext(ctx, scope)
			c.Request = c.Request.WithContext(ctx)
		}
	}
}

type RBACMiddleware struct {
}

func (mw *RBACMiddleware) MiddlewareFunc(h rest.HandlerFunc) rest.HandlerFunc {
	return func(w rest.ResponseWriter, r *rest.Request) {
		if scope := ExtractScopeFromHeader(r.Request); scope != nil {
			ctx := r.Context()
			ctx = WithContext(ctx, scope)
			r.Request = r.WithContext(ctx)
		}

		h(w, r)
	}
}
