// Copyright 2022 Northern.tech AS
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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mendersoftware/mender-server/pkg/rest.utils"

	mapp "github.com/mendersoftware/mender-server/services/reporting/app/reporting/mocks"
)

var contextMatcher = mock.MatchedBy(func(_ context.Context) bool { return true })

func TestStatus(t *testing.T) {
	t.Parallel()
	router := NewRouter(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, URIInternal+URIAlive, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestHealth(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name string

		Error      error
		StatusCode int
	}{{
		Name: "ok",

		StatusCode: http.StatusNoContent,
	}, {
		Name: "error, from application layer",

		Error:      errors.New("mongo: Connection refused"),
		StatusCode: http.StatusInternalServerError,
	}}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			app := new(mapp.App)
			app.On("HealthCheck",
				mock.MatchedBy(func(_ context.Context) bool {
					return true
				}),
			).Return(tc.Error)
			defer app.AssertExpectations(t)
			router := NewRouter(app)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", URIInternal+URIHealth, nil)
			req.Header.Set("X-Men-Requestid", "test")

			router.ServeHTTP(w, req)
			assert.Equal(t, tc.StatusCode, w.Code)
			if tc.Error == nil {
				assert.Nil(t, w.Body.Bytes())
			} else {
				err := rest.Error{
					Err:       tc.Error.Error(),
					RequestID: "test",
				}
				b, _ := json.Marshal(err)
				assert.Equal(t,
					string(b),
					w.Body.String(),
				)
			}
		})
	}
}
