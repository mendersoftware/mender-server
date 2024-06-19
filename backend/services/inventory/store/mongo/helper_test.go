// Copyright 2020 Northern.tech AS
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
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	mstore "github.com/mendersoftware/mender-server/pkg/store"
	"github.com/mendersoftware/mender-server/services/inventory/model"
)

func DeviceFindById(ctx context.Context, c *mongo.Collection, id model.DeviceID, dst *model.Device) error {
	cur, err := c.Find(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	cur.Next(ctx)

	// create a value into which the single document can be decoded
	err = cur.Decode(dst)
	if err != nil {
		return err
	}

	return nil
}

// test funcs
func TestMongoDeviceFindById(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMongoGetDevice in short mode.")
	}

	testCases := map[string]struct {
		InputID     model.DeviceID
		InputDevice *model.Device
		tenant      string
		OutputError error
	}{
		"device with given ID exists, no error": {
			InputID: model.DeviceID("0002"),
			InputDevice: &model.Device{
				ID: model.DeviceID("0002"),
				Attributes: model.DeviceAttributes{
					{Name: "mac", Value: "0002-mac", Scope: "inventory"},
				},
			},
			OutputError: nil,
		},
		"device with given ID does not exist": {
			InputID:     model.DeviceID("123"),
			InputDevice: nil,
			OutputError: errors.New("EOF"),
		},
	}

	for name, testCase := range testCases {
		t.Logf("test case: %s", name)

		// Make sure we start test with empty database
		db.Wipe()

		client := db.Client()

		if testCase.InputDevice != nil {
			_, _ = client.Database(mstore.DbFromContext(db.CTX(), DbName)).Collection(DbDevicesColl).InsertOne(db.CTX(), testCase.InputDevice)
		}

		var dbdev model.Device
		err := DeviceFindById(db.CTX(), client.Database(mstore.DbFromContext(db.CTX(), DbName)).Collection(DbDevicesColl), testCase.InputID, &dbdev)

		if testCase.InputDevice != nil {
			assert.NoError(t, err, "error getting device")
			assert.NotNil(t, dbdev, "expected to device of ID %s to be found", testCase.InputDevice.ID)
			assert.Equal(t, testCase.InputID, dbdev.ID)
		} else {
			assert.EqualError(t, err, testCase.OutputError.Error())
		}
	}
}
