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
package restutil

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"testing"

	"github.com/gin-gonic/gin"
	rtest "github.com/mendersoftware/mender-server/pkg/testing/rest"
	"github.com/stretchr/testify/assert"
)

func TestOptionsHandle(t *testing.T) {

	t.Parallel()
	router := gin.Default()
	router.OPTIONS("/r", NewOptionsHandler(http.MethodGet, http.MethodGet))

	req := rtest.MakeTestRequest(&rtest.TestRequest{
		Method: http.MethodOptions,
		Path:   "http://1.2.3.4/r",
	})
	recorded := RunRequest(t, router, req)

	assert.Equal(t, http.StatusOK, recorded.Recorder.Code)

	if len(recorded.Recorder.Header()[HttpHeaderAllow]) != 2 {
		t.FailNow()
	}

	for _, method := range recorded.Recorder.Header()[HttpHeaderAllow] {
		switch method {
		case http.MethodGet:
			continue
		case http.MethodOptions:
			continue
		default:
			t.FailNow()
		}
	}
}

func TestAutogenOptionsRoutes(t *testing.T) {

	t.Parallel()

	type expHandler map[string]gin.HandlerFunc

	dummy := func(c *gin.Context) {
		c.Status(http.StatusOK)
	}
	genOptions := func(c *gin.Context) {
		c.Status(http.StatusOK)
	}
	createHandler := func(methods ...string) gin.HandlerFunc {
		return genOptions
	}
	testList := []struct {
		in       *gin.Engine
		out      map[string]expHandler
		expCount int
	}{
		{gin.Default(), map[string]expHandler{}, 0},
		{gin.Default(), map[string]expHandler{}, 0},
		{
			in: func() *gin.Engine {
				router := gin.Default()
				router.GET("/path", dummy)
				return router
			}(),
			out: map[string]expHandler{
				"/path": {
					http.MethodGet:     dummy,
					http.MethodOptions: genOptions,
				},
			},
			expCount: 2,
		},
		{
			in: func() *gin.Engine {
				router := gin.Default()
				router.GET("/path", dummy)
				router.POST("/path", dummy)
				return router
			}(),
			out: map[string]expHandler{
				"/path": {
					http.MethodGet:     dummy,
					http.MethodPost:    dummy,
					http.MethodOptions: genOptions,
				},
			},
			expCount: 3,
		},
		{
			in: func() *gin.Engine {
				router := gin.Default()
				router.GET("/path", dummy)
				router.POST("/path/path", dummy)
				return router
			}(),
			out: map[string]expHandler{
				"/path": {
					http.MethodGet:     dummy,
					http.MethodOptions: genOptions,
				},
				"/path/path": {
					http.MethodPost:    dummy,
					http.MethodOptions: genOptions,
				},
			},
			expCount: 4,
		},
		{
			in: func() *gin.Engine {
				router := gin.Default()
				router.GET("/path", dummy)
				router.POST("/path/path", dummy)
				router.PUT("/path", dummy)
				return router
			}(),
			out: map[string]expHandler{
				"/path": {
					http.MethodGet:     dummy,
					http.MethodPut:     dummy,
					http.MethodOptions: genOptions,
				},
				"/path/path": {
					http.MethodPost:    dummy,
					http.MethodOptions: genOptions,
				},
			},
			expCount: 5,
		},
		{
			in: func() *gin.Engine {
				router := gin.Default()
				router.GET("/path", dummy)
				router.OPTIONS("/path", dummy)
				return router
			}(),
			out: map[string]expHandler{
				"/path": {
					http.MethodGet:     dummy,
					http.MethodOptions: dummy,
				},
			},
			expCount: 2,
		},
	}

	for _, test := range testList {
		AutogenOptionsRoutes(createHandler, test.in)

		routes := test.in.Routes()
		if len(routes) != test.expCount {
			t.FailNow()
		}

		for _, r := range routes {
			v, ok := test.out[r.Path]
			assert.Equal(t, ok, true, "failed with route "+r.Path+", route not present")

			h, ok := v[r.Method]
			assert.Equal(t, ok, true, fmt.Sprintf("failed with route %s, method %s not present", r.Path, r.Method))

			assert.Equal(t, r.Handler, funcName(h), "failed with route "+r.Path+", different handler")
		}
	}
}

// We can't compare functions, so let's take the hard way and extract
// func name from runtime
func funcName(f interface{}) string {
	p := reflect.ValueOf(f).Pointer()
	rfunc := runtime.FuncForPC(p)
	return rfunc.Name()
}
