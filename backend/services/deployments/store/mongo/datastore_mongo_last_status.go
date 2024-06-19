// Copyright 2023 Northern.tech AS
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
package mongo

import (
	"context"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	mopts "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mendersoftware/mender-server/pkg/identity"
	mstore "github.com/mendersoftware/mender-server/pkg/store"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

var (
	ErrTenantRequired = errors.New("tenant id is required")
)

func (db *DataStoreMongo) SaveLastDeviceDeploymentStatus(
	ctx context.Context,
	deviceDeployment model.DeviceDeployment,
) error {
	tenantId := ""
	id := identity.FromContext(ctx)
	if id != nil {
		tenantId = id.Tenant
	}
	filter := bson.M{
		"_id": deviceDeployment.DeviceId,
	}

	lastStatus := model.DeviceDeploymentLastStatus{
		DeviceId:               deviceDeployment.DeviceId,
		DeploymentId:           deviceDeployment.DeploymentId,
		DeviceDeploymentId:     deviceDeployment.Id,
		DeviceDeploymentStatus: deviceDeployment.Status,
		TenantId:               tenantId,
	}

	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDevs := database.Collection(CollectionDevicesLastStatus)
	var err error
	replaceOptions := mopts.Replace()
	replaceOptions.SetUpsert(true)
	_, err = collDevs.ReplaceOne(ctx, filter, lastStatus, replaceOptions)
	return err
}

func (db *DataStoreMongo) GetLastDeviceDeploymentStatus(
	ctx context.Context,
	devicesIds []string,
) ([]model.DeviceDeploymentLastStatus, error) {
	database := db.client.Database(mstore.DbFromContext(ctx, DatabaseName))
	collDevs := database.Collection(CollectionDevicesLastStatus)

	tenantId := ""
	id := identity.FromContext(ctx)
	if id != nil && id.Tenant != "" {
		tenantId = id.Tenant
	}
	filter := bson.M{
		"_id":              bson.M{"$in": devicesIds},
		StorageKeyTenantId: tenantId,
	}
	var statuses []model.DeviceDeploymentLastStatus
	cursor, err := collDevs.Find(ctx, filter)
	if err != nil {
		return statuses, err
	}

	err = cursor.All(ctx, &statuses)
	if err != nil {
		return statuses, err
	}

	return statuses, err
}
