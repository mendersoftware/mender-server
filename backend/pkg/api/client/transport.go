// Copyright 2025 Northern.tech AS
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

// Package client provides shared utilities for generated API clients.
package client

import (
	"net/http"

	ctxhttpheader "github.com/mendersoftware/mender-server/pkg/context/httpheader"
	"github.com/mendersoftware/mender-server/pkg/requestid"
)

// MenderTransport is an http.RoundTripper that automatically propagates
// Mender-specific headers from the request context:
//   - X-MEN-RequestID: for request tracing
//   - Authorization: for forwarding auth tokens between services
type MenderTransport struct {
	// Base is the underlying RoundTripper. If nil, http.DefaultTransport is used.
	Base http.RoundTripper
}

// RoundTrip implements http.RoundTripper and adds Mender headers from context.
func (t *MenderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	// Propagate X-MEN-RequestID from context
	if reqID := requestid.FromContext(ctx); reqID != "" {
		if req.Header.Get(requestid.RequestIdHeader) == "" {
			req.Header.Set(requestid.RequestIdHeader, reqID)
		}
	}

	// Propagate Authorization from context
	if auth := ctxhttpheader.FromContext(ctx, "Authorization"); auth != "" {
		if req.Header.Get("Authorization") == "" {
			req.Header.Set("Authorization", auth)
		}
	}

	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}
