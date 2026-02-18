// Copyright 2021 Northern.tech AS
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
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	mopts "go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"
	mstore "github.com/mendersoftware/mender-server/pkg/store"

	"github.com/mendersoftware/mender-server/services/inventory/model"
)

// batchSize for performing upserts during maintenance.
const batchSize = 1024

type migration_1_0_0 struct {
	ms  *DataStoreMongo
	ctx context.Context
	// maintenance determines whether to run migration in "maintenance mode"
	// NOTE: While maintenance is in progress, the following APIs
	//       MUST be DISABLED:
	//        - DELETE /devices/{id}       (management)
	//        - DELETE /devices/{id}/group (management)
	//        - PUT /devices/{id}/group    (management)
	//        - PATCH /devices/attributes  (devices)
	// Until we have an actual mesh system and a circuit breaker that can handle
	// disabling of the API internally, we'll have to do maintenance manually.
	maintenance bool
}

const (
	createdNameField = DbDevAttributes + "." + model.AttrScopeSystem +
		"-" + model.AttrNameCreated + "." + DbDevAttributesName
	createdValueField = DbDevAttributes + "." + model.AttrScopeSystem +
		"-" + model.AttrNameCreated + "." + DbDevAttributesValue
	createdScopeField = DbDevAttributes + "." + model.AttrScopeSystem +
		"-" + model.AttrNameCreated + "." + DbDevAttributesScope
	updatedNameField = DbDevAttributes + "." + model.AttrScopeSystem +
		"-" + model.AttrNameUpdated + "." + DbDevAttributesName
	updatedValueField = DbDevAttributes + "." + model.AttrScopeSystem +
		"-" + model.AttrNameUpdated + "." + DbDevAttributesValue
	updatedScopeField = DbDevAttributes + "." + model.AttrScopeSystem +
		"-" + model.AttrNameUpdated + "." + DbDevAttributesScope
	groupNameField = DbDevAttributes + "." + model.AttrScopeSystem +
		"-" + model.AttrNameGroup + "." + DbDevAttributesName
	groupScopeField = DbDevAttributes + "." + model.AttrScopeSystem +
		"-" + model.AttrNameGroup + "." + DbDevAttributesScope
)

// doMaintenance applies the migration, but instead of replacing document fields
// the new fields are upserted.
func (m *migration_1_0_0) doMaintenance(from migrate.Version) error {
	var N int
	database := m.ms.client.Database(mstore.DbFromContext(m.ctx, DbName))
	collDevs := database.Collection(DbDevicesColl)
	collMgrInfo := database.Collection("migration_info")
	l := log.FromContext(m.ctx)

	findOpts := mopts.Find()
	findOpts.SetProjection(bson.M{
		"_id":               1,
		model.AttrNameGroup: 1,
		"updated_ts":        1,
		"created_ts":        1,
	})
	findOpts.SetBatchSize(batchSize)
	type partialDevice struct {
		ID        string    `bson:"_id"`
		Group     string    `bson:"group,omitempty"`
		UpdatedTS time.Time `bson:"updated_ts"`
		CreatedTS time.Time `bson:"created_ts"`
	}
	bulkUpdates := make([]mongo.WriteModel, 0, batchSize)
	cur, err := collDevs.Find(m.ctx, bson.M{}, findOpts)
	if err != nil {
		return err
	}
	// Loop until inner loop updates a partial batch
	for i := batchSize; i >= batchSize; {
		for i = 0; cur.Next(m.ctx) && i < batchSize; i++ {
			var device partialDevice
			err = cur.Decode(&device)
			if err != nil {
				return err
			}
			updateModel := mongo.NewUpdateOneModel()
			updateModel.SetUpsert(true)
			updateModel.SetFilter(bson.M{"_id": device.ID})
			update := bson.M{
				createdNameField:  model.AttrNameCreated,
				createdScopeField: model.AttrScopeSystem,
				createdValueField: device.CreatedTS,
				updatedNameField:  model.AttrNameUpdated,
				updatedScopeField: model.AttrScopeSystem,
				updatedValueField: device.UpdatedTS,
			}
			if device.Group != "" {
				update[groupNameField] = model.AttrNameGroup
				update[groupScopeField] = model.AttrScopeSystem
				update[DbDevAttributesGroupValue] = device.Group
			}
			updateModel.SetUpdate(bson.M{"$set": update})
			bulkUpdates = append(bulkUpdates, updateModel)
		}
		_, err = collDevs.BulkWrite(m.ctx, bulkUpdates)
		if err != nil {
			return err
		}
		bulkUpdates = bulkUpdates[:0]
		N += i
	}
	l.Infof("Successfully upserted %d devices", N)
	filter := bson.M{"maintenance": bson.M{"version": m.Version()}}
	doc := bson.M{
		"maintenance": filter["maintenance"],
		"timestamp":   time.Now(),
	}
	// In case maintenance document already exists, replace timestamp
	replOpt := mopts.FindOneAndReplace()
	replOpt.SetUpsert(true)
	collMgrInfo.FindOneAndReplace(m.ctx, filter, doc, replOpt)
	return nil
}

func (m *migration_1_0_0) doMigrate(from migrate.Version) error {
	databaseName := mstore.DbFromContext(m.ctx, DbName)
	collDevs := m.ms.client.Database(databaseName).Collection(DbDevicesColl)

	// Move timestamps to identity scope.
	_, err := collDevs.UpdateMany(m.ctx, bson.M{}, bson.M{
		"$rename": bson.M{
			model.AttrNameCreated: createdValueField,
			model.AttrNameUpdated: updatedValueField,
		},
		"$set": bson.M{
			createdNameField:  model.AttrNameCreated,
			createdScopeField: model.AttrScopeSystem,
			updatedNameField:  model.AttrNameUpdated,
			updatedScopeField: model.AttrScopeSystem,
		},
	})
	if err != nil {
		return nil
	}

	// For devices in a group: move to attribute.
	_, err = collDevs.UpdateMany(
		m.ctx,
		bson.M{DbDevGroup: bson.M{"$exists": true}},
		bson.M{
			"$rename": bson.M{
				model.AttrNameGroup: DbDevAttributesGroupValue,
			},
			"$set": bson.M{
				groupNameField:  model.AttrNameGroup,
				groupScopeField: model.AttrScopeSystem,
			},
		})
	return err
}

func (m *migration_1_0_0) doCleanup() error {
	databaseName := mstore.DbFromContext(m.ctx, DbName)
	collDevs := m.ms.client.Database(databaseName).Collection(DbDevicesColl)

	update := bson.M{"$unset": bson.M{
		"updated_ts": "",
		"created_ts": "",
		DbDevGroup:   "",
	}}
	_, err := collDevs.UpdateMany(m.ctx, bson.M{}, update)
	return err
}

// Up creates timestamp and group attributes under the identity scope.
// The values reflect the values previously held in the root of the document.
func (m *migration_1_0_0) Up(from migrate.Version) error {
	l := log.FromContext(m.ctx)
	tenantDB := mstore.DbFromContext(m.ctx, DbName)
	if !migrate.VersionIsLess(from, m.Version()) {
		l.Infof("db '%s' already migrated", tenantDB)
		return nil
	}

	if m.maintenance {
		// Perform upserts
		return m.doMaintenance(from)
	}
	database := m.ms.client.Database(tenantDB)
	collMgrInfo := database.Collection("migration_info")
	// Check if maintenance operation has run
	res := collMgrInfo.FindOne(m.ctx, bson.M{"maintenance.version": m.Version()})
	if res.Err() == nil {
		// Perform maintenance cleanup
		err := m.doCleanup()
		if err == nil {
			_, err = collMgrInfo.DeleteOne(m.ctx, bson.M{"maintenance.version": m.Version()})
		}
		return err
	}
	// Run ordinary migration
	return m.doMigrate(from)
}

func (m *migration_1_0_0) Version() migrate.Version {
	return migrate.MakeVersion(1, 0, 0)
}
