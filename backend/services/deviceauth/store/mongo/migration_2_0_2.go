// Copyright 2026 Northern.tech AS
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	    http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.

package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"
	"github.com/mendersoftware/mender-server/services/deviceauth/model"
)

type migration_2_0_2 struct {
	ds  *DataStoreMongo
	ctx context.Context
}

func (m *migration_2_0_2) Up(from migrate.Version) error {
	ctx := m.ctx
	client := m.ds.client

	limitsCollection := client.Database(DbName).Collection(DbLimitsColl)
	_, err := limitsCollection.UpdateMany(
		ctx,
		bson.D{
			{
				Key:   dbFieldName,
				Value: model.LimitMaxDevicesCount,
			},
			{
				Key:   dbFieldValue,
				Value: 0,
			},
		},
		bson.D{
			{
				Key: "$set",
				Value: bson.M{
					dbFieldValue: int64(-1),
				},
			},
		},
	)

	if err != nil {
		return errors.Wrap(err, "failed to migrate unlimited standard device limits from 0 to -1")
	}
	return nil
}

func (m *migration_2_0_2) Version() migrate.Version {
	return migrate.MakeVersion(2, 0, 2)
}
