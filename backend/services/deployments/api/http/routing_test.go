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

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	rtest "github.com/mendersoftware/mender-server/pkg/testing/rest"
	mapp "github.com/mendersoftware/mender-server/services/deployments/app/mocks"
)

func TestNewRouter(t *testing.T) {
	t.Parallel()
	type testCase struct {
		Name string

		App func(t *testing.T) *mapp.App

		StatusCode        int
		BodyAssertionFunc func(t *testing.T, body string) bool
	}
	testCases := map[string]struct {
		cfg        *Config
		statusCode int
	}{
		"default": {
			cfg: &Config{
				DisableNewReleasesFeature: false,
			},
			statusCode: http.StatusBadRequest,
		},
		"disable new releases features": {
			cfg: &Config{
				DisableNewReleasesFeature: true,
			},
			statusCode: http.StatusServiceUnavailable,
		},
	}

	for name, _ := range testCases {
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			app := new(mapp.App)
			defer app.AssertExpectations(t)

			apiHandler := NewRouter(
				ctx,
				app,
				nil,
				tc.cfg,
			)

			req := rtest.MakeTestRequest(&rtest.TestRequest{
				Method: http.MethodPost,
				Path:   "https://localhost:8443" + ApiUrlManagementArtifactsGenerate,
				Auth:   true,
			})

			w := httptest.NewRecorder()
			apiHandler.ServeHTTP(w, req)

			assert.Equal(t, tc.statusCode, w.Code, "Unexpected HTTP status code")
		})
	}
}
