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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	ctxstore "github.com/mendersoftware/mender-server/pkg/store"

	"github.com/mendersoftware/mender-server/services/deviceauth/model"
)

func TestMigration_1_4_0(t *testing.T) {
	var err error
	pubKey := `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzogVU7RGDilbsoUt/DdH
VJvcepl0A5+xzGQ50cq1VE/Dyyy8Zp0jzRXCnnu9nu395mAFSZGotZVr+sWEpO3c
yC3VmXdBZmXmQdZqbdD/GuixJOYfqta2ytbIUPRXFN7/I7sgzxnXWBYXYmObYvdP
okP0mQanY+WKxp7Q16pt1RoqoAd0kmV39g13rFl35muSHbSBoAW3GBF3gO+mF5Ty
1ddp/XcgLOsmvNNjY+2HOD5F/RX0fs07mWnbD7x+xz7KEKjF+H7ZpkqCwmwCXaf0
iyYyh1852rti3Afw4mDxuVSD7sd9ggvYMc0QHIpQNkD4YWOhNiE1AB0zH57VbUYG
UwIDAQAB
-----END PUBLIC KEY-----`

	ts := time.Now()

	ctx := identity.WithContext(context.Background(), &identity.Identity{
		Tenant: "foo",
	})
	db.Wipe()
	db := NewDataStoreMongoWithClient(db.Client())
	devsColl := db.client.Database(ctxstore.DbFromContext(ctx, DbName)).Collection(DbDevicesColl)
	authSetsColl := db.client.Database(ctxstore.DbFromContext(ctx, DbName)).Collection(DbAuthSetColl)

	// prep base version
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
	err = mig110.Up(migrate.MakeVersion(1, 1, 0))
	assert.NoError(t, err)
	err = mig120.Up(migrate.MakeVersion(1, 2, 0))
	assert.NoError(t, err)
	err = mig130.Up(migrate.MakeVersion(1, 3, 0))
	assert.NoError(t, err)

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
			Status:          "rejected",
			Decommissioning: false,
			CreatedTs:       ts,
			UpdatedTs:       ts,
		},
		model.Device{
			Id:              "3",
			IdData:          "{\"sn\":\"0003\",\"attr\":\"foo3\",\"mac\":\"00:00:00:03\"}",
			Status:          "rejected",
			Decommissioning: false,
			CreatedTs:       ts,
			UpdatedTs:       ts,
		},
		model.Device{
			Id:              "4",
			IdData:          "{\"sn\":\"0004\",\"attr\":\"foo4\",\"mac\":\"00:00:00:04\"}",
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
			PubKey:    pubKey,
			Timestamp: &ts,
		},
		model.AuthSet{
			Id:        "2",
			DeviceId:  "2",
			IdData:    "{\"sn\":\"0002\",\"attr\":\"foo\",\"mac\":\"00:00:00:02\"}",
			Status:    "rejected",
			PubKey:    pubKey,
			Timestamp: &ts,
		},
		model.AuthSet{
			Id:        "3",
			DeviceId:  "2",
			IdData:    "{\"sn\":\"0002\",\"attr\":\"foo1\",\"mac\":\"00:00:00:02\"}",
			Status:    "pending",
			PubKey:    pubKey,
			Timestamp: &ts,
		},
		model.AuthSet{
			Id:        "4",
			DeviceId:  "3",
			IdData:    "{\"sn\":\"0003\",\"attr\":\"foo3\",\"mac\":\"00:00:00:03\"}",
			Status:    "rejected",
			PubKey:    pubKey,
			Timestamp: &ts,
		},
	}

	_, err = devsColl.InsertMany(ctx, devs)
	assert.NoError(t, err)

	_, err = authSetsColl.InsertMany(ctx, asets)
	assert.NoError(t, err)

	// test new version
	mig140 := migration_1_4_0{
		ms:  db,
		ctx: ctx,
	}
	err = mig140.Up(migrate.MakeVersion(1, 4, 0))
	assert.NoError(t, err)

	var dev model.Device
	var status string
	var noAuthSets bool

	for _, ds := range devs {
		d := ds.(model.Device)

		res := []model.AuthSet{}
		cursor, err := authSetsColl.Find(ctx, model.AuthSet{DeviceId: d.Id})
		assert.NoError(t, err)
		err = cursor.All(ctx, &res)
		if (err != nil && err == mongo.ErrNoDocuments) || len(res) == 0 {
			noAuthSets = true
		} else {
			noAuthSets = false
			assert.NoError(t, err)
		}

		if noAuthSets {
			status = model.DevStatusRejected
		} else {
			status, err = getDeviceStatusDB(
				db,
				ctxstore.DbFromContext(ctx, DbName),
				"",
				ctx,
				d.Id,
			)
			assert.NoError(t, err)
		}

		err = devsColl.FindOne(ctx, bson.M{"_id": d.Id}).Decode(&dev)
		assert.NoError(t, err)

		assert.Equal(t, status, dev.Status)

		d.Status = status
		d.UpdatedTs = dev.UpdatedTs

		compareDevices(&d, &dev, t)
	}
}
