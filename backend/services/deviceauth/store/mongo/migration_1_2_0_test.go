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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"
	ctxstore "github.com/mendersoftware/mender-server/pkg/store"

	"github.com/mendersoftware/mender-server/services/deviceauth/model"
	"github.com/mendersoftware/mender-server/services/deviceauth/utils"
)

func TestMigration_1_2_0(t *testing.T) {
	ctx := identity.WithContext(context.Background(), &identity.Identity{
		Tenant: "foo",
	})
	db.Wipe()
	db := NewDataStoreMongoWithClient(db.Client())

	// prep base version
	mig110 := migration_1_1_0{
		ms:  db,
		ctx: ctx,
	}
	err := mig110.Up(migrate.MakeVersion(0, 1, 0))
	assert.NoError(t, err)

	ts := time.Now()

	devsColl := db.client.Database(ctxstore.DbFromContext(ctx, DbName)).Collection(DbDevicesColl)
	authSetsColl := db.client.Database(ctxstore.DbFromContext(ctx, DbName)).Collection(DbAuthSetColl)

	devs := []interface{}{
		model.Device{
			Id:              "1",
			IdData:          "{\"sn\":\"0001\",\"mac\":\"00:00:00:01\"}",
			Status:          "pending",
			Decommissioning: false,
			CreatedTs:       ts,
			UpdatedTs:       ts,
		},
		model.Device{
			Id:              "2",
			IdData:          "{\"sn\":\"0002\",\"attr\":\"foo1\",\"mac\":\"00:00:00:02\"}",
			Status:          "active",
			Decommissioning: false,
			CreatedTs:       ts,
			UpdatedTs:       ts,
		},
	}

	asets := []interface{}{
		model.AuthSet{
			Id:        "1",
			DeviceId:  "1",
			IdData:    "{\"sn\":\"0001\",\"mac\":\"00:00:00:01\"}",
			PubKey:    "pubkey-1",
			Timestamp: &ts,
			Status:    "pending",
		},
		model.AuthSet{
			Id:        "2",
			DeviceId:  "2",
			IdData:    "{\"sn\":\"0002\",\"attr\":\"foo\",\"mac\":\"00:00:00:02\"}",
			PubKey:    "pubkey-2",
			Timestamp: &ts,
			Status:    "active",
		},
		model.AuthSet{
			Id:        "3",
			DeviceId:  "2",
			IdData:    "{\"sn\":\"0002\",\"attr\":\"foo1\",\"mac\":\"00:00:00:02\"}",
			PubKey:    "pubkey-3",
			Timestamp: &ts,
			Status:    "active",
		},
	}

	_, err = devsColl.InsertMany(ctx, devs)
	assert.NoError(t, err)

	_, err = authSetsColl.InsertMany(ctx, asets)
	assert.NoError(t, err)

	// test new version
	mig120 := migration_1_2_0{
		ms:  db,
		ctx: ctx,
	}
	// here as well there was an error it was 1,1,0
	err = mig120.Up(migrate.MakeVersion(1, 2, 0))
	assert.NoError(t, err)

	var dev model.Device
	c := db.client.Database(ctxstore.DbFromContext(ctx, DbName)).Collection(DbDevicesColl)
	for _, ds := range devs {
		d := ds.(model.Device)
		err = c.FindOne(ctx, bson.M{"_id": d.Id}).Decode(&dev)
		assert.NoError(t, err)

		id, err := utils.JsonSort(d.IdData)
		assert.NoError(t, err)

		d.IdData = id

		compareDevices(&d, &dev, t)
	}

	var set model.AuthSet
	c = db.client.Database(ctxstore.DbFromContext(ctx, DbName)).Collection(DbAuthSetColl)
	for _, as := range asets {
		a := as.(model.AuthSet)
		err = c.FindOne(ctx, bson.M{"_id": a.Id}).Decode(&set)
		assert.NoError(t, err)

		id, err := utils.JsonSort(a.IdData)
		assert.NoError(t, err)

		a.IdData = id

		compareAuthSet(&a, &set, t)
	}
}
