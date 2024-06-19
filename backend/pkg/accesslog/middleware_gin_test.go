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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware(t *testing.T) {
	testCases := []struct {
		Name string

		HandlerFunc gin.HandlerFunc

		Fields       []string
		ExpectedBody string
	}{{
		Name: "ok",

		HandlerFunc: func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		},
		Fields: []string{
			"status=204",
			`path=/test`,
			`qs="foo=bar"`,
			"method=GET",
			"useragent=tester",
			"responsetime=",
			"ts=",
		},
	}, {
		Name: "ok, pushed error",

		HandlerFunc: func(c *gin.Context) {
			err := errors.New("internal error")
			_ = c.Error(err)
			c.Status(http.StatusInternalServerError)
			_, _ = c.Writer.Write([]byte(err.Error()))
		},
		Fields: []string{
			"status=500",
			`path=/test`,
			`qs="foo=bar"`,
			"method=GET",
			"responsetime=",
			"byteswritten=14",
			"ts=",
			`error="internal error"`,
		},
	}, {
		Name: "ok, pushed multiple errors",

		HandlerFunc: func(c *gin.Context) {
			err := errors.New("internal error 1")
			_ = c.Error(err)
			err = errors.New("internal error 2")
			_ = c.Error(err)
			c.Status(http.StatusInternalServerError)
			c.Writer.Write([]byte(err.Error()))
		},
		Fields: []string{
			"status=500",
			`path=/test`,
			`qs="foo=bar"`,
			"useragent=tester",
			"method=GET",
			"responsetime=",
			"byteswritten=16",
			"ts=",
			`error="#01: internal error 1\\n#02: internal error 2\\n"`,
		},
	}, {
		Name: "error, panic in handler",

		HandlerFunc: func(c *gin.Context) {
			panic("!!!!!")
		},

		Fields: []string{
			"status=500",
			`path=/test`,
			`qs="foo=bar"`,
			"method=GET",
			"responsetime=",
			"useragent=tester",
			"ts=",
			// First three entries in the trace should match this:
			`trace=".+TestMiddleware\.func[0-9]*@middleware_gin_test\.go:[0-9]+\\n`,
		},
		ExpectedBody: `{"error": "internal error"}`,
	}}

	gin.SetMode(gin.ReleaseMode)
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			var logBuf = bytes.NewBuffer(nil)
			router := gin.New()
			router.Use(func(c *gin.Context) {
				logger := log.NewEmpty()
				logger.Logger.SetLevel(logrus.InfoLevel)
				logger.Logger.SetOutput(logBuf)
				logger.Logger.SetFormatter(&logrus.TextFormatter{
					DisableColors: true,
					FullTimestamp: true,
				})
				ctx := c.Request.Context()
				ctx = log.WithContext(ctx, logger)
				c.Request = c.Request.WithContext(ctx)
			})
			router.Use(Middleware())
			router.GET("/test", tc.HandlerFunc)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(
				http.MethodGet,
				"http://localhost/test?foo=bar",
				nil,
			)
			req.Header.Set("User-Agent", "tester")

			router.ServeHTTP(w, req)

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
