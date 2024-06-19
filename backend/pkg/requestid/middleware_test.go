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

package requestid

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.ReleaseMode) // please just shut up
}

func TestGinMiddleware(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		Options *MiddlewareOptions

		Headers http.Header
	}{{
		Name: "Request with ID",

		Headers: func() http.Header {
			hdr := http.Header{}
			hdr.Set(RequestIdHeader, "test")
			return hdr
		}(),
	}, {
		Name: "Request generated ID",

		Options: NewMiddlewareOptions().
			SetGenerateRequestID(true),
	}}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			router := gin.New()
			router.Use(Middleware(tc.Options))
			router.GET("/test", func(c *gin.Context) {})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "http://mender.io/test", nil)
			for k, v := range tc.Headers {
				for _, vv := range v {
					req.Header.Add(k, vv)
				}
			}
			router.ServeHTTP(w, req)

			rsp := w.Result()

			if id := tc.Headers.Get(RequestIdHeader); id != "" {
				rspID := rsp.Header.Get(RequestIdHeader)
				assert.Equal(t, id, rspID)
			} else {
				if tc.Options.GenerateRequestID != nil &&
					*tc.Options.GenerateRequestID {
					_, err := uuid.Parse(rsp.Header.Get(RequestIdHeader))
					assert.NoError(t, err, "Generated requestID is not a UUID")
				} else {
					assert.Empty(t, rsp.Header.Get(RequestIdHeader))
				}
			}
		})
	}
}

func TestRequestIdMiddlewareWithReqID(t *testing.T) {
	api := rest.NewApi()

	api.Use(&RequestIdMiddleware{})

	reqid := "4420a5b9-dbf2-4e5d-8b4f-3cf2013d04af"
	api.SetApp(rest.AppSimple(func(w rest.ResponseWriter, r *rest.Request) {
		assert.Equal(t, reqid, FromContext(r.Context()))
		w.WriteJson(map[string]string{"foo": "bar"})
	}))

	handler := api.MakeHandler()

	req := test.MakeSimpleRequest("GET", "http://localhost/", nil)
	req.Header.Set(RequestIdHeader, reqid)

	recorded := test.RunRequest(t, handler, req)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()
	recorded.HeaderIs(RequestIdHeader, reqid)

}

func TestRequestIdMiddlewareNoReqID(t *testing.T) {
	api := rest.NewApi()

	api.Use(&RequestIdMiddleware{})

	api.SetApp(rest.AppSimple(func(w rest.ResponseWriter, r *rest.Request) {
		reqid := FromContext(r.Context())
		_, err := uuid.Parse(reqid)
		assert.NoError(t, err)
		w.WriteJson(map[string]string{"foo": "bar"})
	}))

	handler := api.MakeHandler()

	req := test.MakeSimpleRequest("GET", "http://localhost/", nil)
	recorded := test.RunRequest(t, handler, req)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()
	outReqIdStr := recorded.Recorder.HeaderMap.Get(RequestIdHeader)
	_, err := uuid.Parse(outReqIdStr)
	assert.NoError(t, err)
}
