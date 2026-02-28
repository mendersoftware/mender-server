// Copyright 2022 Northern.tech AS
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
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/mendersoftware/mender-server/pkg/identity"
	mstore "github.com/mendersoftware/mender-server/pkg/store"

	"github.com/mendersoftware/mender-server/services/deployments/model"
	. "github.com/mendersoftware/mender-server/services/deployments/utils/pointers"
)

func TimePtr(t time.Time) *time.Time {
	return &t
}

func TestDeploymentStorageInsert(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping TestDeploymentStorageInsert in short mode.")
	}

	newDepl, err := model.NewDeployment()
	assert.NoError(t, err)

	newDeplFromConstr, err := model.NewDeploymentFromConstructor(&model.DeploymentConstructor{

		Name:         "NYC Production",
		ArtifactName: "App 123",
		Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
	})
	assert.NoError(t, err)

	testCases := []struct {
		InputDeployment *model.Deployment
		InputTenant     string
		OutputError     error
	}{
		{
			InputDeployment: nil,
			OutputError:     ErrDeploymentStorageInvalidDeployment,
		},
		{
			InputDeployment: newDepl,
			OutputError:     errors.New("artifact_name: cannot be blank; name: cannot be blank."),
		},
		{
			InputDeployment: newDeplFromConstr,
			OutputError:     nil,
		},
		{
			InputDeployment: newDeplFromConstr,
			InputTenant:     "acme",
		},
	}

	for testCaseNumber, testCase := range testCases {
		t.Run(fmt.Sprintf("test case %d", testCaseNumber+1), func(t *testing.T) {

			// Make sure we start test with empty database
			db.Wipe()

			client := db.Client()
			ctx := db.CTX()
			store := NewDataStoreMongoWithClient(client)

			if testCase.InputTenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: testCase.InputTenant,
				})
			} else {
				ctx = context.Background()
			}

			err := store.InsertDeployment(ctx, testCase.InputDeployment)

			if testCase.OutputError != nil {
				assert.EqualError(t, err, testCase.OutputError.Error())
			} else {
				assert.NoError(t, err)

				collDep := client.Database(mstore.
					DbFromContext(ctx, DatabaseName)).
					Collection(CollectionDeployments)
				count, err := collDep.CountDocuments(
					ctx, bson.D{})
				assert.NoError(t, err)
				assert.Equal(t, 1, int(count))

				if testCase.InputTenant != "" {
					collDefaultDep := client.
						Database(DatabaseName).
						Collection(CollectionDeployments)
					indefault, err := collDefaultDep.
						CountDocuments(ctx, bson.D{})
					assert.NoError(t, err)
					assert.Equal(t, 0, int(indefault))
				}
			}
		})
	}
}

func TestDeploymentStorageDelete(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping TestDeploymentStorageDelete in short mode.")
	}

	testCases := []struct {
		InputID                    string
		InputDeploymentsCollection []interface{}
		InputTenant                string

		OutputError error
	}{
		{
			InputID:     "",
			OutputError: ErrStorageInvalidID,
		},
		{
			InputID:     "b532b01a-9313-404f-8d19-e7fcbe5cc347",
			OutputError: nil,
		},
		{
			InputID: "b532b01a-9313-404f-8d19-e7fcbe5cc347",
			InputDeploymentsCollection: []interface{}{
				model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: "b532b01a-9313-404f-8d19-e7fcbe5cc347",
				},
			},
			OutputError: nil,
		},
		{
			InputID: "b532b01a-9313-404f-8d19-e7fcbe5cc347",
			InputDeploymentsCollection: []interface{}{
				model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: "b532b01a-9313-404f-8d19-e7fcbe5cc347",
				},
			},
			InputTenant: "acme",
		},
	}

	for testCaseNumber, testCase := range testCases {
		t.Run(fmt.Sprintf("test case %d", testCaseNumber+1), func(t *testing.T) {

			// Make sure we start test with empty database
			db.Wipe()

			client := db.Client()
			ctx := db.CTX()
			store := NewDataStoreMongoWithClient(client)

			if testCase.InputTenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: testCase.InputTenant,
				})
			} else {
				ctx = context.Background()
			}

			collDep := client.Database(mstore.
				DbFromContext(ctx, DatabaseName)).
				Collection(CollectionDeployments)
			if testCase.InputDeploymentsCollection != nil {
				_, err := collDep.InsertMany(ctx,
					testCase.InputDeploymentsCollection)
				assert.NoError(t, err)
			}

			err := store.DeleteDeployment(ctx, testCase.InputID)

			if testCase.OutputError != nil {
				assert.EqualError(t, err, testCase.OutputError.Error())
			} else {
				assert.NoError(t, err)

				count, err := collDep.CountDocuments(ctx,
					bson.M{"_id": testCase.InputID})
				assert.NoError(t, err)
				assert.Equal(t, 0, int(count))

				if testCase.InputTenant != "" {
					collDefaultDep := client.
						Database(DatabaseName).
						Collection(CollectionDeployments)
					indefault, err := collDefaultDep.
						CountDocuments(ctx, bson.D{})
					assert.NoError(t, err)
					assert.Equal(t, 0, int(indefault))
				}
			}
		})
	}
}

func TestDeploymentStorageFindByID(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping TestDeploymentStorageFindByID in short mode.")
	}

	testCases := map[string]struct {
		InputID                    string
		InputDeploymentsCollection bson.A
		InputTenant                string

		OutputError      error
		OutputDeployment *model.Deployment
	}{
		"invalid ID": {
			InputID:     "",
			OutputError: ErrStorageInvalidID,
		},
		"no deployments": {
			InputID:          "b532b01a-9313-404f-8d19-e7fcbe5cc347",
			OutputError:      nil,
			OutputDeployment: nil,
		},
		"not found": {
			InputID: "b532b01a-9313-404f-8d19-e7fcbe5cc347",
			InputDeploymentsCollection: bson.A{
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				},
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: "d1804903-5caa-4a73-a3ae-0efcc3205405",
				},
			},
			OutputError:      nil,
			OutputDeployment: nil,
		},
		"ok": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeploymentsCollection: bson.A{
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
					Stats: model.Stats{
						model.DeviceDeploymentStatusDownloadingStr: 0,
						model.DeviceDeploymentStatusInstallingStr:  0,
						model.DeviceDeploymentStatusRebootingStr:   0,
						model.DeviceDeploymentStatusPendingStr:     10,
						model.DeviceDeploymentStatusSuccessStr:     15,
						model.DeviceDeploymentStatusFailureStr:     1,
						model.DeviceDeploymentStatusNoArtifactStr:  0,
						model.DeviceDeploymentStatusAlreadyInstStr: 0,
						model.DeviceDeploymentStatusAbortedStr:     0,
					},
				},
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: "d1804903-5caa-4a73-a3ae-0efcc3205405",
					Stats: model.Stats{
						model.DeviceDeploymentStatusDownloadingStr: 0,
						model.DeviceDeploymentStatusInstallingStr:  0,
						model.DeviceDeploymentStatusRebootingStr:   0,
						model.DeviceDeploymentStatusPendingStr:     5,
						model.DeviceDeploymentStatusSuccessStr:     10,
						model.DeviceDeploymentStatusFailureStr:     3,
						model.DeviceDeploymentStatusNoArtifactStr:  0,
						model.DeviceDeploymentStatusAlreadyInstStr: 0,
						model.DeviceDeploymentStatusAbortedStr:     0,
					},
				},
			},
			OutputError: nil,
			OutputDeployment: &model.Deployment{
				DeploymentConstructor: &model.DeploymentConstructor{
					Name:         "NYC Production",
					ArtifactName: "App 123",
					//Devices is not kept around!
				},
				Id:     "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				Active: true,
				Stats: model.Stats{
					model.DeviceDeploymentStatusDownloadingStr: 0,
					model.DeviceDeploymentStatusInstallingStr:  0,
					model.DeviceDeploymentStatusRebootingStr:   0,
					model.DeviceDeploymentStatusPendingStr:     10,
					model.DeviceDeploymentStatusSuccessStr:     15,
					model.DeviceDeploymentStatusFailureStr:     1,
					model.DeviceDeploymentStatusNoArtifactStr:  0,
					model.DeviceDeploymentStatusAlreadyInstStr: 0,
					model.DeviceDeploymentStatusAbortedStr:     0,
				},
			},
		},
		"ok with tenant": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeploymentsCollection: bson.A{
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id:    "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
					Stats: model.Stats{},
				},
			},
			InputTenant: "acme",

			OutputDeployment: &model.Deployment{
				DeploymentConstructor: &model.DeploymentConstructor{
					Name:         "NYC Production",
					ArtifactName: "App 123",
					//Devices is not kept around!
				},
				Id:     "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				Active: true,
				Stats:  model.Stats{},
			},
		},
	}

	for testCaseName, testCase := range testCases {
		t.Run(fmt.Sprintf("test case %s", testCaseName), func(t *testing.T) {

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

			collDep := client.Database(mstore.
				DbFromContext(ctx, DatabaseName)).
				Collection(CollectionDeployments)
			if testCase.InputDeploymentsCollection != nil {
				_, err := collDep.InsertMany(
					ctx, testCase.InputDeploymentsCollection)
				assert.NoError(t, err)
			}

			deployment, err := store.FindDeploymentByID(ctx, testCase.InputID)

			if testCase.OutputError != nil {
				assert.EqualError(t, err, testCase.OutputError.Error())
			} else {
				assert.NoError(t, err)
				if deployment != nil && assert.Equal(t, 0, len(deployment.Artifacts)) {
					deployment.Artifacts = nil
				}
				assert.Equal(t, testCase.OutputDeployment, deployment)
			}

			// tenant is set, verify that deployment is not present in default DB
			if testCase.InputTenant != "" {
				deployment, err := store.FindDeploymentByID(context.Background(),
					testCase.InputID)
				assert.Nil(t, deployment)
				assert.Nil(t, err)
			}
		})
	}
}

func TestDeploymentStorageFindUnfinishedByID(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping TestDeploymentStorageFindUnfinishedByID in short mode.")
	}
	now := time.Now()

	testCases := map[string]struct {
		InputID                    string
		InputDeploymentsCollection []interface{}
		InputTenant                string

		OutputError      error
		OutputDeployment *model.Deployment
	}{
		"empty ID": {
			InputID:     "",
			OutputError: ErrStorageInvalidID,
		},
		"empty database": {
			InputID:          "b532b01a-9313-404f-8d19-e7fcbe5cc347",
			OutputError:      nil,
			OutputDeployment: nil,
		},
		"no deployments with given ID": {
			InputID: "b532b01a-9313-404f-8d19-e7fcbe5cc347",
			InputDeploymentsCollection: []interface{}{
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				},
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: "d1804903-5caa-4a73-a3ae-0efcc3205405",
				},
			},
			OutputError:      nil,
			OutputDeployment: nil,
		},
		"all correct": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeploymentsCollection: []interface{}{
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
					Stats: model.Stats{
						model.DeviceDeploymentStatusDownloadingStr: 0,
						model.DeviceDeploymentStatusInstallingStr:  0,
						model.DeviceDeploymentStatusRebootingStr:   0,
						model.DeviceDeploymentStatusPendingStr:     10,
						model.DeviceDeploymentStatusSuccessStr:     15,
						model.DeviceDeploymentStatusFailureStr:     1,
						model.DeviceDeploymentStatusNoArtifactStr:  0,
						model.DeviceDeploymentStatusAlreadyInstStr: 0,
						model.DeviceDeploymentStatusAbortedStr:     0,
					},
				},
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: "d1804903-5caa-4a73-a3ae-0efcc3205405",
					Stats: model.Stats{
						model.DeviceDeploymentStatusDownloadingStr: 0,
						model.DeviceDeploymentStatusInstallingStr:  0,
						model.DeviceDeploymentStatusRebootingStr:   0,
						model.DeviceDeploymentStatusPendingStr:     5,
						model.DeviceDeploymentStatusSuccessStr:     10,
						model.DeviceDeploymentStatusFailureStr:     3,
						model.DeviceDeploymentStatusNoArtifactStr:  0,
						model.DeviceDeploymentStatusAlreadyInstStr: 0,
						model.DeviceDeploymentStatusAbortedStr:     0,
					},
				},
			},
			OutputError: nil,
			OutputDeployment: &model.Deployment{
				DeploymentConstructor: &model.DeploymentConstructor{
					Name:         "NYC Production",
					ArtifactName: "App 123",
					//Devices is not kept around!
				},
				Id:     "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				Active: true,
				Stats: model.Stats{
					model.DeviceDeploymentStatusDownloadingStr: 0,
					model.DeviceDeploymentStatusInstallingStr:  0,
					model.DeviceDeploymentStatusRebootingStr:   0,
					model.DeviceDeploymentStatusPendingStr:     10,
					model.DeviceDeploymentStatusSuccessStr:     15,
					model.DeviceDeploymentStatusFailureStr:     1,
					model.DeviceDeploymentStatusNoArtifactStr:  0,
					model.DeviceDeploymentStatusAlreadyInstStr: 0,
					model.DeviceDeploymentStatusAbortedStr:     0,
				},
			},
		},
		"deployment already finished": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeploymentsCollection: []interface{}{
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id:       "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
					Finished: &now,
				},
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: "d1804903-5caa-4a73-a3ae-0efcc3205405",
				},
			},
			OutputError:      nil,
			OutputDeployment: nil,
		},
		"multi tenant, deployment found": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeploymentsCollection: []interface{}{
				&model.Deployment{
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
					Stats: model.Stats{
						model.DeviceDeploymentStatusPendingStr: 10,
						model.DeviceDeploymentStatusSuccessStr: 15,
						model.DeviceDeploymentStatusFailureStr: 1,
					},
				},
			},
			InputTenant: "acme",
			OutputError: nil,
			OutputDeployment: &model.Deployment{
				DeploymentConstructor: &model.DeploymentConstructor{
					Name:         "NYC Production",
					ArtifactName: "App 123",
					//Devices is not kept around!
				},
				Id:     "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				Active: true,
				Stats: model.Stats{
					model.DeviceDeploymentStatusPendingStr: 10,
					model.DeviceDeploymentStatusSuccessStr: 15,
					model.DeviceDeploymentStatusFailureStr: 1,
				},
			},
		},
	}

	for testCaseName, testCase := range testCases {
		t.Run(fmt.Sprintf("test case %s", testCaseName), func(t *testing.T) {

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

			collDep := client.Database(mstore.
				DbFromContext(ctx, DatabaseName)).
				Collection(CollectionDeployments)

			if testCase.InputDeploymentsCollection != nil {
				_, err := collDep.InsertMany(
					ctx, testCase.InputDeploymentsCollection)
				assert.NoError(t, err)
			}

			deployment, err := store.FindUnfinishedByID(ctx, testCase.InputID)

			if testCase.OutputError != nil {
				assert.EqualError(t, err, testCase.OutputError.Error())
			} else {
				assert.NoError(t, err)
				if deployment != nil && assert.Equal(t, 0, len(deployment.Artifacts)) {
					deployment.Artifacts = nil
				}
				assert.Equal(t, testCase.OutputDeployment, deployment)
			}

			// tenant is set, verify that deployment is not present in default DB
			if testCase.InputTenant != "" {
				deployment, err := store.FindUnfinishedByID(context.Background(),
					testCase.InputID)
				assert.Nil(t, deployment)
				assert.Nil(t, err)
			}
		})
	}
}

func TestDeploymentStorageUpdateStatsInc(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestDeploymentStorageUpdateStatsInc in short mode.")
	}

	testCases := map[string]struct {
		InputID         string
		InputDeployment *model.Deployment
		InputTenant     string

		InputStateFrom model.DeviceDeploymentStatus
		InputStateTo   model.DeviceDeploymentStatus

		OutputError error
		OutputStats model.Stats
	}{
		"pending -> finished": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeployment: &model.Deployment{
				Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				Stats: model.Stats{
					model.DeviceDeploymentStatusDownloadingStr: 1,
					model.DeviceDeploymentStatusInstallingStr:  2,
					model.DeviceDeploymentStatusRebootingStr:   3,
					model.DeviceDeploymentStatusPendingStr:     10,
					model.DeviceDeploymentStatusSuccessStr:     15,
					model.DeviceDeploymentStatusFailureStr:     4,
					model.DeviceDeploymentStatusNoArtifactStr:  5,
					model.DeviceDeploymentStatusAlreadyInstStr: 0,
					model.DeviceDeploymentStatusAbortedStr:     0,
				},
			},
			InputStateFrom: model.DeviceDeploymentStatusPending,
			InputStateTo:   model.DeviceDeploymentStatusSuccess,

			OutputError: nil,
			OutputStats: model.Stats{
				model.DeviceDeploymentStatusDownloadingStr: 1,
				model.DeviceDeploymentStatusInstallingStr:  2,
				model.DeviceDeploymentStatusRebootingStr:   3,
				model.DeviceDeploymentStatusPendingStr:     9,
				model.DeviceDeploymentStatusSuccessStr:     16,
				model.DeviceDeploymentStatusFailureStr:     4,
				model.DeviceDeploymentStatusNoArtifactStr:  5,
				model.DeviceDeploymentStatusAlreadyInstStr: 0,
				model.DeviceDeploymentStatusAbortedStr:     0,
			},
		},
		"rebooting -> failed": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeployment: &model.Deployment{
				Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				Stats: model.Stats{
					model.DeviceDeploymentStatusDownloadingStr: 1,
					model.DeviceDeploymentStatusInstallingStr:  2,
					model.DeviceDeploymentStatusRebootingStr:   3,
					model.DeviceDeploymentStatusPendingStr:     10,
					model.DeviceDeploymentStatusSuccessStr:     15,
					model.DeviceDeploymentStatusFailureStr:     4,
					model.DeviceDeploymentStatusNoArtifactStr:  5,
					model.DeviceDeploymentStatusAlreadyInstStr: 0,
					model.DeviceDeploymentStatusAbortedStr:     0,
				},
			},
			InputStateFrom: model.DeviceDeploymentStatusRebooting,
			InputStateTo:   model.DeviceDeploymentStatusFailure,

			OutputError: nil,
			OutputStats: model.Stats{
				model.DeviceDeploymentStatusDownloadingStr: 1,
				model.DeviceDeploymentStatusInstallingStr:  2,
				model.DeviceDeploymentStatusRebootingStr:   2,
				model.DeviceDeploymentStatusPendingStr:     10,
				model.DeviceDeploymentStatusSuccessStr:     15,
				model.DeviceDeploymentStatusFailureStr:     5,
				model.DeviceDeploymentStatusNoArtifactStr:  5,
				model.DeviceDeploymentStatusAlreadyInstStr: 0,
				model.DeviceDeploymentStatusAbortedStr:     0,
			},
		},
		"invalid deployment id": {
			InputID:         "",
			InputDeployment: nil,
			InputStateFrom:  model.DeviceDeploymentStatusRebooting,
			InputStateTo:    model.DeviceDeploymentStatusFailure,

			OutputError: ErrStorageInvalidID,
			OutputStats: nil,
		},
		"wrong deployment id": {
			InputID:         "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeployment: nil,
			InputStateFrom:  model.DeviceDeploymentStatusRebooting,
			InputStateTo:    model.DeviceDeploymentStatusFailure,

			OutputError: ErrStorageInvalidID,
			OutputStats: nil,
		},
		"no old state": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeployment: &model.Deployment{
				Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				Stats: model.Stats{
					model.DeviceDeploymentStatusDownloadingStr: 1,
					model.DeviceDeploymentStatusInstallingStr:  2,
					model.DeviceDeploymentStatusRebootingStr:   3,
					model.DeviceDeploymentStatusPendingStr:     10,
					model.DeviceDeploymentStatusSuccessStr:     15,
					model.DeviceDeploymentStatusFailureStr:     4,
					model.DeviceDeploymentStatusNoArtifactStr:  5,
					model.DeviceDeploymentStatusAlreadyInstStr: 0,
					model.DeviceDeploymentStatusAbortedStr:     0,
				},
			},
			InputStateFrom: model.DeviceDeploymentStatusNull,
			InputStateTo:   model.DeviceDeploymentStatusPending,

			OutputStats: model.Stats{
				model.DeviceDeploymentStatusDownloadingStr: 1,
				model.DeviceDeploymentStatusInstallingStr:  2,
				model.DeviceDeploymentStatusRebootingStr:   3,
				model.DeviceDeploymentStatusPendingStr:     11,
				model.DeviceDeploymentStatusSuccessStr:     15,
				model.DeviceDeploymentStatusFailureStr:     4,
				model.DeviceDeploymentStatusNoArtifactStr:  5,
				model.DeviceDeploymentStatusAlreadyInstStr: 0,
				model.DeviceDeploymentStatusAbortedStr:     0,
			},
		},
		"install install": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeployment: &model.Deployment{
				Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				Stats: model.Stats{
					model.DeviceDeploymentStatusDownloadingStr: 1,
					model.DeviceDeploymentStatusInstallingStr:  2,
					model.DeviceDeploymentStatusRebootingStr:   3,
					model.DeviceDeploymentStatusPendingStr:     10,
					model.DeviceDeploymentStatusSuccessStr:     15,
					model.DeviceDeploymentStatusFailureStr:     4,
					model.DeviceDeploymentStatusNoArtifactStr:  5,
					model.DeviceDeploymentStatusAlreadyInstStr: 0,
					model.DeviceDeploymentStatusAbortedStr:     0,
				},
			},
			InputStateFrom: model.DeviceDeploymentStatusInstalling,
			InputStateTo:   model.DeviceDeploymentStatusInstalling,

			OutputError: nil,
			OutputStats: nil,
		},
		"tenant, pending -> finished": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeployment: &model.Deployment{
				Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				Stats: model.Stats{
					model.DeviceDeploymentStatusDownloadingStr: 1,
					model.DeviceDeploymentStatusInstallingStr:  2,
					model.DeviceDeploymentStatusRebootingStr:   3,
					model.DeviceDeploymentStatusPendingStr:     10,
					model.DeviceDeploymentStatusSuccessStr:     15,
					model.DeviceDeploymentStatusFailureStr:     4,
					model.DeviceDeploymentStatusNoArtifactStr:  5,
					model.DeviceDeploymentStatusAlreadyInstStr: 0,
					model.DeviceDeploymentStatusAbortedStr:     0,
				},
			},
			InputTenant: "acme",

			InputStateFrom: model.DeviceDeploymentStatusPending,
			InputStateTo:   model.DeviceDeploymentStatusSuccess,

			OutputError: nil,
			OutputStats: model.Stats{
				model.DeviceDeploymentStatusDownloadingStr: 1,
				model.DeviceDeploymentStatusInstallingStr:  2,
				model.DeviceDeploymentStatusRebootingStr:   3,
				model.DeviceDeploymentStatusPendingStr:     9,
				model.DeviceDeploymentStatusSuccessStr:     16,
				model.DeviceDeploymentStatusFailureStr:     4,
				model.DeviceDeploymentStatusNoArtifactStr:  5,
				model.DeviceDeploymentStatusAlreadyInstStr: 0,
				model.DeviceDeploymentStatusAbortedStr:     0,
			},
		},
	}

	for testCaseName, tc := range testCases {
		t.Run(fmt.Sprintf("test case %s", testCaseName), func(t *testing.T) {

			db.Wipe()

			client := db.Client()
			store := NewDataStoreMongoWithClient(client)

			ctx := context.Background()
			if tc.InputTenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.InputTenant,
				})
			} else {
				ctx = context.Background()
			}

			if tc.InputDeployment != nil {
				collDefaultDep := client.Database(DatabaseName).
					Collection(CollectionDeployments)
				_, err := collDefaultDep.InsertOne(ctx,
					tc.InputDeployment)
				assert.NoError(t, err)
				// multi tenant test only makes sense if there
				// is a deployment to input, if there's one
				// we'll add it to tenant's DB
				if tc.InputTenant != "" {
					collDep := client.Database(mstore.
						DbFromContext(ctx,
							DatabaseName)).
						Collection(CollectionDeployments)

					_, err = collDep.InsertOne(ctx,
						tc.InputDeployment)
					assert.NoError(t, err)
				}
			}

			stats, err := store.UpdateStatsInc(ctx,
				tc.InputID, tc.InputStateFrom, tc.InputStateTo)

			if tc.OutputError != nil {
				assert.EqualError(t, err, tc.OutputError.Error())
			} else {
				assert.Equal(t, tc.OutputStats, stats)

				// if there's a tenant, verify that deployment
				// in default DB remains unchanged, again only
				// makes sense if there's an input deployment
				if tc.InputTenant != "" && tc.InputDeployment != nil {
					var defDeployment *model.Deployment
					collDefaultDep := client.
						Database(DatabaseName).
						Collection(CollectionDeployments)
					err := collDefaultDep.FindOne(ctx,
						bson.M{"_id": tc.InputID}).
						Decode(&defDeployment)
					assert.NoError(t, err)
					assert.Equal(t, defDeployment.Stats, tc.InputDeployment.Stats)
				}

			}
		})
	}
}

func TestDeploymentStorageUpdateStats(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestDeploymentStorageUpdateStats in short mode.")
	}

	testCases := map[string]struct {
		id     string
		dep    *model.Deployment
		stats  model.Stats
		tenant string

		err error
	}{
		"all correct": {
			id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			dep: &model.Deployment{
				Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				Stats: model.Stats{
					model.DeviceDeploymentStatusDownloadingStr: 1,
					model.DeviceDeploymentStatusInstallingStr:  2,
					model.DeviceDeploymentStatusRebootingStr:   3,
					model.DeviceDeploymentStatusPendingStr:     3,
					model.DeviceDeploymentStatusSuccessStr:     6,
					model.DeviceDeploymentStatusFailureStr:     8,
					model.DeviceDeploymentStatusNoArtifactStr:  4,
					model.DeviceDeploymentStatusAlreadyInstStr: 2,
					model.DeviceDeploymentStatusAbortedStr:     5,
				},
			},
			stats: model.Stats{
				model.DeviceDeploymentStatusDownloadingStr: 1,
				model.DeviceDeploymentStatusInstallingStr:  2,
				model.DeviceDeploymentStatusRebootingStr:   3,
				model.DeviceDeploymentStatusPendingStr:     10,
				model.DeviceDeploymentStatusSuccessStr:     15,
				model.DeviceDeploymentStatusFailureStr:     4,
				model.DeviceDeploymentStatusNoArtifactStr:  5,
				model.DeviceDeploymentStatusAlreadyInstStr: 0,
				model.DeviceDeploymentStatusAbortedStr:     5,
			},
		},
		"invalid deployment id": {
			id:    "",
			dep:   nil,
			stats: nil,

			err: ErrStorageInvalidID,
		},
		"wrong deployment id": {
			id:    "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			dep:   nil,
			stats: nil,

			err: ErrStorageInvalidID,
		},
		"tenant, all correct": {
			id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			dep: &model.Deployment{
				Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				Stats: newTestStats(model.Stats{
					model.DeviceDeploymentStatusRebootingStr: 3,
				}),
			},
			stats: newTestStats(model.Stats{
				model.DeviceDeploymentStatusRebootingStr: 3,
			}),
			tenant: "acme",
		},
	}

	for name, tc := range testCases {
		t.Run(fmt.Sprintf("test case %s", name), func(t *testing.T) {

			db.Wipe()

			client := db.Client()
			store := NewDataStoreMongoWithClient(client)

			ctx := context.Background()
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.tenant,
				})
			}

			collDep := client.Database(mstore.
				DbFromContext(ctx, DatabaseName)).
				Collection(CollectionDeployments)
			if tc.dep != nil {
				_, err := collDep.InsertOne(
					ctx, tc.dep)
				assert.NoError(t, err)
			}

			err := store.UpdateStats(ctx, tc.id, tc.stats)

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				var deployment *model.Deployment
				err := collDep.FindOne(ctx,
					bson.M{"_id": tc.id}).
					Decode(&deployment)
				assert.NoError(t, err)
				assert.Equal(t, tc.stats, deployment.Stats)
			}

			if tc.tenant != "" && tc.dep != nil {
				err := store.UpdateStats(context.Background(),
					tc.id, tc.stats)
				assert.EqualError(t, err, ErrStorageInvalidID.Error())
			}
		})
	}

}

func newTestStats(stats model.Stats) model.Stats {
	st := model.NewDeviceDeploymentStats()
	for k, v := range stats {
		st[k] = v
	}
	return st
}

func TestDeploymentStorageFindBy(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping TestDeploymentStorageFindBy in short mode.")
	}

	now := time.Now()

	someDeployments := []*model.Deployment{
		{
			DeploymentConstructor: &model.DeploymentConstructor{
				Name:         "NYC Production Inc.",
				ArtifactName: "App 123",
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: "a108ae14-bb4e-455f-9b40-000000000001",
			Stats: newTestStats(model.Stats{
				model.DeviceDeploymentStatusNoArtifactStr: 1,
			}),
			Status:   model.DeploymentStatusFinished,
			Finished: &now,
		},
		{
			DeploymentConstructor: &model.DeploymentConstructor{
				Name:         "NYC Production Inc.",
				ArtifactName: "App 123",
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: "a108ae14-bb4e-455f-9b40-000000000002",
			Stats: newTestStats(model.Stats{
				model.DeviceDeploymentStatusNoArtifactStr: 1,
			}),
			Status:   model.DeploymentStatusFinished,
			Finished: &now,
		},
		{
			DeploymentConstructor: &model.DeploymentConstructor{
				Name:         "foo",
				ArtifactName: "bar",
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: "a108ae14-bb4e-455f-9b40-000000000003",
			Stats: newTestStats(model.Stats{
				model.DeviceDeploymentStatusFailureStr: 2,
			}),
			Status:   model.DeploymentStatusFinished,
			Finished: &now,
		},
		{
			DeploymentConstructor: &model.DeploymentConstructor{
				Name:         "foo",
				ArtifactName: "bar",
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: "a108ae14-bb4e-455f-9b40-000000000004",
			Stats: newTestStats(model.Stats{
				model.DeviceDeploymentStatusNoArtifactStr: 1,
			}),
			Status:   model.DeploymentStatusFinished,
			Finished: &now,
		},
		{
			DeploymentConstructor: &model.DeploymentConstructor{
				Name:         "foo",
				ArtifactName: "bar",
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: "a108ae14-bb4e-455f-9b40-000000000005",
			Stats: newTestStats(model.Stats{
				model.DeviceDeploymentStatusDownloadingStr: 1,
			}),
			Status: model.DeploymentStatusInProgress,
		},
		{
			DeploymentConstructor: &model.DeploymentConstructor{
				Name:         "zed",
				ArtifactName: "daz",
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: "a108ae14-bb4e-455f-9b40-000000000006",
			Stats: newTestStats(model.Stats{
				model.DeviceDeploymentStatusDownloadingStr: 1,
				model.DeviceDeploymentStatusPendingStr:     1,
			}),
			Status: model.DeploymentStatusInProgress,
		},
		{
			DeploymentConstructor: &model.DeploymentConstructor{
				Name:         "zed",
				ArtifactName: "daz",
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: "a108ae14-bb4e-455f-9b40-000000000007",
			Stats: newTestStats(model.Stats{
				model.DeviceDeploymentStatusPendingStr: 1,
			}),
			Status: model.DeploymentStatusPending,
		},
		{
			DeploymentConstructor: &model.DeploymentConstructor{
				Name:         "zed",
				ArtifactName: "daz",
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: "a108ae14-bb4e-455f-9b40-000000000008",
			Stats: newTestStats(model.Stats{
				model.DeviceDeploymentStatusNoArtifactStr: 1,
				model.DeviceDeploymentStatusSuccessStr:    1,
			}),
			Status:   model.DeploymentStatusFinished,
			Finished: &now,
		},
		{
			DeploymentConstructor: &model.DeploymentConstructor{
				Name:         "123",
				ArtifactName: "dfs",
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc34a"},
			},
			Id: "a108ae14-bb4e-455f-9b40-000000000009",
			Stats: newTestStats(model.Stats{
				model.DeviceDeploymentStatusAbortedStr: 1,
			}),
			Status:   model.DeploymentStatusFinished,
			Finished: &now,
		},

		//in progress deployment, with only pending and already-installed counters > 0
		{
			DeploymentConstructor: &model.DeploymentConstructor{
				Name:         "baz",
				ArtifactName: "asdf",
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: "a108ae14-bb4e-455f-9b40-000000000010",
			Stats: newTestStats(model.Stats{
				model.DeviceDeploymentStatusPendingStr:     1,
				model.DeviceDeploymentStatusAlreadyInstStr: 1,
			}),
			Status: model.DeploymentStatusInProgress,
		},
		//in progress deployment, with only pending and success counters > 0
		{
			DeploymentConstructor: &model.DeploymentConstructor{
				Name:         "baz",
				ArtifactName: "asdf",
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: "a108ae14-bb4e-455f-9b40-000000000011",
			Stats: newTestStats(model.Stats{
				model.DeviceDeploymentStatusPendingStr: 1,
				model.DeviceDeploymentStatusSuccessStr: 1,
			}),
			Status: model.DeploymentStatusInProgress,
		},
		//in progress deployment, with only pending and failure counters > 0
		{
			DeploymentConstructor: &model.DeploymentConstructor{
				Name:         "baz",
				ArtifactName: "asdf",
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: "a108ae14-bb4e-455f-9b40-000000000012",
			Stats: newTestStats(model.Stats{
				model.DeviceDeploymentStatusPendingStr: 1,
				model.DeviceDeploymentStatusFailureStr: 1,
			}),
			Status: model.DeploymentStatusInProgress,
		},
		//in progress deployment, with only pending and noartifact counters > 0
		{
			DeploymentConstructor: &model.DeploymentConstructor{
				Name:         "baz",
				ArtifactName: "asdf",
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: "a108ae14-bb4e-455f-9b40-000000000013",
			Stats: newTestStats(model.Stats{
				model.DeviceDeploymentStatusPendingStr:    1,
				model.DeviceDeploymentStatusNoArtifactStr: 1,
			}),
			Status: model.DeploymentStatusInProgress,
		},
		//finished deployment, with only already installed counter > 0
		{
			DeploymentConstructor: &model.DeploymentConstructor{
				Name:         "baz",
				ArtifactName: "asdf",
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: "a108ae14-bb4e-455f-9b40-000000000014",
			Stats: newTestStats(model.Stats{
				model.DeviceDeploymentStatusAlreadyInstStr: 1,
			}),
			Status:   model.DeploymentStatusFinished,
			Finished: &now,
		},
		//configuration deployment
		{
			DeploymentConstructor: &model.DeploymentConstructor{
				Name:         "zed",
				ArtifactName: "daz",
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: "a108ae14-bb4e-455f-9b40-000000000015",
			Stats: newTestStats(model.Stats{
				model.DeviceDeploymentStatusPendingStr: 1,
			}),
			Status: model.DeploymentStatusPending,
			Type:   model.DeploymentTypeConfiguration,
		},
	}

	testCases := []struct {
		InputModelQuery            model.Query
		InputDeploymentsCollection []*model.Deployment
		InputTenant                string

		OutputError error
		OutputID    []string
	}{
		{
			InputModelQuery: model.Query{
				SearchText: "foobar-empty-db",
			},
			OutputError: ErrDeploymentStorageCannotExecQuery,
		},
		{
			InputModelQuery: model.Query{
				SearchText: "foobar-no-match",
			},
			InputDeploymentsCollection: []*model.Deployment{
				{
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				},
			},
		},
		{
			InputModelQuery: model.Query{
				SearchText: "NYC",
			},
			InputDeploymentsCollection: someDeployments,
			OutputError:                nil,
			OutputID: []string{
				"a108ae14-bb4e-455f-9b40-000000000002",
				"a108ae14-bb4e-455f-9b40-000000000001",
			},
		},
		{
			InputModelQuery: model.Query{
				SearchText: "NYC Production Inc.",
			},
			InputDeploymentsCollection: someDeployments,
			OutputError:                nil,
			OutputID: []string{
				"a108ae14-bb4e-455f-9b40-000000000002",
				"a108ae14-bb4e-455f-9b40-000000000001",
			},
		},
		{
			InputModelQuery: model.Query{
				SearchText: "bar",
			},
			InputDeploymentsCollection: someDeployments,
			OutputError:                nil,
			OutputID: []string{
				"a108ae14-bb4e-455f-9b40-000000000005",
				"a108ae14-bb4e-455f-9b40-000000000004",
				"a108ae14-bb4e-455f-9b40-000000000003",
			},
		},
		{
			InputModelQuery: model.Query{
				SearchText: "bar",
				Status:     model.StatusQueryInProgress,
			},
			InputDeploymentsCollection: someDeployments,
			OutputError:                nil,
			OutputID: []string{
				"a108ae14-bb4e-455f-9b40-000000000005",
			},
		},
		{
			InputModelQuery: model.Query{
				SearchText: "bar",
				Status:     model.StatusQueryFinished,
			},
			InputDeploymentsCollection: someDeployments,
			OutputError:                nil,
			OutputID: []string{
				"a108ae14-bb4e-455f-9b40-000000000004",
				"a108ae14-bb4e-455f-9b40-000000000003",
			},
		},
		{
			InputModelQuery: model.Query{
				Status: model.StatusQueryInProgress,
			},
			InputDeploymentsCollection: someDeployments,
			OutputError:                nil,
			OutputID: []string{
				"a108ae14-bb4e-455f-9b40-000000000013",
				"a108ae14-bb4e-455f-9b40-000000000012",
				"a108ae14-bb4e-455f-9b40-000000000011",
				"a108ae14-bb4e-455f-9b40-000000000010",
				"a108ae14-bb4e-455f-9b40-000000000006",
				"a108ae14-bb4e-455f-9b40-000000000005",
			},
		},
		{
			InputModelQuery: model.Query{
				Status: model.StatusQueryPending,
			},
			InputDeploymentsCollection: someDeployments,
			OutputError:                nil,
			OutputID: []string{
				"a108ae14-bb4e-455f-9b40-000000000015",
				"a108ae14-bb4e-455f-9b40-000000000007",
			},
		},
		{
			InputModelQuery: model.Query{
				Status: model.StatusQueryFinished,
			},
			InputDeploymentsCollection: someDeployments,
			OutputError:                nil,
			OutputID: []string{
				"a108ae14-bb4e-455f-9b40-000000000014",
				"a108ae14-bb4e-455f-9b40-000000000009",
				"a108ae14-bb4e-455f-9b40-000000000008",
				"a108ae14-bb4e-455f-9b40-000000000004",
				"a108ae14-bb4e-455f-9b40-000000000003",
				"a108ae14-bb4e-455f-9b40-000000000002",
				"a108ae14-bb4e-455f-9b40-000000000001",
			},
		},
		{
			InputModelQuery: model.Query{
				// whatever name
				SearchText: "",
				// any status
				Status: model.StatusQueryAny,
			},
			InputDeploymentsCollection: someDeployments,
			OutputError:                nil,
			OutputID: []string{
				"a108ae14-bb4e-455f-9b40-000000000015",
				"a108ae14-bb4e-455f-9b40-000000000014",
				"a108ae14-bb4e-455f-9b40-000000000013",
				"a108ae14-bb4e-455f-9b40-000000000012",
				"a108ae14-bb4e-455f-9b40-000000000011",
				"a108ae14-bb4e-455f-9b40-000000000010",
				"a108ae14-bb4e-455f-9b40-000000000009",
				"a108ae14-bb4e-455f-9b40-000000000008",
				"a108ae14-bb4e-455f-9b40-000000000007",
				"a108ae14-bb4e-455f-9b40-000000000006",
				"a108ae14-bb4e-455f-9b40-000000000005",
				"a108ae14-bb4e-455f-9b40-000000000004",
				"a108ae14-bb4e-455f-9b40-000000000003",
				"a108ae14-bb4e-455f-9b40-000000000002",
				"a108ae14-bb4e-455f-9b40-000000000001",
			},
		},
		{
			InputModelQuery: model.Query{
				// whatever name
				SearchText: "",
				// any status
				Status: model.StatusQueryAny,
				Limit:  2,
			},
			InputDeploymentsCollection: someDeployments,
			OutputError:                nil,
			OutputID: []string{
				"a108ae14-bb4e-455f-9b40-000000000015",
				"a108ae14-bb4e-455f-9b40-000000000014",
			},
		},
		{
			InputModelQuery: model.Query{
				// whatever name
				SearchText: "",
				// any status
				Status: model.StatusQueryAny,
				Limit:  2,
				Skip:   2,
			},
			InputDeploymentsCollection: someDeployments,
			OutputError:                nil,
			OutputID: []string{
				"a108ae14-bb4e-455f-9b40-000000000013",
				"a108ae14-bb4e-455f-9b40-000000000012",
			},
		},
		{
			InputModelQuery: model.Query{
				SearchText: "NYC",
			},
			InputDeploymentsCollection: someDeployments,
			InputTenant:                "acme",
			OutputError:                nil,
			OutputID: []string{
				"a108ae14-bb4e-455f-9b40-000000000002",
				"a108ae14-bb4e-455f-9b40-000000000001",
			},
		},
		{
			InputModelQuery: model.Query{
				Type: model.DeploymentTypeConfiguration,
			},
			InputDeploymentsCollection: someDeployments,
			OutputError:                nil,
			OutputID: []string{
				"a108ae14-bb4e-455f-9b40-000000000015",
			},
		},
	}

	for testCaseNumber, testCase := range testCases {
		t.Run(fmt.Sprintf("test case %02d search %s", testCaseNumber+1,
			testCase.InputModelQuery.SearchText), func(t *testing.T) {

			t.Logf("testing search: '%s'", testCase.InputModelQuery.SearchText)
			t.Logf("        status: %v", testCase.InputModelQuery.Status)

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

			createdTime := time.Now().UTC()

			for _, d := range testCase.InputDeploymentsCollection {
				// setup created time so that input deployments
				// are created with 'created time' within a
				// minute from each other, this will ensure
				// proper ordering

				// Ensure that storage indexes are present
				// before inserting deployment.
				err := store.EnsureIndexes(
					mstore.DbFromContext(ctx, DatabaseName),
					CollectionDeployments, StorageIndexes)
				assert.NoError(t, err)
				d.Created = &createdTime
				assert.NoError(t, store.InsertDeployment(ctx, d))
				createdTime = createdTime.Add(time.Minute)
			}

			deps, _, err := store.FindDeployments(ctx,
				testCase.InputModelQuery)

			if testCase.OutputError != nil {
				assert.EqualError(t, err,
					testCase.OutputError.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, deps, len(testCase.OutputID))
				if out := assert.Len(t, deps, len(testCase.OutputID)); !out {
					t.FailNow()
				}

				// verify that order is as expected
				for idx, dep := range deps {
					// deployments must be listed in the same order
					assert.Equal(t, testCase.OutputID[idx],
						dep.Id,
						"unexpected deployment %v at position %v, expected %v",
						dep.Id, idx, testCase.OutputID[idx])
				}

				// output result should be stable
				otherDeps, _, _ := store.FindDeployments(ctx,
					testCase.InputModelQuery)
				assert.Equal(t, deps, otherDeps)

			}

			if testCase.InputTenant != "" {
				// have to add a deployment, otherwise, it won't
				// be possible to run find queries

				// Ensure that storage indexes are present
				// before creating deployment
				err := store.EnsureIndexes(DatabaseName,
					CollectionDeployments, StorageIndexes)
				assert.NoError(t, err)
				err = store.InsertDeployment(context.Background(),
					&model.Deployment{
						DeploymentConstructor: &model.DeploymentConstructor{
							Name:         "foo-" + testCase.InputTenant,
							ArtifactName: "bar-" + testCase.InputTenant,
							Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc399"},
						},
						Id:      "e8c32ff6-7c1b-43c7-aa31-2e4fc3a3c199",
						Stats:   newTestStats(model.Stats{}),
						Created: TimeToPointer(time.Now().UTC()),
					})
				assert.NoError(t, err)

				// tenant is set, so only tenant's DB was set
				// up, verify that we cannot find anything in
				// default DB
				deps, _, err := store.FindDeployments(context.Background(),
					testCase.InputModelQuery)
				assert.Len(t, deps, 0)
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeviceDeploymentCounting(t *testing.T) {
	testCases := []struct {
		InputDeploymentID     string
		InputDeviceDeployment []*model.DeviceDeployment
		DeviceCount           int
	}{
		{
			InputDeploymentID: "foo",
			InputDeviceDeployment: []*model.DeviceDeployment{
				{
					Id:           "foo",
					DeploymentId: "foo",
				},
			},
			DeviceCount: 1,
		},
		{
			InputDeploymentID: "bar",
			InputDeviceDeployment: []*model.DeviceDeployment{
				{
					Id:           "996cf733-a7d9-4e8c-823e-122be04d9e39",
					DeploymentId: "bar",
				},
				{
					Id:           "ced2feba-d0a9-4f89-8cda-dd6f749c67a1",
					DeploymentId: "bar",
				},
				{
					Id:           "9d333d96-80ee-45d3-96ef-1dd2776e0994",
					DeploymentId: "bar",
				},
				{
					Id:           "bba8dd14-7980-474f-a791-449f4dc67cf6",
					DeploymentId: "bar",
				},
			},
			DeviceCount: 4,
		},
		{
			InputDeploymentID:     "notfound",
			InputDeviceDeployment: []*model.DeviceDeployment{},
			DeviceCount:           0,
		},
	}

	for idx, tc := range testCases {
		t.Run(fmt.Sprintf("test_%d", idx), func(t *testing.T) {
			db.Wipe()
			client := db.Client()

			store := NewDataStoreMongoWithClient(client)
			ctx := context.Background()

			collDep := client.Database(mstore.
				DbFromContext(ctx, DatabaseName)).
				Collection(CollectionDevices)
			for _, d := range tc.InputDeviceDeployment {
				_, err := collDep.InsertOne(ctx, d)
				assert.NoError(t, err)
			}

			actualCount, err := store.DeviceCountByDeployment(ctx, tc.InputDeploymentID)
			assert.NoError(t, err)
			assert.Equal(t, tc.DeviceCount, actualCount)
		})
	}
}

func TestDeploymentSetStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestDeploymentSetStatus in short mode.")
	}

	id := "a108ae14-bb4e-455f-9b40-2ef4bab97bb7"
	deployment := &model.Deployment{
		Id: id,
	}

	now := time.Now().UTC()
	testCases := map[string]struct {
		tenant string

		status model.DeploymentStatus
	}{
		"pending": {
			status: model.DeploymentStatusPending,
		},
		"inprogress": {
			status: model.DeploymentStatusInProgress,
		},
		"finished, mt": {
			status: model.DeploymentStatusInProgress,
			tenant: "foo",
		},
		"finished": {
			status: model.DeploymentStatusFinished,
		},
	}

	db.Wipe()
	client := db.Client()
	store := NewDataStoreMongoWithClient(client)

	for testCaseName, tc := range testCases {
		t.Run(fmt.Sprintf("test case %s", testCaseName), func(t *testing.T) {

			db.Wipe()

			ctx := context.Background()
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.tenant,
				})
			}

			collDep := client.Database(mstore.
				DbFromContext(ctx, DatabaseName)).
				Collection(CollectionDeployments)

			_, err := collDep.InsertOne(ctx, deployment)
			assert.NoError(t, err)

			err = store.SetDeploymentStatus(ctx, id, tc.status, now)

			var deployment *model.Deployment
			err = collDep.FindOne(ctx,
				bson.M{"_id": id}).
				Decode(&deployment)
			assert.NoError(t, err)

			assert.Equal(t, tc.status, deployment.Status)
			if tc.status == model.DeploymentStatusFinished {
				// mongo trims time, no true equality
				assert.WithinDuration(t, now, *deployment.Finished, time.Second)
			}

			if tc.tenant != "" {
				err := store.SetDeploymentStatus(context.Background(), id, tc.status, now)
				assert.EqualError(t, err, ErrStorageInvalidID.Error())
			}
		})
	}
}
