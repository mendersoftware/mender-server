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
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/mendersoftware/mender-server/pkg/identity"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

var (
	ErrTenantRequired = errors.New("tenant id is required")
)

func (db *DataStoreMongo) GetLastDeviceDeploymentStatus(
	ctx context.Context,
	devicesIds []string,
) ([]model.DeviceDeploymentLastStatus, error) {
	collDevs := db.database.Collection(CollectionDevices)

	tenantId := ""
	id := identity.FromContext(ctx)
	if id != nil && id.Tenant != "" {
		tenantId = id.Tenant
	}
	cursor, err := collDevs.Aggregate(ctx, mongo.Pipeline{
		// For each device ID
		{{Key: "$match", Value: bson.M{
			StorageKeyDeviceDeploymentDeviceId: bson.M{"$in": devicesIds},
			StorageKeyTenantId:                 tenantId,
		}}},
		// Sort deployments by last one finished first
		{{Key: "$sort", Value: bson.M{
			StorageKeyDeviceDeploymentFinished: -1,
		}}},
		// Pick the first result
		{{Key: "$group", Value: bson.M{
			"_id": "$" + StorageKeyDeviceDeploymentDeviceId,
			"status": bson.M{
				"$first": "$$ROOT",
			},
		}}},
		// Mapping only relevant field to result
		{{Key: "$project", Value: bson.M{
			"device_id":                "$_id",
			"deployment_id":            "$status." + StorageKeyDeviceDeploymentDeploymentID,
			"device_deployment_id":     "$status._id",
			"device_deployment_status": "$status." + StorageKeyDeviceDeploymentStatus,
		}}},
	})
	if err != nil {
		return nil, err
	}
	var statuses []model.DeviceDeploymentLastStatus
	err = cursor.All(ctx, &statuses)
	if err != nil {
		return nil, err
	}
	return statuses, err
}
