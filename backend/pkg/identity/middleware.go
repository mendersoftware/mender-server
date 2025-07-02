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

package identity

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"

	"github.com/mendersoftware/mender-server/pkg/log"
	urest "github.com/mendersoftware/mender-server/pkg/rest.utils"
)

type MiddlewareOptions struct {
	// PathRegex sets the regex for the path for which this middleware
	// applies. Defaults to "^/api/management/v[0-9.]{1,6}/.+".
	PathRegex *string

	// UpdateLogger adds the decoded identity to the log context.
	UpdateLogger *bool
}

func NewMiddlewareOptions() *MiddlewareOptions {
	return new(MiddlewareOptions)
}

func (opts *MiddlewareOptions) SetPathRegex(regex string) *MiddlewareOptions {
	opts.PathRegex = &regex
	return opts
}

func (opts *MiddlewareOptions) SetUpdateLogger(updateLogger bool) *MiddlewareOptions {
	opts.UpdateLogger = &updateLogger
	return opts
}

func middlewareWithLogger(c *gin.Context) {
	var (
		err    error
		jwt    string
		idty   Identity
		logCtx = log.Ctx{}
		key    = "sub"
		ctx    = c.Request.Context()
		l      = log.FromContext(ctx)
	)
	if FromContext(ctx) != nil {
		c.Next()
		return
	}
	jwt, err = ExtractJWTFromHeader(c.Request)
	if err != nil {
		goto exitUnauthorized
	}
	idty, err = ExtractIdentity(jwt)
	if err != nil {
		goto exitUnauthorized
	}
	ctx = WithContext(ctx, &idty)
	if idty.IsDevice {
		key = "device_id"
	} else if idty.IsUser {
		key = "user_id"
	}
	logCtx[key] = idty.Subject
	if idty.Tenant != "" {
		logCtx["tenant_id"] = idty.Tenant
	}
	if idty.Plan != "" {
		logCtx["plan"] = idty.Plan
	}
	ctx = log.WithContext(ctx, l.F(logCtx))

	c.Request = c.Request.WithContext(ctx)
	return
exitUnauthorized:
	c.Header("WWW-Authenticate", `Bearer realm="ManagementJWT"`)
	urest.RenderError(c, http.StatusUnauthorized, err)
	c.Abort()
}

func middlewareBase(c *gin.Context) {
	var (
		err  error
		jwt  string
		idty Identity
		ctx  = c.Request.Context()
	)
	jwt, err = ExtractJWTFromHeader(c.Request)
	if err != nil {
		goto exitUnauthorized
	}
	idty, err = ExtractIdentity(jwt)
	if err != nil {
		goto exitUnauthorized
	}
	ctx = WithContext(ctx, &idty)
	c.Request = c.Request.WithContext(ctx)
	return
exitUnauthorized:
	c.Header("WWW-Authenticate", `Bearer realm="ManagementJWT"`)
	urest.RenderError(c, http.StatusUnauthorized, err)
	c.Abort()
}

func Middleware(opts ...*MiddlewareOptions) gin.HandlerFunc {

	var middleware gin.HandlerFunc

	// Initialize default options
	opt := NewMiddlewareOptions().
		SetUpdateLogger(true)
	for _, o := range opts {
		if o == nil {
			continue
		}
		if o.PathRegex != nil {
			opt.PathRegex = o.PathRegex
		}
		if o.UpdateLogger != nil {
			opt.UpdateLogger = o.UpdateLogger
		}
	}

	if *opt.UpdateLogger {
		middleware = middlewareWithLogger
	} else {
		middleware = middlewareBase
	}

	if opt.PathRegex != nil {
		pathRegex := regexp.MustCompile(*opt.PathRegex)
		return func(c *gin.Context) {
			if !pathRegex.MatchString(c.FullPath()) {
				return
			}
			middleware(c)
		}
	}
	return middleware
}
