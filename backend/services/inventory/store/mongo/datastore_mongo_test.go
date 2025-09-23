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

package mongo

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mopts "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mendersoftware/mender-server/services/inventory/model"
	"github.com/mendersoftware/mender-server/services/inventory/store"
	"github.com/mendersoftware/mender-server/services/inventory/utils"

	"github.com/mendersoftware/mender-server/pkg/identity"

	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
	"github.com/mendersoftware/mender-server/pkg/mongo/oid"
	mstore "github.com/mendersoftware/mender-server/pkg/store"
)

func TestPing(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestPing in short mode")
	}
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()
	dataStore := NewDataStoreMongoWithSession(db.Client())
	err := dataStore.Ping(ctx)
	assert.NoError(t, err)
}

// test funcs
func TestMongoGetDevices(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMongoGetDevices in short mode.")
	}

	timeValue, _ := time.Parse("2006-01-02T15:04:05Z", "2014-11-12T11:45:26.371Z")

	inputDevs := []model.Device{
		{ID: model.DeviceID("0")},
		{ID: model.DeviceID("1"), Group: model.GroupName("1")},
		{ID: model.DeviceID("2"), Group: model.GroupName("2")},
		{
			ID: model.DeviceID("3"),
			Attributes: model.DeviceAttributes{
				{Name: "attrString", Value: "val3", Description: strPtr("desc1"), Scope: model.AttrScopeInventory},
				{Name: "attrFloat", Value: 3.0, Description: strPtr("desc2"), Scope: model.AttrScopeInventory},
			},
		},
		{
			ID: model.DeviceID("4"),
			Attributes: model.DeviceAttributes{
				{Name: "attrString", Value: "val4", Description: strPtr("desc1"), Scope: model.AttrScopeInventory},
				{Name: "attrFloat", Value: 4.0, Description: strPtr("desc2"), Scope: model.AttrScopeInventory},
			},
		},
		{
			ID: model.DeviceID("5"),
			Attributes: model.DeviceAttributes{
				{Name: "attrString", Value: "val5", Description: strPtr("desc1"), Scope: model.AttrScopeInventory},
				{Name: "attrFloat", Value: 5.0, Description: strPtr("desc2"), Scope: model.AttrScopeInventory},
				{Name: "attrTime", Value: timeValue, Description: strPtr("desc3"), Scope: model.AttrScopeInventory},
			},
			Group: model.GroupName("2"),
		},
		{
			ID: model.DeviceID("6"),
			Attributes: model.DeviceAttributes{
				{Name: "attrString", Value: "val6", Description: strPtr("desc1"), Scope: model.AttrScopeInventory},
				{Name: "attrFloat", Value: 4.0, Description: strPtr("desc2"), Scope: model.AttrScopeInventory},
			},
		},
		{
			ID: model.DeviceID("7"),
			Attributes: model.DeviceAttributes{
				{Name: "attrString", Value: "val4", Description: strPtr("desc1"), Scope: model.AttrScopeInventory},
				{Name: "attrFloat", Value: 6.0, Description: strPtr("desc2"), Scope: model.AttrScopeInventory},
			},
		},
	}
	floatVal4 := 4.0
	floatVal5 := 5.0

	testCases := map[string]struct {
		expected  []model.Device
		devTotal  int
		skip      int
		limit     int
		filters   []store.Filter
		sort      *store.Sort
		hasGroup  *bool
		groupName string
		tenant    string
	}{
		"get device from group 1": {
			expected:  []model.Device{inputDevs[1]},
			devTotal:  1,
			skip:      0,
			filters:   nil,
			sort:      nil,
			groupName: "1",
		},
		"all devs, no skip, no limit": {
			expected: inputDevs,
			devTotal: len(inputDevs),
			skip:     0,
			limit:    20,
			filters:  nil,
			sort:     nil,
		},
		"all devs, no skip, no limit; with tenant": {
			expected: inputDevs,
			devTotal: len(inputDevs),
			skip:     0,
			limit:    20,
			filters:  nil,
			sort:     nil,
			tenant:   "foo",
		},
		"all devs, with skip": {
			expected: []model.Device{inputDevs[4], inputDevs[5], inputDevs[6], inputDevs[7]},
			devTotal: len(inputDevs),
			skip:     4,
			limit:    20,
			filters:  nil,
			sort:     nil,
		},
		"all devs, no skip, with limit": {
			expected: []model.Device{inputDevs[0], inputDevs[1], inputDevs[2]},
			devTotal: len(inputDevs),
			skip:     0,
			limit:    3,
			filters:  nil,
			sort:     nil,
		},
		"skip + limit": {
			expected: []model.Device{inputDevs[3], inputDevs[4]},
			devTotal: len(inputDevs),
			skip:     3,
			limit:    2,
			filters:  nil,
			sort:     nil,
		},
		"filter on attribute (equal attribute)": {
			expected: []model.Device{inputDevs[3]},
			devTotal: 1,
			skip:     0,
			limit:    20,
			filters: []store.Filter{
				{
					AttrName:  "attrString",
					AttrScope: model.AttrScopeInventory,
					Value:     "val3", Operator: store.Eq,
				},
			},
			sort: nil,
		},
		"filter on attribute (equal attribute float)": {
			expected: []model.Device{inputDevs[5]},
			devTotal: 1,
			skip:     0,
			limit:    20,
			filters: []store.Filter{
				{
					AttrName:   "attrFloat",
					AttrScope:  model.AttrScopeInventory,
					Value:      "5.0",
					ValueFloat: &floatVal5,
					Operator:   store.Eq,
				},
			},
			sort: nil,
		},
		"filter on attribute (equal attribute time)": {
			expected: []model.Device{inputDevs[5]},
			devTotal: 1,
			skip:     0,
			limit:    20,
			filters: []store.Filter{
				{
					AttrName:  "attrTime",
					AttrScope: model.AttrScopeInventory,
					Value:     "2014-11-12T11:45:26.371Z",
					ValueTime: &timeValue,
					Operator:  store.Eq,
				},
			},
			sort: nil,
		},
		"filter on two attributes (equal)": {
			expected: []model.Device{inputDevs[4]},
			devTotal: 1,
			skip:     0,
			limit:    20,
			filters: []store.Filter{
				{
					AttrName:  "attrString",
					AttrScope: model.AttrScopeInventory,
					Value:     "val4",
					Operator:  store.Eq,
				},
				{
					AttrName:   "attrFloat",
					AttrScope:  model.AttrScopeInventory,
					Value:      "4.0",
					ValueFloat: &floatVal4,
					Operator:   store.Eq,
				},
			},
			sort: nil,
		},
		"sort, limit": {
			expected: []model.Device{inputDevs[5], inputDevs[4], inputDevs[3]},
			devTotal: len(inputDevs),
			skip:     0,
			limit:    3,
			filters:  nil,
			sort: &store.Sort{
				AttrName:  "attrFloat",
				AttrScope: model.AttrScopeInventory,
				Ascending: false,
			},
		},
		"hasGroup = true": {
			expected: []model.Device{inputDevs[1], inputDevs[2], inputDevs[5]},
			devTotal: 3,
			skip:     0,
			limit:    20,
			filters:  nil,
			sort:     nil,
			hasGroup: boolPtr(true),
		},
		"hasGroup = false": {
			expected: []model.Device{inputDevs[0], inputDevs[3], inputDevs[4], inputDevs[6], inputDevs[7]},
			devTotal: 5,
			skip:     0,
			limit:    20,
			filters:  nil,
			sort:     nil,
			hasGroup: boolPtr(false),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Make sure we start test with empty database
			db.Wipe()

			client := db.Client()

			var ctx context.Context
			if tc.tenant != "" {
				ctx = identity.WithContext(db.CTX(), &identity.Identity{
					Tenant: tc.tenant,
				})
			} else {
				ctx = identity.WithContext(db.CTX(), &identity.Identity{
					Tenant: "",
				})
			}

			mongoStore := NewDataStoreMongoWithSession(client)
			for _, d := range inputDevs {
				err := mongoStore.AddDevice(ctx, &d)
				assert.NoError(t, err, "failed to setup input data")
			}

			//test
			devs, totalCount, err := mongoStore.GetDevices(ctx,
				store.ListQuery{
					Skip:      tc.skip,
					Limit:     tc.limit,
					Filters:   tc.filters,
					Sort:      tc.sort,
					HasGroup:  tc.hasGroup,
					GroupName: tc.groupName})
			assert.NoError(t, err, "failed to get devices")

			assert.Equal(t, tc.devTotal, totalCount)
			assert.Equal(t, len(tc.expected), len(devs))
		})
	}
}

func TestMongoGetAllAttributeNames(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMongoGetAllAttributeNames in short mode.")
	}

	testCases := map[string]struct {
		inDevs []model.Device
		tenant string

		outAttrs []string
	}{
		"single dev": {
			inDevs: []model.Device{
				{
					ID: model.DeviceID("1"),
					Attributes: model.DeviceAttributes{
						{Name: "mac", Value: "foo", Description: strPtr("desc"), Scope: model.AttrScopeInventory},
						{Name: "sn", Value: "bar", Description: strPtr("desc"), Scope: model.AttrScopeInventory},
					},
				},
			},
			outAttrs: []string{"mac", "sn", "created_ts"},
		},
		"two devs, non-overlapping attrs": {
			inDevs: []model.Device{
				{
					ID: model.DeviceID("1"),
					Attributes: model.DeviceAttributes{
						{Name: "mac", Value: "foo", Description: strPtr("desc"), Scope: model.AttrScopeInventory},
						{Name: "sn", Value: "bar", Description: strPtr("desc"), Scope: model.AttrScopeInventory},
					},
				},
				{
					ID: model.DeviceID("2"),
					Attributes: model.DeviceAttributes{
						{Name: "foo", Value: "foo", Description: strPtr("desc"), Scope: model.AttrScopeInventory},
						{Name: "bar", Value: "bar", Description: strPtr("desc"), Scope: model.AttrScopeInventory},
					},
				},
			},
			outAttrs: []string{"mac", "sn", "foo", "bar", "created_ts"},
		},
		"two devs, overlapping attrs": {
			inDevs: []model.Device{
				{
					ID: model.DeviceID("1"),
					Attributes: model.DeviceAttributes{
						{Name: "mac", Value: "foo", Description: strPtr("desc"), Scope: model.AttrScopeInventory},
						{Name: "sn", Value: "bar", Description: strPtr("desc"), Scope: model.AttrScopeInventory},
					},
				},
				{
					ID: model.DeviceID("2"),
					Attributes: model.DeviceAttributes{
						{Name: "mac", Value: "foo", Description: strPtr("desc"), Scope: model.AttrScopeInventory},
						{Name: "foo", Value: "foo", Description: strPtr("desc"), Scope: model.AttrScopeInventory},
						{Name: "bar", Value: "bar", Description: strPtr("desc"), Scope: model.AttrScopeInventory},
					},
				},
			},
			outAttrs: []string{"mac", "sn", "foo", "bar", "created_ts"},
		},
		"single dev, tenant": {
			inDevs: []model.Device{
				{
					ID: model.DeviceID("1"),
					Attributes: model.DeviceAttributes{
						{Name: "mac", Value: "foo", Description: strPtr("desc"), Scope: model.AttrScopeInventory},
						{Name: "sn", Value: "bar", Description: strPtr("desc"), Scope: model.AttrScopeInventory},
					},
				},
			},
			outAttrs: []string{"mac", "sn", "created_ts"},
			tenant:   "tenant1",
		},
		"no devs": {
			outAttrs: []string{},
		},
	}

	for name, tc := range testCases {
		t.Logf("test case: %s", name)

		// Make sure we start test with empty database
		db.Wipe()

		client := db.Client()

		var ctx context.Context
		if tc.tenant != "" {
			ctx = identity.WithContext(db.CTX(), &identity.Identity{
				Tenant: tc.tenant,
			})
		} else {
			ctx = identity.WithContext(db.CTX(), &identity.Identity{
				Tenant: "",
			})
		}

		mongoStore := NewDataStoreMongoWithSession(client)
		for _, d := range tc.inDevs {
			err := mongoStore.AddDevice(ctx, &d)
			assert.NoError(t, err, "failed to setup input data")
		}

		//test
		names, err := mongoStore.GetAllAttributeNames(ctx)
		assert.NoError(t, err, "failed to get devices")

		assert.ElementsMatch(t, tc.outAttrs, names)
	}
}

func TestMongoGetDevice(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMongoGetDevice in short mode.")
	}

	testCases := map[string]struct {
		InputID     model.DeviceID
		InputDevice *model.Device
		tenant      string
		OutputError error
	}{
		"no device and no ID given": {
			InputID:     model.DeviceID(""),
			InputDevice: nil,
		},
		"no device and no ID given; with tenant": {
			InputID:     model.DeviceID(""),
			InputDevice: nil,
			tenant:      "foo",
		},
		"device with given ID not exists": {
			InputID:     model.DeviceID("123"),
			InputDevice: nil,
		},
		"device with given ID not exists; with tenant": {
			InputID:     model.DeviceID("123"),
			InputDevice: nil,
			tenant:      "foo",
		},
		"device with given ID exists, no error": {
			InputID: model.DeviceID("0002"),
			InputDevice: &model.Device{
				ID: model.DeviceID("0002"),
				Attributes: model.DeviceAttributes{
					{Name: "mac", Value: "0002-mac", Scope: model.AttrScopeInventory},
				},
			},
		},
		"device with given ID exists, no error; with tenant": {
			InputID: model.DeviceID("0002"),
			InputDevice: &model.Device{
				ID: model.DeviceID("0002"),
				Attributes: model.DeviceAttributes{
					{Name: "mac", Value: "0002-mac", Scope: model.AttrScopeInventory},
				},
			},
			tenant: "foo",
		},
	}

	for name, testCase := range testCases {
		t.Logf("test case: %s", name)

		// Make sure we start test with empty database
		db.Wipe()

		client := db.Client()
		store := NewDataStoreMongoWithSession(client)

		var ctx context.Context
		if testCase.tenant != "" {
			ctx = identity.WithContext(db.CTX(), &identity.Identity{
				Tenant: testCase.tenant,
			})
		} else {
			ctx = identity.WithContext(db.CTX(), &identity.Identity{
				Tenant: "",
			})
		}

		if testCase.InputDevice != nil {
			_, _ = client.Database(mstore.DbFromContext(ctx, DbName)).Collection(DbDevicesColl).InsertOne(ctx, testCase.InputDevice)
		}

		dbdev, err := store.GetDevice(ctx, testCase.InputID)

		if testCase.InputDevice != nil {
			assert.NotNil(t, dbdev, "expected to device of ID %s to be found", testCase.InputDevice.ID)
			assert.Equal(t, testCase.InputID, dbdev.ID)
		} else {
			assert.Nil(t, dbdev, "expected no device to be found")
		}

		assert.NoError(t, err, "expected no error")
	}
}

func TestMongoAddDevice(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMongoAddDevice in short mode.")
	}

	existingID := "0000"

	existing := model.Device{
		ID: model.DeviceID(existingID),
		Attributes: model.DeviceAttributes{
			{Name: "mac", Value: "0000-mac", Scope: model.AttrScopeInventory},
			{Name: "sn", Value: "0000-sn", Scope: model.AttrScopeInventory},
		},
	}

	testCases := map[string]struct {
		InputDevice  *model.Device
		OutputDevice *model.Device
		tenant       string
		OutputError  error
	}{
		"valid device with one attribute, no error": {
			InputDevice: &model.Device{
				ID: model.DeviceID("0002"),
				Attributes: model.DeviceAttributes{
					{Name: "mac", Value: "0002-mac", Scope: model.AttrScopeInventory},
				},
			},
			OutputDevice: &model.Device{
				ID: model.DeviceID("0002"),
				Attributes: model.DeviceAttributes{
					{Name: "mac", Value: "0002-mac", Scope: model.AttrScopeInventory},
				},
			},
			OutputError: nil,
		},
		"valid device with one attribute, no error; with tenant": {
			InputDevice: &model.Device{
				ID: model.DeviceID("0002"),
				Attributes: model.DeviceAttributes{
					{Name: "mac", Value: "0002-mac", Scope: model.AttrScopeInventory},
				},
			},
			OutputDevice: &model.Device{
				ID: model.DeviceID("0002"),
				Attributes: model.DeviceAttributes{
					{Name: "mac", Value: "0002-mac", Scope: model.AttrScopeInventory},
				},
			},
			tenant:      "foo",
			OutputError: nil,
		},
		"valid device with two attributes, no error": {
			InputDevice: &model.Device{
				ID: model.DeviceID("0003"),
				Attributes: model.DeviceAttributes{
					{Name: "mac", Value: "0002-mac", Scope: model.AttrScopeInventory},
					{Name: "sn", Value: "0002-sn", Scope: model.AttrScopeInventory},
				},
			},
			OutputDevice: &model.Device{
				ID: model.DeviceID("0003"),
				Attributes: model.DeviceAttributes{
					{Name: "mac", Value: "0002-mac", Scope: model.AttrScopeInventory},
					{Name: "sn", Value: "0002-sn", Scope: model.AttrScopeInventory},
				},
			},
			OutputError: nil,
		},
		"valid device with attribute without value, no error": {
			InputDevice: &model.Device{
				ID: model.DeviceID("0004"),
				Attributes: model.DeviceAttributes{
					{Name: "mac", Scope: model.AttrScopeInventory},
				},
			},
			OutputDevice: &model.Device{
				ID: model.DeviceID("0004"),
				Attributes: model.DeviceAttributes{
					{Name: "mac", Scope: model.AttrScopeInventory},
				},
			},
			OutputError: nil,
		},
		"valid device with array in attribute value, no error": {
			InputDevice: &model.Device{
				ID: model.DeviceID("0005"),
				Attributes: model.DeviceAttributes{
					{Name: "mac", Value: primitive.A{float64(123), float64(456)}, Scope: model.AttrScopeInventory},
				},
			},
			OutputDevice: &model.Device{
				ID: model.DeviceID("0005"),
				Attributes: model.DeviceAttributes{
					{Name: "mac", Value: primitive.A{float64(123), float64(456)}, Scope: model.AttrScopeInventory},
				},
			},
			OutputError: nil,
		},
		"valid device without attributes, no error": {
			InputDevice: &model.Device{
				ID: model.DeviceID("0007"),
				Attributes: model.DeviceAttributes{
					{Name: "mac", Scope: model.AttrScopeInventory},
				},
			},
			OutputDevice: &model.Device{
				ID: model.DeviceID("0007"),
				Attributes: model.DeviceAttributes{
					{Name: "mac", Scope: model.AttrScopeInventory},
				},
			},
			OutputError: nil,
		},
		"valid device with upsert, all attrs updated, no error": {
			InputDevice: &model.Device{
				ID: model.DeviceID(existingID),
				Attributes: model.DeviceAttributes{
					{Name: "mac", Value: "0000-mac-new", Scope: model.AttrScopeInventory},
					{Name: "sn", Value: "0000-sn-new", Scope: model.AttrScopeInventory},
				},
			},
			OutputDevice: &model.Device{
				ID: model.DeviceID(existingID),
				Attributes: model.DeviceAttributes{
					{Name: "mac", Value: "0000-mac-new", Scope: model.AttrScopeInventory},
					{Name: "sn", Value: "0000-sn-new", Scope: model.AttrScopeInventory},
				},
			},
			OutputError: nil,
		},
		"valid device with upsert, one attr updated, no error": {
			InputDevice: &model.Device{
				ID: model.DeviceID(existingID),
				Attributes: model.DeviceAttributes{
					{Name: "mac", Value: "0000-mac-new", Scope: model.AttrScopeInventory},
				},
			},
			OutputDevice: &model.Device{
				ID: model.DeviceID(existingID),
				Attributes: model.DeviceAttributes{
					{Name: "mac", Value: "0000-mac-new", Scope: model.AttrScopeInventory},
					{Name: "sn", Value: "0000-sn", Scope: model.AttrScopeInventory},
				},
			},
			OutputError: nil,
		},
		"valid device with upsert, no attrs updated, new upserted, no error": {
			InputDevice: &model.Device{
				ID: model.DeviceID(existingID),
				Attributes: model.DeviceAttributes{
					{Name: "other-param", Value: "other-param-value", Scope: model.AttrScopeInventory},
				},
			},
			OutputDevice: &model.Device{
				ID: model.DeviceID(existingID),
				Attributes: model.DeviceAttributes{
					{Name: "mac", Value: "0000-mac", Scope: model.AttrScopeInventory},
					{Name: "sn", Value: "0000-sn", Scope: model.AttrScopeInventory},
					{Name: "other-param", Value: "other-param-value", Scope: model.AttrScopeInventory},
				},
			},
			OutputError: nil,
		},
		"valid device with upsert, no attrs updated, many new upserted, no error": {
			InputDevice: &model.Device{
				ID: model.DeviceID(existingID),
				Attributes: model.DeviceAttributes{
					{Name: "other-param", Value: "other-param-value", Scope: model.AttrScopeInventory},
					{Name: "other-param-2", Value: "other-param-2-value", Scope: model.AttrScopeInventory},
				},
			},
			OutputDevice: &model.Device{
				ID: model.DeviceID(existingID),
				Attributes: model.DeviceAttributes{
					{Name: "mac", Value: "0000-mac", Scope: model.AttrScopeInventory},
					{Name: "sn", Value: "0000-sn", Scope: model.AttrScopeInventory},
					{Name: "other-param", Value: "other-param-value", Scope: model.AttrScopeInventory},
					{Name: "other-param-2", Value: "other-param-2-value", Scope: model.AttrScopeInventory},
				},
			},
			OutputError: nil,
		},
	}

	for name, testCase := range testCases {
		t.Logf("test case: %s", name)

		// Make sure we start test with empty database
		db.Wipe()

		client := db.Client()
		store := NewDataStoreMongoWithSession(client)

		var ctx context.Context
		if testCase.tenant != "" {
			ctx = identity.WithContext(db.CTX(), &identity.Identity{
				Tenant: testCase.tenant,
			})
		} else {
			ctx = identity.WithContext(db.CTX(), &identity.Identity{
				Tenant: "",
			})
		}

		c := client.Database(mstore.DbFromContext(ctx, DbName)).Collection(DbDevicesColl)
		_, err := c.InsertOne(ctx, existing)
		assert.NoError(t, err)

		err = store.AddDevice(ctx, testCase.InputDevice)

		if testCase.OutputError != nil {
			assert.EqualError(t, err, testCase.OutputError.Error())
		} else {
			assert.NoError(t, err, "expected no error inserting to data store")

			var dbdev model.Device
			devsColl := client.Database(mstore.DbFromContext(ctx, DbName)).Collection(DbDevicesColl)
			err := DeviceFindById(ctx, devsColl, testCase.InputDevice.ID, &dbdev)
			assert.NoError(t, err, "error getting device")

			compareDevsWithoutTimestamps(t, testCase.OutputDevice, &dbdev)
		}
	}
}

func compareDevsWithoutTimestamps(t *testing.T, expected, actual *model.Device) {
	assert.Equal(t, expected.ID, actual.ID)
	// Sort attribute slices (we don't care about ordering)
	filterAndSortAttrs := func(attrs model.DeviceAttributes) model.DeviceAttributes {
		for i := 0; i < len(attrs); i++ {
			switch attrs[i].Name {
			case "updated_ts", "created_ts":
				attrs = append(attrs[:i], attrs[i+1:]...)
				i--
			}
		}
		sort.Slice(attrs, func(i, j int) bool {
			if attrs[i].Scope == attrs[j].Scope {
				return attrs[i].Name < attrs[j].Name
			}
			return attrs[i].Scope < attrs[j].Scope
		})
		return attrs
	}
	expectedAttrs := make(model.DeviceAttributes, len(expected.Attributes))
	actualAttrs := make(model.DeviceAttributes, len(actual.Attributes))
	copy(expectedAttrs, expected.Attributes)
	copy(actualAttrs, actual.Attributes)
	expectedAttrs = filterAndSortAttrs(expectedAttrs)
	actualAttrs = filterAndSortAttrs(actualAttrs)
	if !reflect.DeepEqual(expectedAttrs, actualAttrs) {
		assert.FailNow(
			t, "",
			"attributes not equal; expected: %v \nactual: %v\n",
			expectedAttrs, actualAttrs,
		)
	}
}

func TestNewDataStoreMongo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestNewDataStoreMongo in short mode.")
	}

	ds, err := NewDataStoreMongo(DataStoreMongoConfig{ConnectionString: "illegal url"})

	assert.Nil(t, ds)
	assert.EqualError(t, err, "failed to open mongo-driver session")
}

func TestMongoUpsertDevicesAttributes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMongoUpsertDeviceAttributes in short mode.")
	}

	//single create timestamp for all inserted devs
	createdTs := time.Now()

	testCases := map[string]struct {
		devs []model.Device

		inDevIDs []model.DeviceID
		inAttrs  model.DeviceAttributes
		inScope  string

		tenant string

		outDevs []model.Device
		err     error
	}{
		"dev exists, attributes exist, update both attrs (descr + val)": {
			devs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:        "mac",
							Value:       "0003-mac",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
						{
							Name:        "sn",
							Value:       "0003-sn",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
					},
					CreatedTs: createdTs,
				},
			},
			inDevIDs: []model.DeviceID{"0003"},
			inAttrs: model.DeviceAttributes{
				{
					Description: strPtr("mac description"),
					Scope:       model.AttrScopeInventory,
					Name:        "mac",
					Value:       "0003-newmac",
				},
				{
					Description: strPtr("sn description"),
					Scope:       model.AttrScopeInventory,
					Name:        "sn",
					Value:       "0003-newsn",
				},
			},
			inScope: model.AttrScopeInventory,

			outDevs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Description: strPtr("mac description"),
							Scope:       model.AttrScopeInventory,
							Name:        "mac",
							Value:       "0003-newmac",
						},
						{
							Description: strPtr("sn description"),
							Scope:       model.AttrScopeInventory,
							Name:        "sn",
							Value:       "0003-newsn",
						},
					},
					CreatedTs: createdTs,
				},
			},
		},
		"dev exists, attributes exist, update both attrs (descr + val); with tenant": {
			devs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:        "mac",
							Value:       "0003-mac",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
						{
							Name:        "sn",
							Value:       "0003-sn",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
					},
					CreatedTs: createdTs,
				},
			},
			inDevIDs: []model.DeviceID{"0003"},
			inAttrs: model.DeviceAttributes{
				{
					Description: strPtr("mac description"),
					Scope:       model.AttrScopeInventory,
					Name:        "mac",
					Value:       "0003-newmac",
				},
				{
					Description: strPtr("sn description"),
					Scope:       model.AttrScopeInventory,
					Name:        "sn",
					Value:       "0003-newsn",
				},
			},
			inScope: model.AttrScopeInventory,

			outDevs: []model.Device{{
				ID: model.DeviceID("0003"),
				Attributes: model.DeviceAttributes{
					{
						Description: strPtr("mac description"),
						Scope:       model.AttrScopeInventory,
						Name:        "mac",
						Value:       "0003-newmac",
					},
					{
						Description: strPtr("sn description"),
						Scope:       model.AttrScopeInventory,
						Name:        "sn",
						Value:       "0003-newsn",
					},
				},
				CreatedTs: createdTs,
			}},
		},
		"dev exists, attributes exist, update one attr (descr + val), scope tags": {
			devs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:        "mac",
							Value:       "0003-mac",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeTags,
						},
						{
							Name:        "sn",
							Value:       "0003-sn",
							Description: strPtr("AttrScopeTags"),
							Scope:       model.AttrScopeTags,
						},
					},
					CreatedTs: createdTs,
				},
			},
			inDevIDs: []model.DeviceID{"0003"},
			inAttrs: model.DeviceAttributes{
				{
					Description: strPtr("sn description"),
					Scope:       model.AttrScopeTags,
					Name:        "sn",
					Value:       "0003-newsn",
				},
			},
			inScope: model.AttrScopeTags,

			outDevs: []model.Device{{
				ID: model.DeviceID("0003"),
				Attributes: model.DeviceAttributes{
					{
						Description: strPtr("descr"),
						Scope:       model.AttrScopeTags,
						Name:        "mac",
						Value:       "0003-mac",
					},
					{
						Description: strPtr("sn description"),
						Scope:       model.AttrScopeTags,
						Name:        "sn",
						Value:       "0003-newsn",
					},
				},
				CreatedTs: createdTs,
			}},
		},
		"dev exists, attributes exist, update one attr (descr + val)": {
			devs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:        "mac",
							Value:       "0003-mac",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
						{
							Name:        "sn",
							Value:       "0003-sn",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
					},
					CreatedTs: createdTs,
				},
			},
			inDevIDs: []model.DeviceID{"0003"},
			inAttrs: model.DeviceAttributes{
				{
					Description: strPtr("sn description"),
					Scope:       model.AttrScopeInventory,
					Name:        "sn",
					Value:       "0003-newsn",
				},
			},
			inScope: model.AttrScopeInventory,

			outDevs: []model.Device{{
				ID: model.DeviceID("0003"),
				Attributes: model.DeviceAttributes{
					{
						Description: strPtr("descr"),
						Scope:       model.AttrScopeInventory,
						Name:        "mac",
						Value:       "0003-mac",
					},
					{
						Description: strPtr("sn description"),
						Scope:       model.AttrScopeInventory,
						Name:        "sn",
						Value:       "0003-newsn",
					},
				},
				CreatedTs: createdTs,
			}},
		},
		"dev exists, attributes exist, update one attr (descr only)": {
			devs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:        "mac",
							Value:       "0003-mac",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
						{
							Name:        "sn",
							Value:       "0003-sn",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
					},
					CreatedTs: createdTs,
				},
			},
			inDevIDs: []model.DeviceID{"0003"},
			inAttrs: model.DeviceAttributes{
				{
					Description: strPtr("sn description"),
					Scope:       model.AttrScopeInventory,
					Name:        "sn",
				},
			},
			inScope: model.AttrScopeInventory,

			outDevs: []model.Device{{
				ID: model.DeviceID("0003"),
				Attributes: model.DeviceAttributes{
					{
						Description: strPtr("descr"),
						Scope:       model.AttrScopeInventory,
						Name:        "mac",
						Value:       "0003-mac",
					},
					{
						Description: strPtr("sn description"),
						Scope:       model.AttrScopeInventory,
						Name:        "sn",
						Value:       "0003-sn",
					},
				},
				CreatedTs: createdTs,
			}},
		},
		"dev exists, attributes exist, update one attr (value only)": {
			devs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:        "mac",
							Value:       "0003-mac",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
						{
							Name:        "sn",
							Value:       "0003-sn",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
					},
					CreatedTs: createdTs,
				},
			},
			inDevIDs: []model.DeviceID{"0003"},
			inAttrs: model.DeviceAttributes{
				{
					Scope: model.AttrScopeInventory,
					Name:  "sn",
					Value: "0003-newsn",
				},
			},
			inScope: model.AttrScopeInventory,

			outDevs: []model.Device{{
				ID: model.DeviceID("0003"),
				Attributes: model.DeviceAttributes{
					{
						Description: strPtr("descr"),
						Scope:       model.AttrScopeInventory,
						Name:        "mac",
						Value:       "0003-mac",
					},
					{
						Description: strPtr("descr"),
						Scope:       model.AttrScopeInventory,
						Name:        "sn",
						Value:       "0003-newsn",
					},
				},
				CreatedTs: createdTs,
			}},
		},
		"dev exists, attributes exist, update one attr (value only, change type)": {
			devs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:        "mac",
							Value:       "0003-mac",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
						{
							Name:        "sn",
							Value:       "0003-sn",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
					},
					CreatedTs: createdTs,
				},
			},
			inDevIDs: []model.DeviceID{"0003"},
			inAttrs: model.DeviceAttributes{
				{
					Value: primitive.A{"0003-sn-1", "0003-sn-2"},
					Scope: model.AttrScopeInventory,
					Name:  "sn",
				},
			},
			inScope: model.AttrScopeInventory,

			outDevs: []model.Device{{
				ID: model.DeviceID("0003"),
				Attributes: model.DeviceAttributes{
					{
						Description: strPtr("descr"),
						Scope:       model.AttrScopeInventory,
						Name:        "mac",
						Value:       "0003-mac",
					},
					{
						Description: strPtr("descr"),
						Scope:       model.AttrScopeInventory,
						Name:        "sn",
						Value:       primitive.A{"0003-sn-1", "0003-sn-2"},
					},
				},
				CreatedTs: createdTs,
			}},
		},
		"dev exists, attributes exist, add(merge) new attrs": {
			devs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:        "mac",
							Value:       "0003-mac",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
						{
							Name:        "sn",
							Value:       "0003-sn",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
					},
					CreatedTs: createdTs,
				},
			},
			inDevIDs: []model.DeviceID{"0003"},
			inAttrs: model.DeviceAttributes{
				{
					Scope: model.AttrScopeInventory,
					Name:  "new-1",
					Value: primitive.A{"new-1-0", "new-1-0"},
				},
				{
					Scope:       model.AttrScopeInventory,
					Name:        "new-2",
					Value:       "new-2-val",
					Description: strPtr("foo"),
				},
			},
			inScope: model.AttrScopeInventory,

			outDevs: []model.Device{{
				ID: model.DeviceID("0003"),
				Attributes: model.DeviceAttributes{
					{
						Scope:       model.AttrScopeInventory,
						Name:        "mac",
						Description: strPtr("descr"),
						Value:       "0003-mac",
					},
					{
						Scope:       model.AttrScopeInventory,
						Name:        "sn",
						Value:       "0003-sn",
						Description: strPtr("descr"),
					},
					{
						Scope: model.AttrScopeInventory,
						Name:  "new-1",
						Value: primitive.A{"new-1-0", "new-1-0"},
					},
					{
						Scope:       model.AttrScopeInventory,
						Name:        "new-2",
						Value:       "new-2-val",
						Description: strPtr("foo"),
					},
				},
				CreatedTs: createdTs,
			}},
		},
		"dev exists, attributes exist, add(merge) new attrs + modify existing": {
			devs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:        "mac",
							Value:       "0003-mac",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
						{
							Name:        "sn",
							Value:       "0003-sn",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
					},
					CreatedTs: createdTs,
				},
			},
			inDevIDs: []model.DeviceID{"0003"},
			inAttrs: model.DeviceAttributes{
				{
					Name:        "mac",
					Value:       "0003-mac-new",
					Description: strPtr("descr-new"),
				},
				{
					Name:  "new-1",
					Value: primitive.A{"new-1-0", "new-1-0"},
				},
				{
					Name:        "new-2",
					Value:       "new-2-val",
					Description: strPtr("foo"),
				},
			},
			inScope: model.AttrScopeInventory,

			outDevs: []model.Device{{
				ID: model.DeviceID("0003"),
				Attributes: model.DeviceAttributes{
					{
						Scope:       model.AttrScopeInventory,
						Name:        "mac",
						Value:       "0003-mac-new",
						Description: strPtr("descr-new"),
					},
					{
						Scope:       model.AttrScopeInventory,
						Name:        "sn",
						Value:       "0003-sn",
						Description: strPtr("descr"),
					},
					{
						Scope: model.AttrScopeInventory,
						Name:  "new-1",
						Value: primitive.A{"new-1-0", "new-1-0"},
					},
					{
						Scope:       model.AttrScopeInventory,
						Name:        "new-2",
						Value:       "new-2-val",
						Description: strPtr("foo"),
					},
				},
				CreatedTs: createdTs,
			}},
		},
		"dev exists, no attributes exist, upsert new attrs (val + descr)": {
			devs: []model.Device{
				{
					ID:        model.DeviceID("0003"),
					CreatedTs: createdTs,
				},
			},
			inDevIDs: []model.DeviceID{"0003"},
			inAttrs: model.DeviceAttributes{
				{
					Scope:       model.AttrScopeInventory,
					Name:        "ip",
					Value:       primitive.A{"1.2.3.4", "1.2.3.5"},
					Description: strPtr("ip addr array"),
				},
				{
					Scope:       model.AttrScopeInventory,
					Name:        "mac",
					Value:       primitive.A{"0006-mac"},
					Description: strPtr("mac addr"),
				},
			},
			inScope: model.AttrScopeInventory,

			outDevs: []model.Device{{
				ID: model.DeviceID("0003"),
				Attributes: model.DeviceAttributes{
					{
						Scope:       model.AttrScopeInventory,
						Name:        "ip",
						Value:       primitive.A{"1.2.3.4", "1.2.3.5"},
						Description: strPtr("ip addr array"),
					},
					{
						Scope:       model.AttrScopeInventory,
						Name:        "mac",
						Value:       primitive.A{"0006-mac"},
						Description: strPtr("mac addr"),
					},
				},
				CreatedTs: createdTs,
			}},
		},
		"dev exists, no attributes exist, upsert new attrs with dots and dolllars (val + descr)": {
			devs: []model.Device{
				{
					ID:        model.DeviceID("0003"),
					CreatedTs: createdTs,
				},
			},
			inDevIDs: []model.DeviceID{"0003"},
			inAttrs: model.DeviceAttributes{
				{
					Scope:       model.AttrScopeInventory,
					Name:        "ip.address",
					Value:       primitive.A{"1.2.3.4", "1.2.3.5"},
					Description: strPtr("ip addr array"),
				},
				{
					Scope:       model.AttrScopeInventory,
					Name:        "mac$addreses",
					Value:       primitive.A{"0006-mac"},
					Description: strPtr("mac addr"),
				},
			},
			inScope: model.AttrScopeInventory,

			outDevs: []model.Device{{
				ID: model.DeviceID("0003"),
				Attributes: model.DeviceAttributes{
					{
						Scope:       model.AttrScopeInventory,
						Name:        "ip.address",
						Value:       primitive.A{"1.2.3.4", "1.2.3.5"},
						Description: strPtr("ip addr array"),
					},
					{
						Scope:       model.AttrScopeInventory,
						Name:        "mac$address",
						Value:       primitive.A{"0006-mac"},
						Description: strPtr("mac addr"),
					},
				},
				CreatedTs: createdTs,
			}},
		},
		"dev doesn't exist, upsert new attr (descr + val)": {
			devs:     []model.Device{},
			inDevIDs: []model.DeviceID{"0099"},
			inAttrs: model.DeviceAttributes{
				{
					Scope:       model.AttrScopeInventory,
					Name:        "ip",
					Description: strPtr("ip addr array"),
					Value:       primitive.A{"1.2.3.4", "1.2.3.5"},
				},
			},
			inScope: model.AttrScopeInventory,

			outDevs: []model.Device{{
				ID: model.DeviceID("0099"),
				Attributes: model.DeviceAttributes{
					{
						Scope:       model.AttrScopeInventory,
						Name:        "ip",
						Description: strPtr("ip addr array"),
						Value:       primitive.A{"1.2.3.4", "1.2.3.5"},
					},
				},
				CreatedTs: createdTs,
			}},
		},
		"dev doesn't exist, upsert new attr (val only)": {
			devs:     []model.Device{},
			inDevIDs: []model.DeviceID{"0099"},
			inAttrs: model.DeviceAttributes{
				{
					Scope: model.AttrScopeInventory,
					Name:  "ip",
					Value: primitive.A{"1.2.3.4", "1.2.3.5"},
				},
			},
			inScope: model.AttrScopeInventory,

			outDevs: []model.Device{{
				ID: model.DeviceID("0099"),
				Attributes: model.DeviceAttributes{
					{
						Scope: model.AttrScopeInventory,
						Name:  "ip",
						Value: primitive.A{"1.2.3.4", "1.2.3.5"},
					},
				},
				CreatedTs: createdTs,
			}},
		},
		"dev doesn't exist, upsert with new attrs (val + descr)": {
			inDevIDs: []model.DeviceID{"0099"},
			inAttrs: model.DeviceAttributes{
				{
					Scope:       model.AttrScopeInventory,
					Name:        "ip",
					Value:       primitive.A{"1.2.3.4", "1.2.3.5"},
					Description: strPtr("ip addr array"),
				},
				{
					Scope:       model.AttrScopeInventory,
					Name:        "mac",
					Value:       primitive.A{"0099-mac"},
					Description: strPtr("mac addr"),
				},
			},
			inScope: model.AttrScopeInventory,

			outDevs: []model.Device{{
				ID: model.DeviceID("0099"),
				Attributes: model.DeviceAttributes{
					{
						Scope:       model.AttrScopeInventory,
						Name:        "ip",
						Value:       primitive.A{"1.2.3.4", "1.2.3.5"},
						Description: strPtr("ip addr array"),
					},
					{
						Scope:       model.AttrScopeInventory,
						Name:        "mac",
						Value:       primitive.A{"0099-mac"},
						Description: strPtr("mac addr"),
					},
				},
				CreatedTs: createdTs,
			}},
		},
		"Update multiple device attributes": {
			devs: []model.Device{{
				ID:        model.DeviceID("0003"),
				CreatedTs: createdTs,
			}, {
				ID: model.DeviceID("0004"),
				Attributes: model.DeviceAttributes{{
					Scope: model.AttrScopeInventory,
					Name:  "Artifact name",
					Value: "acmeware2.0",
				}},
				CreatedTs: createdTs,
			}},
			inDevIDs: []model.DeviceID{"0003", "0004", "0005"},
			inAttrs: model.DeviceAttributes{{
				Scope:       model.AttrScopeIdentity,
				Name:        "status",
				Value:       "accepted",
				Description: strPtr("deviceauth status"),
			}},
			inScope: model.AttrScopeIdentity,

			outDevs: []model.Device{{
				ID: model.DeviceID("0003"),
				Attributes: model.DeviceAttributes{{
					Scope:       model.AttrScopeIdentity,
					Name:        "status",
					Value:       "accepted",
					Description: strPtr("deviceauth status"),
				}},
				CreatedTs: createdTs,
			}, {
				ID: model.DeviceID("0004"),
				Attributes: model.DeviceAttributes{{
					Scope: model.AttrScopeInventory,
					Name:  "Artifact name",
					Value: "acmeware2.0",
				}, {
					Scope:       model.AttrScopeIdentity,
					Name:        "status",
					Value:       "accepted",
					Description: strPtr("deviceauth status"),
				}},
				CreatedTs: createdTs,
			}, {
				ID: model.DeviceID("0005"),
				Attributes: model.DeviceAttributes{{
					Scope:       model.AttrScopeIdentity,
					Name:        "status",
					Value:       "accepted",
					Description: strPtr("deviceauth status"),
				}},
			}},
		},
		"empty device list": {
			devs: []model.Device{{
				ID:        model.DeviceID("0003"),
				CreatedTs: createdTs,
			}},
			outDevs: []model.Device{{
				ID:        model.DeviceID("0003"),
				CreatedTs: createdTs,
			}},
		},
		"error, missing attribute name": {
			inAttrs: model.DeviceAttributes{{
				Scope: model.AttrScopeInventory,
				Value: "foo",
			}},
			err: errors.New("attribute name not present"),
		},
	}

	withOrWithoutUpdated := []bool{true, false}

	for name, tc := range testCases {
		for _, withUpdated := range withOrWithoutUpdated {
			if withUpdated {
				name += " withUpdated"
			}
			t.Run(name, func(t *testing.T) {
				//setup
				db.Wipe()

				s := db.Client()

				var ctx context.Context
				if tc.tenant != "" {
					ctx = identity.WithContext(db.CTX(), &identity.Identity{
						Tenant: tc.tenant,
					})
				} else {
					ctx = identity.WithContext(db.CTX(), &identity.Identity{
						Tenant: "",
					})
				}

				//test
				d := NewDataStoreMongoWithSession(s)
				for _, dev := range tc.devs {
					err := d.AddDevice(ctx, &dev)
					assert.NoError(t, err, "failed to setup input data")
				}

				var err error
				if withUpdated {
					_, err = d.UpsertDevicesAttributesWithUpdated(ctx, tc.inDevIDs, tc.inAttrs, tc.inScope, "")
				} else {
					_, err = d.UpsertDevicesAttributes(ctx, tc.inDevIDs, tc.inAttrs, nil)
				}
				if tc.err != nil {
					assert.EqualError(t, err, tc.err.Error())
				} else {
					assert.NoError(t, err, "UpsertDevicesAttributesWithUpdated failed")
				}

				//get the device back
				var devs []model.Device
				cur, err := s.Database(DbName).
					Collection(DbDevicesColl).
					Find(
						context.Background(),
						bson.M{},
						mopts.Find().SetSort(bson.M{"_id": 1}),
					)
				if err == nil {
					err = cur.All(nil, &devs)
				}
				if !assert.NoError(t, err) {
					t.FailNow()
				}

				if assert.Len(t, devs, len(tc.outDevs)) {
					for i, dev := range tc.outDevs {
						assert.Equal(t, dev.ID, devs[i].ID)
						if !compareAttrsWithoutTimestamp(
							dev.Attributes,
							devs[i].Attributes,
						) {
							t.Errorf("attributes mismatch, have: %v\nwant: %v",
								devs[i].Attributes,
								dev.Attributes,
							)
						}
						if withUpdated {
							// check timestamp validity; new devices (len(tc.devs) == 0) or different scope than inventory
							// won't update the update_ts attribute; updating attributes in the inventory scope will do.
							assert.Condition(t,
								func() bool {
									return (tc.inScope != model.AttrScopeInventory || len(tc.devs) == 0) ||
										(tc.inScope == model.AttrScopeInventory && len(tc.devs) > 0) && devs[i].UpdatedTs.After(devs[i].CreatedTs)

								}, devs[i])
						}
					}
				}
			})
		}
	}
}

func TestMongoUpsertRemoveDeviceAttributes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMongoUpsertRemoveDeviceAttributes in short mode.")
	}

	//single create timestamp for all inserted devs
	testStart := time.Now()

	testCases := map[string]struct {
		devs []model.Device

		inDevID       model.DeviceID
		inUpsertAttrs model.DeviceAttributes
		inRemoveAttrs model.DeviceAttributes

		tenant string

		outDevs []model.Device
		err     error

		scope string
		etag  string

		upsertMatchedCount  int64
		isCorrectEtagWanted bool
	}{
		"dev exists, attributes exist, update both attrs (descr + val)": {
			devs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:        "mac",
							Value:       "0003-mac",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
						{
							Name:        "sn",
							Value:       "0003-sn",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
					},
					CreatedTs: testStart,
				},
			},
			inDevID: model.DeviceID("0003"),
			inUpsertAttrs: model.DeviceAttributes{
				{
					Description: strPtr("mac description"),
					Scope:       model.AttrScopeInventory,
					Name:        "mac",
					Value:       "0003-newmac",
				},
				{
					Description: strPtr("sn description"),
					Scope:       model.AttrScopeInventory,
					Name:        "sn",
					Value:       "0003-newsn",
				},
			},

			outDevs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Description: strPtr("mac description"),
							Scope:       model.AttrScopeInventory,
							Name:        "mac",
							Value:       "0003-newmac",
						},
						{
							Description: strPtr("sn description"),
							Scope:       model.AttrScopeInventory,
							Name:        "sn",
							Value:       "0003-newsn",
						},
					},
					CreatedTs: testStart,
				},
			},

			scope: model.AttrScopeInventory,
		},
		"dev exists, attributes exist, update both attrs (descr + val); with tenant": {
			devs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:        "mac",
							Value:       "0003-mac",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
						{
							Name:        "sn",
							Value:       "0003-sn",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
					},
					CreatedTs: testStart,
				},
			},
			inDevID: model.DeviceID("0003"),
			inUpsertAttrs: model.DeviceAttributes{
				{
					Description: strPtr("mac description"),
					Scope:       model.AttrScopeInventory,
					Name:        "mac",
					Value:       "0003-newmac",
				},
				{
					Description: strPtr("sn description"),
					Scope:       model.AttrScopeInventory,
					Name:        "sn",
					Value:       "0003-newsn",
				},
			},

			outDevs: []model.Device{{
				ID: model.DeviceID("0003"),
				Attributes: model.DeviceAttributes{
					{
						Description: strPtr("mac description"),
						Scope:       model.AttrScopeInventory,
						Name:        "mac",
						Value:       "0003-newmac",
					},
					{
						Description: strPtr("sn description"),
						Scope:       model.AttrScopeInventory,
						Name:        "sn",
						Value:       "0003-newsn",
					},
				},
				CreatedTs: testStart,
			}},

			scope: model.AttrScopeInventory,
		},
		"dev exists, attributes exist, update one attr (descr + val)": {
			devs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:        "mac",
							Value:       "0003-mac",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
						{
							Name:        "sn",
							Value:       "0003-sn",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
					},
					CreatedTs: testStart,
				},
			},
			inDevID: model.DeviceID("0003"),
			inUpsertAttrs: model.DeviceAttributes{
				{
					Description: strPtr("sn description"),
					Scope:       model.AttrScopeInventory,
					Name:        "sn",
					Value:       "0003-newsn",
				},
			},

			outDevs: []model.Device{{
				ID: model.DeviceID("0003"),
				Attributes: model.DeviceAttributes{
					{
						Description: strPtr("descr"),
						Scope:       model.AttrScopeInventory,
						Name:        "mac",
						Value:       "0003-mac",
					},
					{
						Description: strPtr("sn description"),
						Scope:       model.AttrScopeInventory,
						Name:        "sn",
						Value:       "0003-newsn",
					},
				},
				CreatedTs: testStart,
			}},

			scope: model.AttrScopeInventory,
		},
		"dev exists, attributes exist, update one attr (descr + val), remove another": {
			devs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:        "mac",
							Value:       "0003-mac",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
						{
							Name:        "sn",
							Value:       "0003-sn",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
					},
					CreatedTs: testStart,
				},
			},
			inDevID: model.DeviceID("0003"),
			inUpsertAttrs: model.DeviceAttributes{
				{
					Description: strPtr("sn description"),
					Scope:       model.AttrScopeInventory,
					Name:        "sn",
					Value:       "0003-newsn",
				},
			},
			inRemoveAttrs: model.DeviceAttributes{
				{
					Name:        "mac",
					Value:       "0003-mac",
					Description: strPtr("descr"),
					Scope:       model.AttrScopeInventory,
				},
			},
			outDevs: []model.Device{{
				ID: model.DeviceID("0003"),
				Attributes: model.DeviceAttributes{
					{
						Description: strPtr("sn description"),
						Scope:       model.AttrScopeInventory,
						Name:        "sn",
						Value:       "0003-newsn",
					},
				},
				CreatedTs: testStart,
			}},
			scope: model.AttrScopeInventory,
		},
		"dev exists, attributes exist, update one attr (descr only)": {
			devs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:        "mac",
							Value:       "0003-mac",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
						{
							Name:        "sn",
							Value:       "0003-sn",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
					},
					CreatedTs: testStart,
				},
			},
			inDevID: model.DeviceID("0003"),
			inUpsertAttrs: model.DeviceAttributes{
				{
					Description: strPtr("sn description"),
					Scope:       model.AttrScopeInventory,
					Name:        "sn",
				},
			},

			outDevs: []model.Device{{
				ID: model.DeviceID("0003"),
				Attributes: model.DeviceAttributes{
					{
						Description: strPtr("descr"),
						Scope:       model.AttrScopeInventory,
						Name:        "mac",
						Value:       "0003-mac",
					},
					{
						Description: strPtr("sn description"),
						Scope:       model.AttrScopeInventory,
						Name:        "sn",
						Value:       "0003-sn",
					},
				},
				CreatedTs: testStart,
			}},
			scope: model.AttrScopeInventory,
		},
		"dev exists, attributes exist, update one attr (value only)": {
			devs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:        "mac",
							Value:       "0003-mac",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
						{
							Name:        "sn",
							Value:       "0003-sn",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
					},
					CreatedTs: testStart,
				},
			},
			inDevID: model.DeviceID("0003"),
			inUpsertAttrs: model.DeviceAttributes{
				{
					Scope: model.AttrScopeInventory,
					Name:  "sn",
					Value: "0003-newsn",
				},
			},

			outDevs: []model.Device{{
				ID: model.DeviceID("0003"),
				Attributes: model.DeviceAttributes{
					{
						Description: strPtr("descr"),
						Scope:       model.AttrScopeInventory,
						Name:        "mac",
						Value:       "0003-mac",
					},
					{
						Description: strPtr("descr"),
						Scope:       model.AttrScopeInventory,
						Name:        "sn",
						Value:       "0003-newsn",
					},
				},
				CreatedTs: testStart,
			}},
			scope: model.AttrScopeInventory,
		},
		"dev exists, attributes exist, update one attr (value only, change type)": {
			devs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:        "mac",
							Value:       "0003-mac",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
						{
							Name:        "sn",
							Value:       "0003-sn",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
					},
					CreatedTs: testStart,
				},
			},
			inDevID: model.DeviceID("0003"),
			inUpsertAttrs: model.DeviceAttributes{
				{
					Value: primitive.A{"0003-sn-1", "0003-sn-2"},
					Scope: model.AttrScopeInventory,
					Name:  "sn",
				},
			},

			outDevs: []model.Device{{
				ID: model.DeviceID("0003"),
				Attributes: model.DeviceAttributes{
					{
						Description: strPtr("descr"),
						Scope:       model.AttrScopeInventory,
						Name:        "mac",
						Value:       "0003-mac",
					},
					{
						Description: strPtr("descr"),
						Scope:       model.AttrScopeInventory,
						Name:        "sn",
						Value:       primitive.A{"0003-sn-1", "0003-sn-2"},
					},
				},
				CreatedTs: testStart,
			}},
			scope: model.AttrScopeInventory,
		},
		"dev exists, attributes exist, add(merge) new attrs": {
			devs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:        "mac",
							Value:       "0003-mac",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
						{
							Name:        "sn",
							Value:       "0003-sn",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
					},
					CreatedTs: testStart,
				},
			},
			inDevID: model.DeviceID("0003"),
			inUpsertAttrs: model.DeviceAttributes{
				{
					Scope: model.AttrScopeInventory,
					Name:  "new-1",
					Value: primitive.A{"new-1-0", "new-1-0"},
				},
				{
					Scope:       model.AttrScopeInventory,
					Name:        "new-2",
					Value:       "new-2-val",
					Description: strPtr("foo"),
				},
			},

			outDevs: []model.Device{{
				ID: model.DeviceID("0003"),
				Attributes: model.DeviceAttributes{
					{
						Scope:       model.AttrScopeInventory,
						Name:        "mac",
						Description: strPtr("descr"),
						Value:       "0003-mac",
					},
					{
						Scope:       model.AttrScopeInventory,
						Name:        "sn",
						Value:       "0003-sn",
						Description: strPtr("descr"),
					},
					{
						Scope: model.AttrScopeInventory,
						Name:  "new-1",
						Value: primitive.A{"new-1-0", "new-1-0"},
					},
					{
						Scope:       model.AttrScopeInventory,
						Name:        "new-2",
						Value:       "new-2-val",
						Description: strPtr("foo"),
					},
				},
				CreatedTs: testStart,
			}},
			scope: model.AttrScopeInventory,
		},
		"dev exists, attributes exist, add(merge) new attrs + modify existing": {
			devs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:        "mac",
							Value:       "0003-mac",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
						{
							Name:        "sn",
							Value:       "0003-sn",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
					},
					CreatedTs: testStart,
				},
			},
			inDevID: model.DeviceID("0003"),
			inUpsertAttrs: model.DeviceAttributes{
				{
					Name:        "mac",
					Value:       "0003-mac-new",
					Description: strPtr("descr-new"),
				},
				{
					Name:  "new-1",
					Value: primitive.A{"new-1-0", "new-1-0"},
				},
				{
					Name:        "new-2",
					Value:       "new-2-val",
					Description: strPtr("foo"),
				},
			},

			outDevs: []model.Device{{
				ID: model.DeviceID("0003"),
				Attributes: model.DeviceAttributes{
					{
						Scope:       model.AttrScopeInventory,
						Name:        "mac",
						Value:       "0003-mac-new",
						Description: strPtr("descr-new"),
					},
					{
						Scope:       model.AttrScopeInventory,
						Name:        "sn",
						Value:       "0003-sn",
						Description: strPtr("descr"),
					},
					{
						Scope: model.AttrScopeInventory,
						Name:  "new-1",
						Value: primitive.A{"new-1-0", "new-1-0"},
					},
					{
						Scope:       model.AttrScopeInventory,
						Name:        "new-2",
						Value:       "new-2-val",
						Description: strPtr("foo"),
					},
				},
				CreatedTs: testStart,
			}},
			scope: model.AttrScopeInventory,
		},
		"dev exists, no attributes exist, upsert new attrs (val + descr)": {
			devs: []model.Device{
				{
					ID:        model.DeviceID("0003"),
					CreatedTs: testStart,
				},
			},
			inDevID: model.DeviceID("0003"),
			inUpsertAttrs: model.DeviceAttributes{
				{
					Scope:       model.AttrScopeInventory,
					Name:        "ip",
					Value:       primitive.A{"1.2.3.4", "1.2.3.5"},
					Description: strPtr("ip addr array"),
				},
				{
					Scope:       model.AttrScopeInventory,
					Name:        "mac",
					Value:       primitive.A{"0006-mac"},
					Description: strPtr("mac addr"),
				},
			},

			outDevs: []model.Device{{
				ID: model.DeviceID("0003"),
				Attributes: model.DeviceAttributes{
					{
						Scope:       model.AttrScopeInventory,
						Name:        "ip",
						Value:       primitive.A{"1.2.3.4", "1.2.3.5"},
						Description: strPtr("ip addr array"),
					},
					{
						Scope:       model.AttrScopeInventory,
						Name:        "mac",
						Value:       primitive.A{"0006-mac"},
						Description: strPtr("mac addr"),
					},
				},
				CreatedTs: testStart,
			}},
			scope: model.AttrScopeInventory,
		},
		"dev exists, no attributes exist, upsert new attrs with dots and dolllars (val + descr)": {
			devs: []model.Device{
				{
					ID:        model.DeviceID("0003"),
					CreatedTs: testStart,
				},
			},
			inDevID: model.DeviceID("0003"),
			inUpsertAttrs: model.DeviceAttributes{
				{
					Scope:       model.AttrScopeInventory,
					Name:        "ip.address",
					Value:       primitive.A{"1.2.3.4", "1.2.3.5"},
					Description: strPtr("ip addr array"),
				},
				{
					Scope:       model.AttrScopeInventory,
					Name:        "mac$addreses",
					Value:       primitive.A{"0006-mac"},
					Description: strPtr("mac addr"),
				},
			},

			outDevs: []model.Device{{
				ID: model.DeviceID("0003"),
				Attributes: model.DeviceAttributes{
					{
						Scope:       model.AttrScopeInventory,
						Name:        "ip.address",
						Value:       primitive.A{"1.2.3.4", "1.2.3.5"},
						Description: strPtr("ip addr array"),
					},
					{
						Scope:       model.AttrScopeInventory,
						Name:        "mac$address",
						Value:       primitive.A{"0006-mac"},
						Description: strPtr("mac addr"),
					},
				},
				CreatedTs: testStart,
			}},
			scope: model.AttrScopeInventory,
		},
		"dev doesn't exist, upsert new attr (descr + val)": {
			devs:    []model.Device{},
			inDevID: model.DeviceID("0099"),
			inUpsertAttrs: model.DeviceAttributes{
				{
					Scope:       model.AttrScopeInventory,
					Name:        "ip",
					Description: strPtr("ip addr array"),
					Value:       primitive.A{"1.2.3.4", "1.2.3.5"},
				},
			},

			outDevs: []model.Device{{
				ID: model.DeviceID("0099"),
				Attributes: model.DeviceAttributes{
					{
						Scope:       model.AttrScopeInventory,
						Name:        "ip",
						Description: strPtr("ip addr array"),
						Value:       primitive.A{"1.2.3.4", "1.2.3.5"},
					},
				},
				CreatedTs: testStart,
			}},
			scope: model.AttrScopeInventory,
		},
		"dev doesn't exist, upsert new attr (val only)": {
			devs:    []model.Device{},
			inDevID: model.DeviceID("0099"),
			inUpsertAttrs: model.DeviceAttributes{
				{
					Scope: model.AttrScopeInventory,
					Name:  "ip",
					Value: primitive.A{"1.2.3.4", "1.2.3.5"},
				},
			},

			outDevs: []model.Device{{
				ID: model.DeviceID("0099"),
				Attributes: model.DeviceAttributes{
					{
						Scope: model.AttrScopeInventory,
						Name:  "ip",
						Value: primitive.A{"1.2.3.4", "1.2.3.5"},
					},
				},
				CreatedTs: testStart,
			}},
			scope: model.AttrScopeInventory,
		},
		"dev doesn't exist, upsert with new attrs (val + descr)": {
			inDevID: model.DeviceID("0099"),
			inUpsertAttrs: model.DeviceAttributes{
				{
					Scope:       model.AttrScopeInventory,
					Name:        "ip",
					Value:       primitive.A{"1.2.3.4", "1.2.3.5"},
					Description: strPtr("ip addr array"),
				},
				{
					Scope:       model.AttrScopeInventory,
					Name:        "mac",
					Value:       primitive.A{"0099-mac"},
					Description: strPtr("mac addr"),
				},
			},

			outDevs: []model.Device{{
				ID: model.DeviceID("0099"),
				Attributes: model.DeviceAttributes{
					{
						Scope:       model.AttrScopeInventory,
						Name:        "ip",
						Value:       primitive.A{"1.2.3.4", "1.2.3.5"},
						Description: strPtr("ip addr array"),
					},
					{
						Scope:       model.AttrScopeInventory,
						Name:        "mac",
						Value:       primitive.A{"0099-mac"},
						Description: strPtr("mac addr"),
					},
				},
				CreatedTs: testStart,
			}},
			scope: model.AttrScopeInventory,
		},
		"error, missing attribute name": {
			inUpsertAttrs: model.DeviceAttributes{{
				Scope: model.AttrScopeInventory,
				Value: "foo",
			}},
			err: errors.New("attribute name not present"),
		},
		"dev exists, attributes exist, update tags, no etag": {
			devs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:        "environment",
							Value:       "test",
							Description: strPtr("test_env"),
							Scope:       model.AttrScopeTags,
						},
						{
							Name:        "region",
							Value:       "EU",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeTags,
						},
					},
					CreatedTs: testStart,
				},
			},
			inDevID: model.DeviceID("0003"),
			inUpsertAttrs: model.DeviceAttributes{
				{
					Name:        "environment",
					Value:       "prod",
					Description: strPtr("prod_env"),
					Scope:       model.AttrScopeTags,
				},
			},
			outDevs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:        "environment",
							Value:       "prod",
							Description: strPtr("prod_env"),
							Scope:       model.AttrScopeTags,
						},
						{
							Name:        "region",
							Value:       "EU",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeTags,
						},
					},
					CreatedTs: testStart,
				},
			},
			scope:              model.AttrScopeTags,
			upsertMatchedCount: 1,
			etag:               "",
		},
		"dev exists, attributes exist, update tags, wrong etag": {
			devs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:        "environment",
							Value:       "test",
							Description: strPtr("test_env"),
							Scope:       model.AttrScopeTags,
						},
						{
							Name:        "region",
							Value:       "EU",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeTags,
						},
					},
					CreatedTs: testStart,
				},
			},
			inDevID: model.DeviceID("0003"),
			inUpsertAttrs: model.DeviceAttributes{
				{
					Name:        "environment",
					Value:       "prod",
					Description: strPtr("prod_env"),
					Scope:       model.AttrScopeTags,
				},
			},
			outDevs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:        "environment",
							Value:       "prod",
							Description: strPtr("prod_env"),
							Scope:       model.AttrScopeTags,
						},
						{
							Name:        "region",
							Value:       "EU",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeTags,
						},
					},
					CreatedTs: testStart,
				},
			},

			scope:              model.AttrScopeTags,
			upsertMatchedCount: 0,
			etag:               "f7238315-062d-4440-875a-676006f84c34",
		},
		"dev exists, attributes exist, update tags, correct etag": {
			devs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:        "environment",
							Value:       "test",
							Description: strPtr("test_env"),
							Scope:       model.AttrScopeTags,
						},
						{
							Name:        "region",
							Value:       "EU",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeTags,
						},
					},
					CreatedTs: testStart,
				},
			},
			inDevID: model.DeviceID("0003"),
			inUpsertAttrs: model.DeviceAttributes{
				{
					Name:        "environment",
					Value:       "prod",
					Description: strPtr("prod_env"),
					Scope:       model.AttrScopeTags,
				},
			},
			outDevs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:        "environment",
							Value:       "prod",
							Description: strPtr("prod_env"),
							Scope:       model.AttrScopeTags,
						},
						{
							Name:        "region",
							Value:       "EU",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeTags,
						},
					},
					CreatedTs: testStart,
				},
			},

			scope:               model.AttrScopeTags,
			upsertMatchedCount:  1,
			isCorrectEtagWanted: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			//setup
			db.Wipe()

			s := db.Client()

			var ctx context.Context
			if tc.tenant != "" {
				ctx = identity.WithContext(db.CTX(), &identity.Identity{
					Tenant: tc.tenant,
				})
			} else {
				ctx = identity.WithContext(db.CTX(), &identity.Identity{
					Tenant: "",
				})
			}

			//test
			d := NewDataStoreMongoWithSession(s)
			for _, dev := range tc.devs {
				err := d.AddDevice(ctx, &dev)
				assert.NoError(t, err, "failed to setup input data")
			}

			_, err := d.UpsertRemoveDeviceAttributes(ctx, tc.inDevID, tc.inUpsertAttrs, tc.inRemoveAttrs, tc.scope, "")
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err, "UpsertRemoveAttributes failed")
			}

			if tc.scope == model.AttrScopeTags {
				tags_etag := tc.etag
				if tc.isCorrectEtagWanted {
					var d []model.Device
					cur, err := s.Database(DbName).
						Collection(DbDevicesColl).
						Find(
							context.Background(),
							bson.M{},
							mopts.Find().SetSort(bson.M{"_id": 1}),
						)
					if err == nil {
						err = cur.All(nil, &d)
					}
					if !assert.NoError(t, err) {
						t.FailNow()
					}
					assert.Equal(t, len(tc.devs), len(d))
					tags_etag = d[len(d)-1].TagsEtag
				}
				_, err := d.UpsertRemoveDeviceAttributes(ctx, tc.inDevID, tc.inUpsertAttrs, tc.inRemoveAttrs, tc.scope, tags_etag)
				if tc.err != nil {
					assert.EqualError(t, err, tc.err.Error())
				}
			}

			//get the device back
			var devs []model.Device
			cur, err := s.Database(DbName).
				Collection(DbDevicesColl).
				Find(
					context.Background(),
					bson.M{},
					mopts.Find().SetSort(bson.M{"_id": 1}),
				)
			if err == nil {
				err = cur.All(nil, &devs)
			}
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			if assert.Len(t, devs, len(tc.outDevs)) {
				for i, dev := range tc.outDevs {
					assert.Equal(t, dev.ID, devs[i].ID)
					if !compareAttrsWithoutTimestamp(
						dev.Attributes,
						devs[i].Attributes,
					) {
						t.Errorf("attributes mismatch, have: %v\nwant: %v",
							devs[i].Attributes,
							dev.Attributes,
						)
					}
					// Check if updated_ts is later (or equal) than created_ts
					assert.Conditionf(t,
						func() bool {
							if devs[i].UpdatedTs == nil || devs[i].UpdatedTs.IsZero() {
								return true
							} else {
								return devs[i].UpdatedTs.Equal(devs[i].CreatedTs) || devs[i].UpdatedTs.After(devs[i].CreatedTs)
							}
						}, "updated timestamp (%s) is before created timestamp (%s)", devs[i].UpdatedTs, devs[i].CreatedTs)
					// Check if created_ts is sane
					assert.Truef(t,
						devs[i].CreatedTs.After(testStart) ||
							testStart.Sub(devs[i].CreatedTs).Abs() < time.Second,
						"The created timestamp %s is before the start of this test %s",
						devs[i].CreatedTs, testStart)
				}
			}
		})
	}
}

func TestGetFiltersAttributes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestGetFiltersAttributes in short mode.")
	}

	// single create timestamp for all inserted devs
	createdTs := time.Now()

	testCases := map[string]struct {
		tenant              string
		devs                []model.Device
		outFilterAttributes []model.FilterAttribute
	}{
		"single tenant": {
			devs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:  "mac",
							Value: "0003-mac",
							Scope: model.AttrScopeInventory,
						},
						{
							Name:  "sn",
							Value: "0003-sn",
							Scope: model.AttrScopeInventory,
						},
					},
					CreatedTs: createdTs,
				},
				{
					ID: model.DeviceID("0004"),
					Attributes: model.DeviceAttributes{
						{
							Name:  "mac",
							Value: "0004-mac",
							Scope: model.AttrScopeInventory,
						},
					},
					CreatedTs: createdTs,
				},
			},
			outFilterAttributes: []model.FilterAttribute{
				{
					Name:  "mac",
					Scope: model.AttrScopeInventory,
					Count: 2,
				},
				{
					Name:  "created_ts",
					Scope: model.AttrScopeSystem,
					Count: 2,
				},
				{
					Name:  "sn",
					Scope: model.AttrScopeInventory,
					Count: 1,
				},
			},
		},
		"multitenant": {
			tenant: "foo",
			devs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:  "mac",
							Value: "0003-mac",
							Scope: model.AttrScopeInventory,
						},
						{
							Name:  "sn",
							Value: "0003-sn",
							Scope: model.AttrScopeInventory,
						},
					},
					CreatedTs: createdTs,
				},
				{
					ID: model.DeviceID("0004"),
					Attributes: model.DeviceAttributes{
						{
							Name:  "mac",
							Value: "0004-mac",
							Scope: model.AttrScopeInventory,
						},
					},
					CreatedTs: createdTs,
				},
			},
			outFilterAttributes: []model.FilterAttribute{
				{
					Name:  "mac",
					Scope: model.AttrScopeInventory,
					Count: 2,
				},
				{
					Name:  "created_ts",
					Scope: model.AttrScopeSystem,
					Count: 2,
				},
				{
					Name:  "sn",
					Scope: model.AttrScopeInventory,
					Count: 1,
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// setup
			db.Wipe()
			s := db.Client()

			var ctx context.Context
			if tc.tenant != "" {
				ctx = identity.WithContext(db.CTX(), &identity.Identity{
					Tenant: tc.tenant,
				})
			} else {
				ctx = identity.WithContext(db.CTX(), &identity.Identity{
					Tenant: "",
				})
			}

			// test
			d := NewDataStoreMongoWithSession(s)
			for _, dev := range tc.devs {
				err := d.AddDevice(ctx, &dev)
				assert.NoError(t, err, "failed to setup input data")
			}

			attributes, err := d.GetFiltersAttributes(ctx)
			assert.NoError(t, err)
			assert.Equal(t, tc.outFilterAttributes, attributes)
		})
	}
}

func TestMongoUpdateDeviceGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMongoUpdateDeviceGroup in short mode.")
	}

	testCases := map[string]struct {
		InputDeviceIDs []model.DeviceID
		InputGroupName model.GroupName
		InputDevices   []model.Device
		tenant         string
		OutputError    error
		Result         model.UpdateResult
	}{
		"update group for device with empty device id": {
			InputDeviceIDs: nil,
			InputGroupName: model.GroupName("abc"),
			InputDevices:   nil,
		},
		"update group for device with empty device id; with tenant": {
			InputDeviceIDs: nil,
			InputGroupName: model.GroupName("abc"),
			InputDevices:   nil,
			tenant:         "foo",
		},
		"update group for device, device not found": {
			InputDeviceIDs: []model.DeviceID{"2"},
			InputGroupName: model.GroupName("abc"),
			InputDevices:   nil,
		},
		"update group for device, group exists": {
			InputDeviceIDs: []model.DeviceID{"1"},
			InputGroupName: model.GroupName("abc"),
			InputDevices: []model.Device{{
				ID:    model.DeviceID("1"),
				Group: model.GroupName("def"),
			}},
			Result: model.UpdateResult{
				MatchedCount: 1,
				UpdatedCount: 1,
			},
		},
		"update group for device, group exists; with tenant": {
			InputDeviceIDs: []model.DeviceID{"1"},
			InputGroupName: model.GroupName("abc"),
			InputDevices: []model.Device{{
				ID:    model.DeviceID("1"),
				Group: model.GroupName("def"),
			}},
			tenant: "foo",
			Result: model.UpdateResult{
				MatchedCount: 1,
				UpdatedCount: 1,
			},
		},
		"update group for device, group does not exist": {
			InputDeviceIDs: []model.DeviceID{"1"},
			InputGroupName: model.GroupName("abc"),
			InputDevices: []model.Device{{
				ID: model.DeviceID("1"),
			}},
			Result: model.UpdateResult{
				MatchedCount: 1,
				UpdatedCount: 1,
			},
		},
		"update group for multiple devices": {
			InputDeviceIDs: []model.DeviceID{"1", "2", "3", "4", "5"},
			InputGroupName: model.GroupName("grp2"),
			InputDevices: []model.Device{{
				ID:    model.DeviceID("1"),
				Group: model.GroupName("grp1"),
			}, {
				ID:    model.DeviceID("2"),
				Group: model.GroupName("grp1"),
			}, {
				ID:    model.DeviceID("3"),
				Group: model.GroupName("grp1"),
			}, {
				ID:    model.DeviceID("4"),
				Group: model.GroupName("grp2"),
			}, {
				ID:    model.DeviceID("6"),
				Group: model.GroupName("grp3"),
			}},
			Result: model.UpdateResult{
				MatchedCount: 4,
				UpdatedCount: 3,
			},
		},
	}

	for name, testCase := range testCases {
		t.Logf("test case: %s", name)

		// Make sure we start test with empty database
		db.Wipe()

		client := db.Client()
		store := NewDataStoreMongoWithSession(client)

		var ctx context.Context
		if testCase.tenant != "" {
			ctx = identity.WithContext(db.CTX(), &identity.Identity{
				Tenant: testCase.tenant,
			})
		} else {
			ctx = identity.WithContext(db.CTX(), &identity.Identity{
				Tenant: "",
			})
		}

		if testCase.InputDevices != nil {
			ins := make(bson.A, len(testCase.InputDevices))
			for i := range testCase.InputDevices {
				ins[i] = &testCase.InputDevices[i]
			}
			_, err := client.Database(mstore.DbFromContext(ctx, DbName)).
				Collection(DbDevicesColl).
				InsertMany(ctx, ins)
			if err != nil {
				panic(err)
			}
		}

		result, err := store.UpdateDevicesGroup(
			ctx, testCase.InputDeviceIDs, testCase.InputGroupName,
		)
		if testCase.OutputError != nil {
			assert.Error(t, err, "expected error")
			assert.EqualError(t, err, testCase.OutputError.Error())
		} else {
			assert.NoError(t, err, "expected no error")
			if assert.NotNil(t, result) {
				assert.Equal(t, testCase.Result, *result)
			}
		}
	}
}

func TestMongoUpdateDeviceText(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping UpdateDeviceText in short mode.")
	}

	testCases := map[string]struct {
		InputDeviceID model.DeviceID
		Text          string
		Tenant        string
	}{
		"ok, single tenant": {
			InputDeviceID: model.DeviceID("1"),
			Text:          "text",
			Tenant:        "",
		},
		"ok, multi tenant": {
			InputDeviceID: model.DeviceID("1"),
			Text:          "text",
			Tenant:        "tenant1",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Make sure we start test with empty database
			db.Wipe()
			client := db.Client()
			store := NewDataStoreMongoWithSession(client)

			var ctx context.Context
			if tc.Tenant != "" {
				ctx = identity.WithContext(db.CTX(), &identity.Identity{
					Tenant: tc.Tenant,
				})
			} else {
				ctx = identity.WithContext(db.CTX(), &identity.Identity{
					Tenant: "",
				})
			}
			err := store.AddDevice(ctx, &model.Device{ID: tc.InputDeviceID})
			assert.NoError(t, err)

			err = store.UpdateDeviceText(ctx, tc.InputDeviceID, tc.Text)
			assert.NoError(t, err)
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func compareAttrsWithoutTimestamp(a, b model.DeviceAttributes) bool {
	la := len(a)
	lb := len(b)
	var i, j int
	for i < la && j < lb {
		va := a[i]
		switch va.Name {
		case "created_ts", "updated_ts":
			i++
			continue
		}
		vb := b[j]
		switch vb.Name {
		case "created_ts", "updated_ts":
			j++
			continue
		}

		if !reflect.DeepEqual(va.Value, vb.Value) {
			return false
		}

		if !reflect.DeepEqual(va.Description, vb.Description) {
			return false
		}
		i++
		j++
	}

	return true
}

func TestMongoDeleteGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMongoUnsetDevicesGroupWithmodel.GroupName in short mode.")
	}

	groupName := model.GroupName("group")
	inputDevices := []model.Device{
		{
			ID:    model.DeviceID("1"),
			Group: groupName,
		},
		{
			ID:    model.DeviceID("2"),
			Group: groupName,
		},
		{
			ID:    model.DeviceID("3"),
			Group: model.GroupName("other-group"),
		},
		{
			ID: model.DeviceID("4"),
		},
	}

	ctx := identity.WithContext(db.CTX(), &identity.Identity{
		Tenant: "",
	})

	// Make sure we start test with empty database
	db.Wipe()

	// insert the devices
	client := db.Client()
	devices := make(bson.A, len(inputDevices))
	for i := range inputDevices {
		devices[i] = &inputDevices[i]
	}
	_, err := client.Database(mstore.DbFromContext(ctx, DbName)).
		Collection(DbDevicesColl).
		InsertMany(ctx, devices)
	if err != nil {
		panic(err)
	}

	// test the DeleteGroup method
	store := NewDataStoreMongoWithSession(client)
	deviceIDs, err := store.DeleteGroup(ctx, groupName)
	assert.NoError(t, err)

	updated := 0
	for deviceID := range deviceIDs {
		assert.Contains(t, []model.DeviceID{inputDevices[0].ID, inputDevices[1].ID}, deviceID)
		updated++
	}
	assert.Equal(t, 2, updated)
}

func TestMongoUnsetDevicesGroupWithGroupName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMongoUnsetDevicesGroupWithmodel.GroupName in short mode.")
	}

	testCases := map[string]struct {
		InputDeviceIDs []model.DeviceID
		InputGroupName model.GroupName
		InputDevices   []model.Device
		tenant         string
		OutputError    error
		Result         model.UpdateResult
	}{
		"unset group for device with group id, device not found": {
			InputDeviceIDs: []model.DeviceID{"1"},
			InputGroupName: model.GroupName("e16c71ec"),
			InputDevices:   nil,
		},
		"unset group for device with group id, device not found; with tenant": {
			InputDeviceIDs: []model.DeviceID{"1"},
			InputGroupName: model.GroupName("e16c71ec"),
			InputDevices:   nil,
			tenant:         "foo",
		},
		"unset group for device, ok": {
			InputDeviceIDs: []model.DeviceID{"1"},
			InputGroupName: model.GroupName("e16c71ec"),
			InputDevices: []model.Device{{
				ID:    model.DeviceID("1"),
				Group: model.GroupName("e16c71ec"),
			}},
			Result: model.UpdateResult{
				MatchedCount: 1,
				UpdatedCount: 1,
			},
		},
		"unset group for device, ok; with tenant": {
			InputDeviceIDs: []model.DeviceID{"1"},
			InputGroupName: model.GroupName("e16c71ec"),
			InputDevices: []model.Device{{
				ID:    model.DeviceID("1"),
				Group: model.GroupName("e16c71ec"),
			}},
			tenant: "foo",
			Result: model.UpdateResult{
				MatchedCount: 1,
				UpdatedCount: 1,
			},
		},
		"unset group for device with incorrect group name provided": {
			InputDeviceIDs: []model.DeviceID{"1"},
			InputGroupName: model.GroupName("other-group-name"),
			InputDevices: []model.Device{{
				ID:    model.DeviceID("1"),
				Group: model.GroupName("e16c71ec"),
			}},
			Result: model.UpdateResult{},
		},
		"unset group for device with incorrect group name provided; with tenant": {
			InputDeviceIDs: []model.DeviceID{"1"},
			InputGroupName: model.GroupName("other-group-name"),
			InputDevices: []model.Device{{
				ID:    model.DeviceID("1"),
				Group: model.GroupName("e16c71ec"),
			}},
			tenant: "foo",
			Result: model.UpdateResult{},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Logf("test case: %s", name)

			// Make sure we start test with empty database
			db.Wipe()

			client := db.Client()
			store := NewDataStoreMongoWithSession(client)

			var ctx context.Context
			if testCase.tenant != "" {
				ctx = identity.WithContext(db.CTX(), &identity.Identity{
					Tenant: testCase.tenant,
				})
			} else {
				ctx = identity.WithContext(db.CTX(), &identity.Identity{
					Tenant: "",
				})
			}

			if testCase.InputDevices != nil {
				ins := make(bson.A, len(testCase.InputDevices))
				for i := range testCase.InputDevices {
					ins[i] = &testCase.InputDevices[i]
				}
				_, err := client.Database(mstore.DbFromContext(ctx, DbName)).
					Collection(DbDevicesColl).
					InsertMany(ctx, ins)
				if err != nil {
					panic(err)
				}
			}

			res, err := store.UnsetDevicesGroup(ctx, testCase.InputDeviceIDs, testCase.InputGroupName)
			if testCase.OutputError != nil {
				assert.Error(t, err, "expected error")

				assert.EqualError(t, err, testCase.OutputError.Error())
			} else {
				assert.NoError(t, err, "expected no error")
				if assert.NotNil(t, res) {
					assert.Equal(t, testCase.Result, *res)
				}
			}
		})
	}
}

func TestMongoListGroups(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMongoListGroups in short mode.")
	}

	testCases := map[string]struct {
		InputDevices []model.Device
		Filters      []model.FilterPredicate
		tenant       string
		OutputGroups []model.GroupName
		Error        error
	}{
		"groups foo, bar": {
			InputDevices: []model.Device{
				{
					ID:    model.DeviceID("1"),
					Group: model.GroupName("foo"),
				},
				{
					ID:    model.DeviceID("2"),
					Group: model.GroupName("foo"),
				},
				{
					ID:    model.DeviceID("3"),
					Group: model.GroupName("foo"),
				},
				{
					ID:    model.DeviceID("4"),
					Group: model.GroupName("bar"),
				},
				{
					ID:    model.DeviceID("5"),
					Group: model.GroupName(""),
				},
			},
			OutputGroups: []model.GroupName{"foo", "bar"},
		},
		"groups foo, bar; with tenant": {
			InputDevices: []model.Device{
				{
					ID:    model.DeviceID("1"),
					Group: model.GroupName("foo"),
				},
				{
					ID:    model.DeviceID("2"),
					Group: model.GroupName("foo"),
				},
				{
					ID:    model.DeviceID("3"),
					Group: model.GroupName("foo"),
				},
				{
					ID:    model.DeviceID("4"),
					Group: model.GroupName("bar"),
				},
				{
					ID:    model.DeviceID("5"),
					Group: model.GroupName(""),
				},
			},
			tenant:       "foo",
			OutputGroups: []model.GroupName{"foo", "bar"},
		},
		"no groups": {
			InputDevices: []model.Device{
				{
					ID:    model.DeviceID("1"),
					Group: model.GroupName(""),
				},
				{
					ID:    model.DeviceID("2"),
					Group: model.GroupName(""),
				},
				{
					ID:    model.DeviceID("3"),
					Group: model.GroupName(""),
				},
			},
			OutputGroups: []model.GroupName{},
		},
		"no groups; with tenant": {
			InputDevices: []model.Device{
				{
					ID:    model.DeviceID("1"),
					Group: model.GroupName(""),
				},
				{
					ID:    model.DeviceID("2"),
					Group: model.GroupName(""),
				},
				{
					ID:    model.DeviceID("3"),
					Group: model.GroupName(""),
				},
			},
			tenant:       "foo",
			OutputGroups: []model.GroupName{},
		},
		"ok, filtered result": {
			InputDevices: []model.Device{
				{
					ID:    model.DeviceID("1"),
					Group: model.GroupName("foo"),
					Attributes: model.DeviceAttributes{{
						Scope: model.AttrScopeIdentity,
						Name:  "status",
						Value: "accepted",
					}},
				},
				{
					ID:    model.DeviceID("2"),
					Group: model.GroupName("foo"),
					Attributes: model.DeviceAttributes{{
						Scope: model.AttrScopeIdentity,
						Name:  "status",
						Value: "accepted",
					}},
				},
				{
					ID:    model.DeviceID("3"),
					Group: model.GroupName("foo"),
					Attributes: model.DeviceAttributes{{
						Scope: model.AttrScopeIdentity,
						Name:  "status",
						Value: "rejected",
					}},
				},
				{
					ID:    model.DeviceID("4"),
					Group: model.GroupName("bar"),
					Attributes: model.DeviceAttributes{{
						Scope: model.AttrScopeIdentity,
						Name:  "status",
						Value: "noauth",
					}},
				},
				{
					ID:    model.DeviceID("5"),
					Group: model.GroupName(""),
					Attributes: model.DeviceAttributes{{
						Scope: model.AttrScopeIdentity,
						Name:  "status",
						Value: "accepted",
					}},
				},
			},
			Filters: []model.FilterPredicate{{
				Attribute: "status",
				Scope:     model.AttrScopeIdentity,
				Type:      "$eq",
				Value:     "accepted",
			}},
			tenant:       "foo",
			OutputGroups: []model.GroupName{"foo"},
		},
		"error, bad filter": {
			InputDevices: []model.Device{{
				ID:    model.DeviceID("1"),
				Group: model.GroupName("foo"),
			}, {
				ID:    model.DeviceID("2"),
				Group: model.GroupName("foo"),
			}, {
				ID:    model.DeviceID("3"),
				Group: model.GroupName("foo"),
			}, {
				ID:    model.DeviceID("4"),
				Group: model.GroupName("bar"),
			}, {
				ID:    model.DeviceID("5"),
				Group: model.GroupName(""),
			}},
			Filters: []model.FilterPredicate{{
				Attribute: "status",
				Scope:     model.AttrScopeIdentity,
				Type:      "$expr",
				Value:     "accepted",
			}},
			tenant:       "foo",
			OutputGroups: []model.GroupName{"foo"},
			Error: errors.New(
				"store: bad filter predicate: " +
					"type: must be a valid value.",
			),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			db.Wipe()

			client := db.Client()

			var ctx context.Context
			if testCase.tenant != "" {
				ctx = identity.WithContext(db.CTX(), &identity.Identity{
					Tenant: testCase.tenant,
				})
			} else {
				ctx = identity.WithContext(db.CTX(), &identity.Identity{
					Tenant: "",
				})
			}

			for _, d := range testCase.InputDevices {
				client.Database(mstore.DbFromContext(ctx, DbName)).
					Collection(DbDevicesColl).
					InsertOne(ctx, d)
			}

			// Make sure we start test with empty database
			store := NewDataStoreMongoWithSession(client)

			groups, err := store.ListGroups(ctx, testCase.Filters)
			if testCase.Error != nil {
				assert.EqualError(t, err, testCase.Error.Error())
				return
			}
			assert.NoError(t, err, "expected no error")

			if testCase.OutputGroups != nil {
				assert.Len(t, groups, len(testCase.OutputGroups))
				for _, eg := range testCase.OutputGroups {
					assert.Contains(t, groups, eg)
				}
			} else {
				assert.Len(t, groups, 0)
			}
		})
	}
}

func TestGetDevicesByGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestGetDevicesByGroup in short mode.")
	}

	devDevices := []model.Device{
		{
			ID:    model.DeviceID("1"),
			Group: model.GroupName("dev"),
		},
		{
			ID:    model.DeviceID("6"),
			Group: model.GroupName("dev"),
		},
		{
			ID:    model.DeviceID("8"),
			Group: model.GroupName("dev"),
		},
	}
	prodDevices := []model.Device{
		{
			ID:    model.DeviceID("2"),
			Group: model.GroupName("prod"),
		},
		{
			ID:    model.DeviceID("4"),
			Group: model.GroupName("prod"),
		},
		{
			ID:    model.DeviceID("5"),
			Group: model.GroupName("prod"),
		},
	}
	testDevices := []model.Device{
		{
			ID:    model.DeviceID("3"),
			Group: model.GroupName("test"),
		},
		{
			ID:    model.DeviceID("7"),
			Group: model.GroupName("test"),
		},
	}

	inputDevices := make([]model.Device, 0, len(devDevices)+len(prodDevices)+len(testDevices))
	inputDevices = append(inputDevices, devDevices...)
	inputDevices = append(inputDevices, prodDevices...)
	inputDevices = append(inputDevices, testDevices...)

	testCases := map[string]struct {
		InputGroupName    model.GroupName
		InputSkip         int
		InputLimit        int
		OutputDevices     []model.DeviceID
		OutputDeviceCount int
		OutputError       error
	}{
		"no skip, no limit": {
			InputGroupName: "dev",
			InputSkip:      0,
			InputLimit:     0,
			OutputDevices: []model.DeviceID{
				model.DeviceID("1"),
				model.DeviceID("6"),
				model.DeviceID("8"),
			},
			OutputDeviceCount: len(devDevices),
			OutputError:       nil,
		},
		"no skip, limit": {
			InputGroupName: "prod",
			InputSkip:      0,
			InputLimit:     2,
			OutputDevices: []model.DeviceID{
				model.DeviceID("2"),
				model.DeviceID("4"),
			},
			OutputDeviceCount: len(prodDevices),
			OutputError:       nil,
		},
		"skip, no limit": {
			InputGroupName: "dev",
			InputSkip:      2,
			InputLimit:     0,
			OutputDevices: []model.DeviceID{
				model.DeviceID("8"),
			},
			OutputDeviceCount: len(devDevices),
			OutputError:       nil,
		},
		"skip + limit": {
			InputGroupName: "prod",
			InputSkip:      1,
			InputLimit:     1,
			OutputDevices: []model.DeviceID{
				model.DeviceID("4"),
			},
			OutputDeviceCount: len(prodDevices),
			OutputError:       nil,
		},
		"no results (past last page)": {
			InputGroupName:    "dev",
			InputSkip:         10,
			InputLimit:        1,
			OutputDevices:     []model.DeviceID{},
			OutputDeviceCount: len(devDevices),
			OutputError:       nil,
		},
		"group doesn't exist": {
			InputGroupName:    "unknown",
			InputSkip:         0,
			InputLimit:        0,
			OutputDevices:     nil,
			OutputDeviceCount: -1,
			OutputError:       store.ErrGroupNotFound,
		},
		"dev group": {
			InputGroupName: "dev",
			InputSkip:      0,
			InputLimit:     10,
			OutputDevices: []model.DeviceID{
				model.DeviceID("1"),
				model.DeviceID("6"),
				model.DeviceID("8"),
			},
			OutputDeviceCount: len(devDevices),
			OutputError:       nil,
		},
		"prod group": {
			InputGroupName: "prod",
			InputSkip:      0,
			InputLimit:     10,
			OutputDevices: []model.DeviceID{
				model.DeviceID("2"),
				model.DeviceID("4"),
				model.DeviceID("5"),
			},
			OutputDeviceCount: len(prodDevices),
			OutputError:       nil,
		},
		"test group": {
			InputGroupName: "test",
			InputSkip:      0,
			InputLimit:     10,
			OutputDevices: []model.DeviceID{
				model.DeviceID("3"),
				model.DeviceID("7"),
			},
			OutputDeviceCount: len(testDevices),
			OutputError:       nil,
		},
	}

	db.Wipe()
	client := db.Client()

	for _, d := range inputDevices {
		_, err := client.Database(DbName).Collection(DbDevicesColl).InsertOne(db.CTX(), d)
		assert.NoError(t, err, "failed to setup input data")
	}

	for name, tc := range testCases {
		t.Logf("test case: %s", name)

		store := NewDataStoreMongoWithSession(client)

		devs, totalCount, err := store.GetDevicesByGroup(db.CTX(), tc.InputGroupName, tc.InputSkip, tc.InputLimit)

		if tc.OutputError != nil {
			assert.EqualError(t, err, tc.OutputError.Error())
		} else {
			assert.NoError(t, err, "expected no error")
			if !reflect.DeepEqual(tc.OutputDevices, devs) {
				assert.Fail(t, "expected outputDevices to match", fmt.Sprintf("Expected: %v but\n have:%v", tc.OutputDevices, devs))
			}
			if !reflect.DeepEqual(tc.OutputDeviceCount, totalCount) {
				assert.Fail(t, "expected outputDeviceCount to match", fmt.Sprintf("Expected: %v but\n have:%v", tc.OutputDeviceCount, totalCount))
			}
		}
	}
}

func TestGetDevicesByGroupWithTenant(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestGetDevicesByGroupWithTenant in short mode.")
	}

	inputDevices := []model.Device{
		{
			ID:    model.DeviceID("1"),
			Group: model.GroupName("dev"),
		},
		{
			ID:    model.DeviceID("2"),
			Group: model.GroupName("prod"),
		},
		{
			ID:    model.DeviceID("3"),
			Group: model.GroupName("test"),
		},
		{
			ID:    model.DeviceID("4"),
			Group: model.GroupName("prod"),
		},
		{
			ID:    model.DeviceID("5"),
			Group: model.GroupName("prod"),
		},
		{
			ID:    model.DeviceID("6"),
			Group: model.GroupName("dev"),
		},
		{
			ID:    model.DeviceID("7"),
			Group: model.GroupName("test"),
		},
		{
			ID:    model.DeviceID("8"),
			Group: model.GroupName("dev"),
		},
	}

	testCases := map[string]struct {
		InputGroupName    model.GroupName
		InputSkip         int
		InputLimit        int
		OutputDevices     []model.DeviceID
		OutputDeviceCount int
		OutputError       error
	}{
		"no skip, no limit": {
			InputGroupName: "dev",
			InputSkip:      0,
			InputLimit:     0,
			OutputDevices: []model.DeviceID{
				model.DeviceID("1"),
				model.DeviceID("6"),
				model.DeviceID("8"),
			},
			OutputDeviceCount: 3,
			OutputError:       nil,
		},
		"no skip, limit": {
			InputGroupName: "prod",
			InputSkip:      0,
			InputLimit:     2,
			OutputDevices: []model.DeviceID{
				model.DeviceID("2"),
				model.DeviceID("4"),
			},
			OutputDeviceCount: 3,
			OutputError:       nil,
		},
		"skip, no limit": {
			InputGroupName: "dev",
			InputSkip:      2,
			InputLimit:     0,
			OutputDevices: []model.DeviceID{
				model.DeviceID("8"),
			},
			OutputDeviceCount: 3,
			OutputError:       nil,
		},
		"skip + limit": {
			InputGroupName: "prod",
			InputSkip:      1,
			InputLimit:     1,
			OutputDevices: []model.DeviceID{
				model.DeviceID("4"),
			},
			OutputDeviceCount: 3,
			OutputError:       nil,
		},
		"no results (past last page)": {
			InputGroupName:    "dev",
			InputSkip:         10,
			InputLimit:        1,
			OutputDevices:     []model.DeviceID{},
			OutputDeviceCount: 3,
			OutputError:       nil,
		},
		"group doesn't exist": {
			InputGroupName:    "unknown",
			InputSkip:         0,
			InputLimit:        0,
			OutputDevices:     nil,
			OutputDeviceCount: -1,
			OutputError:       store.ErrGroupNotFound,
		},
	}

	db.Wipe()
	client := db.Client()

	for _, d := range inputDevices {
		_, err := client.Database(DbName+"-foo").Collection(DbDevicesColl).InsertOne(db.CTX(), d)
		assert.NoError(t, err, "failed to setup input data")
	}

	for name, tc := range testCases {
		t.Logf("test case: %s", name)

		store := NewDataStoreMongoWithSession(client)

		ctx := identity.WithContext(db.CTX(), &identity.Identity{
			Tenant: "foo",
		})
		devs, totalCount, err := store.GetDevicesByGroup(ctx, tc.InputGroupName, tc.InputSkip, tc.InputLimit)

		if tc.OutputError != nil {
			assert.EqualError(t, err, tc.OutputError.Error())
		} else {
			assert.NoError(t, err, "expected no error")
			if !reflect.DeepEqual(tc.OutputDevices, devs) {
				assert.Fail(t, "expected outputDevices to match", fmt.Sprintf("Expected: %v but\n have:%v", tc.OutputDevices, devs))
			}
			if !reflect.DeepEqual(tc.OutputDeviceCount, totalCount) {
				assert.Fail(t, "expected outputDeviceCount to match", fmt.Sprintf("Expected: %v but\n have:%v", tc.OutputDeviceCount, totalCount))
			}
		}
	}
}

func TestGetDeviceGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestGetDeviceGroup in short mode.")
	}

	inputDevices := []model.Device{
		{
			ID:    model.DeviceID("1"),
			Group: model.GroupName("dev"),
		},
		{
			ID: model.DeviceID("2"),
		},
	}

	testCases := map[string]struct {
		InputDeviceID model.DeviceID
		OutputGroup   model.GroupName
		OutputError   error
	}{
		"dev has group": {
			InputDeviceID: model.DeviceID("1"),
			OutputGroup:   model.GroupName("dev"),
			OutputError:   nil,
		},
		"dev has no group": {
			InputDeviceID: model.DeviceID("2"),
			OutputGroup:   "",
			OutputError:   nil,
		},
		"dev doesn't exist": {
			InputDeviceID: model.DeviceID("3"),
			OutputGroup:   "",
			OutputError:   store.ErrDevNotFound,
		},
	}

	db.Wipe()
	client := db.Client()
	store := NewDataStoreMongoWithSession(client)
	for _, dev := range inputDevices {
		err := store.AddDevice(db.CTX(), &dev)
		assert.NoError(t, err, "failed to setup input data")
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			group, err := store.GetDeviceGroup(
				db.CTX(), tc.InputDeviceID,
			)
			if tc.OutputError != nil {
				assert.EqualError(t, err, tc.OutputError.Error())
			} else {
				assert.NoError(t, err, "expected no error")
				if !assert.Equal(t, tc.OutputGroup, group) {
					time.Sleep(time.Minute * 5)
				}
			}
		})
	}
}

func TestGetDeviceGroupWithTenant(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestGetDeviceGroupWithTenant in short mode.")
	}

	inputDevices := []model.Device{
		{
			ID:    model.DeviceID("1"),
			Group: model.GroupName("dev"),
		},
		{
			ID: model.DeviceID("2"),
		},
	}

	testCases := map[string]struct {
		InputDeviceID model.DeviceID
		OutputGroup   model.GroupName
		OutputError   error
	}{
		"dev has group": {
			InputDeviceID: model.DeviceID("1"),
			OutputGroup:   model.GroupName("dev"),
			OutputError:   nil,
		},
		"dev has no group": {
			InputDeviceID: model.DeviceID("2"),
			OutputGroup:   "",
			OutputError:   nil,
		},
		"dev doesn't exist": {
			InputDeviceID: model.DeviceID("3"),
			OutputGroup:   "",
			OutputError:   store.ErrDevNotFound,
		},
	}

	db.Wipe()
	client := db.Client()

	for _, d := range inputDevices {
		_, err := client.Database(DbName+"-foo").Collection(DbDevicesColl).InsertOne(db.CTX(), d)
		assert.NoError(t, err, "failed to setup input data")
	}

	for name, tc := range testCases {
		t.Logf("test case: %s", name)

		store := NewDataStoreMongoWithSession(client)

		ctx := identity.WithContext(db.CTX(), &identity.Identity{
			Tenant: "foo",
		})
		group, err := store.GetDeviceGroup(ctx, tc.InputDeviceID)

		if tc.OutputError != nil {
			assert.EqualError(t, err, tc.OutputError.Error())
		} else {
			assert.NoError(t, err, "expected no error")
			assert.Equal(t, tc.OutputGroup, group)
		}
	}
}

func TestMigrate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMigrate in short mode.")
	}

	someDevs := []model.Device{
		{
			ID: model.DeviceID("0"),
			Attributes: model.DeviceAttributes{
				{Name: "mac", Value: "foo", Description: strPtr("desc"), Scope: model.AttrScopeInventory},
				{Name: "sn", Value: "bar", Description: strPtr("desc"), Scope: model.AttrScopeInventory},
			},
		},
		{
			ID: model.DeviceID("1"),
			Attributes: model.DeviceAttributes{
				{Name: "mac", Value: "foo", Description: strPtr("desc"), Scope: model.AttrScopeInventory},
				{Name: "foo", Value: "foo", Description: strPtr("desc"), Scope: model.AttrScopeInventory},
				{Name: "bar", Value: "bar", Description: strPtr("desc"), Scope: model.AttrScopeInventory},
			},
		},
		{
			ID: model.DeviceID("2"),
			Attributes: model.DeviceAttributes{
				{Name: "baz", Value: "baz", Description: strPtr("desc"), Scope: model.AttrScopeInventory},
			},
		},
	}

	testCases := map[string]struct {
		versionFrom string
		inDevs      []model.Device
		automigrate bool
		tenant      string

		outVers []string
		err     error
	}{
		"from no version (fresh db)": {
			versionFrom: "",
			automigrate: true,

			outVers: []string{
				"0.2.0",
				"1.0.0",
				"1.0.1",
				"1.0.2",
				DbVersion,
			},
		},
		"from 0.1.0 (first, dummy migration)": {
			versionFrom: "0.1.0",
			automigrate: true,

			outVers: []string{
				"0.1.0",
				"0.2.0",
				"1.0.0",
				"1.0.1",
				"1.0.2",
				DbVersion,
			},
		},
		"from 0.1.0, no-automigrate": {
			versionFrom: "0.0.0",

			err: errors.New("failed to apply migrations: db needs migration: inventory has version 0.0.0, needs version " + DbVersion),
		},
		"with devices, from 0.1.0": {
			versionFrom: "0.1.0",
			inDevs:      someDevs,
			automigrate: true,

			outVers: []string{
				"0.1.0",
				"0.2.0",
				"1.0.0",
				"1.0.1",
				"1.0.2",
				DbVersion,
			},
		},
		"with devices, from 0.1.0, with tenant": {
			versionFrom: "",
			inDevs:      someDevs,
			automigrate: true,
			tenant:      "tenant",

			outVers: []string{
				"0.2.0",
				"1.0.0",
				"1.0.1",
				"1.0.2",
				DbVersion,
			},
		},
		"with devices, from 0.1.0, with tenant, other devs": {
			versionFrom: "0.1.0",
			inDevs:      []model.Device{someDevs[0], someDevs[2]},
			automigrate: true,
			tenant:      "tenant",

			outVers: []string{
				"0.1.0",
				"0.2.0",
				"1.0.0",
				"1.0.1",
				"1.0.2",
				DbVersion,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			db.Wipe()

			client := db.Client()

			var ctx context.Context
			if tc.tenant != "" {
				ctx = identity.WithContext(db.CTX(), &identity.Identity{
					Tenant: tc.tenant,
				})
			} else {
				ctx = identity.WithContext(db.CTX(), &identity.Identity{
					Tenant: "",
				})
			}

			store := NewDataStoreMongoWithSession(client)

			if tc.automigrate {
				store = store.WithAutomigrate()
			}

			// prep input data
			// input migrations
			if tc.versionFrom != "" {
				v, err := migrate.NewVersion(tc.versionFrom)
				assert.NoError(t, err)

				entry := migrate.MigrationEntry{
					Version: *v,
				}
				assert.NoError(t, err)

				_, err = client.
					Database(mstore.DbFromContext(ctx, DbName)).
					Collection(migrate.DbMigrationsColl).
					InsertOne(ctx, entry)
				assert.NoError(t, err)
			}

			// input devices
			for _, d := range tc.inDevs {
				err := store.AddDevice(ctx, &d)
				assert.NoError(t, err)
			}

			err := store.Migrate(ctx, DbVersion)
			if tc.err == nil {
				assert.NoError(t, err)

				// verify migration entries
				var out []migrate.MigrationEntry
				cursor, _ := client.
					Database(mstore.DbFromContext(ctx, DbName)).
					Collection(migrate.DbMigrationsColl).
					Find(ctx, bson.M{})

				count := 0
				for cursor.Next(db.CTX()) {
					var res migrate.MigrationEntry
					count++
					err = cursor.Decode(&res)
					out = append(out, res)
				}

				assert.Equal(t, len(tc.outVers), count)
				for i, v := range tc.outVers {
					assert.Equal(t, v, out[i].Version.String())
				}
			} else {
				assert.EqualError(t, err, tc.err.Error())
			}
		})
	}
}

func TestMongoDeleteDevices(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMongoDeleteDevice in short mode.")
	}

	inputDevs := []model.Device{
		{ID: model.DeviceID("0")},
		{ID: model.DeviceID("1")},
	}

	testCases := map[string]struct {
		inputIDs []model.DeviceID
		expected []model.Device
		err      error
		result   model.UpdateResult
	}{
		"existing 1": {
			inputIDs: []model.DeviceID{"0"},
			expected: []model.Device{
				{ID: model.DeviceID("1")},
			},
			err: nil,
			result: model.UpdateResult{
				DeletedCount: 1,
			},
		},
		"delete all": {
			inputIDs: []model.DeviceID{"0", "1"},
			expected: []model.Device(nil),
			err:      nil,
			result: model.UpdateResult{
				DeletedCount: 2,
			},
		},
		"doesn't exist": {
			inputIDs: []model.DeviceID{"3"},
			expected: []model.Device{
				{ID: model.DeviceID("0")},
				{ID: model.DeviceID("1")},
			},
			result: model.UpdateResult{
				DeletedCount: 0,
			},
		},
	}

	for name, tc := range testCases {
		t.Logf("test case: %s", name)
		t.Run(name, func(t *testing.T) {

			// Make sure we start test with empty database
			db.Wipe()

			client := db.Client()

			for _, d := range inputDevs {
				_, err := client.Database(DbName).
					Collection(DbDevicesColl).
					InsertOne(db.CTX(), d)
				assert.NoError(t, err, "failed to setup input data")
			}

			store := NewDataStoreMongoWithSession(client)

			//test
			result, err := store.DeleteDevices(db.CTX(), tc.inputIDs)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err, "failed to delete device")
				if assert.NotNil(t, result) {
					assert.Equal(t, tc.result, *result)
				}

				var outDevs []model.Device
				cursor, err := client.Database(DbName).
					Collection(DbDevicesColl).
					Find(db.CTX(), bson.M{})
				assert.NoError(t, err, "failed to verify devices")
				cursor.All(nil, &outDevs)
				assert.Equal(t, tc.expected, outDevs)
			}
		})
	}
}

func TestMongoSearchDevices(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMongoSearchDevices in short mode.")
	}

	now := time.Now()
	before := now.Add(-time.Hour)
	inputDevs := []model.Device{
		{
			ID: model.DeviceID("0"),
			Attributes: model.DeviceAttributes{
				{Name: "MAC", Value: "000", Description: strPtr("MAC"), Scope: model.AttrScopeInventory},
				{Name: "SN", Value: float64(100), Description: strPtr("SN"), Scope: model.AttrScopeInventory},
				{Name: "group", Value: "foo", Description: strPtr("group"), Scope: model.AttrScopeInventory},
				{Name: "ip.address", Value: "1.2.3.4", Scope: model.AttrScopeInventory},
				{Name: "name", Value: "device0", Scope: model.AttrScopeTags, Timestamp: &now},
			},
			Group:     "foo",
			CreatedTs: now,
			UpdatedTs: &now,
		},
		{
			ID: model.DeviceID("1"),
			Attributes: model.DeviceAttributes{
				{Name: "MAC", Value: "001", Description: strPtr("MAC"), Scope: model.AttrScopeInventory},
				{Name: "SN", Value: float64(111), Description: strPtr("SN"), Scope: model.AttrScopeInventory},
				{Name: "group", Value: "foo", Description: strPtr("group"), Scope: model.AttrScopeInventory},
				{Name: "name", Value: "device1", Scope: model.AttrScopeTags, Timestamp: &before},
			},
			Group:     "foo",
			CreatedTs: now,
			UpdatedTs: &now,
		},
		{
			ID: model.DeviceID("2"),
			Attributes: model.DeviceAttributes{
				{Name: "MAC", Value: "002", Description: strPtr("MAC"), Scope: model.AttrScopeInventory},
				{Name: "SN", Value: float64(122), Description: strPtr("SN"), Scope: model.AttrScopeInventory},
				{Name: "group", Value: "foo", Description: strPtr("group"), Scope: model.AttrScopeInventory},
				{Name: "name", Value: "device2", Scope: model.AttrScopeTags, Timestamp: &now},
			},
			Group:     "foo",
			CreatedTs: now,
			UpdatedTs: &now,
		},
		{
			ID: model.DeviceID("3"),
			Attributes: model.DeviceAttributes{
				{Name: "MAC", Value: "003", Description: strPtr("MAC"), Scope: model.AttrScopeInventory},
				{Name: "SN", Value: float64(133), Description: strPtr("SN"), Scope: model.AttrScopeInventory},
				{Name: "group", Value: "bar", Description: strPtr("group"), Scope: model.AttrScopeInventory},
				{Name: "name", Value: "device3", Scope: model.AttrScopeTags, Timestamp: &now},
			},
			Group:     "bar",
			CreatedTs: now,
			UpdatedTs: &now,
		},
		{
			ID: model.DeviceID("4"),
			Attributes: model.DeviceAttributes{
				{Name: "MAC", Value: "003", Description: strPtr("MAC"), Scope: model.AttrScopeInventory},
				{Name: "SN", Value: float64(144), Description: strPtr("SN"), Scope: model.AttrScopeInventory},
				{Name: "group", Value: "bar", Description: strPtr("group"), Scope: model.AttrScopeInventory},
				{Name: "text", Value: "this is a free-text searchable attribute", Scope: model.AttrScopeInventory},
				{Name: "name", Value: "device4", Scope: model.AttrScopeTags, Timestamp: &now},
			},
			Group:     "bar",
			CreatedTs: now,
			UpdatedTs: &now,
		},
	}

	testCases := map[string]struct {
		expected           []model.Device
		expectedAttributes model.DeviceAttributes
		devTotal           int
		searchParams       model.SearchParams
		tenant             string
		dbError            error
	}{
		"single filter, single device": {
			expected: []model.Device{inputDevs[0]},
			devTotal: 1,
			searchParams: model.SearchParams{
				Page:    1,
				PerPage: 5,
				Filters: []model.FilterPredicate{
					{
						Scope:     "inventory",
						Attribute: "MAC",
						Type:      "$eq",
						Value:     "000",
					},
				},
				Sort: []model.SortCriteria{},
			},
		},
		"single filter, single device, select single attribute": {
			expected: []model.Device{inputDevs[0]},
			expectedAttributes: []model.DeviceAttribute{
				inputDevs[0].Attributes[0],
			},
			devTotal: 1,
			searchParams: model.SearchParams{
				Page:    1,
				PerPage: 5,
				Filters: []model.FilterPredicate{
					{
						Scope:     "inventory",
						Attribute: "MAC",
						Type:      "$eq",
						Value:     "000",
					},
				},
				Attributes: []model.SelectAttribute{
					{
						Scope:     "inventory",
						Attribute: "MAC",
					},
				},
				Sort: []model.SortCriteria{},
			},
		},
		"single filter, single device, select single two attributes": {
			expected: []model.Device{inputDevs[0]},
			expectedAttributes: []model.DeviceAttribute{
				inputDevs[0].Attributes[0],
				inputDevs[0].Attributes[1],
			},
			devTotal: 1,
			searchParams: model.SearchParams{
				Page:    1,
				PerPage: 5,
				Filters: []model.FilterPredicate{
					{
						Scope:     "inventory",
						Attribute: "MAC",
						Type:      "$eq",
						Value:     "000",
					},
				},
				Attributes: []model.SelectAttribute{
					{
						Scope:     "inventory",
						Attribute: "MAC",
					},
					{
						Scope:     "inventory",
						Attribute: "SN",
					},
				},
				Sort: []model.SortCriteria{},
			},
		},
		"single filter, single device, with dots": {
			expected: []model.Device{inputDevs[0]},
			devTotal: 1,
			searchParams: model.SearchParams{
				Page:    1,
				PerPage: 5,
				Filters: []model.FilterPredicate{
					{
						Scope:     "inventory",
						Attribute: "ip.address",
						Type:      "$eq",
						Value:     "1.2.3.4",
					},
				},
				Sort: []model.SortCriteria{},
			},
		},
		"two filters, single device": {
			expected: []model.Device{inputDevs[0]},
			devTotal: 1,
			searchParams: model.SearchParams{
				Page:    1,
				PerPage: 5,
				Filters: []model.FilterPredicate{
					{
						Scope:     "inventory",
						Attribute: "MAC",
						Type:      "$eq",
						Value:     "000",
					},
					{
						Scope:     "inventory",
						Attribute: "SN",
						Type:      "$eq",
						Value:     100,
					},
				},
				Sort: []model.SortCriteria{},
			},
		},
		"two filters, three devices, sorted": {
			expected: []model.Device{inputDevs[2], inputDevs[1], inputDevs[0]},
			devTotal: 3,
			searchParams: model.SearchParams{
				Page:    1,
				PerPage: 5,
				Filters: []model.FilterPredicate{
					{
						Scope:     "inventory",
						Attribute: "MAC",
						Type:      "$in",
						Value:     []interface{}{"000", "001", "002"},
					},
					{
						Scope:     "inventory",
						Attribute: "SN",
						Type:      "$ne",
						Value:     200,
					},
				},
				Sort: []model.SortCriteria{
					{
						Scope:     "inventory",
						Attribute: "MAC",
						Order:     "desc",
					},
				},
			},
		},
		"one filter, two devices, sorted by two attrs": {
			expected: []model.Device{inputDevs[4], inputDevs[3]},
			devTotal: 2,
			searchParams: model.SearchParams{
				Page:    1,
				PerPage: 5,
				Filters: []model.FilterPredicate{
					{
						Scope:     "inventory",
						Attribute: "MAC",
						Type:      "$in",
						Value:     []interface{}{"003", "004"},
					},
				},
				Sort: []model.SortCriteria{
					{
						Scope:     "inventory",
						Attribute: "MAC",
						Order:     "asc",
					},
					{
						Scope:     "inventory",
						Attribute: "SN",
						Order:     "desc",
					},
				},
			},
		},
		"one filter, all devices, page and perPage": {
			expected: []model.Device{inputDevs[2], inputDevs[3]},
			devTotal: 5,
			searchParams: model.SearchParams{
				Page:    2,
				PerPage: 2,
				Filters: []model.FilterPredicate{
					{
						Scope:     "inventory",
						Attribute: "MAC",
						Type:      "$ne",
						Value:     "foo",
					},
				},
				Sort: []model.SortCriteria{
					{
						Scope:     "inventory",
						Attribute: "MAC",
						Order:     "asc",
					},
				},
			},
		},
		"one filter, device IDs param": {
			expected: []model.Device{inputDevs[0], inputDevs[2]},
			devTotal: 2,
			searchParams: model.SearchParams{
				Page:      1,
				PerPage:   5,
				DeviceIDs: []string{"0", "2"},
				Filters: []model.FilterPredicate{
					{
						Scope:     "inventory",
						Attribute: "SN",
						Type:      "$ne",
						Value:     200,
					},
				},
			},
		},
		"one filter, id": {
			expected: []model.Device{inputDevs[0]},
			devTotal: 1,
			searchParams: model.SearchParams{
				Page:    1,
				PerPage: 5,
				Filters: []model.FilterPredicate{
					{
						Scope:     "identity",
						Attribute: "id",
						Type:      "$eq",
						Value:     "0",
					},
				},
			},
		},
		"one filter, id not found": {
			expected: []model.Device{},
			devTotal: 0,
			searchParams: model.SearchParams{
				Page:    1,
				PerPage: 5,
				Filters: []model.FilterPredicate{
					{
						Scope:     "identity",
						Attribute: "id",
						Type:      "$eq",
						Value:     "not found",
					},
				},
			},
		},
		"$in, bad value": {
			expected: []model.Device{inputDevs[2], inputDevs[3]},
			devTotal: 5,
			searchParams: model.SearchParams{
				Page:    2,
				PerPage: 2,
				Filters: []model.FilterPredicate{
					{
						Scope:     "inventory",
						Attribute: "MAC",
						Type:      "$in",
						Value:     "foo",
					},
				},
			},
			dbError: errors.New("(BadValue) $in needs an array"),
		},
		"no filter": {
			expected: inputDevs,
			devTotal: 5,
			searchParams: model.SearchParams{
				Page:    1,
				PerPage: 5,
			},
		},
		"no filter, sort by ID desc": {
			expected: []model.Device{
				inputDevs[4],
				inputDevs[3],
				inputDevs[2],
				inputDevs[1],
				inputDevs[0],
			},
			devTotal: 5,
			searchParams: model.SearchParams{
				Page:    1,
				PerPage: 5,
				Sort: []model.SortCriteria{
					{
						Scope:     "identity",
						Attribute: "id",
						Order:     "desc",
					},
				},
			},
		},
		"no filter, sort by name desc": {
			expected: []model.Device{
				inputDevs[4],
				inputDevs[3],
				inputDevs[2],
				inputDevs[1],
				inputDevs[0],
			},
			devTotal: 5,
			searchParams: model.SearchParams{
				Page:    1,
				PerPage: 5,
				Sort: []model.SortCriteria{
					{
						Scope:     "tags",
						Attribute: "name",
						Order:     "desc",
					},
				},
			},
		},
		"free text search": {
			expected: []model.Device{inputDevs[4]},
			devTotal: 1,
			searchParams: model.SearchParams{
				Page:    1,
				PerPage: 5,
				Text:    utils.TextToKeywords("free-text attribute"),
				Sort:    []model.SortCriteria{},
			},
		},
	}

	for name, tc := range testCases {
		t.Logf("test case: %s", name)

		// Make sure we start test with empty database
		db.Wipe()

		client := db.Client()

		var ctx context.Context
		if tc.tenant != "" {
			ctx = identity.WithContext(db.CTX(), &identity.Identity{
				Tenant: tc.tenant,
			})
		} else {
			ctx = identity.WithContext(db.CTX(), &identity.Identity{
				Tenant: "",
			})
		}

		mongoStore := NewDataStoreMongoWithSession(client)

		for _, d := range inputDevs {
			err := mongoStore.AddDevice(ctx, &d)
			assert.NoError(t, err, "failed to setup input data")
		}

		//we need the $text index when testing the full-text search; apply migrations
		if tc.searchParams.Text != "" {
			mongoStore = mongoStore.WithAutomigrate()
			err := mongoStore.Migrate(ctx, DbVersion)
			assert.NoError(t, err)
		}

		//test
		devs, totalCount, err := mongoStore.SearchDevices(ctx, tc.searchParams)

		if tc.dbError != nil {
			assert.Error(t, tc.dbError, err)
		} else {
			assert.NoError(t, err, "failed to get devices")
			assert.Equal(t, len(tc.expected), len(devs))
			assert.Equal(t, tc.devTotal, totalCount)
			if len(tc.searchParams.Sort) > 0 {
				for i, dev := range devs {
					assert.Equal(t, tc.expected[i].ID, dev.ID)
				}
			}
			if len(tc.searchParams.Attributes) > 0 {
				for _, dev := range devs {
					assert.Equal(t, len(tc.expectedAttributes), len(dev.Attributes))
					assert.Equal(t, tc.expectedAttributes, dev.Attributes[:len(tc.expectedAttributes)])
				}
			}
		}

	}
}

func TestUpdateDevicesGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestUpdateDevicesGroup in short mode.")
	}

	deviceSet := bson.A{
		model.Device{
			ID: model.DeviceID(oid.NewUUIDv5("1").String()),
			Attributes: model.DeviceAttributes{{
				Name:  "foo",
				Scope: "bar",
				Value: "baz",
			}},
		},
		model.Device{
			ID:    model.DeviceID(oid.NewUUIDv5("2").String()),
			Group: "foo",
		},
		model.Device{
			ID:    model.DeviceID(oid.NewUUIDv5("3").String()),
			Group: "bar",
		},
		model.Device{
			ID:    model.DeviceID(oid.NewUUIDv5("4").String()),
			Group: "baz",
		},
	}

	testCases := []struct {
		Name string

		Tenant    string
		DeviceIDs []model.DeviceID
		model.GroupName

		Result     model.UpdateResult
		MongoError bool
	}{{
		Name: "ok, all matched updated",

		DeviceIDs: []model.DeviceID{
			model.DeviceID(oid.NewUUIDv5("1").String()),
			model.DeviceID(oid.NewUUIDv5("2").String()),
			model.DeviceID(oid.NewUUIDv5("3").String()),
		},
		GroupName: "baz",
		Result: model.UpdateResult{
			UpdatedCount: 3,
			MatchedCount: 3,
		},
	}, {
		Name: "ok, partial update (tenant)",

		Tenant: oid.NewBSONID().String(),
		DeviceIDs: []model.DeviceID{
			model.DeviceID(oid.NewUUIDv5("1").String()),
			model.DeviceID(oid.NewUUIDv5("2").String()),
			model.DeviceID(oid.NewUUIDv5("3").String()),
			model.DeviceID(oid.NewUUIDv5("4").String()),
			model.DeviceID(oid.NewUUIDv5("5").String()),
		},
		GroupName: "baz",
		Result: model.UpdateResult{
			UpdatedCount: 3,
			MatchedCount: 4,
		},
	}, {
		Name: "ok, no match",

		DeviceIDs: []model.DeviceID{
			model.DeviceID(oid.NewUUIDv5("10").String()),
			model.DeviceID(oid.NewUUIDv5("11").String()),
			model.DeviceID(oid.NewUUIDv5("12").String()),
			model.DeviceID(oid.NewUUIDv5("13").String()),
		},
		GroupName: "foo",
		Result: model.UpdateResult{
			UpdatedCount: 0,
			MatchedCount: 0,
		},
	}, {
		Name: "error, nil array - noop",

		DeviceIDs: nil,
		GroupName: "foo",
		Result:    model.UpdateResult{},
	}}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db.Wipe()
			ctx := context.Background()
			client := db.Client()
			store := NewDataStoreMongoWithSession(client)
			if testCase.Tenant != "" {
				ctx = identity.WithContext(
					ctx, &identity.Identity{
						Tenant: testCase.Tenant,
					},
				)
			}
			collDevs := client.
				Database(mstore.DbFromContext(ctx, DbName)).
				Collection(DbDevicesColl)
			if _, err := collDevs.InsertMany(ctx, deviceSet); err != nil {
				t.Fatalf("Failed to initialize test context, error: %v", err)
			}
			result, err := store.UpdateDevicesGroup(
				ctx, testCase.DeviceIDs, testCase.GroupName,
			)

			if testCase.MongoError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if assert.NotNil(t, result) {
					assert.Equal(t, testCase.Result, *result)
				}
			}
		})
	}
}

func TestClearDevicesGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestClearDevicesGroup in short mode.")
	}

	deviceSet := bson.A{
		model.Device{
			ID:    model.DeviceID(oid.NewUUIDv5("1").String()),
			Group: "foo",
		},
		model.Device{
			ID:    model.DeviceID(oid.NewUUIDv5("2").String()),
			Group: "foo",
		},
		model.Device{
			ID:    model.DeviceID(oid.NewUUIDv5("3").String()),
			Group: "bar",
		},
		model.Device{
			ID:    model.DeviceID(oid.NewUUIDv5("4").String()),
			Group: "baz",
		},
	}

	testCases := []struct {
		Name string

		Tenant    string
		DeviceIDs []model.DeviceID
		model.GroupName

		Result     model.UpdateResult
		MongoError bool
	}{{
		Name: "ok, all matched updated",

		DeviceIDs: []model.DeviceID{
			model.DeviceID(oid.NewUUIDv5("1").String()),
			model.DeviceID(oid.NewUUIDv5("2").String()),
		},
		GroupName: "foo",
		Result: model.UpdateResult{
			MatchedCount: 2,
			UpdatedCount: 2,
		},
	}, {
		Name: "ok, partial update (tenant)",

		Tenant: oid.NewBSONID().String(),
		DeviceIDs: []model.DeviceID{
			model.DeviceID(oid.NewUUIDv5("1").String()),
			model.DeviceID(oid.NewUUIDv5("2").String()),
			model.DeviceID(oid.NewUUIDv5("3").String()),
			model.DeviceID(oid.NewUUIDv5("4").String()),
		},
		GroupName: "baz",
		Result: model.UpdateResult{
			MatchedCount: 1,
			UpdatedCount: 1,
		},
	}, {
		Name: "ok, no match",

		DeviceIDs: []model.DeviceID{
			model.DeviceID(oid.NewUUIDv5("10").String()),
			model.DeviceID(oid.NewUUIDv5("11").String()),
			model.DeviceID(oid.NewUUIDv5("12").String()),
			model.DeviceID(oid.NewUUIDv5("13").String()),
		},
		GroupName: "foo",
		Result:    model.UpdateResult{},
	}, {
		Name: "empty input array - noop",

		DeviceIDs: nil,
		GroupName: "foo",
		Result:    model.UpdateResult{},
	}}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db.Wipe()
			ctx := context.Background()
			client := db.Client()
			store := NewDataStoreMongoWithSession(client)
			if testCase.Tenant != "" {
				ctx = identity.WithContext(
					ctx, &identity.Identity{
						Tenant: testCase.Tenant,
					},
				)
			}
			collDevs := client.
				Database(mstore.DbFromContext(ctx, DbName)).
				Collection(DbDevicesColl)
			if _, err := collDevs.InsertMany(ctx, deviceSet); err != nil {
				t.Fatalf("Failed to initialize test context, error: %v", err)
			}
			updated, err := store.UnsetDevicesGroup(
				ctx, testCase.DeviceIDs, testCase.GroupName,
			)

			if testCase.MongoError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if assert.NotNil(t, updated) {
					assert.Equal(t, testCase.Result, *updated)
				}
			}
		})
	}
}

func TestWithAutomigrate(t *testing.T) {
	db.Wipe()

	client := db.Client()

	store := NewDataStoreMongoWithSession(client)

	newStore := store.WithAutomigrate()

	assert.NotEqual(t, store, newStore)
}

func TestMongoUpsertDevicesAttributesWithRevision(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMongoUpsertDevicesAttributesWithRevision in short mode.")
	}

	//single create timestamp for all inserted devs
	createdTs := time.Now()

	testCases := map[string]struct {
		devs []model.Device

		inDevs  []model.DeviceUpdate
		inAttrs model.DeviceAttributes

		tenant string

		outDevs []model.Device
		err     error
	}{
		"dev exists, attributes exist, update both attrs (descr + val)": {
			devs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:        "mac",
							Value:       "0003-mac",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
						{
							Name:        "sn",
							Value:       "0003-sn",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
					},
					CreatedTs: createdTs,
				},
			},
			inDevs: []model.DeviceUpdate{{Id: model.DeviceID("0003"), Revision: 1}},
			inAttrs: model.DeviceAttributes{
				{
					Description: strPtr("mac description"),
					Scope:       model.AttrScopeInventory,
					Name:        "mac",
					Value:       "0003-newmac",
				},
				{
					Description: strPtr("sn description"),
					Scope:       model.AttrScopeInventory,
					Name:        "sn",
					Value:       "0003-newsn",
				},
			},

			outDevs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Description: strPtr("mac description"),
							Scope:       model.AttrScopeInventory,
							Name:        "mac",
							Value:       "0003-newmac",
						},
						{
							Description: strPtr("sn description"),
							Scope:       model.AttrScopeInventory,
							Name:        "sn",
							Value:       "0003-newsn",
						},
					},
					CreatedTs: createdTs,
				},
			},
		},
		"dev exists, attributes exist, same revision": {
			devs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:        "mac",
							Value:       "0003-mac",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
						{
							Name:        "sn",
							Value:       "0003-sn",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
					},
					CreatedTs: createdTs,
				},
			},
			inDevs: []model.DeviceUpdate{{Id: model.DeviceID("0003"), Revision: 0}},
			inAttrs: model.DeviceAttributes{
				{
					Description: strPtr("mac description"),
					Scope:       model.AttrScopeInventory,
					Name:        "mac",
					Value:       "0003-newmac",
				},
				{
					Description: strPtr("sn description"),
					Scope:       model.AttrScopeInventory,
					Name:        "sn",
					Value:       "0003-newsn",
				},
			},

			outDevs: []model.Device{
				{
					ID: model.DeviceID("0003"),
					Attributes: model.DeviceAttributes{
						{
							Name:        "mac",
							Value:       "0003-mac",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
						{
							Name:        "sn",
							Value:       "0003-sn",
							Description: strPtr("descr"),
							Scope:       model.AttrScopeInventory,
						},
					},
					CreatedTs: createdTs,
				},
			},

			err: store.ErrWriteConflict,
		},
		"Update multiple device attributes": {
			devs: []model.Device{{
				ID:        model.DeviceID("0003"),
				CreatedTs: createdTs,
			}, {
				ID: model.DeviceID("0004"),
				Attributes: model.DeviceAttributes{{
					Scope: model.AttrScopeInventory,
					Name:  "Artifact name",
					Value: "acmeware2.0",
				}},
				CreatedTs: createdTs,
			}},
			inDevs: []model.DeviceUpdate{
				{Id: model.DeviceID("0003"), Revision: 1},
				{Id: model.DeviceID("0004"), Revision: 0},
				{Id: model.DeviceID("0005"), Revision: 1},
			},
			inAttrs: model.DeviceAttributes{{
				Scope:       model.AttrScopeIdentity,
				Name:        "status",
				Value:       "accepted",
				Description: strPtr("deviceauth status"),
			}},
			outDevs: []model.Device{{
				ID: model.DeviceID("0003"),
				Attributes: model.DeviceAttributes{{
					Scope:       model.AttrScopeIdentity,
					Name:        "status",
					Value:       "accepted",
					Description: strPtr("deviceauth status"),
				}},
				CreatedTs: createdTs,
			}, {
				ID: model.DeviceID("0004"),
				Attributes: model.DeviceAttributes{{
					Scope: model.AttrScopeInventory,
					Name:  "Artifact name",
					Value: "acmeware2.0",
				}},
				CreatedTs: createdTs,
			}, {
				ID: model.DeviceID("0005"),
				Attributes: model.DeviceAttributes{{
					Scope:       model.AttrScopeIdentity,
					Name:        "status",
					Value:       "accepted",
					Description: strPtr("deviceauth status"),
				}},
			}},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			//setup
			db.Wipe()

			s := db.Client()

			var ctx context.Context
			if tc.tenant != "" {
				ctx = identity.WithContext(db.CTX(), &identity.Identity{
					Tenant: tc.tenant,
				})
			} else {
				ctx = identity.WithContext(db.CTX(), &identity.Identity{
					Tenant: "",
				})
			}

			//test
			d := NewDataStoreMongoWithSession(s)
			for _, dev := range tc.devs {
				err := d.AddDevice(ctx, &dev)
				assert.NoError(t, err, "failed to setup input data")
			}

			_, err := d.UpsertDevicesAttributesWithRevision(ctx, tc.inDevs, tc.inAttrs)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err, "UpsertDevicesAttributesWithRevision failed")
			}

			//get the device back
			var devs []model.Device
			cur, err := s.Database(DbName).
				Collection(DbDevicesColl).
				Find(
					context.Background(),
					bson.M{},
					mopts.Find().SetSort(bson.M{"_id": 1}),
				)
			if err == nil {
				err = cur.All(nil, &devs)
			}
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			if assert.Len(t, devs, len(tc.outDevs)) {
				for i, dev := range tc.outDevs {
					assert.Equal(t, dev.ID, devs[i].ID)
					compareDevsWithoutTimestamps(t, &dev, &devs[i])
				}
			}
		})
	}
}
