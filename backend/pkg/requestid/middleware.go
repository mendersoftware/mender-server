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
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/requestlog"
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

// RequestIdMiddleware sets the X-MEN-RequestID header if it's not present,
// and adds the request id to the request logger's context.
type RequestIdMiddleware struct {
}

// MiddlewareFunc makes RequestIdMiddleware implement the Middleware interface.
func (mw *RequestIdMiddleware) MiddlewareFunc(h rest.HandlerFunc) rest.HandlerFunc {
	return func(w rest.ResponseWriter, r *rest.Request) {
		logger := requestlog.GetRequestLogger(r)

		reqId := r.Header.Get(RequestIdHeader)
		if reqId == "" {
			uid, _ := uuid.NewRandom()
			reqId = uid.String()
		}

		r = SetReqId(r, reqId)

		// enrich log context
		if logger != nil {
			logger = logger.F(log.Ctx{"request_id": reqId})
			r = requestlog.SetRequestLogger(r, logger)
		}

		//return the reuqest ID in response too, the client can log it
		//for end-to-end req tracing
		w.Header().Add(RequestIdHeader, reqId)

		h(w, r)
	}
}
