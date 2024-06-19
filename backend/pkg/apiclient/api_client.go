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
package apiclient

import (
	"net/http"

	"github.com/mendersoftware/mender-server/pkg/requestid"

	ctxhttpheader "github.com/mendersoftware/mender-server/pkg/context/httpheader"
)

// maybeSetHeader sets HTTP header `hdr` to value `val` if `val` is not empty or
// the header is not yet set.
func maybeSetHeader(hdrs http.Header, hdr string, val string) {
	if val == "" {
		return
	}

	if hdrs.Get(hdr) == "" {
		hdrs.Add(hdr, val)
	}
}

// HttpApi is an http.Client wrapper tailored to use with mender's APIs.
type HttpApi struct {
}

// Do behaves similarly to http.Client.Do(), but will also automatically add
// mender related headers, if these can be built based on request's context. The
// headers are:
// - X-Mender-RequestId - extracted with requestid.FromContext()
// - Authorization - extracted with httpheader.FromContext()
// If given header is already set, the value from context will not be used
func (a *HttpApi) Do(r *http.Request) (*http.Response, error) {
	client := &http.Client{}
	ctx := r.Context()

	maybeSetHeader(r.Header, requestid.RequestIdHeader,
		requestid.FromContext(ctx))
	maybeSetHeader(r.Header, "Authorization",
		ctxhttpheader.FromContext(ctx, "Authorization"))

	return client.Do(r)
}
