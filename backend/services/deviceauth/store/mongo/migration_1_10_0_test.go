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

	"github.com/mendersoftware/mender-server/services/deviceauth/model"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	ctxstore "github.com/mendersoftware/mender-server/pkg/store"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMigration_1_10_0(t *testing.T) {
	ctx := identity.WithContext(context.Background(), &identity.Identity{
		Tenant: "foo",
	})
	db.Wipe()

	client := db.Client()
	db := NewDataStoreMongoWithClient(client)

	prep_1_9_0(t, ctx, db)

	mig1100 := migration_1_10_0{
		ds:  db,
		ctx: ctx,
	}
	err := mig1100.Up(migrate.MakeVersion(1, 10, 0))
	assert.NoError(t, err)

	_false := false
	_true := true

	dbName := ctxstore.DbFromContext(ctx, DbName)
	verifyIndexes(t, db.client.Database(dbName).Collection(DbDevicesColl),
		[]mongo.IndexModel{
			{
				Keys: bson.D{
					{Key: model.DevKeyIdDataSha256, Value: 1},
				},
				Options: &options.IndexOptions{
					Background: &_false,
					Name:       &indexDevices_IdentityDataSha256,
					Unique:     &_true,
				},
			},
			{
				Keys: bson.D{
					{Key: model.DevKeyStatus, Value: 1},
					{Key: model.DevKeyId, Value: 1},
				},
				Options: &options.IndexOptions{
					Background: &_false,
					Name:       &indexDevices_Status,
				},
			},
		},
	)
}

func prep_1_9_0(t *testing.T, ctx context.Context, db *DataStoreMongo) {

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
	mig190 := migration_1_9_0{
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
	err = mig190.Up(migrate.MakeVersion(1, 9, 0))
	assert.NoError(t, err)
}
