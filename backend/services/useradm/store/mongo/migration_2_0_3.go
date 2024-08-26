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
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mopts "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
)

const IndexNameTokenExpire = "exp_time"

type migration_2_0_3 struct {
	ds     *DataStoreMongo
	dbName string
}

func (m *migration_2_0_3) Up(from migrate.Version) error {

	const errCodeIndexNotFound = 27
	var err error

	ctx := context.Background()

	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: DbTokenExpireTime, Value: 1},
		},
		Options: mopts.Index().
			SetExpireAfterSeconds(0).
			SetName(IndexNameTokenExpire),
	}
	if m.dbName == DbName {
		logger := log.FromContext(ctx)
		iw := m.ds.client.Database(DbName).
			Collection(DbTokensColl).
			Indexes()
		// drop existing bad TTL index (if it exists)
		_, err = iw.DropOne(ctx, brokenDbTokenExpirationIndexName)
		if err != nil {
			var srvErr mongo.ServerError
			if !errors.As(err, &srvErr) || !srvErr.HasErrorCode(errCodeIndexNotFound) {
				logger.Warnf("failed to drop index '%s': %s",
					brokenDbTokenExpirationIndexName,
					err.Error())
			}
		}

		_, err = iw.CreateOne(ctx, indexModel)
		if err != nil {
			err = fmt.Errorf("failed to create index '%s': %w", IndexNameTokenExpire, err)
		}
	}

	return err
}

func (m *migration_2_0_3) Version() migrate.Version {
	return migrate.MakeVersion(2, 0, 3)
}
