// Copyright 2019 Northern.tech AS
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

package view

import (
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	mt "github.com/mendersoftware/mender-server/pkg/testing"
	rtest "github.com/mendersoftware/mender-server/pkg/testing/rest"

	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/services/deployments/utils/restutil"
)

func TestRenderPost(t *testing.T) {
	router := gin.New()
	router.POST("/test", func(c *gin.Context) {
		new(RESTView).RenderSuccessPost(c, "test_id")
	})

	req := rtest.MakeTestRequest(&rtest.TestRequest{
		Method: "POST",
		Path:   "http://localhost/test",
		Body:   "blah",
	})
	checker := mt.NewJSONResponse(
		http.StatusCreated,
		map[string]string{
			HttpHeaderLocation: "/test/test_id",
		},
		nil,
	)
	recorded := restutil.RunRequest(t, router, req)

	mt.CheckHTTPResponse(t, checker, recorded)
}

func TestRenderSuccessGet(t *testing.T) {

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		new(RESTView).RenderSuccessGet(c, "test")
	})

	req := rtest.MakeTestRequest(&rtest.TestRequest{
		Method: "GET",
		Path:   "http://localhost/test",
	})
	checker := mt.NewJSONResponse(
		http.StatusOK,
		map[string]string{
			"Content-type": "application/json; charset=utf-8",
		},
		"test",
	)
	recorded := restutil.RunRequest(t, router, req)

	mt.CheckHTTPResponse(t, checker, recorded)
}

func TestRenderSuccessDelete(t *testing.T) {

	router := gin.New()
	router.DELETE("/test", func(c *gin.Context) {
		new(RESTView).RenderSuccessDelete(c)
	})

	req := rtest.MakeTestRequest(&rtest.TestRequest{
		Method: "DELETE",
		Path:   "http://localhost/test",
	})
	checker := mt.NewJSONResponse(
		http.StatusNoContent,
		nil,
		nil,
	)
	recorded := restutil.RunRequest(t, router, req)

	mt.CheckHTTPResponse(t, checker, recorded)
}

func TestRenderSuccessPut(t *testing.T) {

	router := gin.New()
	router.PUT("/test", func(c *gin.Context) {
		new(RESTView).RenderSuccessPut(c)
	})

	req := rtest.MakeTestRequest(&rtest.TestRequest{
		Method: "PUT",
		Path:   "http://localhost/test",
	})
	checker := mt.NewJSONResponse(
		http.StatusNoContent,
		nil,
		nil,
	)
	recorded := restutil.RunRequest(t, router, req)

	mt.CheckHTTPResponse(t, checker, recorded)
}

func TestRenderErrorNotFound(t *testing.T) {

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		new(RESTView).RenderErrorNotFound(c)
	})

	req := rtest.MakeTestRequest(&rtest.TestRequest{
		Method: "GET",
		Path:   "http://localhost/test",
	})
	checker := mt.NewJSONResponse(
		http.StatusNotFound,
		nil,
		map[string]string{
			"error": "Resource not found",
		},
	)
	recorded := restutil.RunRequest(t, router, req)

	mt.CheckHTTPResponse(t, checker, recorded)
}

func TestRenderNoUpdateForDevice(t *testing.T) {

	t.Parallel()

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		new(RESTView).RenderNoUpdateForDevice(c)
	})

	req := rtest.MakeTestRequest(&rtest.TestRequest{
		Method: "GET",
		Path:   "http://localhost/test",
	})
	checker := mt.NewJSONResponse(
		http.StatusNoContent,
		nil,
		nil,
	)
	recorded := restutil.RunRequest(t, router, req)

	mt.CheckHTTPResponse(t, checker, recorded)
}

func parseTime(t *testing.T, value string) *time.Time {
	tm, err := time.Parse(time.RFC3339, value)
	if assert.NoError(t, err) == false {
		t.Fatalf("failed to parse time %s", value)
	}

	return &tm
}

func TestRenderDeploymentLog(t *testing.T) {

	t.Parallel()

	tref := parseTime(t, "2006-01-02T15:04:05-07:00")

	messages := []model.LogMessage{
		{
			Timestamp: tref,
			Message:   "foo",
			Level:     "notice",
		},
		{
			Timestamp: tref,
			Message:   "zed zed zed",
			Level:     "debug",
		},
		{
			Timestamp: tref,
			Message:   "bar bar bar",
			Level:     "info",
		},
	}

	tcs := []struct {
		Log  model.DeploymentLog
		Body string
	}{
		{
			// all correct
			Log: model.DeploymentLog{
				DeploymentID: "f826484e-1157-4109-af21-304e6d711560",
				DeviceID:     "device-id-1",
				Messages:     messages,
			},
			Body: `2006-01-02 22:04:05 +0000 UTC notice: foo
2006-01-02 22:04:05 +0000 UTC debug: zed zed zed
2006-01-02 22:04:05 +0000 UTC info: bar bar bar
`,
		},
	}

	for _, tc := range tcs {
		router := gin.New()
		router.GET("/test", func(c *gin.Context) {
			new(RESTView).RenderDeploymentLog(c, tc.Log)
		})

		req := rtest.MakeTestRequest(&rtest.TestRequest{
			Method: "GET",
			Path:   "http://localhost/test",
		})

		recorded := restutil.RunRequest(t, router, req)

		assert.Equal(t, http.StatusOK, recorded.Recorder.Code)
		assert.Equal(t, tc.Body, recorded.Recorder.Body.String())
	}
}
