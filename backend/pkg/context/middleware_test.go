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
	"net/http"
	"testing"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/requestid"
	"github.com/mendersoftware/mender-server/pkg/requestlog"
)

func TestUpdateContextMiddlewareMiddleware(t *testing.T) {
	api := rest.NewApi()

	api.Use(
		&requestlog.RequestLogMiddleware{
			BaseLogger: log.NewEmpty().Logger,
			LogContext: log.Ctx{"foo": "bar"},
		},
		&requestid.RequestIdMiddleware{},
		&UpdateContextMiddleware{
			Updates: []UpdateContextFunc{
				RepackLoggerToContext,
				RepackRequestIdToContext,
			},
		},
	)

	api.SetApp(rest.AppSimple(func(w rest.ResponseWriter, r *rest.Request) {
		cl := log.FromContext(r.Context())

		assert.Contains(t, cl.Data, "foo")
		assert.NotEmpty(t, requestid.FromContext(r.Context()))

		w.WriteHeader(http.StatusNoContent)
	}))

	handler := api.MakeHandler()

	req := test.MakeSimpleRequest("GET", "http://localhost/", nil)

	_ = test.RunRequest(t, handler, req)
}
