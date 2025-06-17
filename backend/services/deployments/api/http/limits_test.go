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

	"github.com/stretchr/testify/mock"

	mt "github.com/mendersoftware/mender-server/pkg/testing"
	rtest "github.com/mendersoftware/mender-server/pkg/testing/rest"
	app_mocks "github.com/mendersoftware/mender-server/services/deployments/app/mocks"
	"github.com/mendersoftware/mender-server/services/deployments/model"
	store_mocks "github.com/mendersoftware/mender-server/services/deployments/store/mocks"
	"github.com/mendersoftware/mender-server/services/deployments/utils/restutil"
	"github.com/mendersoftware/mender-server/services/deployments/utils/restutil/view"
	deployments_testing "github.com/mendersoftware/mender-server/services/deployments/utils/testing"
)

func contextMatcher() interface{} {
	return mock.MatchedBy(func(_ context.Context) bool {
		return true
	})
}
func TestGetLimits(t *testing.T) {

	testCases := []struct {
		name  string
		code  int
		body  map[string]interface{}
		err   error
		limit *model.Limit
	}{
		{
			name: "storage",
			code: http.StatusOK,
			body: map[string]interface{}{"limit": 200, "usage": 0},
			limit: &model.Limit{
				Name:  "storage",
				Value: 200,
			},
		},
		{
			name: "storage",
			code: http.StatusInternalServerError,
			body: deployments_testing.RestError("internal error"),
			err:  errors.New("failed"),
		},
		{
			name: "foobar",
			code: http.StatusBadRequest,
			body: deployments_testing.RestError("unsupported limit foobar"),
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			store := &store_mocks.DataStore{}
			restView := new(view.RESTView)
			app := &app_mocks.App{}

			d := NewDeploymentsApiHandlers(store, restView, app)
			router := setUpTestRouter()
			router.GET("/api/0.0.1/limits/:name", d.GetLimit)

			if tc.err != nil || tc.limit != nil {
				app.On("GetLimit", contextMatcher(), tc.name).
					Return(tc.limit, tc.err)
			}
			req := rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "GET",
				Path:   "http://localhost/api/0.0.1/limits/" + tc.name,
			})
			req.Header.Set("X-MEN-RequestID", "test")
			checker := mt.NewJSONResponse(tc.code, nil, tc.body)

			recorded := restutil.RunRequest(t, router, req)

			mt.CheckHTTPResponse(t, checker, recorded)

			app.AssertExpectations(t)
		})
	}
}
