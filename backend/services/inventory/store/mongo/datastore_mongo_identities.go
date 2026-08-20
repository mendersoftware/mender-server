// Copyright 2026 Northern.tech AS
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

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"

	mstore "github.com/mendersoftware/mender-server/pkg/store"

	"github.com/mendersoftware/mender-server/services/inventory/model"
)

// SetIdentity sets given identities for a device
func (db *DataStoreMongo) SetIdentity(
	ctx context.Context,
	deviceId model.DeviceID,
	identities []any,
) error {
	c := db.client.
		Database(mstore.DbFromContext(ctx, DbName)).
		Collection(DbDevicesColl)
	_, err := c.UpdateByID(
		ctx,
		deviceId,
		bson.M{
			"$set": bson.M{
				"identities": identities,
			},
		},
	)
	if err != nil {
		return errors.Wrap(err, "failed to update identities")
	}

	return nil
}
