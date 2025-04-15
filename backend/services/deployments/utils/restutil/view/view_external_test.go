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

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

func TestRenderPost(t *testing.T) {

	router, err := rest.MakeRouter(rest.Post("/test", func(w rest.ResponseWriter, r *rest.Request) {
		new(RESTView).RenderSuccessPost(w, r, "test_id")
	}))

	if err != nil {
		assert.NoError(t, err)
	}

	api := rest.NewApi()
	api.SetApp(router)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/test", "blah"))

	recorded.CodeIs(http.StatusCreated)
	recorded.ContentTypeIsJson()
	recorded.HeaderIs(HttpHeaderLocation, "/test/test_id")
}

func TestRenderSuccessGet(t *testing.T) {

	router, err := rest.MakeRouter(rest.Get("/test", func(w rest.ResponseWriter, r *rest.Request) {
		new(RESTView).RenderSuccessGet(w, "test")
	}))

	if err != nil {
		assert.NoError(t, err)
	}

	api := rest.NewApi()
	api.SetApp(router)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("GET", "http://localhost/test", nil))

	recorded.CodeIs(http.StatusOK)
	recorded.ContentTypeIsJson()
	recorded.BodyIs(`"test"`)
}

func TestRenderSuccessDelete(t *testing.T) {

	router, err := rest.MakeRouter(rest.Delete("/test", func(w rest.ResponseWriter, r *rest.Request) {
		new(RESTView).RenderSuccessDelete(w)
	}))

	if err != nil {
		assert.NoError(t, err)
	}

	api := rest.NewApi()
	api.SetApp(router)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("DELETE", "http://localhost/test", nil))

	recorded.CodeIs(http.StatusNoContent)
}

func TestRenderSuccessPut(t *testing.T) {

	router, err := rest.MakeRouter(rest.Put("/test", func(w rest.ResponseWriter, r *rest.Request) {
		new(RESTView).RenderSuccessPut(w)
	}))

	if err != nil {
		assert.NoError(t, err)
	}

	api := rest.NewApi()
	api.SetApp(router)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("PUT", "http://localhost/test", nil))

	recorded.CodeIs(http.StatusNoContent)
}

func TestRenderErrorNotFound(t *testing.T) {

	router, err := rest.MakeRouter(rest.Get("/test", func(w rest.ResponseWriter, r *rest.Request) {

		l := log.New(log.Ctx{})
		new(RESTView).RenderErrorNotFound(w, r, l)
	}))

	if err != nil {
		assert.NoError(t, err)
	}

	api := rest.NewApi()
	api.SetApp(router)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("GET", "http://localhost/test", nil))

	recorded.CodeIs(http.StatusNotFound)
	recorded.BodyIs(`{"error":"Resource not found","request_id":""}`)
}

func TestRenderNoUpdateForDevice(t *testing.T) {

	t.Parallel()

	router, err := rest.MakeRouter(rest.Get("/test", func(w rest.ResponseWriter, r *rest.Request) {
		view := &RESTView{}
		view.RenderNoUpdateForDevice(w)
	}))

	if err != nil {
		assert.NoError(t, err)
	}

	api := rest.NewApi()
	api.SetApp(router)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("GET", "http://localhost/test", nil))

	recorded.CodeIs(http.StatusNoContent)
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
		router, err := rest.MakeRouter(rest.Get("/test", func(w rest.ResponseWriter, r *rest.Request) {
			view := &RESTView{}
			view.RenderDeploymentLog(w, tc.Log)
		}))

		assert.NoError(t, err)

		api := rest.NewApi()
		api.SetApp(router)

		recorded := test.RunRequest(t, api.MakeHandler(),
			test.MakeSimpleRequest("GET", "http://localhost/test", nil))

		recorded.CodeIs(http.StatusOK)
		assert.Equal(t, tc.Body, recorded.Recorder.Body.String())
	}
}
