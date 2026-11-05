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
	"github.com/stretchr/testify/require"

	"github.com/mendersoftware/mender-server/pkg/identity"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

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

			err := store.InsertDeployment(ctx, *testCase.InputDeployment)

			if testCase.OutputError != nil {
				assert.EqualError(t, err, testCase.OutputError.Error())
			} else {
				assert.NoError(t, err)
				deployments, _, err := store.FindDeployments(ctx, model.Query{})
				require.NoError(t, err)
				require.Len(t, deployments, 1)
				assert.Equal(t, deployments[0].Id, testCase.InputDeployment.Id)
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
		InputDeploymentsCollection []model.Deployment
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
			InputDeploymentsCollection: []model.Deployment{
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
			InputDeploymentsCollection: []model.Deployment{
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

			for _, deployment := range testCase.InputDeploymentsCollection {
				store.InsertDeployment(ctx, deployment)
			}

			err := store.DeleteDeployment(ctx, testCase.InputID)

			if testCase.OutputError != nil {
				assert.EqualError(t, err, testCase.OutputError.Error())
			} else {
				assert.NoError(t, err)

				_, count, err := store.FindDeployments(ctx, model.Query{IDs: []string{testCase.InputID}})
				assert.NoError(t, err)
				assert.Equal(t, 0, int(count))

				if testCase.InputTenant != "" {
					_, count, err := store.FindDeployments(context.Background(), model.Query{})
					assert.NoError(t, err)
					assert.Equal(t, 0, int(count))
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
		InputDeploymentsCollection []*model.Deployment
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
			InputDeploymentsCollection: []*model.Deployment{
				&model.Deployment{
					Created: new(time.Now()),
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				},
				&model.Deployment{
					Created: new(time.Now()),
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
			InputDeploymentsCollection: []*model.Deployment{
				&model.Deployment{
					Created: new(time.Now()),
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
					Statistics: model.DeploymentStatistics{Status: model.Stats{
						Pending: 10,
						Success: 15,
						Failure: 1,
					}},
				},
				&model.Deployment{
					Created: new(time.Now()),
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: "d1804903-5caa-4a73-a3ae-0efcc3205405",
					Statistics: model.DeploymentStatistics{Status: model.Stats{
						Pending: 5,
						Success: 10,
						Failure: 3,
					}},
				},
			},
			OutputError: nil,
			OutputDeployment: &model.Deployment{
				DeploymentConstructor: &model.DeploymentConstructor{
					Name:         "NYC Production",
					ArtifactName: "App 123",
					//Devices is not kept around!
				},
				Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				Statistics: model.DeploymentStatistics{Status: model.Stats{
					Pending: 10,
					Success: 15,
					Failure: 1,
				}},
			},
		},
		"ok with tenant": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeploymentsCollection: []*model.Deployment{
				&model.Deployment{
					Created: new(time.Now()),
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id:         "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
					Statistics: model.DeploymentStatistics{Status: model.Stats{}},
				},
			},
			InputTenant: "acme",

			OutputDeployment: &model.Deployment{
				DeploymentConstructor: &model.DeploymentConstructor{
					Name:         "NYC Production",
					ArtifactName: "App 123",
					//Devices is not kept around!
				},
				Id:         "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				Statistics: model.DeploymentStatistics{Status: model.Stats{}},
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

			for _, deployment := range testCase.InputDeploymentsCollection {
				require.NoError(t, store.InsertDeployment(ctx, *deployment))
			}

			deployment, err := store.FindDeploymentByID(ctx, testCase.InputID)

			if testCase.OutputError != nil {
				assert.EqualError(t, err, testCase.OutputError.Error())
			} else {
				assert.NoError(t, err)
				if deployment != nil {
					if assert.Equal(t, 0, len(deployment.Artifacts)) {
						deployment.Artifacts = nil
					}
					deployment.Created = nil
					deployment.Configuration = nil
					deployment.Groups = nil
					deployment.DeviceList = nil
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
		InputDeploymentsCollection []*model.Deployment
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
			InputDeploymentsCollection: []*model.Deployment{
				&model.Deployment{
					Created: new(time.Now()),
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				},
				&model.Deployment{
					Created: new(time.Now()),
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
			InputDeploymentsCollection: []*model.Deployment{
				&model.Deployment{
					Created: new(time.Now()),
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
					Statistics: model.DeploymentStatistics{Status: model.Stats{
						Pending: 10,
						Success: 15,
						Failure: 1,
					}},
				},
				&model.Deployment{
					Created: new(time.Now()),
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: "d1804903-5caa-4a73-a3ae-0efcc3205405",
					Statistics: model.DeploymentStatistics{Status: model.Stats{
						Pending: 5,
						Success: 10,
						Failure: 3,
					}},
				},
			},
			OutputError: nil,
			OutputDeployment: &model.Deployment{
				DeploymentConstructor: &model.DeploymentConstructor{
					Name:         "NYC Production",
					ArtifactName: "App 123",
					//Devices is not kept around!
				},
				Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				Statistics: model.DeploymentStatistics{Status: model.Stats{
					Pending: 10,
					Success: 15,
					Failure: 1,
				}},
			},
		},
		"deployment already finished": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeploymentsCollection: []*model.Deployment{
				&model.Deployment{
					Created: new(time.Now()),
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id:       "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
					Finished: &now,
				},
				&model.Deployment{
					Created: new(time.Now()),
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
			InputDeploymentsCollection: []*model.Deployment{
				&model.Deployment{
					Created: new(time.Now()),
					DeploymentConstructor: &model.DeploymentConstructor{
						Name:         "NYC Production",
						ArtifactName: "App 123",
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
					Statistics: model.DeploymentStatistics{Status: model.Stats{
						Pending: 10,
						Success: 15,
						Failure: 1,
					}},
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
				Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				Statistics: model.DeploymentStatistics{Status: model.Stats{
					Pending: 10,
					Success: 15,
					Failure: 1,
				}},
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

			for _, deployment := range testCase.InputDeploymentsCollection {
				require.NoError(t, store.InsertDeployment(ctx, *deployment))
			}

			deployment, err := store.FindUnfinishedByID(ctx, testCase.InputID)

			if testCase.OutputError != nil {
				assert.EqualError(t, err, testCase.OutputError.Error())
			} else {
				assert.NoError(t, err)
				if deployment != nil {
					assert.Equal(t, testCase.OutputDeployment.Id, deployment.Id)
				} else {
					assert.Equal(t, testCase.OutputDeployment, deployment)
				}
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
		OutputStats *model.Stats
	}{
		"pending -> finished": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeployment: &model.Deployment{
				Created: new(time.Now()),
				DeploymentConstructor: &model.DeploymentConstructor{
					Name:         "foo",
					ArtifactName: "foo",
				},
				Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				Statistics: model.DeploymentStatistics{Status: model.Stats{
					Downloading: 1,
					Installing:  2,
					Rebooting:   3,
					Pending:     10,
					Success:     15,
					Failure:     4,
					NoArtifact:  5,
				}},
			},
			InputStateFrom: model.DeviceDeploymentStatusPending,
			InputStateTo:   model.DeviceDeploymentStatusSuccess,

			OutputError: nil,
			OutputStats: &model.Stats{
				Downloading: 1,
				Installing:  2,
				Rebooting:   3,
				Pending:     9,
				Success:     16,
				Failure:     4,
				NoArtifact:  5,
			},
		},
		"rebooting -> failed": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeployment: &model.Deployment{
				Created: new(time.Now()),
				DeploymentConstructor: &model.DeploymentConstructor{
					Name:         "foo",
					ArtifactName: "foo",
				},
				Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				Statistics: model.DeploymentStatistics{Status: model.Stats{
					Downloading: 1,
					Installing:  2,
					Rebooting:   3,
					Pending:     10,
					Success:     15,
					Failure:     4,
					NoArtifact:  5,
				}},
			},
			InputStateFrom: model.DeviceDeploymentStatusRebooting,
			InputStateTo:   model.DeviceDeploymentStatusFailure,

			OutputError: nil,
			OutputStats: &model.Stats{
				Downloading: 1,
				Installing:  2,
				Rebooting:   2,
				Pending:     10,
				Success:     15,
				Failure:     5,
				NoArtifact:  5,
			},
		},
		"invalid deployment id": {
			InputID:         "",
			InputDeployment: nil,
			InputStateFrom:  model.DeviceDeploymentStatusRebooting,
			InputStateTo:    model.DeviceDeploymentStatusFailure,

			OutputError: ErrStorageInvalidID,
		},
		"wrong deployment id": {
			InputID:         "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeployment: nil,
			InputStateFrom:  model.DeviceDeploymentStatusRebooting,
			InputStateTo:    model.DeviceDeploymentStatusFailure,

			OutputError: ErrStorageInvalidID,
		},
		"no old state": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeployment: &model.Deployment{
				Created: new(time.Now()),
				DeploymentConstructor: &model.DeploymentConstructor{
					Name:         "foo",
					ArtifactName: "foo",
				},
				Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				Statistics: model.DeploymentStatistics{Status: model.Stats{
					Downloading: 1,
					Installing:  2,
					Rebooting:   3,
					Pending:     10,
					Success:     15,
					Failure:     4,
					NoArtifact:  5,
				}},
			},
			InputStateFrom: model.DeviceDeploymentStatusNull,
			InputStateTo:   model.DeviceDeploymentStatusPending,

			OutputStats: &model.Stats{
				Downloading: 1,
				Installing:  2,
				Rebooting:   3,
				Pending:     11,
				Success:     15,
				Failure:     4,
				NoArtifact:  5,
			},
		},
		"install install": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeployment: &model.Deployment{
				Created: new(time.Now()),
				DeploymentConstructor: &model.DeploymentConstructor{
					Name:         "foo",
					ArtifactName: "foo",
				},
				Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				Statistics: model.DeploymentStatistics{Status: model.Stats{
					Downloading: 1,
					Installing:  2,
					Rebooting:   3,
					Pending:     10,
					Success:     15,
					Failure:     4,
					NoArtifact:  5,
				}},
			},
			InputStateFrom: model.DeviceDeploymentStatusInstalling,
			InputStateTo:   model.DeviceDeploymentStatusInstalling,

			OutputError: nil,
		},
		"tenant, pending -> finished": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeployment: &model.Deployment{
				DeploymentConstructor: &model.DeploymentConstructor{
					Name:         "foo",
					ArtifactName: "foo",
				},
				Created: new(time.Now()),
				Id:      "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				Statistics: model.DeploymentStatistics{Status: model.Stats{
					Downloading: 1,
					Installing:  2,
					Rebooting:   3,
					Pending:     10,
					Success:     15,
					Failure:     4,
					NoArtifact:  5,
				}},
			},
			InputTenant: "acme",

			InputStateFrom: model.DeviceDeploymentStatusPending,
			InputStateTo:   model.DeviceDeploymentStatusSuccess,

			OutputError: nil,
			OutputStats: &model.Stats{
				Downloading: 1,
				Installing:  2,
				Rebooting:   3,
				Pending:     9,
				Success:     16,
				Failure:     4,
				NoArtifact:  5,
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
				require.NoError(t, store.InsertDeployment(ctx, *tc.InputDeployment))
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
					defDeployment, err := store.FindDeploymentByID(context.Background(), tc.InputID)
					require.NoError(t, err)
					assert.Empty(t, defDeployment)
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
				Created: new(time.Now()),
				DeploymentConstructor: &model.DeploymentConstructor{
					Name:         "foo",
					ArtifactName: "foo",
				},
				Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				Statistics: model.DeploymentStatistics{Status: model.Stats{
					Downloading:      1,
					Installing:       2,
					Rebooting:        3,
					Pending:          3,
					Success:          6,
					Failure:          8,
					NoArtifact:       4,
					AlreadyInstalled: 2,
					Aborted:          5,
				}},
			},
			stats: model.Stats{
				Downloading:      1,
				Installing:       2,
				Rebooting:        3,
				Pending:          10,
				Success:          15,
				Failure:          4,
				NoArtifact:       5,
				AlreadyInstalled: 0,
				Aborted:          5,
			},
		},
		"invalid deployment id": {
			id:  "",
			dep: nil,

			err: ErrStorageInvalidID,
		},
		"wrong deployment id": {
			id:  "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			dep: nil,

			err: ErrStorageInvalidID,
		},
		"tenant, all correct": {
			id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			dep: &model.Deployment{
				Created: new(time.Now()),
				DeploymentConstructor: &model.DeploymentConstructor{
					Name:         "foo",
					ArtifactName: "foo",
				},
				Id: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				Statistics: model.DeploymentStatistics{Status: model.Stats{
					Rebooting: 3,
				}},
			},
			stats: model.Stats{
				Rebooting: 3,
			},
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

			if tc.dep != nil {
				require.NoError(t, store.InsertDeployment(ctx, *tc.dep))
			}

			err := store.UpdateStats(ctx, tc.id, tc.stats)

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				var deployment *model.Deployment
				deployment, err = store.FindDeploymentByID(ctx, tc.id)
				assert.NoError(t, err)
				if deployment != nil {
					assert.Equal(t, tc.stats, deployment.Statistics.Status)
				}
			}

			if tc.tenant != "" && tc.dep != nil {
				err := store.UpdateStats(context.Background(),
					tc.id, tc.stats)
				assert.EqualError(t, err, ErrStorageInvalidID.Error())
			}
		})
	}

}

func TestDeploymentStorageFindBy(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping TestDeploymentStorageFindBy in short mode.")
	}
	t.Skip("FIXME: No text index for SearchText")
	t.SkipNow()

	now := time.Now()

	someDeployments := []*model.Deployment{
		{
			DeploymentConstructor: &model.DeploymentConstructor{
				Name:         "NYC Production Inc.",
				ArtifactName: "App 123",
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: "a108ae14-bb4e-455f-9b40-000000000001",
			Statistics: model.DeploymentStatistics{Status: model.Stats{
				NoArtifact: 1,
			}},
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
			Statistics: model.DeploymentStatistics{Status: model.Stats{
				NoArtifact: 1,
			}},
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
			Statistics: model.DeploymentStatistics{Status: model.Stats{
				Failure: 2,
			}},
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
			Statistics: model.DeploymentStatistics{Status: model.Stats{
				NoArtifact: 1,
			}},
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
			Statistics: model.DeploymentStatistics{Status: model.Stats{
				Downloading: 1,
			}},
			Status: model.DeploymentStatusInProgress,
		},
		{
			DeploymentConstructor: &model.DeploymentConstructor{
				Name:         "zed",
				ArtifactName: "daz",
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: "a108ae14-bb4e-455f-9b40-000000000006",
			Statistics: model.DeploymentStatistics{Status: model.Stats{
				Downloading: 1,
				Pending:     1,
			}},
			Status: model.DeploymentStatusInProgress,
		},
		{
			DeploymentConstructor: &model.DeploymentConstructor{
				Name:         "zed",
				ArtifactName: "daz",
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: "a108ae14-bb4e-455f-9b40-000000000007",
			Statistics: model.DeploymentStatistics{Status: model.Stats{
				Pending: 1,
			}},
			Status: model.DeploymentStatusPending,
		},
		{
			DeploymentConstructor: &model.DeploymentConstructor{
				Name:         "zed",
				ArtifactName: "daz",
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: "a108ae14-bb4e-455f-9b40-000000000008",
			Statistics: model.DeploymentStatistics{Status: model.Stats{
				NoArtifact: 1,
				Success:    1,
			}},
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
			Statistics: model.DeploymentStatistics{Status: model.Stats{
				Aborted: 1,
			}},
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
			Statistics: model.DeploymentStatistics{Status: model.Stats{
				Pending:          1,
				AlreadyInstalled: 1,
			}},
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
			Statistics: model.DeploymentStatistics{Status: model.Stats{
				Pending: 1,
				Success: 1,
			}},
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
			Statistics: model.DeploymentStatistics{Status: model.Stats{
				Pending: 1,
				Failure: 1,
			}},
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
			Statistics: model.DeploymentStatistics{Status: model.Stats{
				Pending:    1,
				NoArtifact: 1,
			}},
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
			Statistics: model.DeploymentStatistics{Status: model.Stats{
				AlreadyInstalled: 1,
			}},
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
			Statistics: model.DeploymentStatistics{Status: model.Stats{
				Pending: 1,
			}},
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

				d.Created = &createdTime
				assert.NoError(t, store.InsertDeployment(ctx, *d))
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

				err = store.InsertDeployment(ctx,
					model.Deployment{
						DeploymentConstructor: &model.DeploymentConstructor{
							Name:         "foo-" + testCase.InputTenant,
							ArtifactName: "bar-" + testCase.InputTenant,
							Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc399"},
						},
						Id:         "e8c32ff6-7c1b-43c7-aa31-2e4fc3a3c199",
						Statistics: model.DeploymentStatistics{Status: model.Stats{}},
						Created:    new(time.Now().UTC()),
					})
				assert.NoError(t, err)

				// tenant is set, so only tenant's DB was set
				// up, verify that we cannot find anything in
				// from default tenant namespace
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
		DeviceCount           int64
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

			for _, d := range tc.InputDeviceDeployment {
				require.NoError(t, store.InsertDeviceDeployment(ctx, d, false))
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
	deployment := model.Deployment{
		Id: id,
		DeploymentConstructor: &model.DeploymentConstructor{
			Name:         "foo",
			ArtifactName: "bar",
		},
		Created: new(time.Now()),
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

			err := store.InsertDeployment(ctx, deployment)
			assert.NoError(t, err)

			err = store.SetDeploymentStatus(ctx, id, tc.status, now)

			deployment, err := store.FindDeploymentByID(ctx, deployment.Id)
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
