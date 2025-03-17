// Copyright 2023 Northern.tech AS
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

package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mendersoftware/mender-server/services/deployments/model"
	fs_mocks "github.com/mendersoftware/mender-server/services/deployments/storage/mocks"
	"github.com/mendersoftware/mender-server/services/deployments/store/mocks"
	"github.com/mendersoftware/mender-server/services/deployments/store/mongo"
)

// separate set of tests for assert if correct deployment status tracking

func TestUpdateDeviceDeploymentStatus(t *testing.T) {
	ctx := context.TODO()

	testCases := map[string]struct {
		ddStatueNew model.DeviceDeploymentState
		deployment  *model.Deployment

		deviceStatus model.DeviceDeploymentStatus

		getDeviceDepErr error

		notFounddDepByID bool
		findDepByIDErr   error

		err error
	}{
		"ok": {
			ddStatueNew: model.DeviceDeploymentState{
				Status: model.DeviceDeploymentStatusInstalling,
			},
			deviceStatus: model.DeviceDeploymentStatusDownloading,
		},
		"error: device deployment not found": {
			ddStatueNew: model.DeviceDeploymentState{
				Status: model.DeviceDeploymentStatusInstalling,
			},
			getDeviceDepErr: mongo.ErrStorageNotFound,
			err:             ErrStorageNotFound,
		},
		"error: deployment not found by id": {
			ddStatueNew: model.DeviceDeploymentState{
				Status: model.DeviceDeploymentStatusInstalling,
			},
			notFounddDepByID: true,
			err:              ErrModelDeploymentNotFound,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			depName := "foo"
			depArtifact := "bar"
			fakeDeployment, err := model.NewDeploymentFromConstructor(
				&model.DeploymentConstructor{
					Name:         depName,
					ArtifactName: depArtifact,
					Devices:      []string{"baz"},
				},
			)
			fakeDeployment.MaxDevices = 1
			assert.NoError(t, err)

			devId := "somedevice"
			fakeDeviceDeployment := model.NewDeviceDeployment(
				devId, fakeDeployment.Id)
			fakeDeviceDeployment.Status = tc.deviceStatus

			fs := &fs_mocks.ObjectStorage{}
			db := mocks.DataStore{}

			db.On("GetDeviceDeployment", ctx,
				fakeDeployment.Id, devId, false).Return(
				fakeDeviceDeployment, tc.getDeviceDepErr).Once()
			if tc.getDeviceDepErr == nil {
				db.On("UpdateDeviceDeploymentStatus", ctx,
					devId,
					fakeDeployment.Id,
					mock.MatchedBy(func(ddStatus model.DeviceDeploymentState) bool {
						assert.Equal(t, tc.ddStatueNew.Status, ddStatus.Status)
						return true
					}),
					mock.AnythingOfType("model.DeviceDeploymentStatus"),
				).Return(tc.deviceStatus, nil).Once()
				db.On("UpdateStatsInc", ctx,
					fakeDeployment.Id,
					tc.deviceStatus,
					tc.ddStatueNew.Status).Run(func(args mock.Arguments) {
					// fake updated stats
					fakeDeployment.Stats.Inc(tc.ddStatueNew.Status)
				}).Return(fakeDeployment.Stats, nil).Once()
				foundDepById := fakeDeployment
				if tc.notFounddDepByID {
					foundDepById = nil
				}
				db.On("FindDeploymentByID", ctx, fakeDeployment.Id).Return(
					foundDepById, tc.findDepByIDErr).Once()
				if tc.findDepByIDErr == nil && foundDepById != nil {
					db.On("SetDeploymentStatus", ctx,
						fakeDeployment.Id,
						model.DeploymentStatusInProgress,
						mock.AnythingOfType("time.Time")).Return(nil).Once()

				}
			}
			ds := NewDeployments(&db, fs, 0, false)

			err = ds.UpdateDeviceDeploymentStatus(ctx, fakeDeployment.Id, fakeDeviceDeployment.DeviceId, tc.ddStatueNew)

			assert.ErrorIs(t, err, tc.err)
		})
	}
}

func TestGetDeploymentForDeviceWithCurrent(t *testing.T) {
	ctx := context.TODO()

	// for simplicity - alreadyinstalled case
	devId := "somedevice"
	devType := "baz"

	depName := "foo"
	depArtifact := "bar"

	request := &model.DeploymentNextRequest{
		DeviceProvides: &model.InstalledDeviceDeployment{
			ArtifactName: depArtifact,
			DeviceType:   devType,
		},
	}

	fakeDeployment, err := model.NewDeploymentFromConstructor(
		&model.DeploymentConstructor{
			Name:         depName,
			ArtifactName: depArtifact,
			Devices:      []string{devType},
		},
	)
	assert.NoError(t, err)
	fakeDeployment.MaxDevices = 1

	fakeImage := &model.Image{
		ArtifactMeta: &model.ArtifactMeta{
			Name:           depArtifact,
			Provides:       map[string]string{"bar": "baz"},
			ClearsProvides: []string{"foo"},
		},
		Size: 5,
	}

	fakeDeviceDeployment := model.NewDeviceDeployment(
		devId, fakeDeployment.Id)
	fakeDeviceDeployment.Status = model.DeviceDeploymentStatusPending

	fs := &fs_mocks.ObjectStorage{}
	db := mocks.DataStore{}

	db.On("FindOldestActiveDeviceDeployment", ctx, devId).Return(
		fakeDeviceDeployment, nil)

	db.On("FindDeploymentByID", ctx, fakeDeployment.Id).Return(
		fakeDeployment, nil).Once()

	db.On("DeviceCountByDeployment", ctx, fakeDeployment.Id).Return(2, nil)
	db.On("GetDeviceDeployment", ctx,
		fakeDeployment.Id, fakeDeviceDeployment.DeviceId, false).Return(
		fakeDeviceDeployment, nil)

	db.On("IncrementDeviceDeploymentAttempts", ctx,
		fakeDeviceDeployment.Id, uint(1)).Return(nil)

	db.On("UpdateDeviceDeploymentStatus", ctx,
		fakeDeviceDeployment.DeviceId,
		fakeDeployment.Id,

		mock.MatchedBy(func(ddStatus model.DeviceDeploymentState) bool {
			assert.Equal(t, model.DeviceDeploymentStatusAlreadyInst, ddStatus.Status)

			return true
		}),
		mock.AnythingOfType("model.DeviceDeploymentStatus"),
	).Return(model.DeviceDeploymentStatusPending, nil)

	db.On("UpdateStatsInc", ctx,
		fakeDeployment.Id,
		model.DeviceDeploymentStatusPending,
		model.DeviceDeploymentStatusAlreadyInst).Run(func(args mock.Arguments) {
		// fake updated stats
		fakeDeployment.Stats.Inc(model.DeviceDeploymentStatusAlreadyInst)
	}).Return(fakeDeployment.Stats, nil).Once()

	// fake updated stats
	fakeDeployment.Stats.Set(model.DeviceDeploymentStatusNoArtifact, 1)
	db.On("FindDeploymentByID", ctx, fakeDeployment.Id).Return(
		fakeDeployment, nil)

	db.On("SetDeploymentStatus", ctx,
		fakeDeployment.Id,
		model.DeploymentStatusFinished,
		mock.AnythingOfType("time.Time")).Return(nil)

	db.On("SaveDeviceDeploymentRequest", ctx,
		mock.AnythingOfType("string"),
		request).Return(nil)

	db.On("ImageByNameAndDeviceType", ctx, depArtifact, devType).Return(
		fakeImage, nil)

	db.On("AssignArtifact", ctx,
		fakeDeviceDeployment.DeviceId,
		fakeDeviceDeployment.DeploymentId,
		fakeImage).Return(nil)

	db.On("SaveLastDeviceDeploymentStatus", ctx,
		mock.AnythingOfType("model.DeviceDeployment"),
	).Return(nil)

	ds := NewDeployments(&db, fs, 0, false)

	_, err = ds.GetDeploymentForDeviceWithCurrent(ctx, devId, request)
	assert.NoError(t, err)
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func intPtr(i int) *int {
	return &i
}

func TestDecommissionDevice(t *testing.T) {
	testCases := map[string]struct {
		inputDeviceId       string
		inputDeploymentId   string
		inputDeploymentName string
		inputArtifactName   string
		inputMaxDevices     int
		inputStats          model.Stats
		inputDevices        []string

		deviceDeployments                                     []model.DeviceDeployment
		findOldestDeploymentForDeviceIDWithStatusesDeployment *model.DeviceDeployment
		findOldestDeploymentForDeviceIDWithStatusesError      error
		getDeviceDeploymentDeployment                         *model.DeviceDeployment
		getDeviceDeploymentError                              error
		updateDeviceDeploymentStatusStatus                    model.DeviceDeploymentStatus
		updateDeviceDeploymentStatusError                     error
		findLatestDeploymentForDeviceIDWithStatusesDeployment *model.DeviceDeployment
		findLatestDeploymentForDeviceIDWithStatusesError      error
		findNewerActiveDeploymentDeployment                   *model.Deployment
		findNewerActiveDeploymentError                        error
		findDeploymentByIDDeployment                          *model.Deployment
		findDeploymentByIDError                               error
		insertDeviceDeploymentError                           error
		updateStatsIncError                                   error
		setDeploymentStatusError                              error

		outputError error
	}{
		"ok": {
			inputDeviceId:       "foo",
			inputDeploymentId:   "bar",
			inputDeploymentName: "foo",
			inputDevices:        []string{"baz"},

			findOldestDeploymentForDeviceIDWithStatusesDeployment: &model.DeviceDeployment{
				Id:           "bar",
				DeviceId:     "foo",
				DeploymentId: "bar",
				Status:       model.DeviceDeploymentStatusDownloading,
			},
			getDeviceDeploymentDeployment: &model.DeviceDeployment{
				Id:           "bar",
				DeviceId:     "foo",
				DeploymentId: "bar",
				Status:       model.DeviceDeploymentStatusDownloading,
			},
			updateDeviceDeploymentStatusStatus: model.DeviceDeploymentStatusDownloading,
			findDeploymentByIDDeployment: &model.Deployment{
				Id:         "bar",
				MaxDevices: 1,
				Stats:      model.Stats{"decommissioned": 1},
			},
		},
		"ok 1": {
			findLatestDeploymentForDeviceIDWithStatusesDeployment: &model.DeviceDeployment{
				Id:           "bar",
				DeploymentId: "bar",
				Status:       model.DeviceDeploymentStatusSuccess,
				Created:      timePtr(time.Now()),
			},
		},
		"ok 2": {},
		"ok 3": {
			findNewerActiveDeploymentDeployment: nil,
		},
		"ok 4": {
			inputDeviceId:     "foo",
			inputDeploymentId: "foo",
			findNewerActiveDeploymentDeployment: &model.Deployment{
				DeviceList:  []string{"foo"},
				Id:          "foo",
				Created:     timePtr(time.Now()),
				DeviceCount: intPtr(0),
				MaxDevices:  1,
				Stats:       model.Stats{},
			},
		},
		"ok, pending": {
			inputDeviceId:     "foo",
			inputDeploymentId: "pending",
			findNewerActiveDeploymentDeployment: &model.Deployment{
				DeviceList:  []string{"foo"},
				Id:          "pending",
				Created:     timePtr(time.Now()),
				DeviceCount: intPtr(0),
				MaxDevices:  2,
				Stats:       model.Stats{},
			},
		},
		"FindOldestActiveDeviceDeployment error": {
			inputDeviceId:       "foo",
			inputDeploymentId:   "bar",
			inputDeploymentName: "foo",
			inputDevices:        []string{"baz"},

			findOldestDeploymentForDeviceIDWithStatusesError: errors.New("foo"),

			outputError: errors.New("Searching for active deployment for the device: foo"),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()
			db := mocks.DataStore{}

			db.On("FindOldestActiveDeviceDeployment",
				ctx, tc.inputDeviceId).
				Return(
					tc.findOldestDeploymentForDeviceIDWithStatusesDeployment,
					tc.findOldestDeploymentForDeviceIDWithStatusesError,
				)

			db.On("GetDeviceDeployment", ctx, tc.inputDeploymentId,
				tc.inputDeviceId, mock.AnythingOfType("bool")).Return(
				tc.getDeviceDeploymentDeployment, tc.getDeviceDeploymentError)

			db.On("UpdateDeviceDeploymentStatus", ctx, tc.inputDeviceId,
				tc.inputDeploymentId, mock.AnythingOfType("model.DeviceDeploymentState"),
				mock.AnythingOfType("model.DeviceDeploymentStatus")).Return(
				tc.updateDeviceDeploymentStatusStatus, tc.updateDeviceDeploymentStatusError)

			db.On("FindLatestInactiveDeviceDeployment",
				ctx, tc.inputDeviceId,
			).Return(
				tc.findLatestDeploymentForDeviceIDWithStatusesDeployment,
				tc.findLatestDeploymentForDeviceIDWithStatusesError,
			)

			db.On("FindNewerActiveDeployment", ctx, mock.AnythingOfType("*time.Time"),
				tc.inputDeviceId).Return(
				tc.findNewerActiveDeploymentDeployment, tc.findNewerActiveDeploymentError)

			db.On("FindNewerActiveDeployment", ctx, mock.AnythingOfType("*time.Time"),
				100, 100).Return(nil, nil)

			db.On("InsertDeviceDeployment", ctx, mock.AnythingOfType("*model.DeviceDeployment"), true).Return(
				tc.insertDeviceDeploymentError)

			db.On("FindDeploymentByID", ctx, tc.inputDeploymentId).Return(
				tc.findDeploymentByIDDeployment, tc.findDeploymentByIDError)

			var stats model.Stats
			if tc.findDeploymentByIDDeployment != nil {
				stats = tc.findDeploymentByIDDeployment.Stats
			}
			db.On("UpdateStatsInc", ctx, tc.inputDeploymentId,
				tc.updateDeviceDeploymentStatusStatus,
				model.DeviceDeploymentStatusDecommissioned).
				Run(func(args mock.Arguments) {
					if stats != nil {
						stats.Inc(model.DeviceDeploymentStatusDecommissioned)
					}
				}).
				Return(stats, tc.updateStatsIncError).
				Once()

			db.On("SetDeploymentStatus", ctx,
				tc.inputDeploymentId,
				model.DeploymentStatusFinished,
				mock.AnythingOfType("time.Time")).
				Return(tc.setDeploymentStatusError).
				Once()

			db.On("SetDeploymentStatus", ctx,
				"pending",
				model.DeploymentStatusPending,
				mock.AnythingOfType("time.Time")).
				Return(tc.setDeploymentStatusError).
				Once()

			db.On("SaveLastDeviceDeploymentStatus", ctx,
				mock.AnythingOfType("model.DeviceDeployment"),
			).Return(nil)

			ds := NewDeployments(&db, nil, 0, false)

			err := ds.DecommissionDevice(ctx, tc.inputDeviceId)
			if tc.outputError != nil {
				assert.EqualError(t, err, tc.outputError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAbortDeviceDeployments(t *testing.T) {
	testCases := map[string]struct {
		inputDeviceId       string
		inputDeploymentId   string
		inputDeploymentName string
		inputArtifactName   string
		inputMaxDevices     int
		inputStats          model.Stats
		inputDevices        []string

		deviceDeployments                                     []model.DeviceDeployment
		findOldestDeploymentForDeviceIDWithStatusesDeployment *model.DeviceDeployment
		findOldestDeploymentForDeviceIDWithStatusesError      error
		getDeviceDeploymentDeployment                         *model.DeviceDeployment
		getDeviceDeploymentError                              error
		updateDeviceDeploymentStatusStatus                    model.DeviceDeploymentStatus
		updateDeviceDeploymentStatusError                     error
		findLatestDeploymentForDeviceIDWithStatusesDeployment *model.DeviceDeployment
		findLatestDeploymentForDeviceIDWithStatusesError      error
		findNewerActiveDeploymentDeployment                   *model.Deployment
		findNewerActiveDeploymentError                        error
		findDeploymentByIDDeployment                          *model.Deployment
		findDeploymentByIDError                               error
		insertDeviceDeploymentError                           error
		updateStatsIncError                                   error
		setDeploymentStatusError                              error
		isDeploymentInProgress                                bool

		outputError error
	}{
		"ok": {
			inputDeviceId:       "foo",
			inputDeploymentId:   "bar",
			inputDeploymentName: "foo",
			inputDevices:        []string{"baz"},

			findOldestDeploymentForDeviceIDWithStatusesDeployment: &model.DeviceDeployment{
				Id:           "bar",
				DeviceId:     "foo",
				DeploymentId: "bar",
				Status:       model.DeviceDeploymentStatusDownloading,
			},
			getDeviceDeploymentDeployment: &model.DeviceDeployment{
				Id:           "bar",
				DeviceId:     "foo",
				DeploymentId: "bar",
				Status:       model.DeviceDeploymentStatusDownloading,
			},
			updateDeviceDeploymentStatusStatus: model.DeviceDeploymentStatusDownloading,
			findDeploymentByIDDeployment: &model.Deployment{
				Id:         "bar",
				MaxDevices: 1,
				Stats:      model.Stats{"decommissioned": 1},
			},
		},
		"ok 1": {
			findLatestDeploymentForDeviceIDWithStatusesDeployment: &model.DeviceDeployment{
				Id:           "bar",
				DeploymentId: "bar",
				Status:       model.DeviceDeploymentStatusSuccess,
				Created:      timePtr(time.Now()),
			},
		},
		"ok 2": {},
		"ok 3": {
			findNewerActiveDeploymentDeployment: nil,
		},
		"ok 4": {
			inputDeviceId:     "foo",
			inputDeploymentId: "foo",
			findNewerActiveDeploymentDeployment: &model.Deployment{
				DeviceList:  []string{"foo"},
				Id:          "foo",
				Created:     timePtr(time.Now()),
				DeviceCount: intPtr(0),
				MaxDevices:  1,
				Stats:       model.Stats{},
			},
		},
		"ok, pending": {
			inputDeviceId:     "foo",
			inputDeploymentId: "pending",
			findNewerActiveDeploymentDeployment: &model.Deployment{
				DeviceList:  []string{"foo"},
				Id:          "pending",
				Created:     timePtr(time.Now()),
				DeviceCount: intPtr(0),
				MaxDevices:  1,
				Stats:       model.Stats{},
			},
		},
		"ok, pending with max devices = 2": {
			inputDeviceId:     "foo",
			inputDeploymentId: "pending",
			findNewerActiveDeploymentDeployment: &model.Deployment{
				DeviceList:  []string{"foo"},
				Id:          "pending",
				Created:     timePtr(time.Now()),
				DeviceCount: intPtr(0),
				MaxDevices:  2,
				Stats:       model.Stats{},
			},
			isDeploymentInProgress: true,
		},
		"FindOldestDeploymentForDeviceIDWithStatuses error": {
			inputDeviceId:       "foo",
			inputDeploymentId:   "bar",
			inputDeploymentName: "foo",
			inputDevices:        []string{"baz"},

			findOldestDeploymentForDeviceIDWithStatusesError: errors.New("foo"),

			outputError: errors.New("Searching for active deployment for the device: foo"),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()
			db := mocks.DataStore{}

			db.On("FindOldestActiveDeviceDeployment",
				ctx, tc.inputDeviceId).
				Return(
					tc.findOldestDeploymentForDeviceIDWithStatusesDeployment,
					tc.findOldestDeploymentForDeviceIDWithStatusesError,
				)

			db.On("GetDeviceDeployment", ctx, tc.inputDeploymentId,
				tc.inputDeviceId, mock.AnythingOfType("bool")).Return(
				tc.getDeviceDeploymentDeployment, tc.getDeviceDeploymentError)

			db.On("UpdateDeviceDeploymentStatus", ctx, tc.inputDeviceId,
				tc.inputDeploymentId, mock.AnythingOfType("model.DeviceDeploymentState"),
				mock.AnythingOfType("model.DeviceDeploymentStatus")).Return(
				tc.updateDeviceDeploymentStatusStatus, tc.updateDeviceDeploymentStatusError)

			db.On("FindLatestInactiveDeviceDeployment",
				ctx, tc.inputDeviceId,
			).Return(
				tc.findLatestDeploymentForDeviceIDWithStatusesDeployment,
				tc.findLatestDeploymentForDeviceIDWithStatusesError,
			)

			db.On("FindNewerActiveDeployment", ctx, mock.AnythingOfType("*time.Time"),
				tc.inputDeviceId).Return(
				tc.findNewerActiveDeploymentDeployment, tc.findNewerActiveDeploymentError).
				Once()

			db.On("FindNewerActiveDeployments", ctx, mock.AnythingOfType("*time.Time"),
				tc.inputDeviceId).Return(nil, nil).
				Once()

			db.On("InsertDeviceDeployment", ctx, mock.AnythingOfType("*model.DeviceDeployment"), true).Return(
				tc.insertDeviceDeploymentError)

			db.On("FindDeploymentByID", ctx, tc.inputDeploymentId).Return(
				tc.findDeploymentByIDDeployment, tc.findDeploymentByIDError)

			var stats model.Stats
			if tc.findDeploymentByIDDeployment != nil {
				stats = tc.findDeploymentByIDDeployment.Stats
			}
			db.On("UpdateStatsInc", ctx, tc.inputDeploymentId,
				tc.updateDeviceDeploymentStatusStatus,
				model.DeviceDeploymentStatusAborted).
				Run(func(args mock.Arguments) {
					if stats != nil {
						stats.Inc(model.DeviceDeploymentStatusAborted)
					}
				}).
				Return(
					stats,
					tc.updateStatsIncError,
				)

			status := model.DeploymentStatusFinished
			if tc.isDeploymentInProgress {
				status = model.DeploymentStatusInProgress
			}
			db.On("SetDeploymentStatus", ctx,
				tc.inputDeploymentId,
				status,
				mock.AnythingOfType("time.Time")).
				Return(tc.setDeploymentStatusError).
				Once()

			db.On("SetDeploymentStatus", ctx,
				"pending",
				model.DeploymentStatusPending,
				mock.AnythingOfType("time.Time")).
				Return(tc.setDeploymentStatusError).
				Once()

			db.On("SaveLastDeviceDeploymentStatus", ctx,
				mock.AnythingOfType("model.DeviceDeployment"),
			).Return(nil)

			ds := NewDeployments(&db, nil, 0, false)

			err := ds.AbortDeviceDeployments(ctx, tc.inputDeviceId)
			if tc.outputError != nil {
				assert.EqualError(t, err, tc.outputError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
