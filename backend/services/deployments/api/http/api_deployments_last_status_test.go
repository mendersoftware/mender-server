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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
	mapp "github.com/mendersoftware/mender-server/services/deployments/app/mocks"
	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/services/deployments/utils/restutil/view"
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
			api := setUpRestTest(
				ApiUrlInternalDeviceDeploymentLastStatusDeployments,
				rest.Post,
				d.GetDeviceDeploymentLastStatus,
			)
			url := strings.ReplaceAll(ApiUrlInternalDeviceDeploymentLastStatusDeployments, "#tenant", tenantId)
			url = "http://localhost" + url
			req := test.MakeSimpleRequest("POST", url, tc.InputBody)

			recorded := test.RunRequest(t, api.MakeHandler(), req)
			recorded.CodeIs(tc.ResponseCode)
			assert.Equal(t, tc.ResponseCode, recorded.Recorder.Code)
			if tc.ResponseCode == http.StatusOK {
				recorded.ContentTypeIsJson()
				var res model.DeviceDeploymentLastStatuses
				recorded.DecodeJsonPayload(&res)
				t.Logf("got: %+v", res)
				assert.Equal(t, len(tc.Statuses), len(res.DeviceDeploymentLastStatuses))
			}
		})
	}
}
