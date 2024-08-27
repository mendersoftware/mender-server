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
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/mendersoftware/mender-server/pkg/identity"
	ctxstore "github.com/mendersoftware/mender-server/pkg/store"

	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/services/deployments/store"
)

func TestDeviceDeploymentStorageInsert(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping TestDeviceDeploymentStorageInsert in short mode.")
	}

	deviceDepl1 := model.NewDeviceDeployment("30b3e62c-9ec2-4312-a7fa-cff24cc7397a", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a")

	deviceDepl2 := model.NewDeviceDeployment("30b3e62c-9ec2-4312-a7fa-cff24cc7397a", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a")

	badDeviceDepl := model.NewDeviceDeployment("bad bad", "bad bad bad")

	testCases := []struct {
		InputDeviceDeployment []*model.DeviceDeployment
		InputTenant           string
		OutputError           error
	}{
		{
			InputDeviceDeployment: nil,
			OutputError:           nil,
		},
		{
			InputDeviceDeployment: []*model.DeviceDeployment{nil, nil},
			OutputError:           ErrStorageInvalidDeviceDeployment,
		},
		{
			InputDeviceDeployment: []*model.DeviceDeployment{
				badDeviceDepl,
				badDeviceDepl,
			},
			OutputError: errors.New("Validating device deployment: DeploymentId: must be a valid UUID."),
		},
		{
			InputDeviceDeployment: []*model.DeviceDeployment{
				deviceDepl1,
				badDeviceDepl,
			},
			OutputError: errors.New("Validating device deployment: DeploymentId: must be a valid UUID."),
		},
		{
			InputDeviceDeployment: []*model.DeviceDeployment{
				deviceDepl1,
				deviceDepl2,
			},
			OutputError: nil,
		},
		{
			// same as previous case, but this time with tenant DB
			InputDeviceDeployment: []*model.DeviceDeployment{
				deviceDepl1,
				deviceDepl2,
			},
			InputTenant: "acme",
			OutputError: nil,
		},
	}

	for testCaseNumber, testCase := range testCases {
		t.Run(fmt.Sprintf("test case %d", testCaseNumber+1), func(t *testing.T) {

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

			err := store.InsertMany(ctx,
				testCase.InputDeviceDeployment...)

			if testCase.OutputError != nil {
				assert.EqualError(t, err, testCase.OutputError.Error())
			} else {
				assert.NoError(t, err)
				collection := client.Database(ctxstore.
					DbFromContext(ctx, DatabaseName)).
					Collection(CollectionDevices)
				count, err := collection.CountDocuments(
					ctx, bson.D{})
				assert.NoError(t, err)
				assert.Equal(t, len(testCase.
					InputDeviceDeployment), int(count))

				if testCase.InputTenant != "" {
					// deployment was added to tenant's DB,
					// make sure it's not in default DB
					collectionDefault := client.Database(
						DatabaseName).Collection(CollectionDevices)
					count, err = collectionDefault.
						CountDocuments(ctx, bson.D{})
					assert.NoError(t, err)
					assert.Equal(t, 0, int(count))
				}
			}
		})
	}
}

func TestUpdateDeviceDeploymentStatus(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping TestUpdateDeviceDeploymentStatus in short mode.")
	}

	now := time.Now()

	deviceDeployments := []model.DeviceDeployment{}

	dds := []struct {
		did   string
		depid string
	}{
		{"456", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a"},
		{"567", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a"},
		{"678", "30b3e62c-9ec2-4312-a7fa-cff24cc7397d"},
		{"12345", "30b3e62c-9ec2-4312-a7fa-cff24cc7397e"},
	}

	for _, dd := range dds {
		newdd := model.NewDeviceDeployment(dd.did, dd.depid)
		deviceDeployments = append(deviceDeployments, *newdd)
	}

	testCases := []struct {
		InputDeviceID         string
		InputDeploymentID     string
		InputStatus           model.DeviceDeploymentStatus
		InputSubState         string
		InputDeviceDeployment []*model.DeviceDeployment
		InputFinishTime       *time.Time
		InputTenant           string

		OutputError     error
		OutputOldStatus model.DeviceDeploymentStatus
	}{
		{
			// null status
			InputDeviceID:     "123",
			InputDeploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
			OutputError:       ErrStorageInvalidInput,
			OutputOldStatus:   model.DeviceDeploymentStatusNull,
		},
		{
			// null deployment ID
			InputDeviceID:   "234",
			InputStatus:     model.DeviceDeploymentStatusNull,
			OutputError:     ErrStorageInvalidID,
			OutputOldStatus: model.DeviceDeploymentStatusNull,
		},
		{
			// null device ID
			InputDeploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
			InputStatus:       model.DeviceDeploymentStatusInstalling,
			OutputError:       ErrStorageInvalidID,
			OutputOldStatus:   model.DeviceDeploymentStatusNull,
		},
		{
			// no deployment/device with this ID
			InputDeviceID:     "345",
			InputDeploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
			InputStatus:       model.DeviceDeploymentStatusAborted,
			OutputError:       ErrStorageNotFound,
			OutputOldStatus:   model.DeviceDeploymentStatusNull,
		},
		{
			InputDeviceID:     "456",
			InputDeploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
			InputStatus:       model.DeviceDeploymentStatusInstalling,
			InputDeviceDeployment: []*model.DeviceDeployment{
				&deviceDeployments[0],
			},
			OutputError:     nil,
			OutputOldStatus: model.DeviceDeploymentStatusPending,
		},
		{
			InputDeviceID:     "567",
			InputDeploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
			InputStatus:       model.DeviceDeploymentStatusFailure,
			InputDeviceDeployment: []*model.DeviceDeployment{
				&deviceDeployments[1],
			},
			InputFinishTime: &now,
			OutputError:     nil,
			OutputOldStatus: model.DeviceDeploymentStatusPending,
		},
		{
			InputDeviceID:     "678",
			InputDeploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397d",
			InputStatus:       model.DeviceDeploymentStatusInstalling,
			InputDeviceDeployment: []*model.DeviceDeployment{
				&deviceDeployments[2],
			},
			InputTenant:     "acme",
			OutputOldStatus: model.DeviceDeploymentStatusPending,
		},
		{
			InputDeviceID:     "12345",
			InputDeploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397e",
			InputStatus:       model.DeviceDeploymentStatusInstalling,
			InputSubState:     "foobar 123",
			InputDeviceDeployment: []*model.DeviceDeployment{
				&deviceDeployments[3],
			},
			OutputError:     nil,
			OutputOldStatus: model.DeviceDeploymentStatusPending,
		},
	}

	for testCaseNumber, testCase := range testCases {
		t.Run(fmt.Sprintf("test case %d", testCaseNumber+1), func(t *testing.T) {

			t.Logf("testing case %s %s %s %v",
				testCase.InputDeviceID, testCase.InputDeploymentID,
				testCase.InputStatus, testCase.OutputError)

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

			// deployments are created with status DeviceDeploymentStatusPending
			err := store.InsertMany(ctx, testCase.InputDeviceDeployment...)
			assert.NoError(t, err)

			old, err := store.UpdateDeviceDeploymentStatus(ctx,
				testCase.InputDeviceID, testCase.InputDeploymentID,
				model.DeviceDeploymentState{
					Status:     testCase.InputStatus,
					SubState:   testCase.InputSubState,
					FinishTime: testCase.InputFinishTime,
				},
				testCase.InputStatus,
			)

			if testCase.OutputError != nil {
				assert.EqualError(t, err, testCase.OutputError.Error())
			} else {
				assert.NoError(t, err)

				if testCase.InputTenant != "" {
					// update in tenant's DB was successful,
					// similar update in default DB should
					// fail because deployments are present
					// in tenant's DB only
					_, err := store.UpdateDeviceDeploymentStatus(
						context.Background(),
						testCase.InputDeviceID, testCase.InputDeploymentID,
						model.DeviceDeploymentState{
							Status:     testCase.InputStatus,
							FinishTime: testCase.InputFinishTime,
							SubState:   testCase.InputSubState,
						},
						testCase.InputStatus,
					)
					t.Logf("error: %+v", err)
					assert.EqualError(t, err, ErrStorageNotFound.Error())
				}
			}

			if testCase.InputDeviceDeployment != nil {
				// these checks only make sense if there are any deployments in database
				var deployment *model.DeviceDeployment
				collDevs := client.Database(ctxstore.
					DbFromContext(ctx, DatabaseName)).
					Collection(CollectionDevices)
				query := bson.M{
					StorageKeyDeviceDeploymentDeviceId:     testCase.InputDeviceID,
					StorageKeyDeviceDeploymentDeploymentID: testCase.InputDeploymentID,
				}
				err := collDevs.FindOne(ctx, query).Decode(&deployment)
				assert.NoError(t, err)
				if testCase.OutputError != nil {
					// status must be unchanged in case of errors
					assert.Equal(t, model.DeviceDeploymentStatusPending, deployment.Status)
				} else {
					if !assert.NotNil(t, deployment) {
						return
					}

					assert.Equal(t, testCase.InputStatus, deployment.Status)
					assert.Equal(t, testCase.OutputOldStatus, old)
					// verify deployment finish time
					if testCase.InputFinishTime != nil && assert.NotNil(t, deployment.Finished) {
						// mongo might have trimmed our
						// time a bit, let's check that
						// we are within a 1s range
						assert.WithinDuration(t, *testCase.InputFinishTime,
							*deployment.Finished, time.Second)
					}

					if testCase.InputSubState != "" {
						assert.Equal(t, testCase.InputSubState, deployment.SubState)
					}
				}
			}
		})
	}
}

func TestUpdateDeviceDeploymentStatusStarted(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping TestUpdateDeviceDeploymentStatus in short mode.")
	}

	deviceDeployments := []model.DeviceDeployment{}

	dds := []struct {
		did   string
		depid string
	}{
		{"456", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a"},
		{"567", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a"},
		{"678", "30b3e62c-9ec2-4312-a7fa-cff24cc7397d"},
		{"12345", "30b3e62c-9ec2-4312-a7fa-cff24cc7397e"},
	}

	for _, dd := range dds {
		newdd := model.NewDeviceDeployment(dd.did, dd.depid)
		deviceDeployments = append(deviceDeployments, *newdd)
	}

	testCases := []struct {
		InputDeviceID         string
		InputDeploymentID     string
		InputStatus           model.DeviceDeploymentStatus
		InputCurrentStatus    model.DeviceDeploymentStatus
		InputSubState         string
		InputDeviceDeployment []*model.DeviceDeployment
		InputFinishTime       *time.Time
		InputTenant           string

		OutputError     error
		OutputOldStatus model.DeviceDeploymentStatus
	}{
		{
			InputDeviceID:      "12345",
			InputDeploymentID:  "30b3e62c-9ec2-4312-a7fa-cff24cc7397e",
			InputCurrentStatus: model.DeviceDeploymentStatusPending,
			InputStatus:        model.DeviceDeploymentStatusDownloading,
			InputSubState:      "foobar 123",
			InputDeviceDeployment: []*model.DeviceDeployment{
				&deviceDeployments[3],
			},
			OutputError:     nil,
			OutputOldStatus: model.DeviceDeploymentStatusPending,
		},
		{
			InputDeviceID:      "12345",
			InputDeploymentID:  "30b3e62c-9ec2-4312-a7fa-cff24cc7397e",
			InputCurrentStatus: model.DeviceDeploymentStatusDownloading,
			InputStatus:        model.DeviceDeploymentStatusDownloading,
			InputSubState:      "foobar 123",
			InputDeviceDeployment: []*model.DeviceDeployment{
				&deviceDeployments[3],
			},
			OutputError:     nil,
			OutputOldStatus: model.DeviceDeploymentStatusPending,
		},
	}

	for testCaseNumber, testCase := range testCases {
		t.Run(fmt.Sprintf("test case %d", testCaseNumber+1), func(t *testing.T) {

			t.Logf("testing case %s %s %s %v",
				testCase.InputDeviceID, testCase.InputDeploymentID,
				testCase.InputStatus, testCase.OutputError)

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

			// deployments are created with status DeviceDeploymentStatusPending
			err := store.InsertMany(ctx, testCase.InputDeviceDeployment...)
			assert.NoError(t, err)

			dt := rand.Intn(15) + 4
			time.Sleep(time.Duration(dt) * time.Second)
			_, err = store.UpdateDeviceDeploymentStatus(ctx,
				testCase.InputDeviceID, testCase.InputDeploymentID,
				model.DeviceDeploymentState{
					Status:     testCase.InputStatus,
					SubState:   testCase.InputSubState,
					FinishTime: testCase.InputFinishTime,
				},
				testCase.InputCurrentStatus,
			)
			collDevs := client.Database(DbName).Collection(CollectionDevices)
			r, err := collDevs.Find(ctx, bson.M{})
			var deviceDeploymentsStored []model.DeviceDeployment
			err = r.All(ctx, &deviceDeploymentsStored)
			if testCase.InputCurrentStatus == testCase.InputStatus {
				assert.Nil(t, deviceDeploymentsStored[0].Started)
			} else {
				assert.NotNil(t, deviceDeploymentsStored[0].Started)
				startedInterval := deviceDeploymentsStored[0].Started.Sub(*deviceDeploymentsStored[0].Created)
				assert.True(
					t,
					startedInterval.Seconds() >= float64(dt),
					"difference between created and started times should be at least "+
						strconv.Itoa(dt)+"s")
			}
		})
	}
}

func TestUpdateDeviceDeploymentLogAvailability(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping TestUpdateDeviceDeploymentLogAvailability in short mode.")
	}

	dd := model.NewDeviceDeployment("456", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a")

	testCases := []struct {
		InputDeviceID         string
		InputDeploymentID     string
		InputLog              bool
		InputDeviceDeployment []*model.DeviceDeployment
		InputTenant           string

		OutputError error
	}{
		{
			// null deployment ID
			InputDeviceID: "234",
			OutputError:   ErrStorageInvalidID,
		},
		{
			// null device ID
			InputDeploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
			OutputError:       ErrStorageInvalidID,
		},
		{
			// no deployment/device with this ID
			InputDeviceID:     "345",
			InputDeploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
			InputLog:          true,
			OutputError:       ErrStorageNotFound,
		},
		{
			InputDeviceID:     "456",
			InputDeploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
			InputLog:          true,
			InputDeviceDeployment: []*model.DeviceDeployment{
				dd,
			},
			OutputError: nil,
		},
		{
			InputDeviceID:     "456",
			InputDeploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
			InputLog:          false,
			InputDeviceDeployment: []*model.DeviceDeployment{
				dd,
			},
			InputTenant: "acme",
		},
	}

	for testCaseNumber, testCase := range testCases {
		t.Run(fmt.Sprintf("test case %d", testCaseNumber+1), func(t *testing.T) {

			t.Logf("testing case %s %s %t %v",
				testCase.InputDeviceID, testCase.InputDeploymentID,
				testCase.InputLog, testCase.OutputError)

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

			// deployments are created with status DeviceDeploymentStatusPending
			err := store.InsertMany(ctx, testCase.InputDeviceDeployment...)
			assert.NoError(t, err)

			err = store.UpdateDeviceDeploymentLogAvailability(ctx,
				testCase.InputDeviceID, testCase.InputDeploymentID, testCase.InputLog)

			if testCase.OutputError != nil {
				assert.EqualError(t, err, testCase.OutputError.Error())
			} else {
				assert.NoError(t, err)

				if testCase.InputTenant != "" {
					// we're using tenant's DB, so acting on default DB should fail
					err := store.UpdateDeviceDeploymentLogAvailability(context.Background(),
						testCase.InputDeviceID, testCase.InputDeploymentID,
						testCase.InputLog)
					assert.EqualError(t, err, ErrStorageNotFound.Error())
				}
			}

			if testCase.InputDeviceDeployment != nil {
				var deployment model.DeviceDeployment
				query := bson.M{
					StorageKeyDeviceDeploymentDeviceId:     testCase.InputDeviceID,
					StorageKeyDeviceDeploymentDeploymentID: testCase.InputDeploymentID,
				}
				collDevs := client.Database(ctxstore.
					DbFromContext(ctx, DatabaseName)).
					Collection(CollectionDevices)

				err := collDevs.FindOne(ctx, query).
					Decode(&deployment)
				assert.NoError(t, err)
				assert.Equal(t, testCase.InputLog, deployment.IsLogAvailable)
			}
		})
	}
}

func newDeviceDeploymentWithStatus(t *testing.T, deviceID string, deploymentID string, status model.DeviceDeploymentStatus) *model.DeviceDeployment {
	d := model.NewDeviceDeployment(deviceID, deploymentID)

	d.Status = status
	d.Active = status.Active()
	return d
}

func TestAggregateDeviceDeploymentByStatus(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping TestAggregateDeviceDeploymentByStatus in short mode.")
	}

	testCases := []struct {
		InputDeploymentID     string
		InputDeviceDeployment []*model.DeviceDeployment
		InputTenant           string
		OutputError           error
		OutputStats           model.Stats
	}{
		{
			InputDeploymentID:     "ee13ea8b-a6d3-4d4c-99a6-bcfcaebc7ec3",
			InputDeviceDeployment: nil,
			OutputError:           nil,
			OutputStats:           nil,
		},
		{
			InputDeploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
			InputDeviceDeployment: []*model.DeviceDeployment{
				newDeviceDeploymentWithStatus(t, "123", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
					model.DeviceDeploymentStatusFailure),
				newDeviceDeploymentWithStatus(t, "234", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
					model.DeviceDeploymentStatusFailure),
				newDeviceDeploymentWithStatus(t, "456", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
					model.DeviceDeploymentStatusSuccess),

				// these 2 count as in progress
				newDeviceDeploymentWithStatus(t, "567", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
					model.DeviceDeploymentStatusDownloading),
				newDeviceDeploymentWithStatus(t, "678", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
					model.DeviceDeploymentStatusRebooting),

				newDeviceDeploymentWithStatus(t, "789", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
					model.DeviceDeploymentStatusPending),
			},
			OutputError: nil,
			OutputStats: model.Stats{
				model.DeviceDeploymentStatusPendingStr:            1,
				model.DeviceDeploymentStatusSuccessStr:            1,
				model.DeviceDeploymentStatusFailureStr:            2,
				model.DeviceDeploymentStatusRebootingStr:          1,
				model.DeviceDeploymentStatusDownloadingStr:        1,
				model.DeviceDeploymentStatusInstallingStr:         0,
				model.DeviceDeploymentStatusNoArtifactStr:         0,
				model.DeviceDeploymentStatusAlreadyInstStr:        0,
				model.DeviceDeploymentStatusAbortedStr:            0,
				model.DeviceDeploymentStatusDecommissionedStr:     0,
				model.DeviceDeploymentStatusPauseBeforeCommitStr:  0,
				model.DeviceDeploymentStatusPauseBeforeInstallStr: 0,
				model.DeviceDeploymentStatusPauseBeforeRebootStr:  0,
			},
		},
		{
			InputDeploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
			InputDeviceDeployment: []*model.DeviceDeployment{
				newDeviceDeploymentWithStatus(t, "123", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
					model.DeviceDeploymentStatusPauseBeforeInstall),

				// these 2 count as in progress
				newDeviceDeploymentWithStatus(t, "567", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
					model.DeviceDeploymentStatusPauseBeforeCommit),
				newDeviceDeploymentWithStatus(t, "678", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
					model.DeviceDeploymentStatusPauseBeforeReboot),
			},
			OutputError: nil,
			OutputStats: model.Stats{
				model.DeviceDeploymentStatusPendingStr:            0,
				model.DeviceDeploymentStatusSuccessStr:            0,
				model.DeviceDeploymentStatusFailureStr:            0,
				model.DeviceDeploymentStatusRebootingStr:          0,
				model.DeviceDeploymentStatusDownloadingStr:        0,
				model.DeviceDeploymentStatusInstallingStr:         0,
				model.DeviceDeploymentStatusNoArtifactStr:         0,
				model.DeviceDeploymentStatusAlreadyInstStr:        0,
				model.DeviceDeploymentStatusAbortedStr:            0,
				model.DeviceDeploymentStatusDecommissionedStr:     0,
				model.DeviceDeploymentStatusPauseBeforeCommitStr:  1,
				model.DeviceDeploymentStatusPauseBeforeInstallStr: 1,
				model.DeviceDeploymentStatusPauseBeforeRebootStr:  1,
			},
		},
		{
			InputDeploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
			InputDeviceDeployment: []*model.DeviceDeployment{
				newDeviceDeploymentWithStatus(t, "123", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
					model.DeviceDeploymentStatusFailure),
				newDeviceDeploymentWithStatus(t, "456", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
					model.DeviceDeploymentStatusSuccess),
			},
			InputTenant: "acme",
			OutputStats: newTestStats(model.Stats{
				model.DeviceDeploymentStatusSuccessStr: 1,
				model.DeviceDeploymentStatusFailureStr: 1,
			}),
		},
	}

	for testCaseNumber, testCase := range testCases {
		t.Run(fmt.Sprintf("test case %d", testCaseNumber+1), func(t *testing.T) {

			t.Logf("testing case %s %v %d", testCase.InputDeploymentID, testCase.OutputError,
				len(testCase.InputDeviceDeployment))

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

			err := store.InsertMany(ctx, testCase.InputDeviceDeployment...)
			assert.NoError(t, err)

			stats, err := store.AggregateDeviceDeploymentByStatus(ctx,
				testCase.InputDeploymentID)
			if testCase.OutputError != nil {
				assert.EqualError(t, err, testCase.OutputError.Error())
			} else {
				assert.NoError(t, err)

				if testCase.InputTenant != "" {
					// data was inserted into tenant's DB,
					// verify that aggregates are all 0
					stats, err := store.AggregateDeviceDeploymentByStatus(context.Background(),
						testCase.InputDeploymentID)
					assert.NoError(t, err)
					assert.Equal(t, newTestStats(model.Stats{}), stats)
				}
			}

			if testCase.OutputStats != nil {
				assert.NotNil(t, stats)
				assert.Equal(t, testCase.OutputStats, stats)
			}
		})
	}
}

func TestGetDeviceStatusesForDeployment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping GetDeviceStatusesForDeployment in short mode.")
	}

	input := []*model.DeviceDeployment{}

	dds := []struct {
		did   string
		depid string
	}{
		{"device0001", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a"},
		{"device0002", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a"},
		{"device0003", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a"},
		{"device0004", "30b3e62c-9ec2-4312-a7fa-cff24cc7397b"},
		{"device0005", "30b3e62c-9ec2-4312-a7fa-cff24cc7397b"},
	}

	for _, dd := range dds {
		newdd := model.NewDeviceDeployment(dd.did, dd.depid)
		input = append(input, newdd)
	}

	testCases := map[string]struct {
		caseId string
		tenant string

		inputDeploymentId string
		outputStatuses    []*model.DeviceDeployment
	}{
		"existing deployments 1": {
			inputDeploymentId: "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
			outputStatuses:    input[:3],
		},
		"existing deployments 2": {
			inputDeploymentId: "30b3e62c-9ec2-4312-a7fa-cff24cc7397b",
			outputStatuses:    input[3:],
		},
		"nonexistent deployment": {
			inputDeploymentId: "aaaaaaaa-9ec2-4312-a7fa-cff24cc7397b",
			outputStatuses:    []*model.DeviceDeployment{},
		},
		"tenant, existing deployments": {
			inputDeploymentId: "30b3e62c-9ec2-4312-a7fa-cff24cc7397b",
			tenant:            "acme",
			outputStatuses:    input[3:],
		},
	}

	for testCaseName, tc := range testCases {
		t.Run(fmt.Sprintf("test case %s", testCaseName), func(t *testing.T) {

			// setup db - once for all cases
			db.Wipe()

			client := db.Client()
			store := NewDataStoreMongoWithClient(client)

			ctx := context.Background()
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.tenant,
				})
			}

			err := store.InsertMany(ctx, input...)
			assert.NoError(t, err)

			statuses, err := store.GetDeviceStatusesForDeployment(ctx,
				tc.inputDeploymentId)
			assert.NoError(t, err)

			assert.Equal(t, len(tc.outputStatuses), len(statuses))
			for i, out := range tc.outputStatuses {
				assert.Equal(t, out.DeviceId, statuses[i].DeviceId)
				assert.Equal(t, out.DeploymentId, statuses[i].DeploymentId)
			}

			if tc.tenant != "" {
				// deployment statuses are present in tenant's
				// DB, verify that listing from default DB
				// yields empty list
				statuses, err := store.GetDeviceStatusesForDeployment(
					context.Background(),
					tc.inputDeploymentId)
				assert.NoError(t, err)
				assert.Len(t, statuses, 0)
			}
		})
	}
}

func TestGetDevicesListForDeployment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping GetDevicesListForDeployment in short mode.")
	}

	dds := []struct {
		did    string
		depid  string
		status model.DeviceDeploymentStatus
	}{{
		did:    "device0001",
		depid:  "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
		status: model.DeviceDeploymentStatusSuccess,
	}, {
		did:    "device0002",
		depid:  "30b3e62c-9ec2-4312-a7fa-cff24cc7397b",
		status: model.DeviceDeploymentStatusFailure,
	}, {
		did:    "device0003",
		depid:  "30b3e62c-9ec2-4312-a7fa-cff24cc7397b",
		status: model.DeviceDeploymentStatusAborted,
	}, {
		did:    "device0004",
		depid:  "30b3e62c-9ec2-4312-a7fa-cff24cc7397b",
		status: model.DeviceDeploymentStatusPauseBeforeInstall,
	}, {
		did:    "device0005",
		depid:  "30b3e62c-9ec2-4312-a7fa-cff24cc7397b",
		status: model.DeviceDeploymentStatusPauseBeforeCommit,
	}, {
		did:    "device0006",
		depid:  "30b3e62c-9ec2-4312-a7fa-cff24cc7397b",
		status: model.DeviceDeploymentStatusPauseBeforeReboot,
	}, {
		did:    "device0007",
		depid:  "30b3e62c-9ec2-4312-a7fa-cff24cc7397b",
		status: model.DeviceDeploymentStatusDownloading,
	}, {
		did:    "device000a",
		depid:  "30b3e62c-9ec2-4312-a7fa-cff24cc7397b",
		status: model.DeviceDeploymentStatusInstalling,
	}, {
		did:    "device0009",
		depid:  "30b3e62c-9ec2-4312-a7fa-cff24cc7397b",
		status: model.DeviceDeploymentStatusRebooting,
	}, {
		did:    "device0008",
		depid:  "30b3e62c-9ec2-4312-a7fa-cff24cc7397b",
		status: model.DeviceDeploymentStatusPending,
	}, {
		did:    "device000b",
		depid:  "30b3e62c-9ec2-4312-a7fa-cff24cc7397b",
		status: model.DeviceDeploymentStatusSuccess,
	}, {
		did:    "device000e",
		depid:  "30b3e62c-9ec2-4312-a7fa-cff24cc7397b",
		status: model.DeviceDeploymentStatusNoArtifact,
	}, {
		did:    "device000d",
		depid:  "30b3e62c-9ec2-4312-a7fa-cff24cc7397b",
		status: model.DeviceDeploymentStatusAlreadyInst,
	}, {
		did:    "device000c",
		depid:  "30b3e62c-9ec2-4312-a7fa-cff24cc7397b",
		status: model.DeviceDeploymentStatusDecommissioned,
	}}
	input := make([]model.DeviceDeployment, len(dds))
	for i, dd := range dds {
		newdd := model.NewDeviceDeployment(dd.did, dd.depid)
		// strip timezone and monotonic time (lost when writing to db)
		notz := newdd.Created.UTC().Round(time.Millisecond)
		newdd.Created = &notz
		newdd.Status = dd.status
		input[i] = *newdd
	}

	testCases := map[string]struct {
		caseId string
		ctx    context.Context

		inputListQuery store.ListQuery
		outputStatuses []model.DeviceDeployment
		Error          error
	}{
		"existing deployments 1": {
			inputListQuery: store.ListQuery{
				DeploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
			},
			outputStatuses: input[:1],
		},
		"existing deployments 2": {
			inputListQuery: store.ListQuery{
				DeploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397b",
			},
			outputStatuses: input[1:],
		},
		"filter by status": {
			inputListQuery: store.ListQuery{
				DeploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397b",
				Status: func() *string {
					s := model.DeviceDeploymentStatusSuccess.String()
					return &s
				}(),
			},
			outputStatuses: input[10:11],
		},
		"range filter pause statuses": {
			inputListQuery: store.ListQuery{
				DeploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397b",
				Status: func() *string {
					s := "pause"
					return &s
				}(),
			},
			outputStatuses: input[3:6],
		},
		"range filter finished status": {
			inputListQuery: store.ListQuery{
				DeploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397b",
				Status: func() *string {
					s := "finished"
					return &s
				}(),
			},
			outputStatuses: []model.DeviceDeployment{
				input[1],
				input[2],
				input[10],
				input[11],
				input[12],
				input[13],
			},
		},
		"nonexistent deployment": {
			inputListQuery: store.ListQuery{
				DeploymentID: "aaaaaaaa-9ec2-4312-a7fa-cff24cc7397b",
			},
			outputStatuses: []model.DeviceDeployment{},
		},
		"tenant, existing deployments": {
			inputListQuery: store.ListQuery{
				DeploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397b",
			},
			ctx: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: "acme",
			}),
			outputStatuses: input[1:],
		},
		"tenant, existing deployments + limit": {
			inputListQuery: store.ListQuery{
				DeploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397b",
				Limit:        2,
			},
			ctx: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: "acme",
			}),
			outputStatuses: input[1:3],
		},
		"tenant, existing deployments + limit + skip": {
			inputListQuery: store.ListQuery{
				DeploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397b",
				Limit:        2,
				Skip:         1,
			},
			ctx: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: "acme",
			}),
			outputStatuses: input[2:4],
		},
		"error: context canceled": {
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.TODO())
				cancel()
				return ctx
			}(),
			Error: context.Canceled,
		},
		"error: bad status filter": {
			inputListQuery: store.ListQuery{
				DeploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397b",
				Status: func() *string {
					s := "foobar"
					return &s
				}(),
			},
			Error: errors.New("invalid status query"),
		},
	}

	for testCaseName, tc := range testCases {
		t.Run(testCaseName, func(t *testing.T) {

			db.Wipe()
			ctx := context.Background()
			if tc.ctx == nil {
				tc.ctx = ctx
			}
			if id := identity.FromContext(tc.ctx); id != nil {
				ctx = identity.WithContext(ctx, id)
			}

			client := db.Client()
			store := NewDataStoreMongoWithClient(client)
			collDevs := client.Database(ctxstore.DbFromContext(ctx, DbName)).
				Collection(CollectionDevices)

			devFaces := make([]interface{}, len(input))
			for i := range input {
				devFaces[i] = &input[i]
			}
			_, err := collDevs.InsertMany(ctx, devFaces)
			assert.NoError(t, err)

			statuses, _, err := store.GetDevicesListForDeployment(tc.ctx,
				tc.inputListQuery)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
				return
			}
			assert.NoError(t, err)

			if tc.inputListQuery.Limit > 0 {
				assert.Equal(t,
					tc.inputListQuery.Limit,
					len(statuses))
			}
			assert.Equal(t, tc.outputStatuses, statuses)

			if id := identity.FromContext(ctx); id != nil {
				// deployment statuses are present in tenant's
				// DB, verify that listing from default DB
				// yields empty list
				statuses, _, err := store.GetDevicesListForDeployment(
					context.Background(),
					tc.inputListQuery,
				)
				assert.NoError(t, err)
				assert.Len(t, statuses, 0)
			}
		})
	}
}

func TestHasDeploymentForDevice(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping GetDeviceStatusesForDeployment in short mode.")
	}

	input := []*model.DeviceDeployment{}

	dds := []struct {
		did   string
		depid string
	}{
		{"device0001", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a"},
		{"device0002", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a"},
		{"device0003", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a"},
	}

	for _, dd := range dds {
		newdd := model.NewDeviceDeployment(dd.did, dd.depid)
		input = append(input, newdd)
	}

	testCases := []struct {
		deviceID     string
		deploymentID string
		tenant       string

		has bool
		err error
	}{
		{
			deviceID:     "device0001",
			deploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
			has:          true,
			err:          nil,
		},
		{
			deviceID:     "device0002",
			deploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
			has:          true,
			err:          nil,
		},
		{
			deviceID:     "device0003",
			deploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397b",
			has:          false,
		},
		{
			deviceID:     "device0004",
			deploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397c",
			has:          false,
		},
		{
			deviceID:     "device0003",
			deploymentID: "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
			has:          true,
			tenant:       "acme",
		},
	}

	for testCaseNumber, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", testCaseNumber+1), func(t *testing.T) {

			t.Logf("testing case: %v %v %v %v", tc.deviceID, tc.deploymentID, tc.has, tc.err)

			db.Wipe()

			client := db.Client()
			store := NewDataStoreMongoWithClient(client)

			ctx := context.Background()
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.tenant,
				})
			}

			err := store.InsertMany(ctx, input...)
			assert.NoError(t, err)

			has, err := store.HasDeploymentForDevice(ctx,
				tc.deploymentID, tc.deviceID)
			if tc.err != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.err.Error())
				assert.False(t, has)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.has, has)

				if tc.tenant != "" {
					// data was added to tenant's DB, verify
					// that there's no deployment if looking
					// in default DB
					has, err := store.HasDeploymentForDevice(context.Background(),
						tc.deploymentID, tc.deviceID)
					assert.False(t, has)
					assert.NoError(t, err)
				}
			}
		})
	}
}

func TestAbortDeviceDeployments(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping TestAbortDeviceDeployments in short mode.")
	}

	input := []*model.DeviceDeployment{}

	dds := []struct {
		did   string
		depid string
	}{
		{"456", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a"},
		{"567", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a"},
	}

	for _, dd := range dds {
		newdd := model.NewDeviceDeployment(dd.did, dd.depid)
		input = append(input, newdd)
	}

	testCases := map[string]struct {
		InputDeploymentID     string
		InputDeviceDeployment []*model.DeviceDeployment

		OutputError error
	}{
		"null deployment id": {
			OutputError: ErrStorageInvalidID,
		},
		"all correct": {
			InputDeploymentID:     "30b3e62c-9ec2-4312-a7fa-cff24cc7397a",
			InputDeviceDeployment: input,
			OutputError:           nil,
		},
	}

	for testCaseName, testCase := range testCases {
		t.Run(fmt.Sprintf("test case %s", testCaseName), func(t *testing.T) {

			// Make sure we start test with empty database
			db.Wipe()

			client := db.Client()
			store := NewDataStoreMongoWithClient(client)

			err := store.InsertMany(context.Background(), testCase.InputDeviceDeployment...)
			assert.NoError(t, err)

			err = store.AbortDeviceDeployments(context.Background(), testCase.InputDeploymentID)

			if testCase.OutputError != nil {
				assert.EqualError(t, err, testCase.OutputError.Error())
			} else {
				assert.NoError(t, err)
			}

			if testCase.InputDeviceDeployment != nil {
				// these checks only make sense if there are any deployments in database
				var deploymentList []model.DeviceDeployment
				collDevs := client.Database(DatabaseName).
					Collection(CollectionDevices)
				query := bson.M{
					StorageKeyDeviceDeploymentDeploymentID: testCase.InputDeploymentID,
				}
				cursor, err := collDevs.Find(db.CTX(), query)
				assert.NoError(t, err)
				err = cursor.All(db.CTX(), &deploymentList)
				assert.NoError(t, err)

				if testCase.OutputError != nil {
					for _, deployment := range deploymentList {
						// status must be unchanged in case of errors
						assert.Equal(t, model.DeviceDeploymentStatusPending,
							deployment.Status)
					}
				} else {
					for _, deployment := range deploymentList {
						assert.Equal(t, model.DeviceDeploymentStatusAborted,
							deployment.Status)
					}
				}
			}
		})
	}
}

func TestDecommissionDeviceDeployments(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping TestDecommissionDeviceDeployments in short mode.")
	}

	input := []*model.DeviceDeployment{}

	dds := []struct {
		did   string
		depid string
	}{
		{"foo", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a"},
		{"bar", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a"},
	}

	for _, dd := range dds {
		newdd := model.NewDeviceDeployment(dd.did, dd.depid)
		input = append(input, newdd)
	}

	testCases := map[string]struct {
		InputDeviceId         string
		InputDeviceDeployment []*model.DeviceDeployment

		OutputError error
	}{
		"null device id": {
			OutputError: ErrStorageInvalidID,
		},
		"all correct": {
			InputDeviceId:         "foo",
			InputDeviceDeployment: input,
			OutputError:           nil,
		},
	}

	for testCaseName, testCase := range testCases {
		t.Run(fmt.Sprintf("test case %s", testCaseName), func(t *testing.T) {

			// Make sure we start test with empty database
			db.Wipe()

			client := db.Client()
			ctx := db.CTX()
			store := NewDataStoreMongoWithClient(client)

			err := store.InsertMany(ctx, testCase.InputDeviceDeployment...)
			assert.NoError(t, err)

			err = store.DecommissionDeviceDeployments(ctx, testCase.InputDeviceId)

			if testCase.OutputError != nil {
				assert.EqualError(t, err, testCase.OutputError.Error())
			} else {
				assert.NoError(t, err)
			}

			if testCase.InputDeviceDeployment != nil {
				// these checks only make sense if there are any deployments in database
				var deploymentList []model.DeviceDeployment
				collDevs := client.Database(DatabaseName).
					Collection(CollectionDevices)
				query := bson.M{
					StorageKeyDeviceDeploymentDeviceId: testCase.InputDeviceId,
				}
				cursor, err := collDevs.Find(ctx, query)
				assert.NoError(t, err)
				err = cursor.All(ctx, &deploymentList)
				assert.NoError(t, err)

				if testCase.OutputError != nil {
					for _, deployment := range deploymentList {
						// status must be unchanged in case of errors
						assert.Equal(t, model.DeviceDeploymentStatusPending,
							deployment.Status)
					}
				} else {
					for _, deployment := range deploymentList {
						assert.Equal(t, model.DeviceDeploymentStatusDecommissioned,
							deployment.Status)
					}
				}
			}
		})
	}
}

func TestSaveDeviceDeploymentRequest(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping TestSaveDeviceDeploymentRequest in short mode.")
	}

	dd := model.NewDeviceDeployment("456", "30b3e62c-9ec2-4312-a7fa-cff24cc7397a")

	testCases := map[string]struct {
		deviceDeployments []*model.DeviceDeployment
		tenant            string
		request           *model.DeploymentNextRequest

		err error
	}{
		"ok": {
			deviceDeployments: []*model.DeviceDeployment{dd},
			request: &model.DeploymentNextRequest{
				DeviceProvides: &model.InstalledDeviceDeployment{
					ArtifactName: "foo",
					DeviceType:   "bar",
				},
			},
		},
		"ok, tenant": {
			deviceDeployments: []*model.DeviceDeployment{dd},
			request: &model.DeploymentNextRequest{
				DeviceProvides: &model.InstalledDeviceDeployment{
					ArtifactName: "foo",
					DeviceType:   "bar",
				},
			},
			tenant: "foo",
		},
		"no device deployments": {
			request: &model.DeploymentNextRequest{
				DeviceProvides: &model.InstalledDeviceDeployment{
					ArtifactName: "foo",
					DeviceType:   "bar",
				},
			},
			err: ErrStorageNotFound,
		},
	}

	for name, tc := range testCases {
		t.Run(fmt.Sprintf("test case %s", name), func(t *testing.T) {

			// Make sure we start test with empty database
			db.Wipe()
			client := db.Client()
			store := NewDataStoreMongoWithClient(client)

			ctx := context.Background()
			if tc.tenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.tenant,
				})
			}

			err := store.InsertMany(ctx, tc.deviceDeployments...)
			assert.NoError(t, err)

			err = store.SaveDeviceDeploymentRequest(ctx, dd.Id, tc.request)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
				var deployment *model.DeviceDeployment
				collDevs := client.Database(ctxstore.
					DbFromContext(ctx, DatabaseName)).
					Collection(CollectionDevices)
				query := bson.M{
					StorageKeyId: dd.Id,
				}
				err := collDevs.FindOne(ctx, query).Decode(&deployment)
				assert.NoError(t, err)
				assert.Equal(t, tc.request, deployment.Request)
			}
		})
	}
}
