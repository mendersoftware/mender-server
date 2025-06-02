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
package context

import (
	"context"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/requestlog"
)

// RepackLoggerToContext can be used to attach a request specific logger
// assigned by RequestLogMiddleware to context ctx. The logger can later be
// accessed using log.FromContext(ctx).
func RepackLoggerToContext(ctx context.Context, r *rest.Request) context.Context {
	return log.WithContext(ctx,
		requestlog.GetRequestLogger(r))
}

// RepackRequestIdToContext can be used to attach a request ID assigned by
// RequestIdMiddleware to context. Request ID can later be accessed using
// requestid.FromContext(ctx).
func RepackRequestIdToContext(ctx context.Context, r *rest.Request) context.Context {
	return requestid.WithContext(ctx,
		requestid.GetReqId(r.Request))
}

// UpdateContextFunc is a function that can update context ctx using data from
// rest.Request and return modified context.
type UpdateContextFunc func(ctx context.Context, r *rest.Request) context.Context

// UpdateContextMiddleware is a middleware that can be used to update
// http.Request context. The middleware will apply user provided context
// modifications listed in Updates. The middleware operates on rest.Request and
// context is owned/assigned to http.Request. Because of this a new rest.Request
// will be allocated before passing it further in the stack.
//
// When combined with RepackRequestIdToContext and RepackLoggerToContext, the
// middleware will populate http.Request context with both request ID and
// request specific logger, Later it is possible to call
// log.FromContext(r.Context()) to obtain context logger or
// requestid.FromContext(r.Context()) to get request ID.
type UpdateContextMiddleware struct {
	Updates []UpdateContextFunc
}

func (ucmw *UpdateContextMiddleware) MiddlewareFunc(h rest.HandlerFunc) rest.HandlerFunc {
	return func(w rest.ResponseWriter, r *rest.Request) {

		ctx := r.Context()
		for _, up := range ucmw.Updates {
			ctx = up(ctx, r)
		}

		r.Request = r.WithContext(ctx)

		h(w, r)
	}
}
