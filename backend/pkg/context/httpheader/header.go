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
package httpheader

import (
	"context"
	"net/http"
	"net/textproto"
)

type headerKeyType string

func makeKeyName(hdr string) headerKeyType {
	key := textproto.CanonicalMIMEHeaderKey(hdr)
	return headerKeyType(key)
}

// WithContext stores HTTP headers from `hdrs` which listed in `which` in a
// context and returns the new context (ctx becomes the parent of newly created
// context). Headers care provided as http.Header. Headers that are unset in
// `hdrs` are skipped. Empty header names are skipped as well. Headers are
// stored using httpheader package specific key namespace.
func WithContext(ctx context.Context, hdrs http.Header, which ...string) context.Context {
	if len(hdrs) == 0 {
		return ctx
	}
	if len(which) == 0 {
		return ctx
	}

	for _, h := range which {
		if h == "" {
			continue
		}
		hv := hdrs.Get(h)
		if hv == "" {
			continue
		}
		ctx = context.WithValue(ctx, makeKeyName(h), hdrs.Get(h))
	}
	return ctx
}

// FromContext extracts httpheader header and returns a string. If header was
// not set in the context, an empty string is returned.
func FromContext(ctx context.Context, hdr string) string {
	v, ok := ctx.Value(makeKeyName(hdr)).(string)
	if !ok {
		return ""
	}
	return v
}
