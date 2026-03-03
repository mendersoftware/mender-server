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
package cmd

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	//"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"
	"github.com/mendersoftware/mender-server/pkg/mongo/v2/oid"
	ctxstore "github.com/mendersoftware/mender-server/pkg/store"

	minv "github.com/mendersoftware/mender-server/services/deviceauth/client/inventory/mocks"
	mwflows "github.com/mendersoftware/mender-server/services/deviceauth/client/orchestrator/mocks"

	"github.com/mendersoftware/mender-server/services/deviceauth/store"

	//dconfig "github.com/mendersoftware/mender-server/services/deviceauth/config"
	"github.com/mendersoftware/mender-server/services/deviceauth/jwt"
	"github.com/mendersoftware/mender-server/services/deviceauth/model"
	mstore "github.com/mendersoftware/mender-server/services/deviceauth/store/mocks"
	"github.com/mendersoftware/mender-server/services/deviceauth/store/mongo"
)

//func TestMaintenance(t *testing.T) {
//	if testing.Short() {
//		t.Skip("skipping TestMaintenance in short mode.")
//	}
//
//	config.SetDefaults(config.Config, dconfig.Defaults)
//	// Enable setting config values by environment variables
//	config.Config.SetEnvPrefix("DEVICEAUTH")
//	config.Config.AutomaticEnv()
//
//	err := Maintenance(true, "", false)
//	assert.NoError(t, err)
//}

func TestMaintenanceWithDataStore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMaintenanceWithDataStore in short mode.")
	}
	datasetDevices := []interface{}{
		model.Device{
			Id:              oid.NewUUIDv5("001").String(),
			IdData:          "001",
			Status:          model.DevStatusPending,
			Decommissioning: false,
		},
		model.Device{
			Id:              oid.NewUUIDv5("002").String(),
			IdData:          "002",
			Status:          model.DevStatusPending,
			Decommissioning: true,
		},
	}

	datasetAuthSets := []interface{}{
		model.AuthSet{
			Id:       oid.NewUUIDv5("001").String(),
			DeviceId: oid.NewUUIDv5("001").String(),
			IdData:   "001",
			PubKey:   "001",
		},
		model.AuthSet{
			Id:       oid.NewUUIDv5("002").String(),
			DeviceId: oid.NewUUIDv5("003").String(),
			IdData:   "001",
			PubKey:   "002",
		},
	}

	datasetTokens := []interface{}{
		jwt.Token{Claims: jwt.Claims{
			ID:        oid.NewUUIDv5("001"),
			Subject:   oid.NewUUIDv5("001"),
			Issuer:    "Tester",
			ExpiresAt: jwt.Time{Time: time.Now().Add(time.Hour)},
		}},
		jwt.Token{Claims: jwt.Claims{
			ID:        oid.NewUUIDv5("002"),
			Subject:   oid.NewUUIDv5("003"),
			Issuer:    "Tester",
			ExpiresAt: jwt.Time{Time: time.Now().Add(time.Hour)},
		}},
	}

	testCases := map[string]struct {
		decommissioningCleanupFlag bool
		tenant                     string
		dryRunFlag                 bool
		withDataSets               bool
	}{
		"do nothing": {
			decommissioningCleanupFlag: false,
		},
		"do nothing with tenant": {
			decommissioningCleanupFlag: false,
			tenant:                     "foo",
		},
		"dry run without data": {
			decommissioningCleanupFlag: true,
			dryRunFlag:                 true,
		},
		"dry run": {
			decommissioningCleanupFlag: true,
			dryRunFlag:                 true,
			withDataSets:               true,
		},
		"dry run with tenant": {
			decommissioningCleanupFlag: true,
			tenant:                     "foo",
			dryRunFlag:                 true,
			withDataSets:               true,
		},
		"run without data": {
			decommissioningCleanupFlag: true,
		},
		"run": {
			decommissioningCleanupFlag: true,
			withDataSets:               true,
		},
		"run with tenant": {
			decommissioningCleanupFlag: true,
			tenant:                     "foo",
			withDataSets:               true,
		},
	}

	for name, tc := range testCases {
		t.Logf("case: %s", name)

		db.Wipe()
		client := db.Client()
		ctx := context.Background()
		if tc.tenant != "" {
			ctx = identity.WithContext(ctx, &identity.Identity{
				Tenant: tc.tenant,
			})
		}
		ds := mongo.NewDataStoreMongoWithClient(client)

		if tc.withDataSets {

			testDbName := mongo.DbName
			if tc.tenant != "" {
				testDbName = ctxstore.DbNameForTenant(tc.tenant, mongo.DbName)
			}

			c := client.Database(testDbName).Collection(mongo.DbDevicesColl)
			_, err := c.InsertMany(ctx, datasetDevices)
			assert.NoError(t, err)
			c = client.Database(testDbName).Collection(mongo.DbAuthSetColl)
			_, err = c.InsertMany(ctx, datasetAuthSets)
			assert.NoError(t, err)
			c = client.Database(testDbName).Collection(mongo.DbTokensColl)
			_, err = c.InsertMany(ctx, datasetTokens)
		}

		err := maintenanceWithDataStore(tc.decommissioningCleanupFlag, tc.tenant, tc.dryRunFlag, ds)
		assert.NoError(t, err)
	}
}

func TestPropagateStatusesInventory(t *testing.T) {
	statuses := []string{"accepted", "pending", "preauthorized", "noauth"}
	cases := map[string]struct {
		devices       []model.Device
		tenantIds     []string
		forcedVersion string

		cmdTenant string
		cmdDryRun bool

		errDbTenants error
		errDbDevices error
		setStatus    error

		err error
	}{
		"ok, default db, no tenant": {
			tenantIds: []string{
				oid.NewUUIDv4().String(),
				oid.NewUUIDv4().String(),
				oid.NewUUIDv4().String(),
			},
			devices: []model.Device{
				{
					Id:              oid.NewUUIDv4().String(),
					IdData:          "somedata",
					IdDataStruct:    map[string]interface{}{"key0": "value0", "key1": "value0"},
					IdDataSha256:    []byte("some"),
					Status:          "",
					Decommissioning: false,
					CreatedTs:       time.Now(),
					UpdatedTs:       time.Now(),
				},
				{
					Id:              oid.NewUUIDv4().String(),
					IdData:          "somedata",
					IdDataStruct:    map[string]interface{}{"key0": "value0", "key1": "value0"},
					IdDataSha256:    []byte("some"),
					Status:          "",
					Decommissioning: false,
					CreatedTs:       time.Now(),
					UpdatedTs:       time.Now(),
				},
			},
		},
		"ok, with forced version": {
			tenantIds: []string{
				oid.NewUUIDv4().String(),
				oid.NewUUIDv4().String(),
				oid.NewUUIDv4().String(),
			},
			devices: []model.Device{
				{
					Id:              oid.NewUUIDv4().String(),
					IdData:          "somedata",
					IdDataStruct:    map[string]interface{}{"key0": "value0", "key1": "value0"},
					IdDataSha256:    []byte("some"),
					Status:          "",
					Decommissioning: false,
					CreatedTs:       time.Now(),
					UpdatedTs:       time.Now(),
				},
				{
					Id:              oid.NewUUIDv4().String(),
					IdData:          "somedata",
					IdDataStruct:    map[string]interface{}{"key0": "value0", "key1": "value0"},
					IdDataSha256:    []byte("some"),
					Status:          "",
					Decommissioning: false,
					CreatedTs:       time.Now(),
					UpdatedTs:       time.Now(),
				},
			},
			cmdTenant:     "tenant1",
			forcedVersion: "1.7.1",
		},
		"error, with bad forced version": {
			tenantIds: []string{
				oid.NewUUIDv4().String(),
				oid.NewUUIDv4().String(),
				oid.NewUUIDv4().String(),
			},
			devices: []model.Device{
				{
					Id:              oid.NewUUIDv4().String(),
					IdData:          "somedata",
					IdDataStruct:    map[string]interface{}{"key0": "value0", "key1": "value0"},
					IdDataSha256:    []byte("some"),
					Status:          "",
					Decommissioning: false,
					CreatedTs:       time.Now(),
					UpdatedTs:       time.Now(),
				},
				{
					Id:              oid.NewUUIDv4().String(),
					IdData:          "somedata",
					IdDataStruct:    map[string]interface{}{"key0": "value0", "key1": "value0"},
					IdDataSha256:    []byte("some"),
					Status:          "",
					Decommissioning: false,
					CreatedTs:       time.Now(),
					UpdatedTs:       time.Now(),
				},
			},
			cmdTenant:     "tenant1",
			forcedVersion: "and what this version might be",
			err:           errors.New("failed to parse Version: expected integer"),
		},
		"error: store get tenant dbs, abort": {
			tenantIds: []string{
				oid.NewUUIDv4().String(),
				oid.NewUUIDv4().String(),
				oid.NewUUIDv4().String(),
			},
			devices: []model.Device{
				{
					Id:              oid.NewUUIDv4().String(),
					IdData:          "somedata",
					IdDataStruct:    map[string]interface{}{"key0": "value0", "key1": "value0"},
					IdDataSha256:    []byte("some"),
					Status:          "",
					Decommissioning: false,
					CreatedTs:       time.Now(),
					UpdatedTs:       time.Now(),
				},
				{
					Id:              oid.NewUUIDv4().String(),
					IdData:          "somedata",
					IdDataStruct:    map[string]interface{}{"key0": "value0", "key1": "value0"},
					IdDataSha256:    []byte("some"),
					Status:          "",
					Decommissioning: false,
					CreatedTs:       time.Now(),
					UpdatedTs:       time.Now(),
				},
			},

			errDbTenants: errors.New("db failure"),
			err:          errors.New("cant list tenants: db failure"),
		},
		"error: store get devices, report but don't abort": {
			tenantIds: []string{
				oid.NewUUIDv4().String(),
				oid.NewUUIDv4().String(),
				oid.NewUUIDv4().String(),
			},
			devices: []model.Device{
				{
					Id:              oid.NewUUIDv4().String(),
					IdData:          "somedata",
					IdDataStruct:    map[string]interface{}{"key0": "value0", "key1": "value0"},
					IdDataSha256:    []byte("some"),
					Status:          "",
					Decommissioning: false,
					CreatedTs:       time.Now(),
					UpdatedTs:       time.Now(),
				},
				{
					Id:              oid.NewUUIDv4().String(),
					IdData:          "somedata",
					IdDataStruct:    map[string]interface{}{"key0": "value0", "key1": "value0"},
					IdDataSha256:    []byte("some"),
					Status:          "",
					Decommissioning: false,
					CreatedTs:       time.Now(),
					UpdatedTs:       time.Now(),
				},
			},
			errDbDevices: errors.New("db failure"),
			err:          errors.New("failed to get devices: db failure"),
		},
		"error: patch devices, report but don't abort": {
			tenantIds: []string{
				oid.NewUUIDv4().String(),
				oid.NewUUIDv4().String(),
				oid.NewUUIDv4().String(),
			},
			devices: []model.Device{
				{
					Id:              oid.NewUUIDv4().String(),
					IdData:          "somedata",
					IdDataStruct:    map[string]interface{}{"key0": "value0", "key1": "value0"},
					IdDataSha256:    []byte("some"),
					Status:          "",
					Decommissioning: false,
					CreatedTs:       time.Now(),
					UpdatedTs:       time.Now(),
				},
				{
					Id:              oid.NewUUIDv4().String(),
					IdData:          "somedata",
					IdDataStruct:    map[string]interface{}{"key0": "value0", "key1": "value0"},
					IdDataSha256:    []byte("some"),
					Status:          "",
					Decommissioning: false,
					CreatedTs:       time.Now(),
					UpdatedTs:       time.Now(),
				},
			},

			setStatus: errors.New("service failure"),
			err:       errors.New("service failure"),
		},
	}

	for k := range cases {
		tc := cases[k]
		t.Run(fmt.Sprintf("tc %s", k), func(t *testing.T) {
			ctxMatcher := mock.Anything
			status := statuses[rand.Intn(len(statuses))]
			if len(tc.tenantIds) > 0 {
				for i := range tc.devices {
					tc.devices[i].Status = status
				}
			}

			db := &mstore.DataStore{}
			v, _ := migrate.NewVersion(tc.forcedVersion)
			db.On("StoreMigrationVersion",
				mock.Anything,
				v).Return(nil)
			db.On("ListTenantsIds",
				mock.MatchedBy(func(context.Context) bool { return true })).
				Return(tc.tenantIds, tc.errDbTenants)

			for j := range model.DevStatuses {
				db.On("GetDevices",
					ctxMatcher,
					uint(0),
					uint(512),
					model.DeviceFilter{Status: []string{model.DevStatuses[j]}},
				).Return(
					tc.devices,
					tc.errDbDevices,
				)
			}

			c := &minv.Client{}

			if tc.cmdDryRun == false {
				devices := make([]model.DeviceInventoryUpdate, len(tc.devices))
				for n, d := range tc.devices {
					devices[n].Id = d.Id
					for _, status := range model.DevStatuses {
						c.On("SetDeviceStatus",
							mock.Anything,
							mock.AnythingOfType("string"),
							devices,
							status).Return(tc.setStatus)
					}
				}
			}

			if tc.cmdDryRun == true {
				c.AssertNotCalled(t, "SetDeviceStatus")
			}

			err := PropagateStatusesInventory(db, c, tc.cmdTenant, tc.forcedVersion, tc.cmdDryRun)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPropagateReporting(t *testing.T) {
	cases := map[string]struct {
		tenantIds []string
		devices   []model.Device

		cmdTenant string
		cmdDryRun bool

		errDbTenants  error
		errDbDevices  error
		workflowError error

		err error
	}{
		"ok": {
			tenantIds: []string{
				oid.NewUUIDv4().String(),
				oid.NewUUIDv4().String(),
				oid.NewUUIDv4().String(),
			},
			devices: []model.Device{
				{
					Id:              oid.NewUUIDv4().String(),
					IdData:          "somedata",
					IdDataStruct:    map[string]interface{}{"key0": "value0", "key1": "value0"},
					IdDataSha256:    []byte("some"),
					Status:          "accepted",
					Decommissioning: false,
					CreatedTs:       time.Now(),
					UpdatedTs:       time.Now(),
				},
				{
					Id:              oid.NewUUIDv4().String(),
					IdData:          "somedata",
					IdDataStruct:    map[string]interface{}{"key0": "value0", "key1": "value0"},
					IdDataSha256:    []byte("some"),
					Status:          "accepted",
					Decommissioning: false,
					CreatedTs:       time.Now(),
					UpdatedTs:       time.Now(),
				},
			},
		},
		"ok, dry run": {
			tenantIds: []string{
				oid.NewUUIDv4().String(),
				oid.NewUUIDv4().String(),
				oid.NewUUIDv4().String(),
			},
			devices: []model.Device{
				{
					Id:              oid.NewUUIDv4().String(),
					IdData:          "somedata",
					IdDataStruct:    map[string]interface{}{"key0": "value0", "key1": "value0"},
					IdDataSha256:    []byte("some"),
					Status:          "accepted",
					Decommissioning: false,
					CreatedTs:       time.Now(),
					UpdatedTs:       time.Now(),
				},
			},
			cmdDryRun: true,
		},
		"error: workflow": {
			tenantIds: []string{
				oid.NewUUIDv4().String(),
				oid.NewUUIDv4().String(),
				oid.NewUUIDv4().String(),
			},
			devices: []model.Device{
				{
					Id:              oid.NewUUIDv4().String(),
					IdData:          "somedata",
					IdDataStruct:    map[string]interface{}{"key0": "value0", "key1": "value0"},
					IdDataSha256:    []byte("some"),
					Status:          "accepted",
					Decommissioning: false,
					CreatedTs:       time.Now(),
					UpdatedTs:       time.Now(),
				},
			},
			workflowError: errors.New("service failure"),
			err:           errors.New("service failure"),
		},
		"error: db failure: devices": {
			tenantIds: []string{
				oid.NewUUIDv4().String(),
				oid.NewUUIDv4().String(),
				oid.NewUUIDv4().String(),
			},
			devices: []model.Device{
				{
					Id:              oid.NewUUIDv4().String(),
					IdData:          "somedata",
					IdDataStruct:    map[string]interface{}{"key0": "value0", "key1": "value0"},
					IdDataSha256:    []byte("some"),
					Status:          "accepted",
					Decommissioning: false,
					CreatedTs:       time.Now(),
					UpdatedTs:       time.Now(),
				},
			},
			errDbDevices: errors.New("db failure"),
			err:          errors.New("failed to get devices: db failure"),
		},
	}

	for k := range cases {
		tc := cases[k]
		t.Run(fmt.Sprintf("tc %s", k), func(t *testing.T) {
			db := &mstore.DataStore{}
			ctxMatcher := mock.MatchedBy(func(c context.Context) bool { return true })

			db.On("GetDevices",
				ctxMatcher,
				uint(0),
				uint(512),
				model.DeviceFilter{},
			).Return(
				tc.devices,
				tc.errDbDevices,
			)
			var (
				actualErr error
			)
			errChannel := make(chan error, 1)
			db.On("ForEachTenant",
				ctxMatcher,
				mock.MatchedBy(func(f store.MapFunc) bool {
					return true
				}),
			).Run(func(args mock.Arguments) {
				go func() {
					// A simplified version of what
					// mongo.ForEachTenant does
					for _, tenant := range tc.tenantIds {
						t.Logf("dbg.%s: here-100", time.Now().Format(time.RFC3339Nano))
						ctx := identity.WithContext(
							args.Get(0).(context.Context),
							&identity.Identity{
								Tenant: tenant,
							},
						)
						mapFun := args.Get(1).(store.MapFunc)
						actualErr = mapFun(ctx)
						if actualErr != nil {
							t.Logf("dbg.%s: here0: actual->channel: %+v", time.Now().Format(time.RFC3339Nano), actualErr)
							errChannel <- actualErr
							break
						}
					}
					var noError error
					errChannel <- noError
				}()
			}).Return(actualErr).Once() // there is no way actualErr will have the value set inside Run, this one is executed first, and has always the value of before the run

			var err error
			wflows := &mwflows.ClientRunner{}
			defer wflows.AssertExpectations(t)

			if !tc.cmdDryRun && tc.errDbTenants == nil && tc.errDbDevices == nil {
				devicesIDs := make([]string, len(tc.devices))
				for i, dev := range tc.devices {
					devicesIDs[i] = dev.Id
				}

				wflows.On("SubmitReindexReportingBatch",
					mock.MatchedBy(func(c context.Context) bool {
						return true
					}),
					devicesIDs,
				).Return(tc.workflowError)
			}

			doneChannel := make(chan bool, 1)
			go func() {
				err = PropagateReporting(db, wflows, tc.cmdTenant, time.Microsecond, tc.cmdDryRun)
				t.Logf("dbg.%s: here0010: about to read from channel", time.Now().Format(time.RFC3339Nano))
				err = <-errChannel
				t.Logf("dbg.%s: here0010: read from channel: %+v", time.Now().Format(time.RFC3339Nano), err)
				if tc.err != nil {
					assert.EqualError(t, err, tc.err.Error())
				} else {
					assert.NoError(t, err)
				}
				doneChannel <- true
			}()
			t.Logf("dbg.%s final wait", time.Now().Format(time.RFC3339Nano))
			ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()
			select {
			case <-ctxTimeout.Done():
			case <-doneChannel:
			}
			t.Logf("dbg.%s exiting", time.Now().Format(time.RFC3339Nano))
		})
	}
}
