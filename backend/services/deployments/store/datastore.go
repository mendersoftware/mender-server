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

package store

import (
	"context"
	"errors"
	"time"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

//go:generate ../utils/mockgen.sh
type DataStore interface {
	Ping(ctx context.Context) error
	//releases
	GetReleases(ctx context.Context, filt *model.ReleaseOrImageFilter) ([]model.Release, int, error)
	UpdateReleaseArtifacts(
		ctx context.Context,
		artifactToAdd *model.Image,
		artifactToRemove *model.Image,
		releaseName string,
	) error
	UpdateReleaseArtifactDescription(
		ctx context.Context,
		artifactToEdit *model.Image,
		releaseName string,
	) error

	//limits
	GetLimit(ctx context.Context, name string) (*model.Limit, error)

	//storage settings
	GetStorageSettings(ctx context.Context) (*model.StorageSettings, error)
	SetStorageSettings(ctx context.Context, storageSettings *model.StorageSettings) error

	//tenants
	ProvisionTenant(ctx context.Context, tenantId string) error

	// images
	Exists(ctx context.Context, id string) (bool, error)
	Update(ctx context.Context, image *model.Image) (bool, error)
	InsertImage(ctx context.Context, image *model.Image) error
	FindImageByID(ctx context.Context, id string) (*model.Image, error)
	IsArtifactUnique(ctx context.Context, artifactName string,
		deviceTypesCompatible []string) (bool, error)
	DeleteImage(ctx context.Context, id string) error
	ListImages(ctx context.Context, filt *model.ReleaseOrImageFilter) ([]*model.Image, int, error)
	DeleteImagesByNames(ctx context.Context, names []string) error

	//artifact getter
	ImagesByName(ctx context.Context,
		artifactName string) ([]*model.Image, error)
	ImageByIdsAndDeviceType(ctx context.Context,
		ids []string, deviceType string) (*model.Image, error)
	ImageByNameAndDeviceType(ctx context.Context,
		name, deviceType string) (*model.Image, error)

	// upload intents
	InsertUploadIntent(ctx context.Context, link *model.UploadLink) error
	UpdateUploadIntentStatus(ctx context.Context, id string, from, to model.LinkStatus) error
	FindUploadLinks(ctx context.Context, expired time.Time) (Iterator[model.UploadLink], error)

	//device deployment log
	SaveDeviceDeploymentLog(ctx context.Context, log model.DeploymentLog) error
	GetDeviceDeploymentLog(ctx context.Context,
		deviceID, deploymentID string) (*model.DeploymentLog, error)

	// device deployments
	InsertDeviceDeployment(ctx context.Context, deviceDeployment *model.DeviceDeployment,
		incrementDeviceCount bool) error
	InsertMany(ctx context.Context,
		deployment ...*model.DeviceDeployment) error
	FindOldestActiveDeviceDeployment(
		ctx context.Context,
		deviceID string,
	) (*model.DeviceDeployment, error)
	FindLatestInactiveDeviceDeployment(
		ctx context.Context,
		deviceID string,
	) (*model.DeviceDeployment, error)
	UpdateDeviceDeploymentStatus(
		ctx context.Context,
		deviceID string,
		deploymentID string,
		state model.DeviceDeploymentState,
		currentStatus model.DeviceDeploymentStatus,
	) (model.DeviceDeploymentStatus, error)
	UpdateDeviceDeploymentLogAvailability(ctx context.Context,
		deviceID string, deploymentID string, log bool) error
	AssignArtifact(
		ctx context.Context,
		deviceID string,
		deploymentID string,
		artifact *model.Image,
	) error
	AggregateDeviceDeploymentByStatus(ctx context.Context,
		id string) (model.Stats, error)
	GetDeviceStatusesForDeployment(ctx context.Context,
		deploymentID string) ([]model.DeviceDeployment, error)
	GetDevicesListForDeployment(ctx context.Context,
		query ListQuery) ([]model.DeviceDeployment, int, error)
	GetDeviceDeploymentsForDevice(ctx context.Context,
		query ListQueryDeviceDeployments) ([]model.DeviceDeployment, int, error)
	HasDeploymentForDevice(ctx context.Context,
		deploymentID string, deviceID string) (bool, error)
	AbortDeviceDeployments(ctx context.Context, deploymentID string) error
	DeleteDeviceDeploymentsHistory(ctx context.Context, deviceId string) error
	DecommissionDeviceDeployments(ctx context.Context, deviceId string) error
	GetDeviceDeployment(ctx context.Context, deploymentID string,
		deviceID string, includeDeleted bool) (*model.DeviceDeployment, error)
	GetDeviceDeployments(
		ctx context.Context,
		skip int,
		limit int,
		deviceID string,
		active *bool,
		includeDeleted bool,
	) ([]model.DeviceDeployment, error)
	SaveDeviceDeploymentRequest(
		ctx context.Context,
		ID string,
		request *model.DeploymentNextRequest,
	) error

	// deployments
	InsertDeployment(ctx context.Context, deployment *model.Deployment) error
	DeleteDeployment(ctx context.Context, id string) error
	FindDeploymentByID(ctx context.Context, id string) (*model.Deployment, error)
	FindDeploymentStatsByIDs(ctx context.Context, ids ...string) ([]*model.DeploymentStats, error)
	FindUnfinishedByID(ctx context.Context,
		id string) (*model.Deployment, error)
	UpdateStatsInc(
		ctx context.Context,
		id string,
		stateFrom,
		stateTo model.DeviceDeploymentStatus,
	) (model.Stats, error)
	UpdateStats(ctx context.Context,
		id string, stats model.Stats) error
	Find(ctx context.Context,
		query model.Query) ([]*model.Deployment, int64, error)
	SetDeploymentStatus(
		ctx context.Context,
		id string,
		status model.DeploymentStatus,
		now time.Time,
	) error
	FindNewerActiveDeployment(ctx context.Context,
		createdAfter *time.Time, deviceID string) (*model.Deployment, error)
	FindNewerActiveDeployments(ctx context.Context,
		createdAfter *time.Time, skip, limit int) ([]*model.Deployment, error)
	ExistUnfinishedByArtifactId(ctx context.Context, id string) (bool, error)
	ExistUnfinishedByArtifactName(ctx context.Context, artifactName string) (bool, error)
	ExistByArtifactId(ctx context.Context, id string) (bool, error)
	SetDeploymentDeviceCount(ctx context.Context, deploymentID string, count int) error
	IncrementDeploymentDeviceCount(ctx context.Context, deploymentID string, increment int) error
	IncrementDeploymentTotalSize(ctx context.Context, deploymentID string, increment int64) error
	DeviceCountByDeployment(ctx context.Context, id string) (int, error)
	UpdateDeploymentsWithArtifactName(
		ctx context.Context,
		artifactName string,
		artifactIDs []string,
	) error
	GetDeploymentIDsByArtifactNames(ctx context.Context, artifactNames []string) ([]string, error)

	GetTenantDbs() ([]string, error)
	SaveLastDeviceDeploymentStatus(
		ctx context.Context,
		deviceDeployment model.DeviceDeployment,
	) error
	GetLastDeviceDeploymentStatus(
		ctx context.Context,
		devicesIds []string,
	) ([]model.DeviceDeploymentLastStatus, error)

	// Releases
	ReplaceReleaseTags(
		ctx context.Context,
		releaseName string,
		tags model.Tags,
	) error
	UpdateRelease(
		ctx context.Context,
		releaseName string,
		release model.ReleasePatch,
	) error
	ListReleaseTags(ctx context.Context) (model.Tags, error)
	SaveUpdateTypes(ctx context.Context, updateTypes []string) error
	GetUpdateTypes(ctx context.Context) ([]string, error)
	DeleteReleasesByNames(ctx context.Context, names []string) error
}

var ErrNotFound = errors.New("document not found")

type Iterator[T interface{}] interface {
	Next(ctx context.Context) (bool, error)
	Decode(value *T) error
	Close(ctx context.Context) error
}
