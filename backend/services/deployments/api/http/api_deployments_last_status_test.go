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
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"

	mt "github.com/mendersoftware/mender-server/pkg/testing"
	rtest "github.com/mendersoftware/mender-server/pkg/testing/rest"
	mapp "github.com/mendersoftware/mender-server/services/deployments/app/mocks"
	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/services/deployments/utils/restutil"
	"github.com/mendersoftware/mender-server/services/deployments/utils/restutil/view"
	deployments_testing "github.com/mendersoftware/mender-server/services/deployments/utils/testing"
)

func TestGetDeviceDeploymentLastStatus(t *testing.T) {
	t.Parallel()

	deviceIds := []string{
		uuid.New().String(),
		uuid.New().String(),
	}
	tenantId := uuid.New().String()
	testCases := []struct {
		Name      string
		InputBody model.DeviceDeploymentLastStatusReq
		Statuses  []model.DeviceDeploymentLastStatus

		AppError     error
		ResponseCode int
		RestErr      map[string]interface{}
	}{
		{
			Name: "ok, device deployments list",
			InputBody: model.DeviceDeploymentLastStatusReq{
				DeviceIds: []string{deviceIds[0]},
			},
			Statuses: []model.DeviceDeploymentLastStatus{
				{
					DeviceId:               deviceIds[0],
					DeploymentId:           uuid.New().String(),
					DeviceDeploymentId:     uuid.New().String(),
					DeviceDeploymentStatus: model.DeviceDeploymentStatusNoArtifact,
					TenantId:               tenantId,
				},
			},
			ResponseCode: http.StatusOK,
		},
		{
			Name: "ok, empty device deployments list",
			InputBody: model.DeviceDeploymentLastStatusReq{
				DeviceIds: deviceIds,
			},
			Statuses:     []model.DeviceDeploymentLastStatus{},
			ResponseCode: http.StatusOK,
		},
		{
			Name: "error: app error",
			InputBody: model.DeviceDeploymentLastStatusReq{
				DeviceIds: deviceIds,
			},
			AppError:     errors.New("some error"),
			ResponseCode: http.StatusInternalServerError,
			RestErr:      deployments_testing.RestError("internal error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			app := &mapp.App{}
			app.On("GetDeviceDeploymentLastStatus", mock.MatchedBy(
				func(ctx interface{}) bool {
					if _, ok := ctx.(context.Context); ok {
						return true
					}
					return false
				}),
				mock.AnythingOfType("[]string"),
			).Return(model.DeviceDeploymentLastStatuses{DeviceDeploymentLastStatuses: tc.Statuses}, tc.AppError)

			restView := new(view.RESTView)
			d := NewDeploymentsApiHandlers(nil, restView, app)
			router := setUpTestRouter()
			router.POST(ApiUrlInternalDeviceDeploymentLastStatusDeployments,
				d.GetDeviceDeploymentLastStatus)
			url := strings.ReplaceAll(ApiUrlInternalDeviceDeploymentLastStatusDeployments, ":tenant", tenantId)
			url = "http://localhost" + url

			req := rtest.MakeTestRequest(&rtest.TestRequest{
				Method: "POST",
				Path:   url,
				Body:   tc.InputBody,
			})
			var body interface{}
			body = model.DeviceDeploymentLastStatuses{DeviceDeploymentLastStatuses: tc.Statuses}
			if tc.RestErr != nil {
				body = tc.RestErr
			}

			checker := mt.NewJSONResponse(tc.ResponseCode,
				map[string]string{
					"Content-Type": "application/json; charset=utf-8",
				},
				body)

			recorded := restutil.RunRequest(t, router, req)

			mt.CheckHTTPResponse(t, checker, recorded)

		})
	}
}
