// Copyright 2026 Northern.tech AS
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

package middlewares

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"
)

// netError is a minimal net.Error implementation for exercising the
// timeout branch of ErrorHandler.
type netError struct {
	msg     string
	timeout bool
}

func (e *netError) Error() string   { return e.msg }
func (e *netError) Timeout() bool   { return e.timeout }
func (e *netError) Temporary() bool { return false }

func TestErrorHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name         string
		handler      gin.HandlerFunc
		expectedCode int
		expectedBody string
		checkBody    bool
		rawBody      bool
	}{
		{
			name:         "no error",
			handler:      func(c *gin.Context) {},
			expectedCode: http.StatusOK,
			expectedBody: "",
			checkBody:    true,
		},
		{
			name: "response already written is left untouched",
			handler: func(c *gin.Context) {
				c.String(http.StatusTeapot, "already written")
				_ = c.Error(errors.New("this error must be ignored"))
			},
			expectedCode: http.StatusTeapot,
			expectedBody: "already written",
			checkBody:    true,
			rawBody:      true,
		},
		{
			name: "net.Error timeout maps to gateway timeout",
			handler: func(c *gin.Context) {
				_ = c.Error(&netError{msg: "i/o timeout", timeout: true})
			},
			expectedCode: http.StatusGatewayTimeout,
			expectedBody: `{"error":"gateway timeout"}`,
			checkBody:    true,
		},
		{
			name: "context.DeadlineExceeded maps to gateway timeout",
			handler: func(c *gin.Context) {
				_ = c.Error(context.DeadlineExceeded)
			},
			expectedCode: http.StatusGatewayTimeout,
			expectedBody: `{"error":"gateway timeout"}`,
			checkBody:    true,
		},
		{
			name: "non-timeout net.Error falls through to internal error",
			handler: func(c *gin.Context) {
				_ = c.Error(&netError{msg: "connection reset", timeout: false})
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"error":"internal error"}`,
			checkBody:    true,
		},
		{
			name: "max bytes error maps to request entity too large",
			handler: func(c *gin.Context) {
				_ = c.Error(&http.MaxBytesError{Limit: 1024})
			},
			expectedCode: http.StatusRequestEntityTooLarge,
			checkBody:    false,
		},
		{
			name: "gin bind error maps to bad request",
			handler: func(c *gin.Context) {
				_ = c.Error(errors.New("invalid request body")).SetType(gin.ErrorTypeBind)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"invalid request body"}`,
			checkBody:    true,
		},
		{
			name: "validation error maps to bad request",
			handler: func(c *gin.Context) {
				_ = c.Error(validation.Errors{
					"name": errors.New("cannot be blank"),
				})
			},
			expectedCode: http.StatusBadRequest,
			checkBody:    false,
		},
		{
			name: "context canceled maps to 499",
			handler: func(c *gin.Context) {
				_ = c.Error(context.Canceled)
			},
			expectedCode: 499,
			expectedBody: "",
			checkBody:    true,
		},
		{
			name: "unhandled error maps to internal error",
			handler: func(c *gin.Context) {
				_ = c.Error(errors.New("boom"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"error":"internal error"}`,
			checkBody:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			engine := gin.New()
			engine.Use(ErrorHandler())
			engine.GET("/test", tc.handler)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			engine.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedCode, w.Code)
			if tc.checkBody {
				switch {
				case tc.expectedBody == "":
					assert.Empty(t, w.Body.String())
				case tc.rawBody:
					assert.Equal(t, tc.expectedBody, w.Body.String())
				default:
					assert.JSONEq(t, tc.expectedBody, w.Body.String())
				}
			}
		})
	}
}
