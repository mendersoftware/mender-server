// Copyright 2021 Northern.tech AS
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

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"
	mstore "github.com/mendersoftware/mender-server/pkg/store"
)

type migration_0_2_0 struct {
	ms  *DataStoreMongo
	ctx context.Context
}

func (m *migration_0_2_0) Up(from migrate.Version) error {
	// get all attribute names
	names, err := m.ms.GetAllAttributeNames(m.ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get attribute names")
	}

	// rename every attribute occurrence to scoped version, add scope field
	// hacky - we're doing it in two runs, but dead simple
	databaseName := mstore.DbFromContext(m.ctx, DbName)
	coll := m.ms.client.Database(databaseName).Collection(DbDevicesColl)
	for _, n := range names {
		nold := fmt.Sprintf("%s.%s", DbDevAttributes, n)
		nnew := fmt.Sprintf("%s.%s-%s", DbDevAttributes, DbScopeInventory, n)

		_, err := coll.UpdateMany(m.ctx, bson.M{}, bson.M{"$rename": bson.M{nold: nnew}})
		if err != nil {
			return errors.Wrapf(err, "failed to update attribute name %s to %s", nold, nnew)
		}

		// get all docs containing a given attribute
		scope := fmt.Sprintf("%s.%s", nnew, DbDevAttributesScope)
		_, err = coll.UpdateMany(
			m.ctx,
			bson.M{nnew: bson.M{"$exists": true}},
			bson.M{"$set": bson.M{scope: DbScopeInventory}},
		)
		if err != nil {
			return errors.Wrapf(err, "failed to update scope for attribute name %s", nold)
		}
	}

	return nil
}

func (m *migration_0_2_0) Version() migrate.Version {
	return migrate.MakeVersion(0, 2, 0)
}
