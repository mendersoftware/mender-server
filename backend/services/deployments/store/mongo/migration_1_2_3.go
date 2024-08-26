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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/mendersoftware/mender-server/pkg/mongo/doc"
	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

type migration_1_2_3 struct {
	client *mongo.Client
	db     string
}

// Up intrduces a unique index on artifact depends_idx and name to ensure unique depends in a
// release, also:
// - drops index on DeviceTypesCompatible, superseded by the above
// - rewrites DeviceTypesCompatible to 'depends.device_type' - even for v1, v2 artifacts
func (m *migration_1_2_3) Up(from migrate.Version) error {
	ctx := context.Background()
	c := m.client.Database(m.db).Collection(CollectionImages)

	// drop old device type + name index
	_, err := c.Indexes().DropOne(ctx, IndexUniqueNameAndDeviceTypeName)

	// the index might not be there - was created only on image inserts (not upfront)
	if err != nil {
		if except, ok := err.(mongo.CommandError); ok {
			if except.Code != errorCodeNamespaceNotFound &&
				except.Code != errorCodeIndexNotFound {
				return err
			}
			// continue
		} else {
			return err
		}
	}

	// transform existing device_types_compatible in v1 and v2 artifacts into 'depends.device_type'
	cursor, err := c.Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	var artifacts []*model.Image
	if err = cursor.All(ctx, &artifacts); err != nil {
		return err
	}

	for _, a := range artifacts {
		// storage.Update is broken, do it manually via the driver
		dtypes := make([]interface{}, len(a.ArtifactMeta.DeviceTypesCompatible))

		for i, d := range a.ArtifactMeta.DeviceTypesCompatible {
			dtypes[i] = interface{}(d)
		}

		depends := bson.M{
			ArtifactDependsDeviceType: dtypes,
		}

		dependsIdx, err := doc.UnwindMap(depends)
		if err != nil {
			return err
		}

		up := bson.M{
			"$set": bson.M{
				StorageKeyImageDepends:    depends,
				StorageKeyImageDependsIdx: dependsIdx,
			},
		}

		res, err := c.UpdateOne(ctx, bson.M{"_id": a.Id}, up)
		if err != nil {
			return err
		}

		if res.MatchedCount != 1 {
			return errors.New(fmt.Sprintf("failed to update artifact %s: not found", a.Id))
		}
	}

	// create new artifact depends + name index
	storage := NewDataStoreMongoWithClient(m.client)
	err = storage.EnsureIndexes(m.db,
		CollectionImages,
		IndexArtifactNameDepends)

	return err
}

func (m *migration_1_2_3) Version() migrate.Version {
	return migrate.MakeVersion(1, 2, 3)
}
