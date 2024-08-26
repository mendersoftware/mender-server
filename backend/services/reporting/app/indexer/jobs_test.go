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

package indexer

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mendersoftware/mender-server/services/reporting/client/deployments"
	deployments_mocks "github.com/mendersoftware/mender-server/services/reporting/client/deployments/mocks"
	"github.com/mendersoftware/mender-server/services/reporting/client/deviceauth"
	deviceauth_mocks "github.com/mendersoftware/mender-server/services/reporting/client/deviceauth/mocks"
	"github.com/mendersoftware/mender-server/services/reporting/client/inventory"
	inventory_mocks "github.com/mendersoftware/mender-server/services/reporting/client/inventory/mocks"
	nats_mocks "github.com/mendersoftware/mender-server/services/reporting/client/nats/mocks"
	"github.com/mendersoftware/mender-server/services/reporting/model"
	store_mocks "github.com/mendersoftware/mender-server/services/reporting/store/mocks"
)

func TestGetJobsSubscriptionError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	jobs := make(chan model.Job, 1)

	subscriptionError := errors.New("subscription error")

	nats := &nats_mocks.Client{}
	nats.On("JetStreamSubscribe",
		ctx,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("chan model.Job"),
	).Return(subscriptionError)

	defer nats.AssertExpectations(t)

	indexer := NewIndexer(nil, nil, nats, nil, nil, nil)
	err := indexer.GetJobs(ctx, jobs)
	assert.Equal(t, "failed to subscribe to the nats JetStream: subscription error", err.Error())

	cancel()
}

func TestGetJobs(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	jobs := make(chan model.Job, 1)

	nats := &nats_mocks.Client{}
	nats.On("JetStreamSubscribe",
		ctx,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string"),
		mock.MatchedBy(func(msgs chan model.Job) bool {
			msgs <- model.Job{Action: model.ActionReindex}
			return true
		}),
	).Return(nil)

	defer nats.AssertExpectations(t)

	indexer := NewIndexer(nil, nil, nats, nil, nil, nil)
	err := indexer.GetJobs(ctx, jobs)
	assert.Nil(t, err)

	time.Sleep(500 * time.Millisecond)

	job := <-jobs
	assert.Equal(t, job.Action, model.ActionReindex)

	cancel()
}

func TestGetJobsError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobs := make(chan model.Job, 1)
	testErr := errors.New("test error")

	nats := &nats_mocks.Client{}
	nats.On("JetStreamSubscribe",
		ctx,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string"),
		mock.MatchedBy(func(msgs chan model.Job) bool {
			return true
		}),
	).Return(testErr)

	defer nats.AssertExpectations(t)

	indexer := NewIndexer(nil, nil, nats, nil, nil, nil)
	err := indexer.GetJobs(ctx, jobs)
	assert.ErrorIs(t, err, testErr)
}

func strptr(s string) *string {
	return &s
}

func TestProcessJobs(t *testing.T) {
	const tenantID = "tenant"

	testCases := map[string]struct {
		jobs []model.Job

		deviceauthDeviceIDs []string
		deviceauthDevices   map[string]deviceauth.DeviceAuthDevice
		deviceauthErr       error

		inventoryDeviceIDs []string
		inventoryDevices   []inventory.Device
		inventoryErr       error

		deploymentsDevices []deployments.LastDeviceDeployment
		deploymentsErr     error

		updateMapping       []string
		updateMappingResult []string

		bulkIndexDevices       []*model.Device
		bulkIndexRemoveDevices []*model.Device
		bulkIndexErr           error
	}{
		"ok": {
			jobs: []model.Job{
				{
					Action:   model.ActionReindex,
					TenantID: tenantID,
					DeviceID: "1",
					Service:  model.ServiceInventory,
				},
				{
					Action:   model.ActionReindex,
					TenantID: tenantID,
					DeviceID: "2",
					Service:  model.ServiceInventory,
				},
				{
					Action:   model.ActionReindex,
					TenantID: tenantID,
					DeviceID: "3",
					Service:  model.ServiceInventory,
				},
			},

			deviceauthDeviceIDs: []string{"1", "2", "3"},
			deviceauthDevices: map[string]deviceauth.DeviceAuthDevice{
				"1": {
					ID:     "1",
					Status: "active",
					IdDataStruct: map[string]string{
						"mac": "00:11:22:33:44",
					},
				},
				"2": {
					ID:     "2",
					Status: "pending",
					IdDataStruct: map[string]string{
						"mac": "00:11:22:33:55",
					},
				},
			},

			inventoryDeviceIDs: []string{"1", "2"},
			inventoryDevices: []inventory.Device{
				{
					ID: "1",
					Attributes: inventory.DeviceAttributes{
						{
							Scope: model.ScopeInventory,
							Name:  "mac",
							Value: "00:11:22:33:55",
						},
					},
				},
				{
					ID: "2",
				},
			},

			updateMapping:       []string{"inventory/mac"},
			updateMappingResult: []string{"inventory/mac"},

			bulkIndexDevices: []*model.Device{
				{
					ID:       strptr("1"),
					TenantID: strptr(tenantID),
					IdentityAttributes: model.InventoryAttributes{
						{
							Scope:  model.ScopeIdentity,
							Name:   model.AttrNameStatus,
							String: []string{"active"},
						},
						{
							Scope:  model.ScopeIdentity,
							Name:   "mac",
							String: []string{"00:11:22:33:44"},
						},
					},
					InventoryAttributes: model.InventoryAttributes{
						{
							Scope:  model.ScopeInventory,
							Name:   "attribute1",
							String: []string{"00:11:22:33:55"},
						},
					},
				},
				{
					ID:       strptr("2"),
					TenantID: strptr(tenantID),
					IdentityAttributes: model.InventoryAttributes{
						{
							Scope:  model.ScopeIdentity,
							Name:   model.AttrNameStatus,
							String: []string{"pending"},
						},
						{
							Scope:  model.ScopeIdentity,
							Name:   "mac",
							String: []string{"00:11:22:33:55"},
						},
					},
				},
			},
			bulkIndexRemoveDevices: []*model.Device{
				{
					ID:       strptr("3"),
					TenantID: strptr(tenantID),
				},
			},
		},
		"ok with latest deployment": {
			jobs: []model.Job{
				{
					Action:   model.ActionReindex,
					TenantID: tenantID,
					DeviceID: "1",
					Service:  model.ServiceInventory,
				},
				{
					Action:   model.ActionReindex,
					TenantID: tenantID,
					DeviceID: "2",
					Service:  model.ServiceInventory,
				},
				{
					Action:   model.ActionReindex,
					TenantID: tenantID,
					DeviceID: "3",
					Service:  model.ServiceInventory,
				},
			},

			deviceauthDeviceIDs: []string{"1", "2", "3"},
			deviceauthDevices: map[string]deviceauth.DeviceAuthDevice{
				"1": {
					ID:     "1",
					Status: "active",
					IdDataStruct: map[string]string{
						"mac": "00:11:22:33:44",
					},
				},
				"2": {
					ID:     "2",
					Status: "pending",
					IdDataStruct: map[string]string{
						"mac": "00:11:22:33:55",
					},
				},
			},

			inventoryDeviceIDs: []string{"1", "2"},
			inventoryDevices: []inventory.Device{
				{
					ID: "1",
				},
				{
					ID: "2",
				},
			},
			deploymentsDevices: []deployments.LastDeviceDeployment{
				{
					DeviceID:               "1",
					DeviceDeploymentStatus: "success",
				},
				{
					DeviceID:               "2",
					DeviceDeploymentStatus: "failure",
				},
			},
			deploymentsErr: nil,

			updateMapping:       []string{},
			updateMappingResult: []string{},

			bulkIndexDevices: []*model.Device{
				{
					ID:       strptr("1"),
					TenantID: strptr(tenantID),
					IdentityAttributes: model.InventoryAttributes{
						{
							Scope:  model.ScopeIdentity,
							Name:   model.AttrNameStatus,
							String: []string{"active"},
						},
						{
							Scope:  model.ScopeIdentity,
							Name:   "mac",
							String: []string{"00:11:22:33:44"},
						},
					},
					SystemAttributes: model.InventoryAttributes{
						{
							Scope:  model.ScopeSystem,
							Name:   model.AttrNameLatestDeploymentStatus,
							String: []string{"success"},
						},
					},
				},
				{
					ID:       strptr("2"),
					TenantID: strptr(tenantID),
					IdentityAttributes: model.InventoryAttributes{
						{
							Scope:  model.ScopeIdentity,
							Name:   model.AttrNameStatus,
							String: []string{"pending"},
						},
						{
							Scope:  model.ScopeIdentity,
							Name:   "mac",
							String: []string{"00:11:22:33:55"},
						},
					},
					SystemAttributes: model.InventoryAttributes{
						{
							Scope:  model.ScopeSystem,
							Name:   model.AttrNameLatestDeploymentStatus,
							String: []string{"failure"},
						},
					},
				},
			},
			bulkIndexRemoveDevices: []*model.Device{
				{
					ID:       strptr("3"),
					TenantID: strptr(tenantID),
				},
			},
		},
		"ko, failure in deviceauth": {
			jobs: []model.Job{
				{
					Action:   model.ActionReindex,
					TenantID: tenantID,
					DeviceID: "1",
					Service:  model.ServiceInventory,
				},
				{
					Action:   model.ActionReindex,
					TenantID: tenantID,
					DeviceID: "2",
					Service:  model.ServiceInventory,
				},
				{
					Action:   model.ActionReindex,
					TenantID: tenantID,
					DeviceID: "3",
					Service:  model.ServiceInventory,
				},
			},

			deviceauthDeviceIDs: []string{"1", "2", "3"},
			deviceauthErr:       errors.New("abc"),
		},
		"ko, failure in inventory": {
			jobs: []model.Job{
				{
					Action:   model.ActionReindex,
					TenantID: tenantID,
					DeviceID: "1",
					Service:  model.ServiceInventory,
				},
				{
					Action:   model.ActionReindex,
					TenantID: tenantID,
					DeviceID: "2",
					Service:  model.ServiceInventory,
				},
				{
					Action:   model.ActionReindex,
					TenantID: tenantID,
					DeviceID: "3",
					Service:  model.ServiceInventory,
				},
			},

			deviceauthDeviceIDs: []string{"1", "2", "3"},
			deviceauthDevices: map[string]deviceauth.DeviceAuthDevice{
				"1": {
					ID:     "1",
					Status: "active",
					IdDataStruct: map[string]string{
						"mac": "00:11:22:33:44",
					},
				},
				"2": {
					ID:     "2",
					Status: "pending",
					IdDataStruct: map[string]string{
						"mac": "00:11:22:33:55",
					},
				},
			},

			inventoryDeviceIDs: []string{"1", "2", "3"},
			inventoryErr:       errors.New("abc"),
		},
		"ko, failure in BulkIndex": {
			jobs: []model.Job{
				{
					Action:   model.ActionReindex,
					TenantID: tenantID,
					DeviceID: "1",
					Service:  model.ServiceInventory,
				},
				{
					Action:   model.ActionReindex,
					TenantID: tenantID,
					DeviceID: "2",
					Service:  model.ServiceInventory,
				},
				{
					Action:   model.ActionReindex,
					TenantID: tenantID,
					DeviceID: "3",
					Service:  model.ServiceInventory,
				},
			},

			deviceauthDeviceIDs: []string{"1", "2", "3"},
			deviceauthDevices: map[string]deviceauth.DeviceAuthDevice{
				"1": {
					ID:     "1",
					Status: "active",
					IdDataStruct: map[string]string{
						"mac": "00:11:22:33:44",
					},
				},
				"2": {
					ID:     "2",
					Status: "pending",
					IdDataStruct: map[string]string{
						"mac": "00:11:22:33:55",
					},
				},
			},

			inventoryDeviceIDs: []string{"1", "2"},
			inventoryDevices: []inventory.Device{
				{
					ID: "1",
				},
				{
					ID: "2",
				},
			},

			updateMapping:       []string{},
			updateMappingResult: []string{},

			bulkIndexDevices: []*model.Device{
				{
					ID:       strptr("1"),
					TenantID: strptr(tenantID),
					IdentityAttributes: model.InventoryAttributes{
						{
							Scope:  model.ScopeIdentity,
							Name:   model.AttrNameStatus,
							String: []string{"active"},
						},
						{
							Scope:  model.ScopeIdentity,
							Name:   "mac",
							String: []string{"00:11:22:33:44"},
						},
					},
				},
				{
					ID:       strptr("2"),
					TenantID: strptr(tenantID),
					IdentityAttributes: model.InventoryAttributes{
						{
							Scope:  model.ScopeIdentity,
							Name:   model.AttrNameStatus,
							String: []string{"pending"},
						},
						{
							Scope:  model.ScopeIdentity,
							Name:   "mac",
							String: []string{"00:11:22:33:55"},
						},
					},
				},
			},
			bulkIndexRemoveDevices: []*model.Device{
				{
					ID:       strptr("3"),
					TenantID: strptr(tenantID),
				},
			},
			bulkIndexErr: errors.New("bulk index error"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			store := &store_mocks.Store{}
			defer store.AssertExpectations(t)

			if len(tc.bulkIndexDevices) > 0 || len(tc.bulkIndexRemoveDevices) > 0 {
				store.On("BulkIndexDevices",
					ctx,
					tc.bulkIndexDevices,
					tc.bulkIndexRemoveDevices,
				).Return(tc.bulkIndexErr)
			}

			devClient := &deviceauth_mocks.Client{}
			defer devClient.AssertExpectations(t)

			devClient.On("GetDevices",
				ctx,
				tenantID,
				mock.MatchedBy(func(ids []string) bool {
					sort.Strings(ids)
					assert.Equal(t, ids, tc.deviceauthDeviceIDs)

					return true
				}),
			).Return(tc.deviceauthDevices, tc.deviceauthErr)

			invClient := &inventory_mocks.Client{}
			defer invClient.AssertExpectations(t)

			if tc.deviceauthErr == nil {
				invClient.On("GetDevices",
					ctx,
					tenantID,
					mock.MatchedBy(func(ids []string) bool {
						sort.Strings(ids)
						assert.Equal(t, ids, tc.deviceauthDeviceIDs)

						return true
					}),
				).Return(tc.inventoryDevices, tc.inventoryErr)
			}

			deplClient := &deployments_mocks.Client{}
			defer deplClient.AssertExpectations(t)
			if tc.deviceauthErr == nil && tc.inventoryErr == nil {
				deplClient.On("GetLatestFinishedDeployment",
					ctx,
					tenantID,
					mock.AnythingOfType("[]string"),
				).Return(tc.deploymentsDevices, tc.deploymentsErr)
			}

			ds := &store_mocks.DataStore{}
			ds.On("UpdateAndGetMapping",
				ctx,
				tenantID,
				tc.updateMapping,
			).Return(&model.Mapping{
				TenantID:  tenantID,
				Inventory: tc.updateMappingResult,
			}, nil)

			indexer := NewIndexer(store, ds, nil, devClient, invClient, deplClient)

			indexer.ProcessJobs(ctx, tc.jobs)
		})
	}
}

func TestProcessJobsDeployments(t *testing.T) {
	const tenantID = "tenant"

	now := time.Now().Truncate(0)
	five_seconds_ago := now.Add(-5 * time.Second)

	testCases := map[string]struct {
		jobs []model.Job

		getDeploymentsIDs []string
		getDeployments    []*deployments.DeviceDeployment
		getDeploymentsErr error

		bulkIndexDeployments []*model.Deployment
		bulkIndexErr         error
	}{
		"ok": {
			jobs: []model.Job{
				{
					Action:   model.ActionReindexDeployment,
					TenantID: tenantID,
					ID:       "92be929e-f924-49d0-9b98-3dec6c504901",
					Service:  model.ServiceDeployments,
				},
				{
					Action:   model.ActionReindexDeployment,
					TenantID: tenantID,
					ID:       "92be929e-f924-49d0-9b98-3dec6c504902",
					Service:  model.ServiceDeployments,
				},
				{
					Action:   model.ActionReindexDeployment,
					TenantID: tenantID,
					ID:       "92be929e-f924-49d0-9b98-3dec6c504903",
					Service:  model.ServiceInventory,
				},
			},

			getDeployments: []*deployments.DeviceDeployment{
				{
					ID:         "92be929e-f924-49d0-9b98-3dec6c504901",
					Deployment: &deployments.Deployment{},
					Device: &deployments.Device{
						Created:  &five_seconds_ago,
						Finished: &now,
						Status:   "finished",
						Image: &deployments.Image{
							Info: &deployments.ArtifactInfo{},
						},
					},
				},
				{
					ID:         "92be929e-f924-49d0-9b98-3dec6c504902",
					Deployment: &deployments.Deployment{},
					Device: &deployments.Device{
						Created:  &five_seconds_ago,
						Finished: &now,
						Status:   "finished",
						Image: &deployments.Image{
							Info: &deployments.ArtifactInfo{},
						},
					},
				},
				{
					ID:         "92be929e-f924-49d0-9b98-3dec6c504903",
					Deployment: &deployments.Deployment{},
					Device: &deployments.Device{
						Created: &five_seconds_ago,
						Status:  "downloading",
						Image: &deployments.Image{
							Info: &deployments.ArtifactInfo{},
						},
					},
				},
			},
			bulkIndexDeployments: []*model.Deployment{
				{
					ID:                   "92be929e-f924-49d0-9b98-3dec6c504901",
					TenantID:             tenantID,
					DeviceCreated:        &five_seconds_ago,
					DeviceFinished:       &now,
					DeviceElapsedSeconds: 5,
					DeviceStatus:         "finished",
				},
				{
					ID:                   "92be929e-f924-49d0-9b98-3dec6c504902",
					TenantID:             tenantID,
					DeviceCreated:        &five_seconds_ago,
					DeviceFinished:       &now,
					DeviceElapsedSeconds: 5,
					DeviceStatus:         "finished",
				},
				{
					ID:            "92be929e-f924-49d0-9b98-3dec6c504903",
					TenantID:      tenantID,
					DeviceCreated: &five_seconds_ago,
					DeviceStatus:  "downloading",
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			store := &store_mocks.Store{}
			defer store.AssertExpectations(t)

			if tc.getDeploymentsErr == nil {
				store.On("BulkIndexDeployments",
					ctx,
					mock.MatchedBy(func(deployments []*model.Deployment) bool {
						for _, i := range deployments {
							found := false
							for _, j := range tc.bulkIndexDeployments {
								if assert.ObjectsAreEqual(i, j) {
									found = true
									break
								}
							}
							if !found {
								return false
							}
						}

						return true
					}),
				).Return(tc.bulkIndexErr)
			}

			deplClient := &deployments_mocks.Client{}
			defer deplClient.AssertExpectations(t)

			deplClient.On("GetDeployments",
				ctx,
				tenantID,
				mock.AnythingOfType("[]string"),
			).Return(tc.getDeployments, tc.getDeploymentsErr)

			indexer := NewIndexer(store, nil, nil, nil, nil, deplClient)
			indexer.ProcessJobs(ctx, tc.jobs)
		})
	}
}
