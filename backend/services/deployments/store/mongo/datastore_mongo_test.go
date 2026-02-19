// Copyright 2024 Northern.tech AS
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
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"
	ctxstore "github.com/mendersoftware/mender-server/pkg/store"

	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/services/deployments/store"
)

func init() {
	log.Log.Out = io.Discard
}

func TestPing(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestPing in short mode.")
	}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
	defer cancel()
	ds := NewDataStoreMongoWithClient(db.Client())
	err := ds.Ping(ctx)
	assert.NoError(t, err)
}

func timePtr(timeStr string) *time.Time {
	t, _ := time.Parse(time.RFC3339, timeStr)
	t = t.UTC()
	return &t
}

func TestInsertUploadIntent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestInsertUploadIntent in short mode.")
	}
	db.Wipe()

	ctx := context.Background()
	mgoClient := db.Client()
	ds := NewDataStoreMongoWithClient(mgoClient)
	link := &model.UploadLink{
		Link: model.Link{
			Expire:   time.Now().Add(time.Minute).Round(time.Second),
			TenantID: "123456789012345678901234",
		},
		ArtifactID: uuid.New().String(),
	}

	err := ds.InsertUploadIntent(ctx, link)
	if assert.NoError(t, err) {
		// verify that record exist in database
		var actual model.UploadLink
		err := mgoClient.
			Database(DatabaseName).
			Collection(CollectionUploadIntents).
			FindOne(ctx, bson.D{}).
			Decode(&actual)
		if assert.NoError(t, err, "error retrieving document from db") {
			assert.Equal(t, link.ArtifactID, actual.ArtifactID)
		}
	}
}

func TestFindUploadLinks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestFindUploadLinks in short mode.")
	}
	db.Wipe()

	ctx := context.Background()
	mgoClient := db.Client()
	ds := NewDataStoreMongoWithClient(mgoClient)
	links := []model.UploadLink{{
		Link: model.Link{
			Expire:   time.Now().Add(-time.Minute).Round(time.Second),
			TenantID: "123456789012345678901234",
		},
		Status:     model.LinkStatusCompleted,
		ArtifactID: uuid.New().String(),
	}, {
		Link: model.Link{
			Expire: time.Now().Add(-time.Minute * 10).Round(time.Second),
		},
		Status:     model.LinkStatusPending,
		ArtifactID: uuid.New().String(),
	}, {
		Link: model.Link{
			Expire:   time.Now().Add(-time.Minute * 15).Round(time.Second),
			TenantID: "123456789012345678901234",
		},
		ArtifactID: uuid.New().String(),
	}, {
		Link: model.Link{
			Expire:   time.Now().Add(-time.Minute * 20).Round(time.Second),
			TenantID: "123456789012345678901234",
		},
		Status:     model.LinkStatusAborted,
		ArtifactID: uuid.New().String(),
	}, {
		Link: model.Link{
			Expire:   time.Now().Add(-time.Minute * 25).Round(time.Second),
			TenantID: "123456789012345678901234",
		},
		Status:     model.LinkStatusCompleted | model.LinkStatusProcessedBit,
		ArtifactID: uuid.Nil.String(),
	}, {
		Link: model.Link{
			Expire:   time.Now().Add(-time.Minute * 30).Round(time.Second),
			TenantID: "123456789012345678901234",
		},
		Status:     model.LinkStatusProcessing,
		ArtifactID: uuid.New().String(),
	}}
	ins := make([]interface{}, len(links))
	for i := range links {
		ins[i] = links[i]
	}
	_, err := mgoClient.Database(DatabaseName).
		Collection(CollectionUploadIntents).
		InsertMany(ctx, ins)
	if err != nil {
		panic(err)
	}

	it, err := ds.FindUploadLinks(ctx, time.Now())
	if assert.NoError(t, err) {
		var (
			i    int
			more bool
		)
		for more, err = it.Next(ctx); more && i < len(links); more, err = it.Next(ctx) {
			for links[i].Status&model.LinkStatusProcessedBit > 0 {
				i++
			}
			var link model.UploadLink
			err = it.Decode(&link)
			if err != nil {
				break
			}
			assert.Equal(t, links[i].ArtifactID, link.ArtifactID)

			i++
		}

		assert.NoError(t, err)
	}
}

func TestUpdateUploadIntentStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestInsertUploadIntent in short mode.")
	}
	db.Wipe()

	const (
		artifactIDOS = "00000000-0000-0000-0000-000000000000"
		artifactIDMT = "00000000-0000-0000-0000-000000000001"
		tenantID     = "123456789012345678901234"
	)

	ctx := context.Background()
	mgoClient := db.Client()
	ds := NewDataStoreMongoWithClient(mgoClient)
	mgoClient.Database(DatabaseName).
		Collection(CollectionUploadIntents).
		InsertMany(ctx, []interface{}{model.UploadLink{
			ArtifactID: artifactIDOS,
			Link: model.Link{
				Uri:    "http://localhost:8080",
				Method: "PUT",
			},
			IssuedAt: time.Now().Add(-time.Minute),
			Status:   model.LinkStatusPending,
		}, model.UploadLink{
			ArtifactID: artifactIDMT,
			Link: model.Link{
				Uri:      "http://localhost:8080",
				Method:   "PUT",
				TenantID: tenantID,
			},
			IssuedAt: time.Now().Add(-time.Second * 30),
			Status:   model.LinkStatusProcessing,
		}})

	t.Run("ok", func(t *testing.T) {
		err := ds.UpdateUploadIntentStatus(
			ctx,
			artifactIDOS,
			model.LinkStatusPending,
			model.LinkStatusProcessing,
		)
		assert.NoError(t, err)
	})
	t.Run("ok/multi-tenant", func(t *testing.T) {
		ctx = identity.WithContext(ctx, &identity.Identity{
			Tenant: tenantID,
		})
		err := ds.UpdateUploadIntentStatus(
			ctx,
			artifactIDMT,
			model.LinkStatusProcessing,
			model.LinkStatusCompleted,
		)
		assert.NoError(t, err)
	})
	t.Run("error/not found status", func(t *testing.T) {
		ctx = identity.WithContext(ctx, &identity.Identity{
			Tenant: tenantID,
		})
		err := ds.UpdateUploadIntentStatus(
			ctx,
			artifactIDMT,
			model.LinkStatusPending,
			model.LinkStatusProcessing,
		)
		assert.ErrorIs(t, err, store.ErrNotFound)
	})
	t.Run("error/not found id", func(t *testing.T) {
		ctx = identity.WithContext(ctx, &identity.Identity{
			Tenant: tenantID,
		})
		err := ds.UpdateUploadIntentStatus(
			ctx,
			"4a54ab54-05b9-4aaa-bfd9-162703ea3232",
			model.LinkStatusPending,
			model.LinkStatusProcessing,
		)
		assert.ErrorIs(t, err, store.ErrNotFound)
	})
	t.Run("error/context cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		cancel()
		err := ds.UpdateUploadIntentStatus(
			ctx,
			artifactIDOS,
			model.LinkStatusProcessing,
			model.LinkStatusAborted,
		)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

func TestFindNewerActiveDeployments(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestFindNewerActiveDeployments in short mode.")
	}
	now := time.Now()

	testCases := map[string]struct {
		InputDeploymentsCollection []interface{}
		InputTenant                string
		InputCreatedAfter          *time.Time
		InputSkip                  int
		InputLimit                 int

		OutputError       error
		OutputDeployments []*model.Deployment
	}{
		"empty database": {
			InputCreatedAfter: &now,
			InputSkip:         0,
			InputLimit:        1,

			OutputError:       nil,
			OutputDeployments: nil,
		},
		"no newer deployments": {
			InputDeploymentsCollection: []interface{}{
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id:      "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
					Created: &now,
				},
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id:      "d1804903-5caa-4a73-a3ae-0efcc3205405",
					Created: &now,
				},
			},
			InputSkip:         0,
			InputLimit:        1,
			InputCreatedAfter: TimePtr(now.Add(time.Hour * 24)),

			OutputError:       nil,
			OutputDeployments: nil,
		},
		"one newer deployments": {
			InputDeploymentsCollection: []interface{}{
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id:      "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
					Created: &now,
				},
			},
			InputSkip:         0,
			InputLimit:        5,
			InputCreatedAfter: TimePtr(now.Add(-time.Hour * 24)),

			OutputError: nil,
			OutputDeployments: []*model.Deployment{
				{
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
					},
					Id:     "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
					Active: true,
				},
			},
		},
		"one older deployment and one newer deployments": {
			InputDeploymentsCollection: []interface{}{
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id:      "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
					Created: TimePtr(now.Add(-time.Hour * 24)),
				},
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id:      "d1804903-5caa-4a73-a3ae-0efcc3205405",
					Active:  true,
					Created: TimePtr(now.Add(time.Hour * 24)),
				},
			},
			InputSkip:         0,
			InputLimit:        5,
			InputCreatedAfter: &now,

			OutputError: nil,
			OutputDeployments: []*model.Deployment{
				{
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
					},
					Id:     "d1804903-5caa-4a73-a3ae-0efcc3205405",
					Active: true,
				},
			},
		},
	}

	for testCaseName, testCase := range testCases {
		t.Run(testCaseName, func(t *testing.T) {

			// Make sure we start test with empty database
			db.Wipe()

			client := db.Client()
			store := NewDataStoreMongoWithClient(client)

			ctx := context.Background()
			if testCase.InputTenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: testCase.InputTenant,
				})
			} else {
				ctx = context.Background()
			}

			collDep := client.Database(ctxstore.
				DbFromContext(ctx, DatabaseName)).
				Collection(CollectionDeployments)

			if testCase.InputDeploymentsCollection != nil {
				_, err := collDep.InsertMany(
					ctx, testCase.InputDeploymentsCollection)
				assert.NoError(t, err)
			}

			deployments, err := store.FindNewerActiveDeployments(ctx,
				testCase.InputCreatedAfter, testCase.InputSkip, testCase.InputLimit)

			for i := range deployments {
				deployments[i].Created = nil
			}

			if testCase.OutputError != nil {
				assert.EqualError(t, err, testCase.OutputError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.OutputDeployments, deployments)
			}
		})
	}
}

func TestInsertDeploymentConflict(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestInsertDeploymentConflict in short mode.")
	}
	testCases := map[string]struct {
		constructor1 *model.DeploymentConstructor
		constructor2 *model.DeploymentConstructor
		not_active   bool
		err          error
	}{
		"ok": {
			constructor1: &model.DeploymentConstructor{
				Name:         "name",
				ArtifactName: "artifact",
				Devices:      []string{"device-1"},
			},
			constructor2: &model.DeploymentConstructor{
				Name:         "name",
				ArtifactName: "artifact",
				Devices:      []string{"device-2"},
			},
		},
		"ko, conflict": {
			constructor1: &model.DeploymentConstructor{
				Name:         "name",
				ArtifactName: "artifact",
				Devices:      []string{"device-1"},
			},
			constructor2: &model.DeploymentConstructor{
				Name:         "name",
				ArtifactName: "artifact",
				Devices:      []string{"device-1"},
			},
			err: ErrConflictingDeployment,
		},
		"ok, conflict but deployment 1 is not active, so no error": {
			constructor1: &model.DeploymentConstructor{
				Name:         "name",
				ArtifactName: "artifact",
				Devices:      []string{"device-1"},
			},
			constructor2: &model.DeploymentConstructor{
				Name:         "name",
				ArtifactName: "artifact",
				Devices:      []string{"device-1"},
			},
			not_active: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			client := db.Client()
			db.Wipe()

			ctx := context.Background()
			ds := NewDataStoreMongoWithClient(client)

			err := MigrateSingle(ctx, DbName, DbVersion, client, true)
			assert.Nil(t, err)

			deployment1, _ := model.NewDeploymentFromConstructor(tc.constructor1)
			if tc.not_active {
				deployment1.Status = model.DeploymentStatusFinished
			}
			err = ds.InsertDeployment(ctx, deployment1)
			assert.Nil(t, err)

			deployment2, _ := model.NewDeploymentFromConstructor(tc.constructor2)
			err = ds.InsertDeployment(ctx, deployment2)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestSetDeploymentDeviceCount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestSetDeploymentDeviceCount in short mode.")
	}

	zero := 0
	one := 1
	now := time.Now()

	testCases := map[string]struct {
		deployment *model.Deployment
		count      int
		expected   int
	}{
		"device_count doesn't exist": {
			deployment: &model.Deployment{
				Id:      "d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
				Created: &now,
				DeploymentConstructor: &model.DeploymentConstructor{
					Name:         "name",
					ArtifactName: "artifact",
					Devices:      []string{"device-1"},
				},
			},
			count:    10,
			expected: 10,
		},
		"device_count is zero": {
			deployment: &model.Deployment{
				Id:          "d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
				DeviceCount: &zero,
				Created:     &now,
				DeploymentConstructor: &model.DeploymentConstructor{
					Name:         "name",
					ArtifactName: "artifact",
					Devices:      []string{"device-1"},
				},
			},
			count:    10,
			expected: zero,
		},
		"device_count is one": {
			deployment: &model.Deployment{
				Id:          "d50eda0d-2cea-4de1-8d42-9cd3e7e86703",
				DeviceCount: &one,
				Created:     &now,
				DeploymentConstructor: &model.DeploymentConstructor{
					Name:         "name",
					ArtifactName: "artifact",
					Devices:      []string{"device-1"},
				},
			},
			count:    10,
			expected: one,
		},
	}

	ctx := context.Background()
	ds := NewDataStoreMongoWithClient(db.Client())

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := ds.InsertDeployment(ctx, tc.deployment)
			assert.Nil(t, err)

			err = ds.SetDeploymentDeviceCount(ctx, tc.deployment.Id, tc.count)
			assert.Nil(t, err)

			deployment, err := ds.FindDeploymentByID(ctx, tc.deployment.Id)
			assert.Nil(t, err)
			assert.NotNil(t, deployment)
			assert.NotNil(t, deployment.DeviceCount)
			assert.Equal(t, *deployment.DeviceCount, tc.expected)
		})
	}

}

func TestSetStorageSettings(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestSetStorageSettings in short mode.")
	}

	testCases := map[string]struct {
		tenantID string
		settings *model.StorageSettings
		err      error
	}{
		"ok": {
			settings: &model.StorageSettings{
				Region: "region",
				Key:    "secretkey",
				Secret: "secret",
				Bucket: "bucket",
				Uri:    "https://example.com",
				Token:  "token",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			ds := NewDataStoreMongoWithClient(db.Client())

			err := ds.SetStorageSettings(ctx, tc.settings)
			assert.NoError(t, err)

			settings, err := ds.GetStorageSettings(ctx)
			assert.NoError(t, err)
			assert.Equal(t, tc.settings, settings)
		})
	}
}

func TestSortDeployments(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestSortDeployments in short mode.")
	}

	// Make sure we start test with empty database
	db.Wipe()

	uuidVal, _ := uuid.NewRandom()
	deploymentOneID := uuidVal.String()
	uuidVal, _ = uuid.NewRandom()
	deploymentTwoID := uuidVal.String()
	now := time.Now()
	startDate := now.AddDate(0, -1, 0)
	deviceCount := 1
	uuidVal, _ = uuid.NewRandom()
	devicesList := []string{uuidVal.String()}
	config := make([]byte, 0)
	inputDeployments := []*model.Deployment{
		{
			DeploymentConstructor: &model.DeploymentConstructor{
				Name:         "deployment 1",
				ArtifactName: "artifact 1",
			},
			Created:       &now,
			Id:            deploymentOneID,
			DeviceCount:   &deviceCount,
			MaxDevices:    1,
			DeviceList:    devicesList,
			Status:        model.DeploymentStatusInProgress,
			Type:          model.DeploymentTypeConfiguration,
			Configuration: config,
		},
		{
			DeploymentConstructor: &model.DeploymentConstructor{
				Name:         "deployment 2",
				ArtifactName: "artifact 2",
			},
			Created:       &startDate,
			Id:            deploymentTwoID,
			DeviceCount:   &deviceCount,
			MaxDevices:    1,
			DeviceList:    devicesList,
			Status:        model.DeploymentStatusFinished,
			Type:          model.DeploymentTypeConfiguration,
			Configuration: config,
		},
	}

	ctx := context.Background()
	ds := NewDataStoreMongoWithClient(db.Client())
	var deploymentsQty int64 = 2

	for _, depl := range inputDeployments {
		err := ds.InsertDeployment(ctx, depl)
		assert.NoError(t, err)
	}

	query := model.Query{
		Sort: model.SortDirectionDescending,
	}
	deployments, count, err := ds.FindDeployments(ctx, query)
	assert.NoError(t, err)
	assert.NotEmpty(t, deployments)
	assert.Equal(t, deploymentsQty, count)
	assert.Equal(t, deploymentOneID, deployments[0].Id)

	query = model.Query{
		Sort: model.SortDirectionAscending,
	}
	deployments, count, err = ds.FindDeployments(ctx, query)
	assert.NoError(t, err)
	assert.NotEmpty(t, deployments)
	assert.Equal(t, deploymentsQty, count)
	assert.Equal(t, deploymentTwoID, deployments[0].Id)
}

func TestFindOldestActiveDeviceDeployment(t *testing.T) {
	db.Wipe()
	const (
		TenantID       = "123456789012345678901234"
		TenantDeviceID = "27d5d258-b880-4157-8eb7-8d68aeb1663d"
		DeviceID       = "1140bc78-b898-4b2a-a4a2-551cb7bd9ac8"
	)
	// Initialize dataset
	ds := NewDataStoreMongoWithClient(db.Client())
	now := time.Now()
	for _, env := range []struct {
		tenantID string
		deviceID string
	}{
		{tenantID: TenantID, deviceID: TenantDeviceID},
		{deviceID: DeviceID},
	} {
		ctx := context.Background()
		if env.tenantID != "" {
			ctx = identity.WithContext(ctx, &identity.Identity{
				Tenant: env.tenantID,
			})
		}
		if err := ds.ProvisionTenant(ctx, env.tenantID); err != nil {
			panic(err)
		}
		for _, depl := range []*model.DeviceDeployment{{
			Id: "0",
			Created: func() *time.Time {
				ret := now.Add(-2 * time.Hour)
				return &ret
			}(),
			Status:       model.DeviceDeploymentStatusSuccess,
			DeviceId:     env.deviceID,
			DeploymentId: uuid.New().String(),
		}, {
			Id: "1",
			Created: func() *time.Time {
				ret := now.Add(-3 * time.Hour / 2)
				return &ret
			}(),
			Status:       model.DeviceDeploymentStatusAlreadyInst,
			DeviceId:     env.deviceID,
			DeploymentId: uuid.New().String(),
		}, {
			Id: "2",
			Created: func() *time.Time {
				ret := now.Add(-time.Hour)
				return &ret
			}(),
			Status:       model.DeviceDeploymentStatusDownloading,
			DeviceId:     env.deviceID,
			DeploymentId: uuid.New().String(),
			Active:       true,
		}, {
			Id: "3",
			Created: func() *time.Time {
				ret := now.Add(-time.Minute)
				return &ret
			}(),
			Status:       model.DeviceDeploymentStatusPending,
			DeviceId:     env.deviceID,
			DeploymentId: uuid.New().String(),
			Active:       true,
		}} {
			if err := ds.InsertDeviceDeployment(ctx, depl, true); err != nil {
				panic(err)
			}
		}
	}

	testCases := []struct {
		Name string

		CTX      context.Context
		TenantID string
		DeviceID string

		ExpectedID *string
		Error      error
	}{{
		Name: "ok",

		CTX:      context.Background(),
		DeviceID: DeviceID,

		ExpectedID: func() *string { s := "2"; return &s }(),
	}, {
		Name: "ok, multi-tenant mode",

		CTX: identity.WithContext(context.Background(), &identity.Identity{
			Tenant: TenantID,
		}),
		DeviceID: TenantDeviceID,

		ExpectedID: func() *string { s := "2"; return &s }(),
	}, {
		Name: "ok, no document",

		CTX: identity.WithContext(context.Background(), &identity.Identity{
			Tenant: TenantID,
		}),
		DeviceID: DeviceID, // NOTE: We're using the tenant database
	}, {
		Name: "error, context canceled",

		CTX: func() context.Context {
			ctx, ccl := context.WithCancel(context.Background())
			ccl()
			return ctx
		}(),
		DeviceID: DeviceID, // NOTE: We're using the tenant database

		Error: context.Canceled,
	}, {
		Name:  "error, empty device id",
		Error: ErrStorageInvalidID,
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			d, err := ds.FindOldestActiveDeviceDeployment(tc.CTX, tc.DeviceID)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				if assert.NoError(t, err) {
					if tc.ExpectedID != nil && assert.NotNil(t, d) {
						assert.Equal(t, *tc.ExpectedID, d.Id,
							"did not receive the expected device deployment id",
						)
					} else {
						assert.Nil(t, d, "did not expect to find a deployment")
					}
				}
			}
		})
	}
}

func TestFindLatestInactiveDeviceDeployment(t *testing.T) {
	db.Wipe()
	const (
		TenantID       = "123456789012345678901234"
		TenantDeviceID = "27d5d258-b880-4157-8eb7-8d68aeb1663d"
		DeviceID       = "1140bc78-b898-4b2a-a4a2-551cb7bd9ac8"
	)
	// Initialize dataset
	ds := NewDataStoreMongoWithClient(db.Client())
	now := time.Now()
	for _, env := range []struct {
		tenantID string
		deviceID string
	}{
		{tenantID: TenantID, deviceID: TenantDeviceID},
		{deviceID: DeviceID},
	} {
		ctx := context.Background()
		if env.tenantID != "" {
			ctx = identity.WithContext(ctx, &identity.Identity{
				Tenant: env.tenantID,
			})
		}
		if err := ds.ProvisionTenant(ctx, env.tenantID); err != nil {
			panic(err)
		}
		for _, depl := range []*model.DeviceDeployment{{
			Id: "0",
			Created: func() *time.Time {
				ret := now.Add(-2 * time.Hour)
				return &ret
			}(),
			Status:       model.DeviceDeploymentStatusSuccess,
			DeviceId:     env.deviceID,
			DeploymentId: uuid.New().String(),
		}, {
			Id: "1",
			Created: func() *time.Time {
				ret := now.Add(-3 * time.Hour / 2)
				return &ret
			}(),
			Status:       model.DeviceDeploymentStatusAlreadyInst,
			DeviceId:     env.deviceID,
			DeploymentId: uuid.New().String(),
		}, {
			Id: "2",
			Created: func() *time.Time {
				ret := now.Add(-time.Hour)
				return &ret
			}(),
			Status:       model.DeviceDeploymentStatusDownloading,
			DeviceId:     env.deviceID,
			DeploymentId: uuid.New().String(),
			Active:       true,
		}, {
			Id: "3",
			Created: func() *time.Time {
				ret := now.Add(-time.Minute)
				return &ret
			}(),
			Status:       model.DeviceDeploymentStatusPending,
			DeviceId:     env.deviceID,
			DeploymentId: uuid.New().String(),
			Active:       true,
		}, {
			Id: "4",
			Created: func() *time.Time {
				ret := now.Add(-time.Second)
				return &ret
			}(),
			Status:       model.DeviceDeploymentStatusPending,
			DeviceId:     env.deviceID,
			DeploymentId: uuid.New().String(),
			Deleted: func() *time.Time {
				return &now
			}(),
		}} {
			if err := ds.InsertDeviceDeployment(ctx, depl, true); err != nil {
				panic(err)
			}
		}
	}

	testCases := []struct {
		Name string

		CTX      context.Context
		TenantID string
		DeviceID string

		ExpectedID *string
		Error      error
	}{{
		Name: "ok",

		CTX:      context.Background(),
		DeviceID: DeviceID,

		ExpectedID: func() *string { s := "1"; return &s }(),
	}, {
		Name: "ok, multi-tenant mode",

		CTX: identity.WithContext(context.Background(), &identity.Identity{
			Tenant: TenantID,
		}),
		DeviceID: TenantDeviceID,

		ExpectedID: func() *string { s := "1"; return &s }(),
	}, {
		Name: "ok, no document",

		CTX: identity.WithContext(context.Background(), &identity.Identity{
			Tenant: TenantID,
		}),
		DeviceID: DeviceID, // NOTE: We're using the tenant database
	}, {
		Name: "error, context canceled",

		CTX: func() context.Context {
			ctx, ccl := context.WithCancel(context.Background())
			ccl()
			return ctx
		}(),
		DeviceID: DeviceID, // NOTE: We're using the tenant database

		Error: context.Canceled,
	}, {
		Name:  "error, empty device id",
		Error: ErrStorageInvalidID,
	}}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			d, err := ds.FindLatestInactiveDeviceDeployment(tc.CTX, tc.DeviceID)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				if assert.NoError(t, err) {
					if tc.ExpectedID != nil && assert.NotNil(t, d) {
						assert.Equal(t, *tc.ExpectedID, d.Id,
							"did not receive the expected device deployment id",
						)
					} else {
						assert.Nil(t, d, "did not expect to find a deployment")
					}
				}
			}
		})
	}
}

func TestFindDeploymentStatsByIDs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestFindDeploymentStatsByIDs in short mode.")
	}

	now := time.Now()

	deployments := []*model.Deployment{
		{
			Id:      "d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
			Created: &now,
			DeploymentConstructor: &model.DeploymentConstructor{
				Name:         "name",
				ArtifactName: "artifact",
				Devices:      []string{"device-1"},
			},
			Stats: model.Stats{},
		},
		{
			Id:      "d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
			Created: &now,
			DeploymentConstructor: &model.DeploymentConstructor{
				Name:         "name",
				ArtifactName: "artifact",
				Devices:      []string{"device-1"},
			},
			Stats: model.NewDeviceDeploymentStats(),
		},
		{
			Id:      "d50eda0d-2cea-4de1-8d42-9cd3e7e86703",
			Created: &now,
			DeploymentConstructor: &model.DeploymentConstructor{
				Name:         "name",
				ArtifactName: "artifact",
				Devices:      []string{"device-1"},
			},
			Stats: model.NewDeviceDeploymentStats(),
		},
	}

	ctx := context.Background()
	ds := NewDataStoreMongoWithClient(db.Client())

	for _, deployment := range deployments {
		assert.NoError(t, ds.InsertDeployment(ctx, deployment))
	}

	testCases := map[string]struct {
		deployments []*model.Deployment
	}{
		"OK - single": {
			deployments: []*model.Deployment{
				deployments[0],
			},
		},
		"OK - multiple": {
			deployments: deployments,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var ids []string
			for _, d := range tc.deployments {
				ids = append(ids, d.Id)
			}
			depStats, err := ds.FindDeploymentStatsByIDs(ctx, ids...)
			assert.Nil(t, err)
			assert.NotNil(t, depStats)
			assert.Equal(t, len(depStats), len(tc.deployments))
		})
	}
}

func str2ptr(s string) *string {
	return &s
}

func TestGetDeviceDeploymentsForDevice(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestGetDeviceDeploymentsForDevice in short mode.")
	}

	now := time.Now()

	ctx := context.Background()
	ds := NewDataStoreMongoWithClient(db.Client())

	const deviceID = "d50eda0d-2cea-4de1-8d42-9cd3e7e86700"
	deviceDeployments := []*model.DeviceDeployment{
		{
			Id: "d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
			Created: func() *time.Time {
				ret := now.Add(3 * time.Hour)
				return &ret
			}(),
			Status:       model.DeviceDeploymentStatusPauseBeforeInstall,
			DeviceId:     deviceID,
			DeploymentId: "d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
		},
		{
			Id: "d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
			Created: func() *time.Time {
				ret := now.Add(2 * time.Hour)
				return &ret
			}(),
			Status:       model.DeviceDeploymentStatusSuccess,
			DeviceId:     deviceID,
			DeploymentId: "d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
		},
		{
			Id: "d50eda0d-2cea-4de1-8d42-9cd3e7e86703",
			Created: func() *time.Time {
				ret := now.Add(1 * time.Hour)
				return &ret
			}(),
			Status:       model.DeviceDeploymentStatusSuccess,
			DeviceId:     deviceID,
			DeploymentId: "d50eda0d-2cea-4de1-8d42-9cd3e7e86703",
		},
	}
	for _, deviceDeployment := range deviceDeployments {
		assert.NoError(t, ds.InsertDeviceDeployment(ctx, deviceDeployment, true))
	}

	testCases := map[string]struct {
		q store.ListQueryDeviceDeployments

		res      []model.DeviceDeployment
		resCount int
		resErr   error
	}{
		"ok": {
			q: store.ListQueryDeviceDeployments{
				DeviceID: deviceID,
				Status:   nil,
				Limit:    10,
				Skip:     0,
			},
			res: []model.DeviceDeployment{
				*deviceDeployments[0],
				*deviceDeployments[1],
				*deviceDeployments[2],
			},
			resCount: 3,
		},
		"ok, IDs": {
			q: store.ListQueryDeviceDeployments{
				IDs:    []string{"d50eda0d-2cea-4de1-8d42-9cd3e7e86701", "d50eda0d-2cea-4de1-8d42-9cd3e7e86702"},
				Status: nil,
				Limit:  10,
				Skip:   0,
			},
			res: []model.DeviceDeployment{
				*deviceDeployments[0],
				*deviceDeployments[1],
			},
			resCount: 2,
		},
		"ok, status pause": {
			q: store.ListQueryDeviceDeployments{
				DeviceID: deviceID,
				Status:   str2ptr(model.DeviceDeploymentStatusPauseStr),
				Limit:    10,
				Skip:     0,
			},
			res: []model.DeviceDeployment{
				*deviceDeployments[0],
			},
			resCount: 1,
		},
		"ok, status finished": {
			q: store.ListQueryDeviceDeployments{
				DeviceID: deviceID,
				Status:   str2ptr(model.DeviceDeploymentStatusFinishedStr),
				Limit:    10,
				Skip:     0,
			},
			res: []model.DeviceDeployment{
				*deviceDeployments[1],
				*deviceDeployments[2],
			},
			resCount: 2,
		},
		"ok, status active": {
			q: store.ListQueryDeviceDeployments{
				DeviceID: deviceID,
				Status:   str2ptr(model.DeviceDeploymentStatusActiveStr),
				Limit:    10,
				Skip:     0,
			},
			res: []model.DeviceDeployment{
				*deviceDeployments[0],
			},
			resCount: 1,
		},
		"ok, status successful": {
			q: store.ListQueryDeviceDeployments{
				DeviceID: deviceID,
				Status:   str2ptr(model.DeviceDeploymentStatusSuccessStr),
				Limit:    10,
				Skip:     0,
			},
			res: []model.DeviceDeployment{
				*deviceDeployments[1],
				*deviceDeployments[2],
			},
			resCount: 2,
		},
		"ok, status active, first page": {
			q: store.ListQueryDeviceDeployments{
				DeviceID: deviceID,
				Status:   str2ptr(model.DeviceDeploymentStatusSuccessStr),
				Limit:    1,
				Skip:     0,
			},
			res: []model.DeviceDeployment{
				*deviceDeployments[1],
			},
			resCount: 2,
		},
		"ok, status active, second page": {
			q: store.ListQueryDeviceDeployments{
				DeviceID: deviceID,
				Status:   str2ptr(model.DeviceDeploymentStatusSuccessStr),
				Limit:    1,
				Skip:     1,
			},
			res: []model.DeviceDeployment{
				*deviceDeployments[2],
			},
			resCount: 2,
		},
		"ok, no results": {
			q: store.ListQueryDeviceDeployments{
				DeviceID: deviceID,
				Status:   str2ptr(model.DeviceDeploymentStatusDownloadingStr),
				Limit:    10,
				Skip:     0,
			},
			res:      []model.DeviceDeployment{},
			resCount: 0,
		},
		"ko, status invalid": {
			q: store.ListQueryDeviceDeployments{
				DeviceID: deviceID,
				Status:   str2ptr("dummy"),
				Limit:    10,
				Skip:     0,
			},
			res:      nil,
			resCount: -1,
			resErr:   errors.New("invalid status query: invalid status for device 'dummy'"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			res, count, err := ds.GetDeviceDeploymentsForDevice(ctx, tc.q)
			assert.Equal(t, tc.resCount, count)

			if tc.resErr != nil {
				assert.EqualError(t, err, tc.resErr.Error())
			} else {
				for i, _ := range res {
					// ignore Created field when comparing the results
					res[i].Created = tc.res[i].Created
					// ignore Started field when comparing the results
					res[i].Started = tc.res[i].Started
				}
				assert.Equal(t, tc.res, res)
				assert.Nil(t, err)
			}
		})
	}
}

func TestGetDeviceDeployments(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestGetDeviceDeployments in short mode.")
	}

	now := time.Now()
	f := false

	ctx := context.Background()
	ds := NewDataStoreMongoWithClient(db.Client())

	const deviceID = "d50eda0d-2cea-4de1-8d42-9cd3e7e86700"
	const differentDeviceID = "d50eda0d-2cea-4de1-8d42-9cd3e7e86701"
	deviceDeployments := []*model.DeviceDeployment{
		{
			Id: "d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
			Created: func() *time.Time {
				ret := now.Add(5 * time.Hour)
				return &ret
			}(),
			Status:       model.DeviceDeploymentStatusPauseBeforeInstall,
			DeviceId:     deviceID,
			DeploymentId: "d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
			Active:       true,
		},
		{
			Id: "d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
			Created: func() *time.Time {
				ret := now.Add(4 * time.Hour)
				return &ret
			}(),
			Status:       model.DeviceDeploymentStatusSuccess,
			DeviceId:     deviceID,
			DeploymentId: "d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
			Deleted:      &now,
		},
		{
			Id: "d50eda0d-2cea-4de1-8d42-9cd3e7e86703",
			Created: func() *time.Time {
				ret := now.Add(3 * time.Hour)
				return &ret
			}(),
			Status:       model.DeviceDeploymentStatusSuccess,
			DeviceId:     deviceID,
			DeploymentId: "d50eda0d-2cea-4de1-8d42-9cd3e7e86703",
		},
		{
			Id: "d50eda0d-2cea-4de1-8d42-9cd3e7e86704",
			Created: func() *time.Time {
				ret := now.Add(2 * time.Hour)
				return &ret
			}(),
			Status:       model.DeviceDeploymentStatusPending,
			DeviceId:     deviceID,
			DeploymentId: "d50eda0d-2cea-4de1-8d42-9cd3e7e86704",
			Active:       true,
		},
		{
			Id: "d50eda0d-2cea-4de1-8d42-9cd3e7e86705",
			Created: func() *time.Time {
				ret := now.Add(1 * time.Hour)
				return &ret
			}(),
			Status:       model.DeviceDeploymentStatusInstalling,
			DeviceId:     differentDeviceID,
			DeploymentId: "d50eda0d-2cea-4de1-8d42-9cd3e7e86705",
			Active:       true,
		},
	}
	// Make sure we start test with empty database
	db.Wipe()
	for _, deviceDeployment := range deviceDeployments {
		assert.NoError(t, ds.InsertDeviceDeployment(ctx, deviceDeployment, true))
	}

	testCases := map[string]struct {
		skip           int
		limit          int
		deviceID       string
		active         *bool
		includeDeleted bool

		res []model.DeviceDeployment
	}{
		"ok": {
			includeDeleted: true,
			res: []model.DeviceDeployment{
				*deviceDeployments[0],
				*deviceDeployments[1],
				*deviceDeployments[2],
				*deviceDeployments[3],
				*deviceDeployments[4],
			},
		},
		"ok, skip and limit": {
			skip:           1,
			limit:          2,
			includeDeleted: true,
			res: []model.DeviceDeployment{
				*deviceDeployments[1],
				*deviceDeployments[2],
			},
		},
		"ok, not active, not deleted": {
			active:         &f,
			includeDeleted: false,
			res: []model.DeviceDeployment{
				*deviceDeployments[2],
			},
		},
		"ok, filter by deviceID": {
			deviceID:       differentDeviceID,
			includeDeleted: false,
			res: []model.DeviceDeployment{
				*deviceDeployments[4],
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			res, err := ds.GetDeviceDeployments(
				ctx, tc.skip, tc.limit, tc.deviceID, tc.active, tc.includeDeleted)
			assert.NoError(t, err)

			for i, _ := range res {
				// ignore Created, Started, and Deleted fields when comparing the results
				res[i].Created = tc.res[i].Created
				res[i].Deleted = tc.res[i].Deleted
				res[i].Started = tc.res[i].Started
			}
			assert.Equal(t, tc.res, res)
			assert.Nil(t, err)
		})
	}
}

func TestExistUnfinishedByArtifactName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestExistUnfinishedByArtifactName in short mode.")
	}

	testCases := map[string]struct {
		inputDeploymentsCollection []interface{}

		artifactName string

		exist bool
		err   error
	}{
		"ok, exist": {
			inputDeploymentsCollection: []interface{}{
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "foo",
					},
					Id:     "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
					Active: true,
				},
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "bar",
					},
					Id:     "d1804903-5caa-4a73-a3ae-0efcc3205405",
					Active: false,
				},
			},
			artifactName: "foo",
			exist:        true,
		},
		"ok, does not exist": {
			inputDeploymentsCollection: []interface{}{
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "foo",
					},
					Id:     "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
					Active: true,
				},
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "bar",
					},
					Id:     "d1804903-5caa-4a73-a3ae-0efcc3205405",
					Active: false,
				},
			},
			artifactName: "baz",
			exist:        false,
		},
		"no deployments": {
			artifactName: "baz",
			exist:        false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Make sure we start test with empty database
			db.Wipe()

			client := db.Client()
			ds := NewDataStoreMongoWithClient(client)

			ctx := context.Background()

			collDep := client.Database(ctxstore.
				DbFromContext(ctx, DatabaseName)).
				Collection(CollectionDeployments)

			if tc.inputDeploymentsCollection != nil {
				_, err := collDep.InsertMany(
					ctx, tc.inputDeploymentsCollection)
				assert.NoError(t, err)
			}

			exist, err := ds.ExistUnfinishedByArtifactName(ctx, tc.artifactName)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.exist, exist)
			}
		})
	}
}

func TestExistUnfinishedByArtifactId(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestExistUnfinishedByArtifactId in short mode.")
	}

	testCases := map[string]struct {
		inputDeploymentsCollection []interface{}

		artifactId string

		exist bool
		err   error
	}{
		"ok, exist": {
			inputDeploymentsCollection: []interface{}{
				&model.Deployment{
					Id:        "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
					Artifacts: []string{"foo", "bar"},
					Active:    true,
				},
				&model.Deployment{
					Id:        "d1804903-5caa-4a73-a3ae-0efcc3205405",
					Artifacts: []string{"baz"},
					Active:    false,
				},
			},
			artifactId: "foo",
			exist:      true,
		},
		"ok, does not exist": {
			inputDeploymentsCollection: []interface{}{
				&model.Deployment{
					Id:        "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
					Artifacts: []string{"foo", "bar"},
					Active:    true,
				},
				&model.Deployment{
					Id:        "d1804903-5caa-4a73-a3ae-0efcc3205405",
					Artifacts: []string{"bar"},
					Active:    false,
				},
			},
			artifactId: "baz",
			exist:      false,
		},
		"no deployments": {
			artifactId: "baz",
			exist:      false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Make sure we start test with empty database
			db.Wipe()

			client := db.Client()
			ds := NewDataStoreMongoWithClient(client)

			ctx := context.Background()

			collDep := client.Database(ctxstore.
				DbFromContext(ctx, DatabaseName)).
				Collection(CollectionDeployments)

			if tc.inputDeploymentsCollection != nil {
				_, err := collDep.InsertMany(
					ctx, tc.inputDeploymentsCollection)
				assert.NoError(t, err)
			}

			exist, err := ds.ExistUnfinishedByArtifactId(ctx, tc.artifactId)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.exist, exist)
			}
		})
	}
}

func TestUpdateDeploymentsWithArtifactName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestUpdateDeploymentsWithArtifactName in short mode.")
	}

	testCases := map[string]struct {
		inputDeploymentsCollection []interface{}

		artifactName string
		artifactIDs  []string

		outputDeployments []*model.Deployment
		err               error
	}{
		"ok": {
			inputDeploymentsCollection: []interface{}{
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "foo",
					},
					Id:        "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
					Artifacts: []string{"foo-1"},

					Active: true,
				},
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "bar",
					},
					Id:     "d1804903-5caa-4a73-a3ae-0efcc3205405",
					Active: true,
				},
			},
			artifactName: "foo",
			artifactIDs:  []string{"foo-1", "foo-2"},
			outputDeployments: []*model.Deployment{
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "foo",
					},
					Id:        "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
					Artifacts: []string{"foo-1", "foo-2"},
					Active:    true,
				},
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "bar",
					},
					Id:     "d1804903-5caa-4a73-a3ae-0efcc3205405",
					Active: true,
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Make sure we start test with empty database
			db.Wipe()

			client := db.Client()
			ds := NewDataStoreMongoWithClient(client)

			ctx := context.Background()

			collDep := client.Database(ctxstore.
				DbFromContext(ctx, DatabaseName)).
				Collection(CollectionDeployments)

			if tc.inputDeploymentsCollection != nil {
				_, err := collDep.InsertMany(
					ctx, tc.inputDeploymentsCollection)
				assert.NoError(t, err)
			}

			err := ds.UpdateDeploymentsWithArtifactName(
				ctx,
				tc.artifactName,
				tc.artifactIDs,
			)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
				deployments, _, err := ds.FindDeployments(ctx, model.Query{})
				assert.NoError(t, err)
				assert.Equal(t, tc.outputDeployments, deployments)
			}
		})
	}
}

func TestDeleteDeviceDeploymentsHistory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestDeleteDeviceDeploymentsHistory in short mode.")
	}

	testCases := map[string]struct {
		inputDeviceDeployments []interface{}

		deviceID string
		assert   func(deviceDeployments []model.DeviceDeployment)
	}{
		"ok": {
			inputDeviceDeployments: []interface{}{
				&model.DeviceDeployment{
					Id:       "a108ae14-bb4e-455f-9b40-2ef4bab97bb0",
					DeviceId: "a108ae14-bb4e-455f-9b40-2ef4bab97bb0",
					Active:   true,
				},
				&model.DeviceDeployment{
					Id:       "a108ae14-bb4e-455f-9b40-2ef4bab97bb1",
					DeviceId: "a108ae14-bb4e-455f-9b40-2ef4bab97bb0",
					Active:   false,
				},
				&model.DeviceDeployment{
					Id:       "a108ae14-bb4e-455f-9b40-2ef4bab97bb2",
					DeviceId: "a108ae14-bb4e-455f-9b40-2ef4bab97bb0",
					Active:   false,
				},
			},
			deviceID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb0",
			assert: func(deviceDeployments []model.DeviceDeployment) {
				assert.Len(t, deviceDeployments, 2)
			},
		},
		"ko, no matches": {
			inputDeviceDeployments: []interface{}{
				&model.DeviceDeployment{
					Id:       "a108ae14-bb4e-455f-9b40-2ef4bab97bb0",
					DeviceId: "a108ae14-bb4e-455f-9b40-2ef4bab97bb0",
					Active:   true,
				},
				&model.DeviceDeployment{
					Id:       "a108ae14-bb4e-455f-9b40-2ef4bab97bb1",
					DeviceId: "a108ae14-bb4e-455f-9b40-2ef4bab97bb0",
					Active:   false,
				},
				&model.DeviceDeployment{
					Id:       "a108ae14-bb4e-455f-9b40-2ef4bab97bb2",
					DeviceId: "a108ae14-bb4e-455f-9b40-2ef4bab97bb0",
					Active:   false,
				},
			},
			deviceID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb1",
			assert: func(deviceDeployments []model.DeviceDeployment) {
				assert.Len(t, deviceDeployments, 0)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db.Wipe()

			client := db.Client()
			ds := NewDataStoreMongoWithClient(client)

			ctx := context.Background()

			collDeviceDeps := client.Database(ctxstore.
				DbFromContext(ctx, DatabaseName)).
				Collection(CollectionDevices)

			if tc.inputDeviceDeployments != nil {
				_, err := collDeviceDeps.InsertMany(
					ctx, tc.inputDeviceDeployments)
				assert.NoError(t, err)
			}

			err := ds.DeleteDeviceDeploymentsHistory(ctx, tc.deviceID)
			assert.NoError(t, err)
			cur, err := collDeviceDeps.Find(ctx, bson.M{
				StorageKeyDeviceDeploymentDeleted: bson.M{"$exists": true},
			})
			assert.NoError(t, err)
			var deviceDeployments []model.DeviceDeployment
			err = cur.All(ctx, &deviceDeployments)
			assert.NoError(t, err)
			tc.assert(deviceDeployments)
		})
	}
}

func TestIncrementDeploymentTotalSize(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestIncrementDeploymentTotalSize in short mode.")
	}

	testCases := map[string]struct {
		inputDeploymentsCollection []interface{}

		artifactSize int64
		deploymentID string

		outputDeployments []*model.Deployment
		err               error
	}{
		"ok": {
			inputDeploymentsCollection: []interface{}{
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "foo",
					},
					Id:        "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
					Artifacts: []string{"foo-1"},
					Statistics: model.DeploymentStatistics{
						TotalSize: 100,
					},
					Active: true,
				},
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "bar",
					},
					Id:     "d1804903-5caa-4a73-a3ae-0efcc3205405",
					Active: true,
				},
			},
			artifactSize: 200,
			deploymentID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			outputDeployments: []*model.Deployment{
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "foo",
					},
					Id:        "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
					Artifacts: []string{"foo-1"},
					Statistics: model.DeploymentStatistics{
						TotalSize: 300,
					},
					Active: true,
				},
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "bar",
					},
					Id:     "d1804903-5caa-4a73-a3ae-0efcc3205405",
					Active: true,
				},
			},
		},
		"ok, no statistics at the beginning": {
			inputDeploymentsCollection: []interface{}{
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "foo",
					},
					Id:        "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
					Artifacts: []string{"foo-1"},
					Statistics: model.DeploymentStatistics{
						TotalSize: 100,
					},
					Active: true,
				},
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "bar",
					},
					Id:     "d1804903-5caa-4a73-a3ae-0efcc3205405",
					Active: true,
				},
			},
			artifactSize: 200,
			deploymentID: "d1804903-5caa-4a73-a3ae-0efcc3205405",
			outputDeployments: []*model.Deployment{
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "foo",
					},
					Id:        "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
					Artifacts: []string{"foo-1"},
					Statistics: model.DeploymentStatistics{
						TotalSize: 100,
					},
					Active: true,
				},
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "bar",
					},
					Id: "d1804903-5caa-4a73-a3ae-0efcc3205405",
					Statistics: model.DeploymentStatistics{
						TotalSize: 200,
					},
					Active: true,
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Make sure we start test with empty database
			db.Wipe()

			client := db.Client()
			ds := NewDataStoreMongoWithClient(client)

			ctx := context.Background()

			collDep := client.Database(ctxstore.
				DbFromContext(ctx, DatabaseName)).
				Collection(CollectionDeployments)

			if tc.inputDeploymentsCollection != nil {
				_, err := collDep.InsertMany(
					ctx, tc.inputDeploymentsCollection)
				assert.NoError(t, err)
			}

			err := ds.IncrementDeploymentTotalSize(
				ctx,
				tc.deploymentID,
				tc.artifactSize,
			)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
				deployments, _, err := ds.FindDeployments(ctx, model.Query{})
				assert.NoError(t, err)
				assert.Equal(t, tc.outputDeployments, deployments)
			}
		})
	}
}

func TestSaveUpdateTypes(t *testing.T) {
	ctx := context.Background()
	client := db.Client()
	db.Wipe()

	type testCase struct {
		Name string

		context.Context

		Init                func(t *testing.T, self *testCase)
		UpdateTypes         []string
		ExpectedUpdateTypes []string
	}

	testCases := []testCase{
		{
			Name: "ok newly added",

			Context: context.Background(),

			Init: func(t *testing.T, self *testCase) {
				t.Helper()
				client.Database(DbName).
					Collection(CollectionUpdateTypes).
					DeleteMany(ctx, bson.M{})
			},
			UpdateTypes:         []string{"type0", "type1", "type2", "type3"},
			ExpectedUpdateTypes: []string{"type0", "type1", "type2", "type3"},
		},
		{
			Name: "ok same set",

			Context: context.Background(),

			Init: func(t *testing.T, self *testCase) {
				t.Helper()
				client.Database(DbName).
					Collection(CollectionUpdateTypes).
					DeleteMany(ctx, bson.M{})
				_, err := client.Database(DbName).
					Collection(CollectionUpdateTypes).
					InsertOne(ctx, bson.M{
						StorageKeyStorageReleaseUpdateTypes: []string{"type0", "type1"},
						StorageKeyTenantId:                  "",
					})
				if err != nil {
					t.Errorf("failed to initialize dataset: %s", err)
					t.FailNow()
				}
			},
			UpdateTypes:         []string{"type0", "type1"},
			ExpectedUpdateTypes: []string{"type0", "type1"},
		},
		{
			Name: "ok add to existing",

			Context: context.Background(),

			Init: func(t *testing.T, self *testCase) {
				t.Helper()
				client.Database(DbName).
					Collection(CollectionUpdateTypes).
					DeleteMany(ctx, bson.M{})
				_, err := client.Database(DbName).
					Collection(CollectionUpdateTypes).
					InsertOne(ctx, bson.M{
						StorageKeyStorageReleaseUpdateTypes: []string{"type1", "type2"},
						StorageKeyTenantId:                  "",
					})
				if err != nil {
					t.Errorf("failed to initialize dataset: %s", err)
					t.FailNow()
				}
			},
			UpdateTypes:         []string{"type0", "type1", "type2", "type3"},
			ExpectedUpdateTypes: []string{"type1", "type2", "type0", "type3"},
		},
		{
			Name: "ok empty add",

			Context: context.Background(),

			Init: func(t *testing.T, self *testCase) {
				t.Helper()
				client.Database(DbName).
					Collection(CollectionUpdateTypes).
					DeleteMany(ctx, bson.M{})
				_, err := client.Database(DbName).
					Collection(CollectionUpdateTypes).
					InsertOne(ctx, bson.M{
						StorageKeyStorageReleaseUpdateTypes: []string{"type1", "type2"},
						StorageKeyTenantId:                  "",
					})
				if err != nil {
					t.Errorf("failed to initialize dataset: %s", err)
					t.FailNow()
				}
			},
			UpdateTypes:         []string{},
			ExpectedUpdateTypes: []string{"type1", "type2"},
		},
		{
			Name: "ok newly added with tenant",

			Context: identity.WithContext(context.Background(),
				&identity.Identity{
					Tenant: "222222222222222222222222",
				},
			),

			Init: func(t *testing.T, self *testCase) {
				t.Helper()
				client.Database(DbName).
					Collection(CollectionUpdateTypes).
					DeleteMany(ctx, bson.M{})
				_, err := client.Database(DbName).
					Collection(CollectionUpdateTypes).
					InsertOne(ctx, bson.M{
						StorageKeyStorageReleaseUpdateTypes: []string{"other_tenant_type0", "other_tenant_type1"},
						StorageKeyTenantId:                  "322222222222222222222222",
					})
				if err != nil {
					t.Errorf("failed to initialize dataset: %s", err)
					t.FailNow()
				}
			},
			UpdateTypes:         []string{"type0", "type1", "type2", "type3"},
			ExpectedUpdateTypes: []string{"type0", "type1", "type2", "type3"},
		},
		{
			Name: "ok same set with tenant",

			Context: identity.WithContext(context.Background(),
				&identity.Identity{
					Tenant: "222222222222222222222222",
				},
			),

			Init: func(t *testing.T, self *testCase) {
				t.Helper()
				client.Database(DbName).
					Collection(CollectionUpdateTypes).
					DeleteMany(ctx, bson.M{})
				_, err := client.Database(DbName).
					Collection(CollectionUpdateTypes).
					InsertOne(ctx, bson.M{
						StorageKeyStorageReleaseUpdateTypes: []string{"type0", "type1"},
						StorageKeyTenantId:                  "222222222222222222222222",
					})
				if err != nil {
					t.Errorf("failed to initialize dataset: %s", err)
					t.FailNow()
				}
			},
			UpdateTypes:         []string{"type0", "type1"},
			ExpectedUpdateTypes: []string{"type0", "type1"},
		},
		{
			Name: "ok add to existing with tenant",

			Context: identity.WithContext(context.Background(),
				&identity.Identity{
					Tenant: "222222222222222222222222",
				},
			),

			Init: func(t *testing.T, self *testCase) {
				t.Helper()
				client.Database(DbName).
					Collection(CollectionUpdateTypes).
					DeleteMany(ctx, bson.M{})
				_, err := client.Database(DbName).
					Collection(CollectionUpdateTypes).
					InsertOne(ctx, bson.M{
						StorageKeyStorageReleaseUpdateTypes: []string{"type1", "type2"},
						StorageKeyTenantId:                  "222222222222222222222222",
					})
				if err != nil {
					t.Errorf("failed to initialize dataset: %s", err)
					t.FailNow()
				}
			},
			UpdateTypes:         []string{"type0", "type1", "type2", "type3"},
			ExpectedUpdateTypes: []string{"type1", "type2", "type0", "type3"},
		},
		{
			Name: "ok empty add with tenant",

			Context: identity.WithContext(context.Background(),
				&identity.Identity{
					Tenant: "222222222222222222222222",
				},
			),

			Init: func(t *testing.T, self *testCase) {
				t.Helper()
				client.Database(DbName).
					Collection(CollectionUpdateTypes).
					DeleteMany(ctx, bson.M{})
				_, err := client.Database(DbName).
					Collection(CollectionUpdateTypes).
					InsertOne(ctx, bson.M{
						StorageKeyStorageReleaseUpdateTypes: []string{"type1", "type2"},
						StorageKeyTenantId:                  "222222222222222222222222",
					})
				if err != nil {
					t.Errorf("failed to initialize dataset: %s", err)
					t.FailNow()
				}
			},
			UpdateTypes:         []string{},
			ExpectedUpdateTypes: []string{"type1", "type2"},
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()
			var tenantID string
			if id := identity.FromContext(tc.Context); id != nil {
				tenantID = id.Tenant
			}
			err := MigrateSingle(ctx,
				ctxstore.DbNameForTenant(tenantID, DbName),
				DbVersion,
				client,
				true)
			if err != nil {
				panic(err)
			}
			tc.Init(t, &tc)

			ds := NewDataStoreMongoWithClient(client)
			err = ds.SaveUpdateTypes(tc.Context, tc.UpdateTypes)
			assert.NoError(t, err)
			updateTypes, err := ds.GetUpdateTypes(tc.Context)
			assert.Equal(t, tc.ExpectedUpdateTypes, updateTypes)
		})
	}
}

func TestInsertDeviceDeployment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestGetDeviceDeployments in short mode.")
	}

	now := time.Now()

	ctx := context.Background()
	client := db.Client()
	ds := NewDataStoreMongoWithClient(client)

	const deviceID = "d50eda0d-2cea-4de1-8d42-9cd3e7e86700"
	const differentDeviceID = "d50eda0d-2cea-4de1-8d42-9cd3e7e86701"
	deviceDeployments := []*model.DeviceDeployment{
		{
			Id: "d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
			Created: func() *time.Time {
				ret := now.Add(5 * time.Hour)
				return &ret
			}(),
			Status:       model.DeviceDeploymentStatusPauseBeforeInstall,
			DeviceId:     deviceID,
			DeploymentId: "d50eda0d-2cea-4de1-8d42-9cd3e7e86701",
			Active:       true,
		},
		{
			Id: "d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
			Created: func() *time.Time {
				ret := now.Add(4 * time.Hour)
				return &ret
			}(),
			Status:       model.DeviceDeploymentStatusSuccess,
			DeviceId:     deviceID,
			DeploymentId: "d50eda0d-2cea-4de1-8d42-9cd3e7e86702",
			Deleted:      &now,
		},
		{
			Id: "d50eda0d-2cea-4de1-8d42-9cd3e7e86703",
			Created: func() *time.Time {
				ret := now.Add(3 * time.Hour)
				return &ret
			}(),
			Status:       model.DeviceDeploymentStatusSuccess,
			DeviceId:     deviceID,
			DeploymentId: "d50eda0d-2cea-4de1-8d42-9cd3e7e86703",
		},
		{
			Id: "d50eda0d-2cea-4de1-8d42-9cd3e7e86704",
			Created: func() *time.Time {
				ret := now.Add(2 * time.Hour)
				return &ret
			}(),
			Status:       model.DeviceDeploymentStatusPending,
			DeviceId:     deviceID,
			DeploymentId: "d50eda0d-2cea-4de1-8d42-9cd3e7e86704",
			Active:       true,
		},
	}

	testCases := map[string]struct {
		DeviceDeploymentsToInsert []*model.DeviceDeployment
		ExpectedStarted           []bool
	}{
		"inserted one pending": {
			DeviceDeploymentsToInsert: []*model.DeviceDeployment{
				deviceDeployments[3],
			},
			ExpectedStarted: []bool{false},
		},
		"inserted one non pending": {
			DeviceDeploymentsToInsert: []*model.DeviceDeployment{
				deviceDeployments[2],
			},
			ExpectedStarted: []bool{true},
		},
		"inserted more than one pending": {
			DeviceDeploymentsToInsert: []*model.DeviceDeployment{
				deviceDeployments[2],
				deviceDeployments[3],
			},
			ExpectedStarted: []bool{true, false},
		},
		"inserted more than one non pending": {
			DeviceDeploymentsToInsert: []*model.DeviceDeployment{
				deviceDeployments[0],
				deviceDeployments[1],
				deviceDeployments[2],
			},
			ExpectedStarted: []bool{true, true, true},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db.Wipe()
			if len(tc.DeviceDeploymentsToInsert) > 1 {
				ds.InsertMany(ctx, tc.DeviceDeploymentsToInsert...)
			} else {
				ds.InsertDeviceDeployment(ctx, tc.DeviceDeploymentsToInsert[0], true)
			}
			d := client.Database(DbName)
			c := d.Collection(CollectionDevices)
			r, _ := c.Find(ctx, bson.M{})
			var results []model.DeviceDeployment
			r.All(ctx, &results)
			assert.Equal(t, len(tc.DeviceDeploymentsToInsert), len(results))
			for i, e := range results {
				assert.Equal(t, tc.ExpectedStarted[i], e.Started != nil)
			}
		})
	}
}

func TestDeleteImage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestDeleteImage in short mode.")
	}

	testCases := map[string]struct {
		inputImages  []interface{}
		id           string
		outputImages []model.Image
	}{
		"ok": {
			inputImages: []interface{}{
				&model.Image{
					Id: "foo",
				},
				&model.Image{
					Id: "bar",
				},
			},
			id: "foo",
			outputImages: []model.Image{
				{
					Id: "bar",
				},
			},
		},
		"not found": {
			inputImages: []interface{}{
				&model.Image{
					Id: "foo",
				},
			},
			id: "bar",
			outputImages: []model.Image{
				{
					Id: "foo",
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db.Wipe()

			client := db.Client()
			ds := NewDataStoreMongoWithClient(client)

			ctx := context.Background()

			collImages := client.Database(ctxstore.
				DbFromContext(ctx, DatabaseName)).
				Collection(CollectionImages)

			if tc.inputImages != nil {
				_, err := collImages.InsertMany(
					ctx, tc.inputImages)
				assert.NoError(t, err)
			}

			err := ds.DeleteImage(ctx, tc.id)
			assert.NoError(t, err)
			cur, err := collImages.Find(ctx, bson.M{})
			assert.NoError(t, err)
			var images []model.Image
			err = cur.All(ctx, &images)
			assert.NoError(t, err)
			assert.Equal(t, tc.outputImages, images)
		})
	}
}

func TestDeleteImagesByNames(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestDeleteImagesByNames in short mode.")
	}

	testCases := map[string]struct {
		inputImages  []interface{}
		names        []string
		outputImages []model.Image
	}{
		"ok": {
			inputImages: []interface{}{
				&model.Image{
					Id: "1",
					ArtifactMeta: &model.ArtifactMeta{
						Name:                  "foo",
						DeviceTypesCompatible: []string{"foo"},
					},
				},
				&model.Image{
					Id: "2",
					ArtifactMeta: &model.ArtifactMeta{
						Name:                  "bar",
						DeviceTypesCompatible: []string{"bar"},
					},
				},
				&model.Image{
					Id: "3",
					ArtifactMeta: &model.ArtifactMeta{
						Name:                  "baz",
						DeviceTypesCompatible: []string{"baz"},
					},
				},
			},
			names: []string{"foo", "baz"},
			outputImages: []model.Image{
				{
					Id: "2",
					ArtifactMeta: &model.ArtifactMeta{
						Name:                  "bar",
						DeviceTypesCompatible: []string{"bar"},
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db.Wipe()

			client := db.Client()
			ds := NewDataStoreMongoWithClient(client)

			ctx := context.Background()

			collImages := client.Database(ctxstore.
				DbFromContext(ctx, DatabaseName)).
				Collection(CollectionImages)

			if tc.inputImages != nil {
				_, err := collImages.InsertMany(
					ctx, tc.inputImages)
				assert.NoError(t, err)
			}

			err := ds.DeleteImagesByNames(ctx, tc.names)
			assert.NoError(t, err)
			cur, err := collImages.Find(ctx, bson.M{})
			assert.NoError(t, err)
			var images []model.Image
			err = cur.All(ctx, &images)
			// clear depends; not relevant for this test
			for i := range images {
				images[i].Depends = nil
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.outputImages, images)
		})
	}
}

func TestGetActiveDeploymentIDsByArtifactNames(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestGetActiveDeploymentIDsByArtifactNames in short mode.")
	}

	testCases := map[string]struct {
		inputDeployments []interface{}

		artifactNames []string

		outputIDs []string
	}{
		"ok": {
			inputDeployments: []interface{}{
				&model.Deployment{
					Status: model.DeploymentStatusInProgress,
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "foo",
					},
					Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				},
				&model.Deployment{
					Status: model.DeploymentStatusInProgress,
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "bar",
					},
					Id: "d1804903-5caa-4a73-a3ae-0efcc3205405",
				},
				&model.Deployment{
					Status: model.DeploymentStatusPending,
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "baz",
					},
					Id: "d1804903-5caa-4a73-a3ae-0efcc3205406",
				},
			},
			artifactNames: []string{"foo", "baz"},
			outputIDs:     []string{"a108ae14-bb4e-455f-9b40-2ef4bab97bb7", "d1804903-5caa-4a73-a3ae-0efcc3205406"},
		},
		"no active deployments:": {
			inputDeployments: []interface{}{
				&model.Deployment{
					Status: model.DeploymentStatusFinished, // sets active to false
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "foo",
					},
					Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				},
				&model.Deployment{
					Status: model.DeploymentStatusFinished,
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "bar",
					},
					Id: "d1804903-5caa-4a73-a3ae-0efcc3205405",
				},
			},
			artifactNames: []string{"foo", "bar"},
			outputIDs:     []string{},
		},
		"no deployments": {
			artifactNames: []string{"foo", "baz"},
			outputIDs:     []string{},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Make sure we start test with empty database
			db.Wipe()

			client := db.Client()
			ds := NewDataStoreMongoWithClient(client)

			ctx := context.Background()

			collDep := client.Database(ctxstore.
				DbFromContext(ctx, DatabaseName)).
				Collection(CollectionDeployments)

			if tc.inputDeployments != nil {
				_, err := collDep.InsertMany(
					ctx, tc.inputDeployments)
				assert.NoError(t, err)
			}

			ids, err := ds.GetActiveDeploymentIDsByArtifactNames(ctx, tc.artifactNames)
			assert.NoError(t, err)
			assert.Equal(t, tc.outputIDs, ids)
		})
	}
}

func TestFindDeployments(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestFindDeployments in short mode.")
	}

	testCases := map[string]struct {
		inputDeployments []interface{}
		query            model.Query

		outputDeployments []*model.Deployment
	}{
		"ok, empty query": {
			inputDeployments: []interface{}{
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "foo",
					},
					Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				},
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "bar",
					},
					Id: "d1804903-5caa-4a73-a3ae-0efcc3205405",
				},
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "baz",
					},
					Id: "d1804903-5caa-4a73-a3ae-0efcc3205406",
				},
			},
			query: model.Query{},
			outputDeployments: []*model.Deployment{
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "foo",
					},
					Id:     "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
					Active: true,
				},
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "bar",
					},
					Id:     "d1804903-5caa-4a73-a3ae-0efcc3205405",
					Active: true,
				},
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "baz",
					},
					Id:     "d1804903-5caa-4a73-a3ae-0efcc3205406",
					Active: true,
				},
			},
		},
		"ok, Ids in query": {
			inputDeployments: []interface{}{
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "foo",
						Name:         "1",
					},
					Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				},
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "bar",
						Name:         "2",
					},
					Id: "d1804903-5caa-4a73-a3ae-0efcc3205405",
				},
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "baz",
						Name:         "3",
					},
					Id: "d1804903-5caa-4a73-a3ae-0efcc3205406",
				},
			},
			query: model.Query{
				IDs: []string{
					"a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
					"d1804903-5caa-4a73-a3ae-0efcc3205405",
				},
			},
			outputDeployments: []*model.Deployment{
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "foo",
						Name:         "1",
					},
					Id:     "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
					Active: true,
				},
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "bar",
						Name:         "2",
					},
					Id:     "d1804903-5caa-4a73-a3ae-0efcc3205405",
					Active: true,
				},
			},
		},
		"ok, names in query": {
			inputDeployments: []interface{}{
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "foo",
						Name:         "1",
					},
					Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				},
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "bar",
						Name:         "2",
					},
					Id: "d1804903-5caa-4a73-a3ae-0efcc3205405",
				},
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "baz",
						Name:         "3",
					},
					Id: "d1804903-5caa-4a73-a3ae-0efcc3205406",
				},
			},
			query: model.Query{
				Names: []string{
					"1",
					"3",
				},
			},
			outputDeployments: []*model.Deployment{
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "foo",
						Name:         "1",
					},
					Id:     "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
					Active: true,
				},
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						ArtifactName: "baz",
						Name:         "3",
					},
					Id:     "d1804903-5caa-4a73-a3ae-0efcc3205406",
					Active: true,
				},
			},
		},
		"no deployments": {
			query:             model.Query{},
			outputDeployments: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Make sure we start test with empty database
			db.Wipe()

			client := db.Client()
			ds := NewDataStoreMongoWithClient(client)

			ctx := context.Background()

			collDep := client.Database(ctxstore.
				DbFromContext(ctx, DatabaseName)).
				Collection(CollectionDeployments)

			if tc.inputDeployments != nil {
				_, err := collDep.InsertMany(
					ctx, tc.inputDeployments)
				assert.NoError(t, err)
			}

			deployments, count, err := ds.FindDeployments(ctx, tc.query)
			assert.NoError(t, err)
			assert.Equal(t, int64(len(tc.outputDeployments)), count)
			assert.Equal(t, tc.outputDeployments, deployments)
		})
	}
}
