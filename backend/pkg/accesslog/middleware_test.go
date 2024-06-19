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

package accesslog

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMiddlewareLegacy(t *testing.T) {
	testCases := []struct {
		Name string
		CTX  context.Context

		HandlerFunc rest.HandlerFunc

		Fields       []string
		ExpectedBody string
	}{{
		Name: "ok",

		HandlerFunc: func(w rest.ResponseWriter, r *rest.Request) {
			w.WriteHeader(http.StatusNoContent)
		},
		Fields: []string{
			"status=204",
			`path=/test`,
			`qs="foo=bar"`,
			"method=GET",
			"responsetime=",
			"ts=",
		},
	}, {
		Name: "canceled context",

		CTX: func() context.Context {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			return ctx
		}(),
		HandlerFunc: func(w rest.ResponseWriter, r *rest.Request) {
			w.WriteHeader(http.StatusNoContent)
		},
		Fields: []string{
			"status=499",
			`path=/test`,
			`qs="foo=bar"`,
			"method=GET",
			"responsetime=",
			"ts=",
		},
	}, {
		Name: "error, panic in handler",

		HandlerFunc: func(w rest.ResponseWriter, r *rest.Request) {
			panic("!!!!!")
		},

		Fields: []string{
			"status=500",
			`path=/test`,
			`qs="foo=bar"`,
			"method=GET",
			"responsetime=",
			"ts=",
			// First three entries in the trace should match this:
			`trace=".+TestMiddlewareLegacy\.func[0-9.]*@middleware_test\.go:[0-9.]+\\n`,
		},
		ExpectedBody: `{"Error": "Internal Server Error"}`,
	}}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			app, err := rest.MakeRouter(rest.Get("/test", tc.HandlerFunc))
			if err != nil {
				t.Error(err)
				t.FailNow()
			}
			api := rest.NewApi()
			var logBuf = bytes.NewBuffer(nil)
			api.Use(rest.MiddlewareSimple(
				func(h rest.HandlerFunc) rest.HandlerFunc {
					logger := log.NewEmpty()
					logger.Logger.SetLevel(logrus.InfoLevel)
					logger.Logger.SetOutput(logBuf)
					logger.Logger.SetFormatter(&logrus.TextFormatter{
						DisableColors: true,
						FullTimestamp: true,
					})
					return func(w rest.ResponseWriter, r *rest.Request) {
						ctx := r.Request.Context()
						ctx = log.WithContext(ctx, logger)
						r.Request = r.Request.WithContext(ctx)
						h(w, r)
						t.Log(r.Env)
					}
				}))
			api.Use(&AccessLogMiddleware{})
			api.SetApp(app)
			handler := api.MakeHandler()
			w := httptest.NewRecorder()
			ctx := context.Background()
			if tc.CTX != nil {
				ctx = tc.CTX
			}
			req, _ := http.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"http://localhost/test?foo=bar",
				nil,
			)
			req.Header.Set("User-Agent", "tester")

			handler.ServeHTTP(w, req)

			logEntry := logBuf.String()
			for _, field := range tc.Fields {
				assert.Regexp(t, field, logEntry)
			}
			if tc.Fields == nil {
				assert.Empty(t, logEntry)
			}
			if tc.ExpectedBody != "" {
				if assert.NotNil(t, w.Body) {
					assert.JSONEq(t, tc.ExpectedBody, w.Body.String())
				}
			}
		})
	}
}
