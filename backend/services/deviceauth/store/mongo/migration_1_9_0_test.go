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
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	ctxstore "github.com/mendersoftware/mender-server/pkg/store"

	"github.com/mendersoftware/mender-server/services/deviceauth/model"
)

func TestMigration_1_9_0(t *testing.T) {
	ts := time.Now()

	ctx := identity.WithContext(context.Background(), &identity.Identity{
		Tenant: "foo",
	})
	db.Wipe()
	db := NewDataStoreMongoWithClient(db.Client())
	devsColl := db.client.Database(ctxstore.DbFromContext(ctx, DbName)).Collection(DbDevicesColl)
	authSetsColl := db.client.Database(ctxstore.DbFromContext(ctx, DbName)).Collection(DbAuthSetColl)

	prep_1_8_0(t, ctx, db)

	devs := []interface{}{
		model.Device{
			Id:              "1",
			IdData:          "{\"sn\":\"0001\",\"mac\":\"00:00:00:01\"}",
			IdDataSha256:    getIdDataHash("{\"sn\":\"0001\",\"mac\":\"00:00:00:01\"}"),
			Status:          "accepted",
			Decommissioning: false,
			CreatedTs:       ts,
			UpdatedTs:       ts,
		},
		model.Device{
			Id:              "2",
			IdData:          "{\"sn\":\"0002\",\"attr\":\"foo1\",\"mac\":\"00:00:00:02\"}",
			IdDataSha256:    getIdDataHash("{\"sn\":\"0002\",\"attr\":\"foo1\",\"mac\":\"00:00:00:02\"}"),
			Status:          "accepted",
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
			Status:    "accepted",
			PubKey:    "key1",
			Timestamp: &ts,
		},
		model.AuthSet{
			Id:        "2",
			DeviceId:  "2",
			IdData:    "{\"sn\":\"0002\",\"attr\":\"foo\",\"mac\":\"00:00:00:02\"}",
			Status:    "accepted",
			PubKey:    "key2",
			Timestamp: &ts,
		},
	}

	_, err := devsColl.InsertMany(ctx, devs)
	assert.NoError(t, err)

	_, err = authSetsColl.InsertMany(ctx, asets)
	assert.NoError(t, err)

	// NOTE: after upgrading past mongodb 4.0 there are no longer
	//       restrictions on Index Key Size
	// ref: https://docs.mongodb.com/manual/reference/limits/#Index-Key-Limit
	// try device or authset with 'too large' id data - should fail
	idData := randstr(4096)
	devTooLarge := model.Device{
		Id:              "3",
		IdData:          idData,
		IdDataSha256:    getIdDataHash(idData),
		Status:          "accepted",
		Decommissioning: false,
		CreatedTs:       ts,
		UpdatedTs:       ts,
	}

	// err := db.AddDevice(ctx, devTooLarge)
	// assert.NotNil(t, err)

	asetTooLarge := model.AuthSet{
		Id:           "3",
		DeviceId:     "3",
		IdData:       idData,
		IdDataSha256: getIdDataHash(idData),
		Status:       "accepted",
		PubKey:       "key3",
		Timestamp:    &ts,
	}
	// err = db.AddAuthSet(ctx, asetTooLarge)
	// assert.NotNil(t, err)

	// test new version, long id data added successfully
	mig190 := migration_1_9_0{
		ds:  db,
		ctx: ctx,
	}
	err = mig190.Up(migrate.MakeVersion(1, 9, 0))
	assert.NoError(t, err)

	_, err = devsColl.InsertOne(ctx, devTooLarge)
	assert.NoError(t, err)

	_, err = authSetsColl.InsertOne(ctx, asetTooLarge)
	assert.NoError(t, err)

	// verify our uniqueness invariants
	// can't add device with duplicated id data sha
	dev := model.Device{
		Id:           "4",
		IdDataSha256: getIdDataHash(idData),
	}
	_, err = devsColl.InsertOne(ctx, dev)
	assert.Contains(t, err.Error(), "duplicate key error collection")

	// can't add authset with duplicated dev id + id data sha + pubkey
	aset := model.AuthSet{
		Id:           "4",
		DeviceId:     "3",
		IdDataSha256: getIdDataHash(idData),
		PubKey:       "key3",
	}
	_, err = authSetsColl.InsertOne(ctx, aset)
	assert.Contains(t, err.Error(), "duplicate key error collection")

	// can't add authset with duplicated id data sha + pubkey
	aset = model.AuthSet{
		Id:           "4",
		IdDataSha256: getIdDataHash(idData),
		PubKey:       "key3",
	}
	_, err = authSetsColl.InsertOne(ctx, aset)
	assert.Contains(t, err.Error(), "duplicate key error collection")
}

func TestMigration_1_9_0_NoFailDelete(t *testing.T) {
	ctx := identity.WithContext(context.Background(), &identity.Identity{
		Tenant: "foo",
	})
	db.Wipe()
	db := NewDataStoreMongoWithClient(db.Client())

	// no previous indexes to delete
	mig190 := migration_1_9_0{
		ds:  db,
		ctx: ctx,
	}
	err := mig190.Up(migrate.MakeVersion(1, 9, 0))
	assert.NoError(t, err)
}

func prep_1_8_0(t *testing.T, ctx context.Context, db *DataStoreMongo) {

	mig110 := migration_1_1_0{
		ms:  db,
		ctx: ctx,
	}
	mig120 := migration_1_2_0{
		ms:  db,
		ctx: ctx,
	}
	mig130 := migration_1_3_0{
		ms:  db,
		ctx: ctx,
	}
	mig140 := migration_1_4_0{
		ms:  db,
		ctx: ctx,
	}
	mig150 := migration_1_5_0{
		ms:  db,
		ctx: ctx,
	}
	mig160 := migration_1_6_0{
		ms:  db,
		ctx: ctx,
	}
	mig170 := migration_1_7_0{
		ms:  db,
		ctx: ctx,
	}
	mig180 := migration_1_8_0{
		ds:  db,
		ctx: ctx,
	}

	err := mig110.Up(migrate.MakeVersion(1, 1, 0))
	assert.NoError(t, err)
	err = mig120.Up(migrate.MakeVersion(1, 2, 0))
	assert.NoError(t, err)
	err = mig130.Up(migrate.MakeVersion(1, 3, 0))
	assert.NoError(t, err)
	err = mig140.Up(migrate.MakeVersion(1, 4, 0))
	assert.NoError(t, err)
	err = mig150.Up(migrate.MakeVersion(1, 5, 0))
	assert.NoError(t, err)
	err = mig160.Up(migrate.MakeVersion(1, 6, 0))
	assert.NoError(t, err)
	err = mig170.Up(migrate.MakeVersion(1, 7, 0))
	assert.NoError(t, err)
	err = mig180.Up(migrate.MakeVersion(1, 8, 0))
	assert.NoError(t, err)
}

func randstr(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
