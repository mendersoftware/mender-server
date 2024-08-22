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
package requestid

import (
	"context"

	"github.com/ant0ine/go-json-rest/rest"
)

type requestIdKeyType int

const (
	requestIdKey requestIdKeyType = 0
)

// GetReqId helper for retrieving current request Id
func GetReqId(r *rest.Request) string {
	return FromContext(r.Context())
}

// SetReqId is a helper for setting request ID in request context
func SetReqId(r *rest.Request, reqid string) *rest.Request {
	ctx := WithContext(r.Context(), reqid)
	r.Request = r.Request.WithContext(ctx)
	return r
}

// FromContext extracts current request Id from context.Context
func FromContext(ctx context.Context) string {
	val := ctx.Value(requestIdKey)
	if v, ok := val.(string); ok {
		return v
	}
	return ""
}

// WithContext adds request to context `ctx` and returns the resulting context.
func WithContext(ctx context.Context, reqid string) context.Context {
	return context.WithValue(ctx, requestIdKey, reqid)
}
