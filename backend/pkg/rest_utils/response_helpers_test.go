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

package rest_utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/pkg/accesslog"
	"github.com/mendersoftware/mender-server/pkg/log"
)

type logCounter struct {
	n int
}

func (logCounter) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (l *logCounter) Fire(*logrus.Entry) error {
	l.n++
	return nil
}

func TestResponseHelpers(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string
		CTX  context.Context

		HandlerFunc rest.HandlerFunc
		NumEntries  int

		Fields       []string
		ExpectedBody string
	}{{
		Name: "internal",

		NumEntries: 1,
		HandlerFunc: func(w rest.ResponseWriter, r *rest.Request) {
			RestErrWithLogInternal(w, r, log.NewEmpty(), errors.New("test error"))
		},
		Fields: []string{
			`level=error`,
			`error="(?P<callerFrame>rest_utils.TestResponseHelpers[^@]+` +
				`@[^:]+:[0-9]+:) internal error: test error"`,
		},
		ExpectedBody: func() string {
			b, _ := json.Marshal(ApiError{Err: "internal error"})
			return string(b)
		}(),
	}, {
		Name: "client error",

		NumEntries: 1,
		HandlerFunc: func(w rest.ResponseWriter, r *rest.Request) {
			RestErrWithWarningMsg(w, r, log.NewEmpty(),
				errors.New("test error"), http.StatusBadRequest, "bad request")
		},
		Fields: []string{
			`level=warn`,
			`error="(?P<callerFrame>rest_utils.TestResponseHelpers[^@]+` +
				`@[^:]+:[0-9]+:) bad request: test error"`,
		},
		ExpectedBody: func() string {
			b, _ := json.Marshal(ApiError{Err: "bad request"})
			return string(b)
		}(),
	}, {
		Name: "fallback to logger",

		NumEntries: 2,
		HandlerFunc: func(w rest.ResponseWriter, r *rest.Request) {
			lc := accesslog.GetContext(r.Request.Context())
			e := errors.New("test")
			i := 0
			for lc.PushError(e) && i < 10000 {
				i++
			}
			if i >= 10000 {
				// Guard against breaking the accesslog
				t.Error("should not be able to push 10000 errors to accesslog")
				t.FailNow()
			}
			RestErrWithWarningMsg(w, r, log.NewEmpty(),
				errors.New("test error"), http.StatusBadRequest, "bad request")
		},
		Fields: []string{
			`level=warn`,
			`msg="bad request: test error"`,
		},
		ExpectedBody: func() string {
			b, _ := json.Marshal(ApiError{Err: "bad request"})
			return string(b)
		}(),
	}}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			app, err := rest.MakeRouter(rest.Get("/test", tc.HandlerFunc))
			if err != nil {
				t.Error(err)
				t.FailNow()
			}
			counter := &logCounter{}
			api := rest.NewApi()
			var logBuf = bytes.NewBuffer(nil)
			api.Use(rest.MiddlewareSimple(
				func(h rest.HandlerFunc) rest.HandlerFunc {
					logger := log.NewEmpty()
					logger.Logger.SetLevel(logrus.DebugLevel)
					logger.Logger.SetOutput(logBuf)
					logger.Logger.SetFormatter(&logrus.TextFormatter{
						DisableColors: true,
						FullTimestamp: true,
					})
					logger.Logger.AddHook(counter)
					return func(w rest.ResponseWriter, r *rest.Request) {
						ctx := r.Request.Context()
						ctx = log.WithContext(ctx, logger)
						r.Request = r.Request.WithContext(ctx)
						h(w, r)
					}
				}))
			api.Use(&accesslog.AccessLogMiddleware{})
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
			assert.Equal(t, tc.NumEntries, counter.n)
		})
	}
}
