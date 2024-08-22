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

package mongo

import (
	"context"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/mendersoftware/mender-server/services/deviceauth/model"
	"github.com/mendersoftware/mender-server/services/deviceauth/store"
)

// Retrieves devices with decommissioning flag set
func (db *DataStoreMongo) GetDevicesBeingDecommissioned(tenantId string) ([]model.Device, error) {
	c := db.client.Database(DbName).Collection(DbDevicesColl)

	devices := []model.Device{}

	cursor, err := c.Find(
		context.Background(),
		model.Device{
			Decommissioning: true,
			TenantID:        tenantId,
		},
	)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.Background(), &devices); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, store.ErrDevNotFound
		}
		return nil, errors.Wrap(err, "failed to fetch devices")
	}

	return devices, nil
}

// Retrieves Ids of the auth sets owned by devices that are in decommissioning state or not owned by
// any device.
func (db *DataStoreMongo) GetBrokenAuthSets(tenantId string) ([]string, error) {
	c := db.client.Database(DbName).Collection(DbAuthSetColl)

	deviceIds := []string{}
	brokenAuthSets := []string{}

	ctx := context.Background()

	// get all auth sets; group by device id
	match := bson.D{
		{
			Key: "$match", Value: bson.D{
				{Key: dbFieldTenantID, Value: tenantId}},
		},
	}
	group := bson.D{
		{
			Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$device_id"}},
		},
	}
	pipeline := []bson.D{
		match,
		group,
	}
	var result []struct {
		DeviceId string `bson:"_id"`
	}

	cursor, err := c.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	if err := cursor.All(ctx, &result); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	for _, res := range result {
		deviceIds = append(deviceIds, res.DeviceId)
	}

	//check if devices exists
	nonexistentDevices, err := db.filterNonExistentDevices(tenantId, deviceIds)
	if err != nil {
		return nil, err
	}

	// fetch auth sets for non exisitent devices
	for _, dev := range nonexistentDevices {

		authSets := []model.AuthSet{}

		cursor, err := c.Find(ctx, bson.M{"device_id": dev, dbFieldTenantID: tenantId})
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch authentication sets")
		}
		if err = cursor.All(ctx, &authSets); err != nil {
			if err != mongo.ErrNoDocuments {
				return nil, errors.Wrap(err, "failed to fetch authentication sets")
			}
		}

		for _, authSet := range authSets {
			brokenAuthSets = append(brokenAuthSets, authSet.Id)
		}
	}

	return brokenAuthSets, nil
}

// Deletes devices with decommissioning flag set
func (db *DataStoreMongo) DeleteDevicesBeingDecommissioned(tenantId string) error {
	c := db.client.Database(DbName).Collection(DbDevicesColl)

	_, err := c.DeleteMany(
		context.Background(),
		model.Device{
			Decommissioning: true,
			TenantID:        tenantId,
		},
	)
	if err != nil {
		return errors.Wrap(err, "failed to remove decommissioned devices")
	}

	return nil
}

// Deletes auth sets owned by devices that are in decommissioning state and auth sets not
// owned by any device.
func (db *DataStoreMongo) DeleteBrokenAuthSets(tenantId string) error {
	database := db.client.Database(DbName)
	collAuthSets := database.Collection(DbAuthSetColl)
	collTokens := database.Collection(DbTokensColl)

	authSets, err := db.GetBrokenAuthSets(tenantId)
	if err != nil {
		return err
	}

	// delete authsets for non-existent devices
	ctx := context.Background()
	for _, as := range authSets {
		_, err := collAuthSets.DeleteOne(ctx, model.AuthSet{Id: as, TenantID: tenantId})
		if err != nil {
			return errors.Wrapf(err, "tenant %s, failed to delete authentication sets", tenantId)
		}
		// Attempt to delete token (may have already expired).
		_, _ = collTokens.DeleteOne(ctx, bson.M{"_id": as, dbFieldTenantClaim: tenantId})
	}

	return nil
}

// Filters list of device ids.
// Result is the list of ids of non-existent devices and devices with decommissioning flag set.
func (db *DataStoreMongo) filterNonExistentDevices(
	tenantId string,
	devIds []string,
) ([]string, error) {
	c := db.client.Database(DbName).Collection(DbDevicesColl)

	nonexistentDevices := []string{}

	//check if device exists
	for _, devId := range devIds {
		res := model.Device{}
		err := c.FindOne(
			context.Background(),
			bson.M{
				"_id":           devId,
				dbFieldTenantID: tenantId,
			},
		).Decode(&res)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				nonexistentDevices = append(nonexistentDevices, devId)
			} else {
				return nil, errors.Wrapf(err, "tenant %s, failed to fetch device", tenantId)
			}
		}

		if res.Decommissioning {
			nonexistentDevices = append(nonexistentDevices, devId)
		}
	}

	return nonexistentDevices, nil
}
