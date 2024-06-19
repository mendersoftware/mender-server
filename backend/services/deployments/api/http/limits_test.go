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

package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	app_mocks "github.com/mendersoftware/mender-server/services/deployments/app/mocks"
	"github.com/mendersoftware/mender-server/services/deployments/model"
	store_mocks "github.com/mendersoftware/mender-server/services/deployments/store/mocks"
	"github.com/mendersoftware/mender-server/services/deployments/utils/restutil/view"
)

type routerTypeHandler func(pathExp string, handlerFunc rest.HandlerFunc) *rest.Route

func contextMatcher() interface{} {
	return mock.MatchedBy(func(_ context.Context) bool {
		return true
	})
}
func TestGetLimits(t *testing.T) {

	testCases := []struct {
		name  string
		code  int
		body  string
		err   error
		limit *model.Limit
	}{
		{
			name: "storage",
			code: http.StatusOK,
			body: `{"limit":200,"usage":0}`,
			limit: &model.Limit{
				Name:  "storage",
				Value: 200,
			},
		},
		{
			name: "storage",
			code: http.StatusInternalServerError,
			err:  errors.New("failed"),
		},
		{
			name: "foobar",
			code: http.StatusBadRequest,
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			store := &store_mocks.DataStore{}
			restView := new(view.RESTView)
			app := &app_mocks.App{}

			d := NewDeploymentsApiHandlers(store, restView, app)

			api := setUpRestTest("/api/0.0.1/limits/:name", rest.Get, d.GetLimit)

			if tc.err != nil || tc.limit != nil {
				app.On("GetLimit", contextMatcher(), tc.name).
					Return(tc.limit, tc.err)
			}

			recorded := test.RunRequest(t, api.MakeHandler(),
				test.MakeSimpleRequest("GET", "http://localhost/api/0.0.1/limits/"+tc.name,
					nil))
			recorded.CodeIs(tc.code)
			if tc.code == http.StatusOK {
				assert.JSONEq(t, tc.body, recorded.Recorder.Body.String())
			}

			app.AssertExpectations(t)
		})
	}
}
