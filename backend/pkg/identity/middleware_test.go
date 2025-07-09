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
package identity

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/pkg/log"
	urest "github.com/mendersoftware/mender-server/pkg/rest.utils"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func makeFakeAuth(idty Identity) string {
	b, _ := json.Marshal(idty)
	claims := base64.RawURLEncoding.EncodeToString(b)
	return "aGVhZGVy." + claims + ".c2lnbg"
}

func TestGinMiddleware(t *testing.T) {
	testCases := []struct {
		Name string

		Request *http.Request
		Options *MiddlewareOptions

		Validator func(t *testing.T,
			w *httptest.ResponseRecorder, req *http.Request,
		)
	}{{
		Name: "ok, user",
		Request: func() *http.Request {
			req, _ := http.NewRequest("GET",
				"http://localhost/api/management/v1/test?foo=bar",
				nil,
			)
			req.Header.Set("Authorization",
				"Bearer "+makeFakeAuth(Identity{
					Subject: "3e955f9d-53bf-47d6-a182-ff27b2c96282",
					Tenant:  "123456789012345678901234",
					IsUser:  true,
					Plan:    "professional",
				}),
			)
			return req
		}(),

		Validator: func(t *testing.T,
			w *httptest.ResponseRecorder, req *http.Request,
		) {
			ctx := req.Context()
			expected := &Identity{
				Subject: "3e955f9d-53bf-47d6-a182-ff27b2c96282",
				Tenant:  "123456789012345678901234",
				IsUser:  true,
				Plan:    "professional",
			}
			actual := FromContext(ctx)
			assert.EqualValues(t, expected, actual)
			logger := log.FromContext(ctx)
			assert.Equal(t,
				"3e955f9d-53bf-47d6-a182-ff27b2c96282",
				logger.Entry.Data["user_id"],
			)
			assert.Equal(t,
				"123456789012345678901234",
				logger.Entry.Data["tenant_id"],
			)
			assert.Equal(t,
				"professional",
				logger.Entry.Data["plan"],
			)
		},
	}, {
		Name: "ok, device",
		Request: func() *http.Request {
			req, _ := http.NewRequest("GET",
				"http://localhost/api/management/v1/test?foo=bar",
				nil,
			)
			req.Header.Set("Authorization",
				"Bearer "+makeFakeAuth(Identity{
					Subject:  "3e955f9d-53bf-47d6-a182-ff27b2c96282",
					Tenant:   "123456789012345678901234",
					IsDevice: true,
				}),
			)
			return req
		}(),

		Validator: func(t *testing.T,
			w *httptest.ResponseRecorder, req *http.Request,
		) {
			ctx := req.Context()
			expected := &Identity{
				Subject:  "3e955f9d-53bf-47d6-a182-ff27b2c96282",
				Tenant:   "123456789012345678901234",
				IsDevice: true,
			}
			actual := FromContext(ctx)
			assert.EqualValues(t, expected, actual)
			logger := log.FromContext(ctx)
			assert.Equal(t,
				"3e955f9d-53bf-47d6-a182-ff27b2c96282",
				logger.Entry.Data["device_id"],
			)
			assert.Equal(t,
				"123456789012345678901234",
				logger.Entry.Data["tenant_id"],
			)
		},
	}, {
		Name: "ok, with option override",
		Request: func() *http.Request {
			req, _ := http.NewRequest("GET",
				"http://localhost/api/management/v1/test?foo=bar",
				nil,
			)
			req.Header.Set("Authorization",
				"Bearer "+makeFakeAuth(Identity{
					Subject: "3e955f9d-53bf-47d6-a182-ff27b2c96282",
					Tenant:  "123456789012345678901234",
				}),
			)
			return req
		}(),
		Options: NewMiddlewareOptions().
			SetPathRegex("^/api/management/v1/test$").
			SetUpdateLogger(false),

		Validator: func(t *testing.T,
			w *httptest.ResponseRecorder, req *http.Request,
		) {
			ctx := req.Context()
			expected := &Identity{
				Subject: "3e955f9d-53bf-47d6-a182-ff27b2c96282",
				Tenant:  "123456789012345678901234",
			}
			actual := FromContext(ctx)
			assert.EqualValues(t, expected, actual)
			logger := log.FromContext(ctx)
			assert.Empty(t, logger.Entry.Data)
		},
	}, {
		Name: "ok, path does not match",
		Request: func() *http.Request {
			req, _ := http.NewRequest("GET",
				"http://localhost/api/management/",
				nil,
			)
			req.Header.Set("Authorization",
				"Bearer "+makeFakeAuth(Identity{
					Subject: "3e955f9d-53bf-47d6-a182-ff27b2c96282",
					Tenant:  "123456789012345678901234",
				}),
			)
			return req
		}(),
		Options: NewMiddlewareOptions().
			SetPathRegex("^/api/management/v1/test$"),

		Validator: func(t *testing.T,
			w *httptest.ResponseRecorder, req *http.Request,
		) {
			ctx := req.Context()
			actual := FromContext(ctx)
			assert.Nil(t, actual)
			logger := log.FromContext(ctx)
			assert.Empty(t, logger.Entry.Data)
		},
	}, {
		Name: "error, token not present (w/logger)",
		Request: func() *http.Request {
			req, _ := http.NewRequest("GET",
				"http://localhost/api/management/v1/test",
				nil,
			)
			return req
		}(),
		Options: NewMiddlewareOptions().
			SetPathRegex("^/api/management/v1/test$"),

		Validator: func(t *testing.T,
			w *httptest.ResponseRecorder, req *http.Request,
		) {
			assert.Equal(t, 401, w.Code)
			var apiErr urest.Error
			_ = json.Unmarshal(w.Body.Bytes(), &apiErr)
			assert.EqualError(t,
				apiErr,
				"Authorization not present in header",
			)
		},
	}, {
		Name: "error, token malformed (w/logger)",
		Request: func() *http.Request {
			req, _ := http.NewRequest("GET",
				"http://localhost/api/management/v1/test",
				nil,
			)
			req.Header.Set("Authorization", "Bearer bruh?==")
			return req
		}(),
		Options: NewMiddlewareOptions().
			SetPathRegex("^/api/management/v1/test$"),

		Validator: func(t *testing.T,
			w *httptest.ResponseRecorder, req *http.Request,
		) {
			assert.Equal(t, 401, w.Code)
			var apiErr urest.Error
			_ = json.Unmarshal(w.Body.Bytes(), &apiErr)
			assert.EqualError(t,
				apiErr,
				"identity: incorrect token format",
			)
		},
	}, {
		Name: "error, token not present (base middleware)",
		Request: func() *http.Request {
			req, _ := http.NewRequest("GET",
				"http://localhost/api/management/v1/test",
				nil,
			)
			return req
		}(),
		Options: NewMiddlewareOptions().
			SetUpdateLogger(false),

		Validator: func(t *testing.T,
			w *httptest.ResponseRecorder, req *http.Request,
		) {
			assert.Equal(t, 401, w.Code)
			var apiErr urest.Error
			_ = json.Unmarshal(w.Body.Bytes(), &apiErr)
			assert.EqualError(t,
				apiErr,
				"Authorization not present in header",
			)
		},
	}, {
		Name: "error, token malformed (base middleware)",
		Request: func() *http.Request {
			req, _ := http.NewRequest("GET",
				"http://localhost/api/management/v1/test",
				nil,
			)
			req.Header.Set("Authorization", "Bearer bruh?==")
			return req
		}(),
		Options: NewMiddlewareOptions().
			SetUpdateLogger(false),

		Validator: func(t *testing.T,
			w *httptest.ResponseRecorder, req *http.Request,
		) {
			assert.Equal(t, 401, w.Code)
			var apiErr urest.Error
			_ = json.Unmarshal(w.Body.Bytes(), &apiErr)
			assert.EqualError(t,
				apiErr,
				"identity: incorrect token format",
			)
		},
	}}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			reqChan := make(chan *http.Request, 1)
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Next()
				c.Writer.Flush()
				reqChan <- c.Request
			})
			router.Use(Middleware(tc.Options))
			router.GET("/api/management/v1/test", func(c *gin.Context) {
				c.Status(200)
			})
			router.NoRoute(func(c *gin.Context) {
				c.Status(200)
			})

			w := httptest.NewRecorder()
			router.ServeHTTP(w, tc.Request)

			var req *http.Request
			select {
			case req = <-reqChan:
				tc.Validator(t, w, req)
			case <-time.After(time.Second):
				panic("[PROG ERR] Bad test case")
			}
		})
	}

}
