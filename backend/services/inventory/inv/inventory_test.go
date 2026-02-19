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

package inv

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"

	mdm "github.com/mendersoftware/mender-server/services/inventory/client/devicemonitor/mocks"
	mworkflows "github.com/mendersoftware/mender-server/services/inventory/client/workflows/mocks"
	"github.com/mendersoftware/mender-server/services/inventory/model"
	"github.com/mendersoftware/mender-server/services/inventory/store"
	mstore "github.com/mendersoftware/mender-server/services/inventory/store/mocks"
	"github.com/mendersoftware/mender-server/services/inventory/store/mongo"
	"github.com/mendersoftware/mender-server/services/inventory/utils"
)

func invForTest(d store.DataStore) InventoryApp {
	return &inventory{db: d}
}

func boolPtr(value bool) *bool {
	return &value
}

func timePtr(value time.Time) *time.Time {
	return &value
}

func TestHealthCheck(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name           string
		DataStoreError error
		WorkflowsError error
	}{{
		Name: "ok",
	}, {
		Name:           "error, error reaching MongoDB",
		DataStoreError: errors.New("connection refused"),
	}, {
		Name:           "error, error reaching workflows",
		WorkflowsError: errors.New("connection refused"),
	}}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.TODO()
			db := &mstore.DataStore{}
			defer db.AssertExpectations(t)
			db.On("Ping", ctx).Return(tc.DataStoreError)

			workflows := &mworkflows.Client{}
			defer workflows.AssertExpectations(t)
			if tc.DataStoreError == nil {
				workflows.On("CheckHealth", ctx).Return(tc.WorkflowsError)
			}

			inv := NewInventory(db).WithReporting(workflows)
			err := inv.HealthCheck(ctx)
			if tc.DataStoreError != nil {
				assert.EqualError(t, err,
					"error reaching MongoDB: "+
						tc.DataStoreError.Error(),
				)
			} else if tc.WorkflowsError != nil {
				assert.EqualError(t, err,
					"error reaching workflows: "+
						tc.WorkflowsError.Error(),
				)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInventoryListDevices(t *testing.T) {
	t.Parallel()

	group := model.GroupName("asd")
	testCases := map[string]struct {
		group           string
		inHasGroup      *bool
		datastoreFilter []store.Filter
		datastoreError  error
		outError        error
		outDevices      []model.Device
		outDeviceCount  int
	}{
		"has group nil": {
			inHasGroup:      nil,
			datastoreFilter: nil,
			datastoreError:  nil,
			outError:        nil,
			outDevices:      []model.Device{{ID: model.DeviceID("1")}},
			outDeviceCount:  1,
		},
		"has group true": {
			inHasGroup:      boolPtr(true),
			datastoreFilter: nil,
			datastoreError:  nil,
			outError:        nil,
			outDevices:      []model.Device{{ID: model.DeviceID("1"), Group: group}},
			outDeviceCount:  1,
		},
		"has group false": {
			inHasGroup:      boolPtr(false),
			datastoreFilter: nil,
			datastoreError:  nil,
			outError:        nil,
			outDevices:      []model.Device{{ID: model.DeviceID("1")}},
			outDeviceCount:  1,
		},
		"datastore error": {
			inHasGroup:      nil,
			datastoreFilter: nil,
			datastoreError:  errors.New("db connection failed"),
			outError:        errors.New("failed to fetch devices: db connection failed"),
			outDevices:      nil,
			outDeviceCount:  -1,
		},
		"get devices from group": {
			group: "asd",
			outDevices: []model.Device{
				{ID: model.DeviceID("1"), Group: group},
				{ID: model.DeviceID("2"), Group: group},
			},
			outDeviceCount: 2,
		},
	}

	for name, tc := range testCases {
		t.Logf("test case: %s", name)

		ctx := context.Background()

		db := &mstore.DataStore{}
		db.On("GetDevices",
			ctx,
			mock.AnythingOfType("store.ListQuery"),
		).Return(tc.outDevices, tc.outDeviceCount, tc.datastoreError)
		i := invForTest(db)

		devs, totalCount, err := i.ListDevices(ctx,
			store.ListQuery{
				Skip:      1,
				Limit:     10,
				Filters:   nil,
				Sort:      nil,
				HasGroup:  tc.inHasGroup,
				GroupName: tc.group})

		if tc.outError != nil {
			if assert.Error(t, err) {
				assert.EqualError(t, err, tc.outError.Error())
			}
		} else {
			assert.NoError(t, err)
			assert.Equal(t, len(devs), len(tc.outDevices))
			assert.Equal(t, totalCount, tc.outDeviceCount)
		}
	}
}

func TestInventoryGetDevice(t *testing.T) {
	t.Parallel()

	group := model.GroupName("asd")
	testCases := map[string]struct {
		devid          model.DeviceID
		datastoreError error
		outError       error
		outDevice      *model.Device
	}{
		"has device": {
			devid:     model.DeviceID("1"),
			outDevice: &model.Device{ID: model.DeviceID("1"), Group: group},
		},
		"no device": {
			devid:     model.DeviceID("2"),
			outDevice: nil,
		},
		"datastore error": {
			devid:          model.DeviceID("3"),
			datastoreError: errors.New("db connection failed"),
			outError:       errors.New("failed to fetch device: db connection failed"),
		},
	}

	for name, tc := range testCases {
		t.Logf("test case: %s", name)

		ctx := context.Background()

		db := &mstore.DataStore{}
		db.On("GetDevice",
			ctx,
			mock.AnythingOfType("model.DeviceID"),
		).Return(tc.outDevice, tc.datastoreError)
		i := invForTest(db)

		dev, err := i.GetDevice(ctx, tc.devid)

		if tc.outError != nil {
			if assert.Error(t, err) {
				assert.EqualError(t, err, tc.outError.Error())
			}
		} else {
			assert.NoError(t, err)
			if tc.outDevice != nil && assert.NotNil(t, dev) {
				assert.Equal(t, *tc.outDevice, *dev)
			} else {
				assert.Nil(t, dev)
			}
		}
	}
}

func TestInventoryAddDevice(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		inDevice       *model.Device
		datastoreError error
		outError       error
	}{
		"nil device": {
			inDevice:       nil,
			datastoreError: nil,
			outError:       errors.New("no device given"),
		},
		"datastore success": {
			inDevice:       &model.Device{ID: "1"},
			datastoreError: nil,
			outError:       nil,
		},
		"datastore error": {
			inDevice:       &model.Device{ID: "1"},
			datastoreError: errors.New("db connection failed"),
			outError:       errors.New("failed to add device: db connection failed"),
		},
	}

	for name, tc := range testCases {
		t.Logf("test case: %s", name)

		ctx := context.Background()

		db := &mstore.DataStore{}
		db.On("AddDevice",
			ctx,
			mock.MatchedBy(func(device *model.Device) bool {
				text := utils.GetTextField(device)
				assert.Equal(t, device.Text, text)

				return true
			})).
			Return(tc.datastoreError)
		i := invForTest(db)

		err := i.AddDevice(ctx, tc.inDevice)

		if tc.outError != nil {
			if assert.Error(t, err) {
				assert.EqualError(t, err, tc.outError.Error())
			}
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestInventoryUpsertAttributes(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		datastoreRes   *model.UpdateResult
		datastoreError error
		outError       error
	}{
		"datastore success": {
			datastoreRes: &model.UpdateResult{
				Devices: []*model.Device{
					{
						ID: "1",
					},
				},
			},
			datastoreError: nil,
			outError:       nil,
		},
		"datastore error": {
			datastoreError: errors.New("db connection failed"),
			outError:       errors.New("failed to upsert attributes in db: db connection failed"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Logf("test case: %s", name)

			ctx := context.Background()

			db := &mstore.DataStore{}
			db.On("UpsertDevicesAttributes",
				ctx,
				mock.AnythingOfType("[]model.DeviceID"),
				mock.AnythingOfType("model.DeviceAttributes"),
				(*time.Time)(nil)).
				Return(tc.datastoreRes, tc.datastoreError)
			if tc.datastoreError == nil {
				db.On("UpdateDeviceText",
					ctx,
					tc.datastoreRes.Devices[0].ID,
					utils.GetTextField(tc.datastoreRes.Devices[0]),
				).Return(nil)
			}

			i := invForTest(db)
			err := i.UpsertAttributes(ctx, "devid", model.DeviceAttributes{}, nil)

			if tc.outError != nil {
				if assert.Error(t, err) {
					assert.EqualError(t, err, tc.outError.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInventoryUpsertAttributesWithUpdated(t *testing.T) {
	t.Parallel()

	dsResultFailed := model.UpdateResult{MatchedCount: 0}
	dsResultsuccess := model.UpdateResult{
		Devices: []*model.Device{
			{
				ID: "1",
			},
		},
		MatchedCount: 1,
	}

	const devID = model.DeviceID("devid")

	testCases := map[string]struct {
		getDevice       *model.Device
		getDeviceErr    error
		limitAttributes int
		limitTags       int
		attributes      model.DeviceAttributes

		datastoreError error
		outError       error

		datastoreResult *model.UpdateResult

		scope string
		etag  string
	}{
		"datastore success": {
			getDevice: &model.Device{
				Attributes: model.DeviceAttributes{},
			},
			attributes: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "name",
					Value: "foo",
					Scope: model.AttrScopeInventory,
				},
			},
			datastoreResult: &dsResultsuccess,
			datastoreError:  nil,
			outError:        nil,
			scope:           model.AttrScopeInventory,
		},
		"no upsert needed": {
			getDevice: &model.Device{
				Attributes: model.DeviceAttributes{
					model.DeviceAttribute{
						Name:  "name",
						Value: "foo",
						Scope: model.AttrScopeInventory,
					},
				},
			},
			attributes: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "name",
					Value: "foo",
					Scope: model.AttrScopeInventory,
				},
			},
			datastoreResult: nil,
			datastoreError:  nil,
			outError:        nil,
			scope:           model.AttrScopeInventory,
		},
		"no upsert needed with slices": {
			getDevice: &model.Device{
				Attributes: model.DeviceAttributes{
					model.DeviceAttribute{
						Name:  "name",
						Value: bson.A{"foo", "bar"},
						Scope: model.AttrScopeInventory,
					},
				},
			},
			attributes: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "name",
					Value: []interface{}{"foo", "bar"},
					Scope: model.AttrScopeInventory,
				},
			},
			datastoreResult: nil,
			datastoreError:  nil,
			outError:        nil,
			scope:           model.AttrScopeInventory,
		},
		"upsert needed, at least once a day": {
			getDevice: &model.Device{
				Attributes: model.DeviceAttributes{
					model.DeviceAttribute{
						Name:  "name",
						Value: "foo",
						Scope: model.AttrScopeInventory,
					},
				},
				// yesterday
				UpdatedTs: timePtr(time.Now().Add(-25 * time.Hour)),
			},
			attributes: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "name",
					Value: "foo",
					Scope: model.AttrScopeInventory,
				},
			},
			datastoreResult: &dsResultsuccess,
			datastoreError:  nil,
			outError:        nil,
			scope:           model.AttrScopeInventory,
		},
		"datastore error": {
			getDevice: &model.Device{
				Attributes: model.DeviceAttributes{},
			},
			attributes: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "name",
					Value: "foo",
					Scope: model.AttrScopeInventory,
				},
			},
			datastoreError: errors.New("db connection failed"),
			outError:       errors.New("failed to upsert attributes in db: db connection failed"),
			scope:          model.AttrScopeInventory,
		},
		"incorrect etag": {
			getDevice: &model.Device{
				Attributes: model.DeviceAttributes{},
			},
			attributes: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "name",
					Value: "foo",
					Scope: model.AttrScopeInventory,
				},
			},
			datastoreResult: &dsResultFailed,
			datastoreError:  errors.New("failed to replace attributes in db"),
			scope:           model.AttrScopeTags,
			etag:            "f7238315-062d-4440-875a-676006f84c34",
		},
		"correct etag": {
			getDevice: &model.Device{
				Attributes: model.DeviceAttributes{},
			},
			attributes: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "name",
					Value: "foo",
					Scope: model.AttrScopeInventory,
				},
			},
			datastoreResult: &dsResultsuccess,
			datastoreError:  nil,
			scope:           model.AttrScopeTags,
			etag:            "f7238315-062d-4440-875a-676006f84c34",
		},
		"limits ko, getDevice error": {
			getDeviceErr:    errors.New("datastore error"),
			limitAttributes: 1,
			datastoreResult: nil,
			datastoreError:  nil,
			outError:        errors.New("failed to get the device: datastore error"),
			scope:           model.AttrScopeInventory,
		},
		"limits ok, attributes": {
			getDevice: &model.Device{
				Attributes: model.DeviceAttributes{
					model.DeviceAttribute{
						Name:  "name",
						Value: "foo",
						Scope: model.AttrScopeInventory,
					},
				},
			},
			attributes: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "name",
					Value: "bar",
					Scope: model.AttrScopeInventory,
				},
			},
			limitAttributes: 1,
			datastoreResult: &dsResultsuccess,
			datastoreError:  nil,
			outError:        nil,
			scope:           model.AttrScopeInventory,
		},
		"limits ok, tags": {
			getDevice: &model.Device{
				Attributes: model.DeviceAttributes{},
			},
			attributes: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "name",
					Value: "foo",
					Scope: model.AttrScopeTags,
				},
			},
			limitTags:       1,
			datastoreResult: &dsResultsuccess,
			datastoreError:  nil,
			outError:        nil,
			scope:           model.AttrScopeTags,
		},
		"limits ko, attributes": {
			getDevice: &model.Device{
				Attributes: model.DeviceAttributes{
					model.DeviceAttribute{
						Name:  "env",
						Value: "dev",
						Scope: model.AttrScopeInventory,
					},
				},
			},
			attributes: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "name",
					Value: "foo",
					Scope: model.AttrScopeInventory,
				},
			},
			limitAttributes: 1,
			datastoreResult: nil,
			datastoreError:  nil,
			outError:        ErrTooManyAttributes,
			scope:           model.AttrScopeInventory,
		},
		"limits ko, tags": {
			getDevice: &model.Device{
				Attributes: model.DeviceAttributes{
					model.DeviceAttribute{
						Name:  "env",
						Value: "dev",
						Scope: model.AttrScopeTags,
					},
				},
			},
			attributes: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "name",
					Value: "foo",
					Scope: model.AttrScopeTags,
				},
			},
			limitTags:       1,
			datastoreResult: nil,
			datastoreError:  nil,
			outError:        ErrTooManyAttributes,
			scope:           model.AttrScopeTags,
		},
		"limits ko, tags with multiple existing attributes": {
			getDevice: &model.Device{
				Attributes: model.DeviceAttributes{
					model.DeviceAttribute{
						Name:  "env",
						Value: "dev",
						Scope: model.AttrScopeTags,
					},
					model.DeviceAttribute{
						Name:  "region",
						Value: "eu",
						Scope: model.AttrScopeTags,
					},
				},
			},
			attributes: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "name",
					Value: "foo",
					Scope: model.AttrScopeTags,
				},
			},
			limitTags:       1,
			datastoreResult: nil,
			datastoreError:  nil,
			outError:        ErrTooManyAttributes,
			scope:           model.AttrScopeTags,
		},
		"limits ko, tags with multiple attributes": {
			getDevice: &model.Device{
				Attributes: model.DeviceAttributes{
					model.DeviceAttribute{
						Name:  "env",
						Value: "dev",
						Scope: model.AttrScopeTags,
					},
				},
			},
			attributes: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "name",
					Value: "foo",
					Scope: model.AttrScopeTags,
				},
				model.DeviceAttribute{
					Name:  "region",
					Value: "eu",
					Scope: model.AttrScopeTags,
				},
			},
			limitTags:       1,
			datastoreResult: nil,
			datastoreError:  nil,
			outError:        ErrTooManyAttributes,
			scope:           model.AttrScopeTags,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Logf("test case: %s", name)

			ctx := context.Background()

			db := &mstore.DataStore{}
			defer db.AssertExpectations(t)

			if tc.getDevice != nil || tc.getDeviceErr != nil {
				db.On("GetDevice",
					ctx,
					devID,
				).Return(tc.getDevice, tc.getDeviceErr)
			}

			if tc.datastoreResult != nil || tc.datastoreError != nil {
				db.On("UpsertDevicesAttributesWithUpdated",
					ctx,
					mock.AnythingOfType("[]model.DeviceID"),
					mock.AnythingOfType("model.DeviceAttributes"),
					tc.scope,
					tc.etag,
				).Return(tc.datastoreResult, tc.datastoreError)
			}

			if tc.datastoreError == nil && tc.datastoreResult != nil {
				db.On("UpdateDeviceText",
					ctx,
					tc.datastoreResult.Devices[0].ID,
					utils.GetTextField(tc.datastoreResult.Devices[0]),
				).Return(nil)
			}

			i := invForTest(db).WithLimits(tc.limitAttributes, tc.limitTags)

			err := i.UpsertAttributesWithUpdated(ctx, devID, tc.attributes, tc.scope, tc.etag)

			if tc.outError != nil {
				if assert.Error(t, err) {
					assert.EqualError(t, err, tc.outError.Error())
				}
			} else {
				if tc.etag != "" && tc.datastoreError != nil {
					assert.EqualError(t, err, "failed to upsert attributes in db: failed to replace attributes in db")
				} else {
					assert.NoError(t, err)
				}
			}

		})
	}
}

func TestReplaceAttributes(t *testing.T) {
	t.Parallel()

	updateResult := &model.UpdateResult{
		Devices: []*model.Device{
			{
				ID: "1",
			},
		},
		MatchedCount: 1,
	}

	testCases := map[string]struct {
		deviceID        model.DeviceID
		getDevice       *model.Device
		getDeviceErr    error
		dataStoreResult *model.UpdateResult
		datastoreError  error
		limitAttributes int
		limitTags       int

		upsertAttrs model.DeviceAttributes
		removeAttrs model.DeviceAttributes
		outError    error

		scope string
		etag  string
	}{
		"ok, device not found": {
			deviceID:     "1",
			getDevice:    nil,
			getDeviceErr: store.ErrDevNotFound,

			upsertAttrs: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "name",
					Value: "foo",
					Scope: model.AttrScopeInventory,
				},
				model.DeviceAttribute{
					Name:  "ip_address",
					Value: "127.0.0.1",
					Scope: model.AttrScopeInventory,
				},
			},
			removeAttrs: model.DeviceAttributes{},

			dataStoreResult: nil,
			datastoreError:  nil,
			outError:        nil,

			scope: model.AttrScopeInventory,
		},
		"ok, device found": {
			deviceID: "1",
			getDevice: &model.Device{
				Attributes: model.DeviceAttributes{},
			},
			getDeviceErr: nil,

			upsertAttrs: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "name",
					Value: "foo",
					Scope: model.AttrScopeInventory,
				},
				model.DeviceAttribute{
					Name:  "ip_address",
					Value: "127.0.0.1",
					Scope: model.AttrScopeInventory,
				},
			},
			removeAttrs: model.DeviceAttributes{},

			dataStoreResult: updateResult,
			datastoreError:  nil,
			outError:        nil,

			scope: model.AttrScopeInventory,
		},
		"ok, device found, replace attributes": {
			deviceID: "1",
			getDevice: &model.Device{
				Attributes: model.DeviceAttributes{
					model.DeviceAttribute{
						Name:  "name",
						Value: "foo",
						Scope: model.AttrScopeInventory,
					},
					model.DeviceAttribute{
						Name:  "custom",
						Value: "bar",
						Scope: model.AttrScopeInventory,
					},
				},
			},
			getDeviceErr: nil,

			upsertAttrs: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "name",
					Value: "foo",
					Scope: model.AttrScopeInventory,
				},
				model.DeviceAttribute{
					Name:  "ip_address",
					Value: "127.0.0.1",
					Scope: model.AttrScopeInventory,
				},
			},
			removeAttrs: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "custom",
					Value: "bar",
					Scope: model.AttrScopeInventory,
				},
			},

			dataStoreResult: updateResult,
			datastoreError:  nil,
			outError:        nil,

			scope: model.AttrScopeInventory,
		},
		"ko, get device error": {
			deviceID:     "1",
			getDevice:    nil,
			getDeviceErr: errors.New("get device error"),

			datastoreError: nil,
			outError:       errors.New("failed to get the device: get device error"),

			scope: model.AttrScopeInventory,
		},
		"ko, datastore error": {
			deviceID:     "1",
			getDevice:    nil,
			getDeviceErr: store.ErrDevNotFound,

			upsertAttrs: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "name",
					Value: "foo",
					Scope: model.AttrScopeInventory,
				},
				model.DeviceAttribute{
					Name:  "ip_address",
					Value: "127.0.0.1",
					Scope: model.AttrScopeInventory,
				},
			},
			removeAttrs: model.DeviceAttributes{},

			datastoreError: errors.New("get device error"),
			outError:       errors.New("failed to replace attributes in db: get device error"),

			scope: model.AttrScopeInventory,
		},
		"ok, add tags, no etag": {
			deviceID: "1",
			getDevice: &model.Device{
				Attributes: model.DeviceAttributes{},
			},
			upsertAttrs: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "name",
					Value: "foo",
					Scope: model.AttrScopeTags,
				},
				model.DeviceAttribute{
					Name:  "region",
					Value: "bar",
					Scope: model.AttrScopeTags,
				},
			},
			removeAttrs: model.DeviceAttributes{},
			scope:       model.AttrScopeTags,

			dataStoreResult: updateResult,
		},
		"ok, replace tags, with etag": {
			deviceID: "1",
			getDevice: &model.Device{
				Attributes: model.DeviceAttributes{
					model.DeviceAttribute{
						Name:  "name",
						Value: "foo",
						Scope: model.AttrScopeTags,
					},
				},
				TagsEtag: "f7238315-062d-4440-875a-676006f84c34",
			},
			upsertAttrs: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "region",
					Value: "bar",
					Scope: model.AttrScopeTags,
				},
			},
			removeAttrs: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "name",
					Value: "foo",
					Scope: model.AttrScopeTags,
				},
			},
			scope: model.AttrScopeTags,
			etag:  "f7238315-062d-4440-875a-676006f84c34",

			dataStoreResult: updateResult,
		},
		"ok, delete tags, no etag": {
			deviceID: "1",
			getDevice: &model.Device{
				Attributes: model.DeviceAttributes{
					model.DeviceAttribute{
						Name:  "name",
						Value: "foo",
						Scope: model.AttrScopeTags,
					},
					model.DeviceAttribute{
						Name:  "region",
						Value: "bar",
						Scope: model.AttrScopeTags,
					},
				},
			},
			upsertAttrs: model.DeviceAttributes{},
			removeAttrs: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "name",
					Value: "foo",
					Scope: model.AttrScopeTags,
				},
				model.DeviceAttribute{
					Name:  "region",
					Value: "bar",
					Scope: model.AttrScopeTags,
				},
			},
			scope: model.AttrScopeTags,

			dataStoreResult: updateResult,
		},
		"ok, delete tags, with etag": {
			deviceID: "1",
			getDevice: &model.Device{
				Attributes: model.DeviceAttributes{
					model.DeviceAttribute{
						Name:  "name",
						Value: "foo",
						Scope: model.AttrScopeTags,
					},
					model.DeviceAttribute{
						Name:  "region",
						Value: "bar",
						Scope: model.AttrScopeTags,
					},
				},
				TagsEtag: "f7238315-062d-4440-875a-676006f84c34",
			},
			upsertAttrs: model.DeviceAttributes{},
			removeAttrs: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "name",
					Value: "foo",
					Scope: model.AttrScopeTags,
				},
				model.DeviceAttribute{
					Name:  "region",
					Value: "bar",
					Scope: model.AttrScopeTags,
				},
			},
			scope: model.AttrScopeTags,
			etag:  "f7238315-062d-4440-875a-676006f84c34",

			dataStoreResult: updateResult,
		},
		"fail, modify tags, etag doesn't match": {
			deviceID: "1",
			getDevice: &model.Device{
				Attributes: model.DeviceAttributes{
					model.DeviceAttribute{
						Name:  "name",
						Value: "foo",
						Scope: model.AttrScopeTags,
					},
					model.DeviceAttribute{
						Name:  "region",
						Value: "bar",
						Scope: model.AttrScopeTags,
					},
				},
				TagsEtag: "f7238315-062d-4440-875a-676006f84c34",
			},
			upsertAttrs: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "name",
					Value: "new-foo",
					Scope: model.AttrScopeTags,
				},
			},
			removeAttrs: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "region",
					Value: "bar",
					Scope: model.AttrScopeTags,
				},
			},
			scope:          model.AttrScopeTags,
			etag:           "e5f05b31-398a-4df9-a0fd-52c38ef77123",
			datastoreError: errors.New("failed to replace attributes in db: failed to replace attributes in db: get device error"),
			outError:       errors.New("failed to replace attributes in db: failed to replace attributes in db: failed to replace attributes in db: get device error"),
		},
		"ok, inventory, limits ok": {
			deviceID:        "1",
			getDevice:       nil,
			getDeviceErr:    store.ErrDevNotFound,
			limitAttributes: 2,

			upsertAttrs: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "name",
					Value: "foo",
					Scope: model.AttrScopeInventory,
				},
				model.DeviceAttribute{
					Name:  "ip_address",
					Value: "127.0.0.1",
					Scope: model.AttrScopeInventory,
				},
			},
			removeAttrs: model.DeviceAttributes{},

			dataStoreResult: updateResult,
			datastoreError:  nil,
			outError:        nil,

			scope: model.AttrScopeInventory,
		},
		"ko, inventory, limits ko": {
			deviceID:        "1",
			getDevice:       nil,
			getDeviceErr:    store.ErrDevNotFound,
			limitAttributes: 1,

			upsertAttrs: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "name",
					Value: "foo",
					Scope: model.AttrScopeInventory,
				},
				model.DeviceAttribute{
					Name:  "ip_address",
					Value: "127.0.0.1",
					Scope: model.AttrScopeInventory,
				},
			},
			removeAttrs: model.DeviceAttributes{},

			datastoreError: nil,
			outError:       ErrTooManyAttributes,

			scope: model.AttrScopeInventory,
		},
		"ok, tags, limits ok": {
			deviceID:     "1",
			getDevice:    nil,
			getDeviceErr: store.ErrDevNotFound,
			limitTags:    2,

			upsertAttrs: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "name",
					Value: "foo",
					Scope: model.AttrScopeTags,
				},
				model.DeviceAttribute{
					Name:  "env",
					Value: "dev",
					Scope: model.AttrScopeTags,
				},
			},
			removeAttrs: model.DeviceAttributes{},

			dataStoreResult: updateResult,
			datastoreError:  nil,
			outError:        nil,

			scope: model.AttrScopeTags,
		},
		"ko, tags, limits ko": {
			deviceID:     "1",
			getDevice:    nil,
			getDeviceErr: store.ErrDevNotFound,
			limitTags:    1,

			upsertAttrs: model.DeviceAttributes{
				model.DeviceAttribute{
					Name:  "name",
					Value: "foo",
					Scope: model.AttrScopeTags,
				},
				model.DeviceAttribute{
					Name:  "env",
					Value: "dev",
					Scope: model.AttrScopeTags,
				},
			},
			removeAttrs: model.DeviceAttributes{},

			dataStoreResult: updateResult,
			datastoreError:  nil,
			outError:        ErrTooManyAttributes,

			scope: model.AttrScopeTags,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			db := &mstore.DataStore{}
			defer db.AssertExpectations(t)

			if tc.outError != ErrTooManyAttributes {
				db.On("GetDevice",
					ctx,
					tc.deviceID,
				).Return(tc.getDevice, tc.getDeviceErr)
			}

			if (tc.getDeviceErr == nil || tc.getDeviceErr == store.ErrDevNotFound) && tc.outError != ErrTooManyAttributes {
				db.On("UpsertRemoveDeviceAttributes",
					ctx,
					tc.deviceID,
					tc.upsertAttrs,
					tc.removeAttrs,
					tc.scope,
					tc.etag,
				).Return(tc.dataStoreResult, tc.datastoreError)

				if (tc.getDeviceErr == nil || tc.getDeviceErr == store.ErrDevNotFound) &&
					tc.datastoreError == nil && tc.dataStoreResult != nil {
					db.On("UpdateDeviceText",
						ctx,
						tc.dataStoreResult.Devices[0].ID,
						utils.GetTextField(tc.dataStoreResult.Devices[0]),
					).Return(nil)
				}
			}

			i := invForTest(db).WithLimits(tc.limitAttributes, tc.limitTags)
			err := i.ReplaceAttributes(ctx, tc.deviceID, tc.upsertAttrs, tc.scope, tc.etag)

			if tc.outError != nil {
				if assert.Error(t, err) {
					assert.EqualError(t, err, tc.outError.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetFiltersAttributes(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attributes []model.FilterAttribute
		err        error
		outErr     error
	}{
		"ok": {
			attributes: []model.FilterAttribute{
				{
					Name:  "name",
					Scope: "scope",
					Count: 100,
				},
				{
					Name:  "other_name",
					Scope: "scope",
					Count: 90,
				},
			},
		},
		"ko": {
			err:    errors.New("error"),
			outErr: errors.New("failed to get filter attributes from the db: error"),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			db := &mstore.DataStore{}
			db.On("GetFiltersAttributes",
				ctx,
			).Return(tc.attributes, tc.err)

			i := invForTest(db)
			attributes, err := i.GetFiltersAttributes(ctx)
			assert.Equal(t, tc.attributes, attributes)
			if tc.err != nil {
				assert.EqualError(t, tc.outErr, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteGroup(t *testing.T) {
	t.Parallel()

	groupName := model.GroupName("group")

	testCases := map[string]struct {
		count  int
		err    error
		outErr error
	}{
		"ok": {
			count: 2,
		},
		"ok with pagination": {
			count: 102,
		},
		"ko": {
			err:    errors.New("error"),
			outErr: errors.New("failed to delete group: error"),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			db := &mstore.DataStore{}

			if tc.err != nil {
				db.On("DeleteGroup",
					ctx,
					groupName,
				).Return(nil, tc.err)
			} else {
				devices := make(chan model.DeviceID)
				go func() {
					for i := 0; i < tc.count; i++ {
						devices <- model.DeviceID(strconv.Itoa(i + 1))
					}
					close(devices)
				}()

				db.On("DeleteGroup",
					ctx,
					groupName,
				).Return(devices, tc.err)
			}

			workflows := &mworkflows.Client{}
			defer workflows.AssertExpectations(t)
			if tc.err == nil {
				if tc.count > 100 {
					workflows.On("StartReindex",
						ctx,
						[]model.DeviceID{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10",
							"11", "12", "13", "14", "15", "16", "17", "18", "19", "20",
							"21", "22", "23", "24", "25", "26", "27", "28", "29", "30",
							"31", "32", "33", "34", "35", "36", "37", "38", "39", "40",
							"41", "42", "43", "44", "45", "46", "47", "48", "49", "50",
							"51", "52", "53", "54", "55", "56", "57", "58", "59", "60",
							"61", "62", "63", "64", "65", "66", "67", "68", "69", "70",
							"71", "72", "73", "74", "75", "76", "77", "78", "79", "80",
							"81", "82", "83", "84", "85", "86", "87", "88", "89", "90",
							"91", "92", "93", "94", "95", "96", "97", "98", "99", "100"},
					).Return(nil).Once()

					workflows.On("StartReindex",
						ctx,
						[]model.DeviceID{"101", "102"},
					).Return(nil).Once()
				} else {
					workflows.On("StartReindex",
						ctx,
						[]model.DeviceID{"1", "2"},
					).Return(nil).Once()
				}
			}

			i := invForTest(db).WithReporting(workflows)
			res, err := i.DeleteGroup(ctx, groupName)
			if tc.err != nil {
				assert.EqualError(t, tc.outErr, err.Error())
			} else {
				assert.Equal(t, int64(tc.count), res.UpdatedCount)
				assert.NoError(t, err)
			}
		})
	}
}

func TestInventoryUnsetDeviceGroup(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		inDeviceID      model.DeviceID
		inGroupName     model.GroupName
		datastoreError  error
		datastoreResult *model.UpdateResult
		outError        error
	}{
		"empty device ID, not found": {
			inDeviceID:  model.DeviceID(""),
			inGroupName: model.GroupName("gr1"),
			datastoreResult: &model.UpdateResult{
				MatchedCount: 0,
				UpdatedCount: 0,
			},
			outError: store.ErrDevNotFound,
		},
		"device group name not matching": {
			inDeviceID:  model.DeviceID("1"),
			inGroupName: model.GroupName("not-matching"),
			datastoreResult: &model.UpdateResult{
				MatchedCount: 0,
				UpdatedCount: 0,
			},
			outError: store.ErrDevNotFound,
		},
		"datastore success": {
			inDeviceID:  model.DeviceID("1"),
			inGroupName: model.GroupName("gr1"),
			datastoreResult: &model.UpdateResult{
				MatchedCount: 1,
				UpdatedCount: 1,
			},
			datastoreError: nil,
			outError:       nil,
		},
		"datastore internal error": {
			inDeviceID:      model.DeviceID("1"),
			inGroupName:     model.GroupName("gr1"),
			datastoreResult: nil,
			datastoreError:  errors.New("internal error"),
			outError:        errors.New("failed to unassign group from device: internal error"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			db := &mstore.DataStore{}
			db.On("UnsetDevicesGroup",
				ctx,
				mock.AnythingOfType("[]model.DeviceID"),
				mock.AnythingOfType("model.GroupName")).
				Return(tc.datastoreResult, tc.datastoreError)
			i := invForTest(db)

			err := i.UnsetDeviceGroup(ctx, tc.inDeviceID, tc.inGroupName)

			if tc.outError != nil {
				if assert.Error(t, err) {
					assert.EqualError(t, err, tc.outError.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInventoryUpdateDeviceGroup(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		inDeviceID      model.DeviceID
		inGroupName     model.GroupName
		datastoreResult *model.UpdateResult
		datastoreError  error
		outError        error
	}{
		"empty device ID, not found": {
			inDeviceID:      model.DeviceID(""),
			inGroupName:     model.GroupName("gr1"),
			datastoreResult: &model.UpdateResult{},
			outError:        errors.New("Device not found"),
		},
		"datastore success": {
			inDeviceID:  model.DeviceID("1"),
			inGroupName: model.GroupName("gr1"),
			datastoreResult: &model.UpdateResult{
				MatchedCount: 1,
				UpdatedCount: 1,
			},
			datastoreError: nil,
			outError:       nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Logf("test case: %s", name)

			ctx := context.Background()

			db := &mstore.DataStore{}
			db.On("UpdateDevicesGroup",
				ctx,
				mock.AnythingOfType("[]model.DeviceID"),
				mock.AnythingOfType("model.GroupName")).
				Return(tc.datastoreResult, tc.datastoreError)
			i := invForTest(db)

			err := i.UpdateDeviceGroup(ctx, tc.inDeviceID, tc.inGroupName)

			if tc.outError != nil {
				if assert.Error(t, err) {
					assert.EqualError(t, err, tc.outError.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInventoryListGroups(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		inputGroups    []model.GroupName
		outputGroups   []model.GroupName
		filters        []model.FilterPredicate
		datastoreError error
		outError       error
	}{
		"some groups": {
			inputGroups:  []model.GroupName{"foo", "bar"},
			outputGroups: []model.GroupName{"foo", "bar"},
		},
		"no groups - nil": {
			inputGroups:  nil,
			outputGroups: []model.GroupName{},
			filters: []model.FilterPredicate{{
				Attribute: "status",
				Scope:     model.AttrScopeIdentity,
				Type:      "$eq",
				Value:     "rejected",
			}},
		},
		"no groups - empty slice": {
			inputGroups:  []model.GroupName{},
			outputGroups: []model.GroupName{},
		},
		"error": {
			datastoreError: errors.New("random error"),
			outError:       errors.New("failed to list groups: random error"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			db := &mstore.DataStore{}

			db.On("ListGroups", ctx, tc.filters).
				Return(tc.inputGroups, tc.datastoreError)
			i := invForTest(db)

			groups, err := i.ListGroups(ctx, tc.filters)
			if tc.outError != nil {
				if assert.Error(t, err) {
					assert.EqualError(t, err, tc.outError.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, tc.outputGroups, groups)
			}
		})
	}
}

func TestInventoryListDevicesByGroup(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		DatastoreError error
		OutError       string
		OutDevices     []model.DeviceID
		OutDeviceCount int
	}{
		"success": {
			DatastoreError: nil,
			OutError:       "",
			OutDevices: []model.DeviceID{
				model.DeviceID("1"),
				model.DeviceID("2"),
				model.DeviceID("3"),
			},
			OutDeviceCount: 3,
		},
		"success - empty list": {
			DatastoreError: nil,
			OutError:       "",
			OutDevices:     []model.DeviceID{},
			OutDeviceCount: 0,
		},
		"datastore error - group not found": {
			DatastoreError: store.ErrGroupNotFound,
			OutError:       "group not found",
			OutDevices:     nil,
			OutDeviceCount: -1,
		},
		"datastore error - generic": {
			DatastoreError: errors.New("datastore error"),
			OutError:       "failed to list devices by group: datastore error",
			OutDevices:     nil,
			OutDeviceCount: -1,
		},
	}

	for name, tc := range testCases {
		t.Logf("test case: %s", name)

		ctx := context.Background()

		db := &mstore.DataStore{}

		db.On("GetDevicesByGroup",
			ctx,
			mock.AnythingOfType("model.GroupName"),
			mock.AnythingOfType("int"),
			mock.AnythingOfType("int"),
		).Return(tc.OutDevices, tc.OutDeviceCount, tc.DatastoreError)

		i := invForTest(db)

		devs, totalCount, err := i.ListDevicesByGroup(ctx, "foo", 1, 1)

		if tc.OutError != "" {
			if assert.Error(t, err) {
				assert.EqualError(t, err, tc.OutError)
			}
		} else {
			assert.NoError(t, err)
			if !reflect.DeepEqual(tc.OutDevices, devs) {
				assert.Fail(t, "expected outDevices to match", fmt.Sprintf("Expected: %v but\n have:%v", tc.OutDevices, devs))
			}
			if !reflect.DeepEqual(tc.OutDeviceCount, totalCount) {
				assert.Fail(t, "expected outDeviceCount to match", fmt.Sprintf("Expected: %v but\n have:%v", tc.OutDeviceCount, totalCount))
			}
		}
	}
}

func TestInventoryGetDeviceGroup(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		DatastoreError error
		DatastoreGroup model.GroupName
		OutError       error
		OutGroup       model.GroupName
	}{
		"success - device has group": {
			DatastoreError: nil,
			DatastoreGroup: model.GroupName("dev"),
			OutError:       nil,
			OutGroup:       model.GroupName("dev"),
		},
		"success - device has no group": {
			DatastoreError: nil,
			DatastoreGroup: model.GroupName(""),
			OutError:       nil,
			OutGroup:       model.GroupName(""),
		},
		"datastore error - device not found": {
			DatastoreError: store.ErrDevNotFound,
			DatastoreGroup: model.GroupName(""),
			OutError:       store.ErrDevNotFound,
			OutGroup:       model.GroupName(""),
		},
		"datastore error - generic": {
			DatastoreError: errors.New("datastore error"),
			DatastoreGroup: model.GroupName(""),
			OutError:       errors.New("failed to get device's group: datastore error"),
			OutGroup:       model.GroupName(""),
		},
	}

	for name, tc := range testCases {
		t.Logf("test case: %s", name)

		ctx := context.Background()

		db := &mstore.DataStore{}

		db.On("GetDeviceGroup",
			ctx,
			mock.AnythingOfType("model.DeviceID"),
		).Return(tc.OutGroup, tc.DatastoreError)

		i := invForTest(db)

		group, err := i.GetDeviceGroup(ctx, "foo")

		if tc.OutError != nil {
			if assert.Error(t, err) {
				assert.EqualError(t, err, tc.OutError.Error())
			}
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tc.OutGroup, group)
		}
	}
}

func TestInventoryDeleteDevice(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		datastoreResult *model.UpdateResult
		datastoreError  error
		outError        error
	}{
		"ok": {
			datastoreResult: &model.UpdateResult{
				DeletedCount: 1,
			},
			outError: nil,
		},
		"no device": {
			datastoreResult: &model.UpdateResult{
				DeletedCount: 0,
			},
			outError: store.ErrDevNotFound,
		},
		"datastore error": {
			datastoreError: errors.New("db connection failed"),
			outError:       errors.New("failed to delete device: db connection failed"),
		},
	}

	for name, tc := range testCases {
		t.Run(fmt.Sprintf("test case: %s", name), func(t *testing.T) {
			ctx := context.Background()

			db := &mstore.DataStore{}
			db.On("DeleteDevices",
				ctx,
				mock.AnythingOfType("[]model.DeviceID"),
			).Return(tc.datastoreResult, tc.datastoreError)
			i := invForTest(db)

			err := i.DeleteDevice(ctx, model.DeviceID("foo"))

			if tc.outError != nil {
				if assert.Error(t, err) {
					assert.EqualError(t, err, tc.outError.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInventoryDeleteDevices(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		datastoreResult *model.UpdateResult
		datastoreError  error
		outError        error
		workflowsError  error
	}{
		"ok": {
			datastoreResult: &model.UpdateResult{
				DeletedCount: 1,
			},
			outError: nil,
		},
		"ok, with workflows (swallowed) error": {
			datastoreResult: &model.UpdateResult{
				DeletedCount: 1,
			},
			workflowsError: errors.New("workflows error"),
			outError:       nil,
		},
		"datastore error": {
			datastoreError: errors.New("db connection failed"),
			outError:       errors.New("db connection failed"),
		},
	}

	for name, tc := range testCases {
		t.Run(fmt.Sprintf("test case: %s", name), func(t *testing.T) {
			ctx := context.Background()

			db := &mstore.DataStore{}
			db.On("DeleteDevices",
				ctx,
				mock.AnythingOfType("[]model.DeviceID"),
			).Return(tc.datastoreResult, tc.datastoreError)

			workflows := &mworkflows.Client{}
			defer workflows.AssertExpectations(t)
			if tc.outError == nil {
				workflows.On("StartReindex",
					ctx,
					mock.AnythingOfType("[]model.DeviceID"),
				).Return(tc.workflowsError)
			}

			i := invForTest(db).WithReporting(workflows)

			_, err := i.DeleteDevices(ctx, []model.DeviceID{"foo"})

			if tc.outError != nil {
				if assert.Error(t, err) {
					assert.EqualError(t, err, tc.outError.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInventoryUpsertDevicesStatuses(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		datastoreResult *model.UpdateResult
		datastoreError  error
		outError        error
		workflowsError  error
	}{
		"ok": {
			datastoreResult: &model.UpdateResult{
				DeletedCount: 1,
			},
			outError: nil,
		},
		"ok, with workflows (swallowed) error": {
			datastoreResult: &model.UpdateResult{
				DeletedCount: 1,
			},
			workflowsError: errors.New("workflows error"),
			outError:       nil,
		},
		"datastore error": {
			datastoreError: errors.New("db connection failed"),
			outError:       errors.New("db connection failed"),
		},
	}

	for name, tc := range testCases {
		t.Run(fmt.Sprintf("test case: %s", name), func(t *testing.T) {
			ctx := context.Background()

			db := &mstore.DataStore{}
			db.On("UpsertDevicesAttributesWithRevision",
				ctx,
				mock.AnythingOfType("[]model.DeviceUpdate"),
				mock.AnythingOfType("model.DeviceAttributes"),
			).Return(tc.datastoreResult, tc.datastoreError)

			workflows := &mworkflows.Client{}
			defer workflows.AssertExpectations(t)
			if tc.outError == nil {
				workflows.On("StartReindex",
					ctx,
					mock.AnythingOfType("[]model.DeviceID"),
				).Return(tc.workflowsError)
			}

			i := invForTest(db).WithReporting(workflows)

			_, err := i.UpsertDevicesStatuses(ctx, []model.DeviceUpdate{{Id: "foo"}}, model.DeviceAttributes{})

			if tc.outError != nil {
				if assert.Error(t, err) {
					assert.EqualError(t, err, tc.outError.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
func TestNewInventory(t *testing.T) {
	t.Parallel()

	i := NewInventory(&mstore.DataStore{})

	assert.NotNil(t, i)
}

func TestUserAdmCreateTenant(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		tenant    string
		tenantErr error
		err       error
	}{
		"ok": {
			tenant: "foobar",
		},
		"error": {
			tenant:    "1234",
			tenantErr: errors.New("migration failed"),
			err:       errors.New("failed to apply migrations for tenant 1234: migration failed"),
		},
	}

	for name := range testCases {
		tc := testCases[name]
		t.Run(fmt.Sprintf("tc %s", name), func(t *testing.T) {

			t.Logf("test case: %s", name)

			ctx := context.Background()

			tenantDb := &mstore.DataStore{}
			tenantDb.On("MigrateTenant",
				ctx, mongo.DbVersion, tc.tenant).
				Return(tc.tenantErr)
			tenantDb.On("WithAutomigrate").Return(tenantDb)

			useradm := NewInventory(tenantDb)

			err := useradm.CreateTenant(ctx, model.NewTenant{ID: tc.tenant})
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInventorySearchDevices(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		searchParams   model.SearchParams
		datastoreError error
		outError       error
		outDevices     []model.Device
		outDeviceCount int
	}{
		"ok": {
			searchParams:   model.SearchParams{},
			datastoreError: nil,
			outError:       nil,
			outDevices:     []model.Device{{ID: model.DeviceID("1")}},
			outDeviceCount: 1,
		},
		"datastore error": {
			searchParams:   model.SearchParams{},
			datastoreError: errors.New("db connection failed"),
			outError:       errors.New("failed to fetch devices: db connection failed"),
			outDevices:     nil,
			outDeviceCount: -1,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			db := &mstore.DataStore{}
			db.On("SearchDevices",
				ctx,
				mock.AnythingOfType("model.SearchParams"),
			).Return(tc.outDevices, tc.outDeviceCount, tc.datastoreError)
			i := invForTest(db)

			devs, totalCount, err := i.SearchDevices(ctx, tc.searchParams)

			if tc.outError != nil {
				if assert.Error(t, err) {
					assert.EqualError(t, err, tc.outError.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(devs), len(tc.outDevices))
				assert.Equal(t, totalCount, tc.outDeviceCount)
			}
		})
	}
}

func TestInventoryUpdateDevicesGroup(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		DeviceIDs []model.DeviceID
		model.GroupName
		*model.UpdateResult
		Err error
	}{{
		Name: "ok",
		DeviceIDs: []model.DeviceID{
			"1", "2", "3",
		},
		GroupName: "foo",
		UpdateResult: &model.UpdateResult{
			UpdatedCount: 2,
			MatchedCount: 2,
		},
		Err: nil,
	}, {
		Name: "datastore error",
		DeviceIDs: []model.DeviceID{
			"1", "2", "3",
		},
		GroupName: "bar",
		Err:       errors.New("doesn't matter"),
	}}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.Background()
			db := &mstore.DataStore{}
			db.On("UpdateDevicesGroup",
				ctx,
				testCase.DeviceIDs,
				testCase.GroupName,
			).Return(
				testCase.UpdateResult,
				testCase.Err,
			)
			i := invForTest(db)
			rsp, err := i.UpdateDevicesGroup(
				ctx, testCase.DeviceIDs, testCase.GroupName,
			)
			assert.Equal(t, testCase.UpdateResult, rsp)
			assert.Equal(t, testCase.Err, err)
		})
	}
}

func TestInventoryUnsetDevicesGroup(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		DeviceIDs []model.DeviceID
		model.GroupName
		*model.UpdateResult
		Err error
	}{{
		Name: "ok",
		DeviceIDs: []model.DeviceID{
			"1", "2", "3",
		},
		GroupName: "foo",
		UpdateResult: &model.UpdateResult{
			MatchedCount: 2,
			UpdatedCount: 2,
		},
		Err: nil,
	}, {
		Name: "datastore error",
		DeviceIDs: []model.DeviceID{
			"1", "2", "3",
		},
		GroupName: "bar",
		Err:       errors.New("doesn't matter"),
	}}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.Background()
			db := &mstore.DataStore{}
			db.On("UnsetDevicesGroup",
				ctx,
				testCase.DeviceIDs,
				testCase.GroupName,
			).Return(
				testCase.UpdateResult,
				testCase.Err,
			)
			i := invForTest(db)
			rsp, err := i.UnsetDevicesGroup(
				ctx, testCase.DeviceIDs, testCase.GroupName,
			)
			assert.Equal(t, testCase.UpdateResult, rsp)
			assert.Equal(t, testCase.Err, err)
		})
	}
}

func TestCheckAlerts(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		deviceId string
		count    int
		err      error
		outCount int
		outErr   error
	}{
		"ok": {
			count:    3,
			outCount: 3,
		},
		"ko": {
			err:    errors.New("error"),
			outErr: errors.New("error"),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			dm := &mdm.Client{}
			dm.On("CheckAlerts",
				ctx, tc.deviceId).Return(tc.count, tc.err)

			i := invForTest(nil)
			i = i.WithDevicemonitor(dm)
			count, err := i.CheckAlerts(ctx, tc.deviceId)
			if tc.err != nil {
				assert.EqualError(t, tc.outErr, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.outCount, count)
			}
		})
	}
}
func TestGetDeviceStatistics(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		status *model.DeviceStatisticsByStatus
		err    error
	}{
		"ok": {
			status: &model.DeviceStatisticsByStatus{
				Accepted: model.DeviceCountPerTier{
					Standard: 1,
					Micro:    2,
					System:   3,
				},
				Pending: model.DeviceCountPerTier{
					Standard: 4,
					Micro:    5,
					System:   6,
				},
			},
			err: nil,
		},
		"nil stats": {
			status: nil,
			err:    nil,
		},
		"db error": {
			status: nil,
			err:    nil,
		},
	}

	for desc, tc := range testCases {
		t.Run(desc, func(t *testing.T) {
			ctx := context.Background()
			db := &mstore.DataStore{}
			db.On("GetDeviceTierStatisticsByStatus",
				ctx,
			).Return(tc.status, tc.err)

			i := invForTest(db)
			res, err := i.GetDeviceStatistics(ctx)
			if tc.status != nil && res != nil {
				require.NoError(t, err, "expected no error")
				assert.Equal(t, *tc.status, res.ByStatus)
			} else if tc.status == nil && tc.err == nil {
				require.Error(t, err, "expected error")
			} else {
				require.Error(t, err, "expected error")
				assert.Equal(t, tc.err, err)
			}

		})
	}
}
