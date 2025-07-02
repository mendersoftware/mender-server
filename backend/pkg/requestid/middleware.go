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
package requestid

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/mendersoftware/mender-server/pkg/log"
)

const RequestIdHeader = "X-MEN-RequestID"

type MiddlewareOptions struct {
	// GenerateRequestID decides whether a request ID should
	// be generated when none exists. (default: true)
	GenerateRequestID *bool
}

func NewMiddlewareOptions() *MiddlewareOptions {
	return new(MiddlewareOptions)
}

func (opt *MiddlewareOptions) SetGenerateRequestID(gen bool) *MiddlewareOptions {
	opt.GenerateRequestID = &gen
	return opt
}

// Middleware provides requestid middleware for the gin-gonic framework.
func Middleware(opts ...*MiddlewareOptions) gin.HandlerFunc {
	opt := NewMiddlewareOptions().
		SetGenerateRequestID(true)
	for _, o := range opts {
		if o == nil {
			continue
		}
		if o.GenerateRequestID != nil {
			opt.GenerateRequestID = o.GenerateRequestID
		}
	}
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		requestID := c.GetHeader(RequestIdHeader)
		if requestID == "" && *opt.GenerateRequestID {
			uid, _ := uuid.NewRandom()
			requestID = uid.String()
		}
		ctx = WithContext(ctx, requestID)

		logger := log.FromContext(ctx)
		if logger != nil {
			logger = logger.F(log.Ctx{"request_id": requestID})
			ctx = log.WithContext(ctx, logger)
		}
		c.Header(RequestIdHeader, requestID)
		c.Request = c.Request.WithContext(ctx)
	}
}
