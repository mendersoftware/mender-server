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

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/mendersoftware/mender-server/pkg/identity"
	mongostore "github.com/mendersoftware/mender-server/pkg/mongo/v2"

	"github.com/mendersoftware/mender-server/services/deviceauth/model"
)

func TestGetDevicesBeingDecommissioned(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestGetDevicesBeingDecommissioned in short mode.")
	}

	testCases := []struct {
		inDevices  bson.A
		outDevices []model.Device
		tenant     string
	}{
		{
			inDevices: bson.A{
				model.Device{
					Id:              "001",
					IdData:          "001",
					Status:          model.DevStatusPending,
					Decommissioning: false,
					IdDataSha256:    getIdDataHash("001"),
				},
				model.Device{
					Id:              "002",
					IdData:          "002",
					Status:          model.DevStatusPending,
					Decommissioning: true,
					IdDataSha256:    getIdDataHash("002"),
				},
			},
			outDevices: []model.Device{
				{
					Id:              "002",
					IdData:          "002",
					Status:          model.DevStatusPending,
					Decommissioning: true,
					IdDataSha256:    getIdDataHash("002"),
				},
			},
		},
		{
			inDevices: bson.A{
				model.Device{
					Id:              "001",
					IdData:          "001",
					Status:          model.DevStatusPending,
					Decommissioning: false,
					IdDataSha256:    getIdDataHash("001"),
					TenantID:        tenant,
				},
				model.Device{
					Id:              "002",
					IdData:          "002",
					Status:          model.DevStatusPending,
					Decommissioning: true,
					IdDataSha256:    getIdDataHash("002"),
					TenantID:        tenant,
				},
			},
			outDevices: []model.Device{
				{
					Id:           "002",
					IdDataSha256: getIdDataHash("002"),
					TenantID:     tenant,
				},
			},
			tenant: tenant,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("tc %d", i), func(t *testing.T) {
			ctx := identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tc.tenant,
			})

			db := getDb(ctx)

			coll := db.client.Database(DbName).Collection(DbDevicesColl)
			_, err := coll.InsertMany(ctx, tc.inDevices)
			assert.NoError(t, err)

			brokenDevices, err := db.GetDevicesBeingDecommissioned(tc.tenant)
			assert.NoError(t, err)
			assert.Equal(t, tc.outDevices[0].Id, brokenDevices[0].Id)
		})
	}
}

func TestDeleteDevicesBeingDecommissioned(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestDeleteDevicesBeingDecommissioned in short mode.")
	}

	testCases := []struct {
		inDevices bson.A
		tenant    string
	}{
		{
			inDevices: bson.A{
				model.Device{
					Id:              "001",
					IdData:          "001",
					Status:          model.DevStatusPending,
					Decommissioning: false,
					IdDataSha256:    getIdDataHash("001"),
				},
				model.Device{
					Id:              "002",
					IdData:          "002",
					Status:          model.DevStatusPending,
					Decommissioning: true,
					IdDataSha256:    getIdDataHash("002"),
				},
			},
		},
		{
			inDevices: bson.A{
				model.Device{
					Id:              "001",
					IdData:          "001",
					Status:          model.DevStatusPending,
					Decommissioning: false,
					IdDataSha256:    getIdDataHash("001"),
				},
				model.Device{
					Id:              "002",
					IdData:          "002",
					Status:          model.DevStatusPending,
					Decommissioning: true,
					IdDataSha256:    getIdDataHash("002"),
				},
			},
			tenant: tenant,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("tc %d", i), func(t *testing.T) {
			ctx := context.Background()
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.tenant,
				})
			}

			db := getDb(ctx)

			coll := db.client.Database(DbName).Collection(DbDevicesColl)
			_, err := coll.InsertMany(ctx, tc.inDevices)
			assert.NoError(t, err)

			err = db.DeleteDevicesBeingDecommissioned(tc.tenant)
			assert.NoError(t, err)

			dbDevs, err := db.GetDevices(ctx, 0, 5, model.DeviceFilter{})
			assert.NoError(t, err)
			for _, dbDev := range dbDevs {
				assert.Equal(t, false, dbDev.Decommissioning)
			}
		})
	}
}

func TestGetBrokenAuthSets(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestGetBrokenAuthSets in short mode.")
	}

	testCases := []struct {
		inAuthSets     bson.A
		inDevices      bson.A
		outAuthSetsIds []string
		tenant         string
		err            string
	}{
		{
			inAuthSets: bson.A{
				model.AuthSet{
					Id:       "001",
					DeviceId: "001",
					IdData:   "001",
					PubKey:   "001",
				},
				model.AuthSet{
					Id:       "002",
					DeviceId: "003",
					IdData:   "001",
					PubKey:   "002",
				},
			},
			inDevices: bson.A{
				model.Device{
					Id:              "001",
					IdData:          "001",
					Status:          model.DevStatusPending,
					Decommissioning: false,
					IdDataSha256:    getIdDataHash("001"),
				},
				model.Device{
					Id:              "002",
					IdData:          "002",
					Status:          model.DevStatusPending,
					Decommissioning: false,
					IdDataSha256:    getIdDataHash("002"),
				},
			},
			outAuthSetsIds: []string{"002"},
			err:            "",
		},
		{
			inAuthSets: bson.A{
				model.AuthSet{
					Id:       "001",
					DeviceId: "001",
					IdData:   "001",
					PubKey:   "001",
					TenantID: tenant,
				},
				model.AuthSet{
					Id:       "002",
					DeviceId: "003",
					IdData:   "001",
					PubKey:   "002",
					TenantID: tenant,
				},
			},
			inDevices: bson.A{
				model.Device{
					Id:              "001",
					IdData:          "001",
					Status:          model.DevStatusPending,
					IdDataSha256:    []byte(fmt.Sprintf("sha-%04d", rand.Int())),
					Decommissioning: false,
					TenantID:        tenant,
				},
				model.Device{
					Id:              "002",
					IdData:          "002",
					IdDataSha256:    []byte(fmt.Sprintf("sha-%04d", rand.Int())),
					Status:          model.DevStatusPending,
					Decommissioning: false,
					TenantID:        tenant,
				},
			},
			tenant:         tenant,
			outAuthSetsIds: []string{"002"},
			err:            "",
		},
		{
			inAuthSets: bson.A{
				model.AuthSet{
					Id:       "001",
					DeviceId: "001",
					IdData:   "001",
					PubKey:   "001",
					TenantID: tenant,
				},
				model.AuthSet{
					Id:       "002",
					DeviceId: "002",
					IdData:   "001",
					PubKey:   "002",
					TenantID: tenant,
				},
			},
			inDevices: bson.A{
				model.Device{
					Id:              "001",
					IdData:          "001",
					IdDataSha256:    []byte(fmt.Sprintf("sha-%04d", rand.Int())),
					Status:          model.DevStatusPending,
					Decommissioning: false,
					TenantID:        tenant,
				},
				model.Device{
					Id:              "002",
					IdData:          "002",
					IdDataSha256:    []byte(fmt.Sprintf("sha-%04d", rand.Int())),
					Status:          model.DevStatusPending,
					Decommissioning: true,
					TenantID:        tenant,
				},
			},
			tenant:         tenant,
			outAuthSetsIds: []string{"002"},
			err:            "",
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("tc %d", i), func(t *testing.T) {
			ctx := identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tc.tenant,
			})

			db := getDb(ctx)

			coll := db.client.Database(DbName).Collection(DbAuthSetColl)
			var err error
			for _, a := range tc.inAuthSets {
				_, err = coll.InsertOne(ctx, mongostore.WithTenantID(ctx, a))
				assert.NoError(t, err)
			}

			coll = db.client.Database(DbName).Collection(DbDevicesColl)
			for _, d := range tc.inDevices {
				_, err = coll.InsertOne(ctx, mongostore.WithTenantID(ctx, d))
				assert.NoError(t, err)
			}

			brokenAuthSetsIds, err := db.GetBrokenAuthSets(tc.tenant)
			if tc.err != "" {
				assert.Equal(t, tc.err, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.outAuthSetsIds, brokenAuthSetsIds)
			}
		})
	}
}

func TestDeleteBrokenAuthSets(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestDeleteBrokenAuthSets in short mode.")
	}

	testCases := []struct {
		inAuthSets bson.A
		inDevices  bson.A
		tenant     string
	}{
		{
			inAuthSets: bson.A{
				model.AuthSet{
					Id:       "001",
					DeviceId: "001",
					IdData:   "001",
					PubKey:   "001",
				},
				model.AuthSet{
					Id:       "002",
					DeviceId: "003",
					IdData:   "001",
					PubKey:   "002",
				},
			},
			inDevices: bson.A{
				model.Device{
					Id:              "001",
					IdData:          "001",
					Status:          model.DevStatusPending,
					Decommissioning: false,
					IdDataSha256:    getIdDataHash("001"),
				},
				model.Device{
					Id:              "002",
					IdData:          "002",
					Status:          model.DevStatusPending,
					Decommissioning: false,
					IdDataSha256:    getIdDataHash("002"),
				},
			},
		},
		{
			inAuthSets: bson.A{
				model.AuthSet{
					Id:       "001",
					DeviceId: "001",
					IdData:   "001",
					PubKey:   "001",
				},
				model.AuthSet{
					Id:       "002",
					DeviceId: "003",
					IdData:   "001",
					PubKey:   "002",
				},
			},
			inDevices: bson.A{
				model.Device{
					Id:              "001",
					IdData:          "001",
					Status:          model.DevStatusPending,
					Decommissioning: false,
					IdDataSha256:    getIdDataHash("001"),
				},
				model.Device{
					Id:              "002",
					IdData:          "002",
					Status:          model.DevStatusPending,
					Decommissioning: false,
					IdDataSha256:    getIdDataHash("002"),
				},
			},
			tenant: tenant,
		},
		{
			inAuthSets: bson.A{
				model.AuthSet{
					Id:       "001",
					DeviceId: "001",
					IdData:   "001",
					PubKey:   "001",
				},
				model.AuthSet{
					Id:       "002",
					DeviceId: "002",
					IdData:   "001",
					PubKey:   "002",
				},
			},
			inDevices: bson.A{
				model.Device{
					Id:              "001",
					IdData:          "001",
					Status:          model.DevStatusPending,
					Decommissioning: false,
					IdDataSha256:    getIdDataHash("001"),
				},
				model.Device{
					Id:              "002",
					IdData:          "002",
					Status:          model.DevStatusPending,
					Decommissioning: true,
					IdDataSha256:    getIdDataHash("002"),
				},
			},
			tenant: tenant,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("tc %d", i), func(t *testing.T) {
			ctx := context.Background()
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.tenant,
				})
			}

			db := getDb(ctx)

			coll := db.client.Database(DbName).Collection(DbAuthSetColl)
			_, err := coll.InsertMany(ctx, tc.inAuthSets)
			assert.NoError(t, err)

			coll = db.client.Database(DbName).Collection(DbDevicesColl)
			_, err = coll.InsertMany(ctx, tc.inDevices)
			assert.NoError(t, err)

			err = db.DeleteBrokenAuthSets(tc.tenant)
			assert.NoError(t, err)

			brokenAuthSetsIds, err := db.GetBrokenAuthSets(tc.tenant)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(brokenAuthSetsIds))
		})
	}
}
