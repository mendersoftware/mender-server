// Copyright 2024 Northern.tech AS
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

// Code generated by mockery v2.45.1. DO NOT EDIT.

package mocks

import (
	context "context"
	io "io"

	mock "github.com/stretchr/testify/mock"

	model "github.com/mendersoftware/mender-server/services/deployments/model"

	store "github.com/mendersoftware/mender-server/services/deployments/store"

	time "time"
)

// App is an autogenerated mock type for the App type
type App struct {
	mock.Mock
}

// AbortDeployment provides a mock function with given fields: ctx, deploymentID
func (_m *App) AbortDeployment(ctx context.Context, deploymentID string) error {
	ret := _m.Called(ctx, deploymentID)

	if len(ret) == 0 {
		panic("no return value specified for AbortDeployment")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, deploymentID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AbortDeviceDeployments provides a mock function with given fields: ctx, deviceID
func (_m *App) AbortDeviceDeployments(ctx context.Context, deviceID string) error {
	ret := _m.Called(ctx, deviceID)

	if len(ret) == 0 {
		panic("no return value specified for AbortDeviceDeployments")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, deviceID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CompleteUpload provides a mock function with given fields: ctx, intentID, skipVerify, metadata
func (_m *App) CompleteUpload(ctx context.Context, intentID string, skipVerify bool, metadata *model.DirectUploadMetadata) error {
	ret := _m.Called(ctx, intentID, skipVerify, metadata)

	if len(ret) == 0 {
		panic("no return value specified for CompleteUpload")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, bool, *model.DirectUploadMetadata) error); ok {
		r0 = rf(ctx, intentID, skipVerify, metadata)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateDeployment provides a mock function with given fields: ctx, constructor
func (_m *App) CreateDeployment(ctx context.Context, constructor *model.DeploymentConstructor) (string, error) {
	ret := _m.Called(ctx, constructor)

	if len(ret) == 0 {
		panic("no return value specified for CreateDeployment")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.DeploymentConstructor) (string, error)); ok {
		return rf(ctx, constructor)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *model.DeploymentConstructor) string); ok {
		r0 = rf(ctx, constructor)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *model.DeploymentConstructor) error); ok {
		r1 = rf(ctx, constructor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateDeviceConfigurationDeployment provides a mock function with given fields: ctx, constructor, deviceID, deploymentID
func (_m *App) CreateDeviceConfigurationDeployment(ctx context.Context, constructor *model.ConfigurationDeploymentConstructor, deviceID string, deploymentID string) (string, error) {
	ret := _m.Called(ctx, constructor, deviceID, deploymentID)

	if len(ret) == 0 {
		panic("no return value specified for CreateDeviceConfigurationDeployment")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.ConfigurationDeploymentConstructor, string, string) (string, error)); ok {
		return rf(ctx, constructor, deviceID, deploymentID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *model.ConfigurationDeploymentConstructor, string, string) string); ok {
		r0 = rf(ctx, constructor, deviceID, deploymentID)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *model.ConfigurationDeploymentConstructor, string, string) error); ok {
		r1 = rf(ctx, constructor, deviceID, deploymentID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateImage provides a mock function with given fields: ctx, multipartUploadMsg
func (_m *App) CreateImage(ctx context.Context, multipartUploadMsg *model.MultipartUploadMsg) (string, error) {
	ret := _m.Called(ctx, multipartUploadMsg)

	if len(ret) == 0 {
		panic("no return value specified for CreateImage")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.MultipartUploadMsg) (string, error)); ok {
		return rf(ctx, multipartUploadMsg)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *model.MultipartUploadMsg) string); ok {
		r0 = rf(ctx, multipartUploadMsg)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *model.MultipartUploadMsg) error); ok {
		r1 = rf(ctx, multipartUploadMsg)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DecommissionDevice provides a mock function with given fields: ctx, deviceID
func (_m *App) DecommissionDevice(ctx context.Context, deviceID string) error {
	ret := _m.Called(ctx, deviceID)

	if len(ret) == 0 {
		panic("no return value specified for DecommissionDevice")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, deviceID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteDeviceDeploymentsHistory provides a mock function with given fields: ctx, deviceId
func (_m *App) DeleteDeviceDeploymentsHistory(ctx context.Context, deviceId string) error {
	ret := _m.Called(ctx, deviceId)

	if len(ret) == 0 {
		panic("no return value specified for DeleteDeviceDeploymentsHistory")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, deviceId)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteImage provides a mock function with given fields: ctx, imageID
func (_m *App) DeleteImage(ctx context.Context, imageID string) error {
	ret := _m.Called(ctx, imageID)

	if len(ret) == 0 {
		panic("no return value specified for DeleteImage")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, imageID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteReleases provides a mock function with given fields: ctx, releaseNames
func (_m *App) DeleteReleases(ctx context.Context, releaseNames []string) ([]string, error) {
	ret := _m.Called(ctx, releaseNames)

	if len(ret) == 0 {
		panic("no return value specified for DeleteReleases")
	}

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []string) ([]string, error)); ok {
		return rf(ctx, releaseNames)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []string) []string); ok {
		r0 = rf(ctx, releaseNames)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []string) error); ok {
		r1 = rf(ctx, releaseNames)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DownloadLink provides a mock function with given fields: ctx, imageID, expire
func (_m *App) DownloadLink(ctx context.Context, imageID string, expire time.Duration) (*model.Link, error) {
	ret := _m.Called(ctx, imageID, expire)

	if len(ret) == 0 {
		panic("no return value specified for DownloadLink")
	}

	var r0 *model.Link
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, time.Duration) (*model.Link, error)); ok {
		return rf(ctx, imageID, expire)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, time.Duration) *model.Link); ok {
		r0 = rf(ctx, imageID, expire)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Link)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, time.Duration) error); ok {
		r1 = rf(ctx, imageID, expire)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EditImage provides a mock function with given fields: ctx, id, constructorData
func (_m *App) EditImage(ctx context.Context, id string, constructorData *model.ImageMeta) (bool, error) {
	ret := _m.Called(ctx, id, constructorData)

	if len(ret) == 0 {
		panic("no return value specified for EditImage")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *model.ImageMeta) (bool, error)); ok {
		return rf(ctx, id, constructorData)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, *model.ImageMeta) bool); ok {
		r0 = rf(ctx, id, constructorData)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, *model.ImageMeta) error); ok {
		r1 = rf(ctx, id, constructorData)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GenerateConfigurationImage provides a mock function with given fields: ctx, deviceType, deploymentID
func (_m *App) GenerateConfigurationImage(ctx context.Context, deviceType string, deploymentID string) (io.Reader, error) {
	ret := _m.Called(ctx, deviceType, deploymentID)

	if len(ret) == 0 {
		panic("no return value specified for GenerateConfigurationImage")
	}

	var r0 io.Reader
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (io.Reader, error)); ok {
		return rf(ctx, deviceType, deploymentID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) io.Reader); ok {
		r0 = rf(ctx, deviceType, deploymentID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.Reader)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, deviceType, deploymentID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GenerateImage provides a mock function with given fields: ctx, multipartUploadMsg
func (_m *App) GenerateImage(ctx context.Context, multipartUploadMsg *model.MultipartGenerateImageMsg) (string, error) {
	ret := _m.Called(ctx, multipartUploadMsg)

	if len(ret) == 0 {
		panic("no return value specified for GenerateImage")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.MultipartGenerateImageMsg) (string, error)); ok {
		return rf(ctx, multipartUploadMsg)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *model.MultipartGenerateImageMsg) string); ok {
		r0 = rf(ctx, multipartUploadMsg)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *model.MultipartGenerateImageMsg) error); ok {
		r1 = rf(ctx, multipartUploadMsg)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDeployment provides a mock function with given fields: ctx, deploymentID
func (_m *App) GetDeployment(ctx context.Context, deploymentID string) (*model.Deployment, error) {
	ret := _m.Called(ctx, deploymentID)

	if len(ret) == 0 {
		panic("no return value specified for GetDeployment")
	}

	var r0 *model.Deployment
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.Deployment, error)); ok {
		return rf(ctx, deploymentID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Deployment); ok {
		r0 = rf(ctx, deploymentID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Deployment)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, deploymentID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDeploymentForDeviceWithCurrent provides a mock function with given fields: ctx, deviceID, request
func (_m *App) GetDeploymentForDeviceWithCurrent(ctx context.Context, deviceID string, request *model.DeploymentNextRequest) (*model.DeploymentInstructions, error) {
	ret := _m.Called(ctx, deviceID, request)

	if len(ret) == 0 {
		panic("no return value specified for GetDeploymentForDeviceWithCurrent")
	}

	var r0 *model.DeploymentInstructions
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *model.DeploymentNextRequest) (*model.DeploymentInstructions, error)); ok {
		return rf(ctx, deviceID, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, *model.DeploymentNextRequest) *model.DeploymentInstructions); ok {
		r0 = rf(ctx, deviceID, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.DeploymentInstructions)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, *model.DeploymentNextRequest) error); ok {
		r1 = rf(ctx, deviceID, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDeploymentStats provides a mock function with given fields: ctx, deploymentID
func (_m *App) GetDeploymentStats(ctx context.Context, deploymentID string) (model.Stats, error) {
	ret := _m.Called(ctx, deploymentID)

	if len(ret) == 0 {
		panic("no return value specified for GetDeploymentStats")
	}

	var r0 model.Stats
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (model.Stats, error)); ok {
		return rf(ctx, deploymentID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) model.Stats); ok {
		r0 = rf(ctx, deploymentID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(model.Stats)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, deploymentID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDeploymentsStats provides a mock function with given fields: ctx, deploymentIDs
func (_m *App) GetDeploymentsStats(ctx context.Context, deploymentIDs ...string) ([]*model.DeploymentStats, error) {
	_va := make([]interface{}, len(deploymentIDs))
	for _i := range deploymentIDs {
		_va[_i] = deploymentIDs[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for GetDeploymentsStats")
	}

	var r0 []*model.DeploymentStats
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, ...string) ([]*model.DeploymentStats, error)); ok {
		return rf(ctx, deploymentIDs...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, ...string) []*model.DeploymentStats); ok {
		r0 = rf(ctx, deploymentIDs...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.DeploymentStats)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, ...string) error); ok {
		r1 = rf(ctx, deploymentIDs...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDeviceDeploymentLastStatus provides a mock function with given fields: ctx, devicesIds
func (_m *App) GetDeviceDeploymentLastStatus(ctx context.Context, devicesIds []string) (model.DeviceDeploymentLastStatuses, error) {
	ret := _m.Called(ctx, devicesIds)

	if len(ret) == 0 {
		panic("no return value specified for GetDeviceDeploymentLastStatus")
	}

	var r0 model.DeviceDeploymentLastStatuses
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []string) (model.DeviceDeploymentLastStatuses, error)); ok {
		return rf(ctx, devicesIds)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []string) model.DeviceDeploymentLastStatuses); ok {
		r0 = rf(ctx, devicesIds)
	} else {
		r0 = ret.Get(0).(model.DeviceDeploymentLastStatuses)
	}

	if rf, ok := ret.Get(1).(func(context.Context, []string) error); ok {
		r1 = rf(ctx, devicesIds)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDeviceDeploymentListForDevice provides a mock function with given fields: ctx, query
func (_m *App) GetDeviceDeploymentListForDevice(ctx context.Context, query store.ListQueryDeviceDeployments) ([]model.DeviceDeploymentListItem, int, error) {
	ret := _m.Called(ctx, query)

	if len(ret) == 0 {
		panic("no return value specified for GetDeviceDeploymentListForDevice")
	}

	var r0 []model.DeviceDeploymentListItem
	var r1 int
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, store.ListQueryDeviceDeployments) ([]model.DeviceDeploymentListItem, int, error)); ok {
		return rf(ctx, query)
	}
	if rf, ok := ret.Get(0).(func(context.Context, store.ListQueryDeviceDeployments) []model.DeviceDeploymentListItem); ok {
		r0 = rf(ctx, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.DeviceDeploymentListItem)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, store.ListQueryDeviceDeployments) int); ok {
		r1 = rf(ctx, query)
	} else {
		r1 = ret.Get(1).(int)
	}

	if rf, ok := ret.Get(2).(func(context.Context, store.ListQueryDeviceDeployments) error); ok {
		r2 = rf(ctx, query)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GetDeviceDeploymentLog provides a mock function with given fields: ctx, deviceID, deploymentID
func (_m *App) GetDeviceDeploymentLog(ctx context.Context, deviceID string, deploymentID string) (*model.DeploymentLog, error) {
	ret := _m.Called(ctx, deviceID, deploymentID)

	if len(ret) == 0 {
		panic("no return value specified for GetDeviceDeploymentLog")
	}

	var r0 *model.DeploymentLog
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*model.DeploymentLog, error)); ok {
		return rf(ctx, deviceID, deploymentID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.DeploymentLog); ok {
		r0 = rf(ctx, deviceID, deploymentID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.DeploymentLog)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, deviceID, deploymentID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDeviceStatusesForDeployment provides a mock function with given fields: ctx, deploymentID
func (_m *App) GetDeviceStatusesForDeployment(ctx context.Context, deploymentID string) ([]model.DeviceDeployment, error) {
	ret := _m.Called(ctx, deploymentID)

	if len(ret) == 0 {
		panic("no return value specified for GetDeviceStatusesForDeployment")
	}

	var r0 []model.DeviceDeployment
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]model.DeviceDeployment, error)); ok {
		return rf(ctx, deploymentID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []model.DeviceDeployment); ok {
		r0 = rf(ctx, deploymentID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.DeviceDeployment)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, deploymentID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDevicesListForDeployment provides a mock function with given fields: ctx, query
func (_m *App) GetDevicesListForDeployment(ctx context.Context, query store.ListQuery) ([]model.DeviceDeployment, int, error) {
	ret := _m.Called(ctx, query)

	if len(ret) == 0 {
		panic("no return value specified for GetDevicesListForDeployment")
	}

	var r0 []model.DeviceDeployment
	var r1 int
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, store.ListQuery) ([]model.DeviceDeployment, int, error)); ok {
		return rf(ctx, query)
	}
	if rf, ok := ret.Get(0).(func(context.Context, store.ListQuery) []model.DeviceDeployment); ok {
		r0 = rf(ctx, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.DeviceDeployment)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, store.ListQuery) int); ok {
		r1 = rf(ctx, query)
	} else {
		r1 = ret.Get(1).(int)
	}

	if rf, ok := ret.Get(2).(func(context.Context, store.ListQuery) error); ok {
		r2 = rf(ctx, query)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GetImage provides a mock function with given fields: ctx, id
func (_m *App) GetImage(ctx context.Context, id string) (*model.Image, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for GetImage")
	}

	var r0 *model.Image
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.Image, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Image); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Image)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetLimit provides a mock function with given fields: ctx, name
func (_m *App) GetLimit(ctx context.Context, name string) (*model.Limit, error) {
	ret := _m.Called(ctx, name)

	if len(ret) == 0 {
		panic("no return value specified for GetLimit")
	}

	var r0 *model.Limit
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.Limit, error)); ok {
		return rf(ctx, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Limit); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Limit)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetRelease provides a mock function with given fields: ctx, releaseName
func (_m *App) GetRelease(ctx context.Context, releaseName string) (*model.Release, error) {
	ret := _m.Called(ctx, releaseName)

	if len(ret) == 0 {
		panic("no return value specified for GetRelease")
	}

	var r0 *model.Release
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.Release, error)); ok {
		return rf(ctx, releaseName)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Release); ok {
		r0 = rf(ctx, releaseName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Release)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, releaseName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetReleasesUpdateTypes provides a mock function with given fields: ctx
func (_m *App) GetReleasesUpdateTypes(ctx context.Context) ([]string, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetReleasesUpdateTypes")
	}

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]string, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []string); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetStorageSettings provides a mock function with given fields: ctx
func (_m *App) GetStorageSettings(ctx context.Context) (*model.StorageSettings, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetStorageSettings")
	}

	var r0 *model.StorageSettings
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*model.StorageSettings, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *model.StorageSettings); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.StorageSettings)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// HasDeploymentForDevice provides a mock function with given fields: ctx, deploymentID, deviceID
func (_m *App) HasDeploymentForDevice(ctx context.Context, deploymentID string, deviceID string) (bool, error) {
	ret := _m.Called(ctx, deploymentID, deviceID)

	if len(ret) == 0 {
		panic("no return value specified for HasDeploymentForDevice")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (bool, error)); ok {
		return rf(ctx, deploymentID, deviceID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) bool); ok {
		r0 = rf(ctx, deploymentID, deviceID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, deploymentID, deviceID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// HealthCheck provides a mock function with given fields: ctx
func (_m *App) HealthCheck(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for HealthCheck")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// IsDeploymentFinished provides a mock function with given fields: ctx, deploymentID
func (_m *App) IsDeploymentFinished(ctx context.Context, deploymentID string) (bool, error) {
	ret := _m.Called(ctx, deploymentID)

	if len(ret) == 0 {
		panic("no return value specified for IsDeploymentFinished")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (bool, error)); ok {
		return rf(ctx, deploymentID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) bool); ok {
		r0 = rf(ctx, deploymentID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, deploymentID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListImages provides a mock function with given fields: ctx, filters
func (_m *App) ListImages(ctx context.Context, filters *model.ReleaseOrImageFilter) ([]*model.Image, int, error) {
	ret := _m.Called(ctx, filters)

	if len(ret) == 0 {
		panic("no return value specified for ListImages")
	}

	var r0 []*model.Image
	var r1 int
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.ReleaseOrImageFilter) ([]*model.Image, int, error)); ok {
		return rf(ctx, filters)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *model.ReleaseOrImageFilter) []*model.Image); ok {
		r0 = rf(ctx, filters)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Image)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *model.ReleaseOrImageFilter) int); ok {
		r1 = rf(ctx, filters)
	} else {
		r1 = ret.Get(1).(int)
	}

	if rf, ok := ret.Get(2).(func(context.Context, *model.ReleaseOrImageFilter) error); ok {
		r2 = rf(ctx, filters)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// ListImagesV2 provides a mock function with given fields: ctx, filters
func (_m *App) ListImagesV2(ctx context.Context, filters *model.ImageFilter) ([]*model.Image, error) {
	ret := _m.Called(ctx, filters)

	if len(ret) == 0 {
		panic("no return value specified for ListImagesV2")
	}

	var r0 []*model.Image
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.ImageFilter) ([]*model.Image, error)); ok {
		return rf(ctx, filters)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *model.ImageFilter) []*model.Image); ok {
		r0 = rf(ctx, filters)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Image)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *model.ImageFilter) error); ok {
		r1 = rf(ctx, filters)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListReleaseTags provides a mock function with given fields: ctx
func (_m *App) ListReleaseTags(ctx context.Context) (model.Tags, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for ListReleaseTags")
	}

	var r0 model.Tags
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (model.Tags, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) model.Tags); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(model.Tags)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LookupDeployment provides a mock function with given fields: ctx, query
func (_m *App) LookupDeployment(ctx context.Context, query model.Query) ([]*model.Deployment, int64, error) {
	ret := _m.Called(ctx, query)

	if len(ret) == 0 {
		panic("no return value specified for LookupDeployment")
	}

	var r0 []*model.Deployment
	var r1 int64
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, model.Query) ([]*model.Deployment, int64, error)); ok {
		return rf(ctx, query)
	}
	if rf, ok := ret.Get(0).(func(context.Context, model.Query) []*model.Deployment); ok {
		r0 = rf(ctx, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Deployment)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, model.Query) int64); ok {
		r1 = rf(ctx, query)
	} else {
		r1 = ret.Get(1).(int64)
	}

	if rf, ok := ret.Get(2).(func(context.Context, model.Query) error); ok {
		r2 = rf(ctx, query)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// ProvisionTenant provides a mock function with given fields: ctx, tenant_id
func (_m *App) ProvisionTenant(ctx context.Context, tenant_id string) error {
	ret := _m.Called(ctx, tenant_id)

	if len(ret) == 0 {
		panic("no return value specified for ProvisionTenant")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, tenant_id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ReplaceReleaseTags provides a mock function with given fields: ctx, releaseName, tags
func (_m *App) ReplaceReleaseTags(ctx context.Context, releaseName string, tags model.Tags) error {
	ret := _m.Called(ctx, releaseName, tags)

	if len(ret) == 0 {
		panic("no return value specified for ReplaceReleaseTags")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.Tags) error); ok {
		r0 = rf(ctx, releaseName, tags)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveDeviceDeploymentLog provides a mock function with given fields: ctx, deviceID, deploymentID, logs
func (_m *App) SaveDeviceDeploymentLog(ctx context.Context, deviceID string, deploymentID string, logs []model.LogMessage) error {
	ret := _m.Called(ctx, deviceID, deploymentID, logs)

	if len(ret) == 0 {
		panic("no return value specified for SaveDeviceDeploymentLog")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, []model.LogMessage) error); ok {
		r0 = rf(ctx, deviceID, deploymentID, logs)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetStorageSettings provides a mock function with given fields: ctx, storageSettings
func (_m *App) SetStorageSettings(ctx context.Context, storageSettings *model.StorageSettings) error {
	ret := _m.Called(ctx, storageSettings)

	if len(ret) == 0 {
		panic("no return value specified for SetStorageSettings")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.StorageSettings) error); ok {
		r0 = rf(ctx, storageSettings)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateDeploymentsWithArtifactName provides a mock function with given fields: ctx, artifactName
func (_m *App) UpdateDeploymentsWithArtifactName(ctx context.Context, artifactName string) error {
	ret := _m.Called(ctx, artifactName)

	if len(ret) == 0 {
		panic("no return value specified for UpdateDeploymentsWithArtifactName")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, artifactName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateDeviceDeploymentStatus provides a mock function with given fields: ctx, deploymentID, deviceID, state
func (_m *App) UpdateDeviceDeploymentStatus(ctx context.Context, deploymentID string, deviceID string, state model.DeviceDeploymentState) error {
	ret := _m.Called(ctx, deploymentID, deviceID, state)

	if len(ret) == 0 {
		panic("no return value specified for UpdateDeviceDeploymentStatus")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, model.DeviceDeploymentState) error); ok {
		r0 = rf(ctx, deploymentID, deviceID, state)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateRelease provides a mock function with given fields: ctx, releaseName, release
func (_m *App) UpdateRelease(ctx context.Context, releaseName string, release model.ReleasePatch) error {
	ret := _m.Called(ctx, releaseName, release)

	if len(ret) == 0 {
		panic("no return value specified for UpdateRelease")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.ReleasePatch) error); ok {
		r0 = rf(ctx, releaseName, release)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UploadLink provides a mock function with given fields: ctx, expire, skipVerify
func (_m *App) UploadLink(ctx context.Context, expire time.Duration, skipVerify bool) (*model.UploadLink, error) {
	ret := _m.Called(ctx, expire, skipVerify)

	if len(ret) == 0 {
		panic("no return value specified for UploadLink")
	}

	var r0 *model.UploadLink
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, time.Duration, bool) (*model.UploadLink, error)); ok {
		return rf(ctx, expire, skipVerify)
	}
	if rf, ok := ret.Get(0).(func(context.Context, time.Duration, bool) *model.UploadLink); ok {
		r0 = rf(ctx, expire, skipVerify)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.UploadLink)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, time.Duration, bool) error); ok {
		r1 = rf(ctx, expire, skipVerify)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewApp creates a new instance of App. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewApp(t interface {
	mock.TestingT
	Cleanup(func())
}) *App {
	mock := &App{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
