// Copyright 2024 Northern.tech AS
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

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	mopts "go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"
)

type migration_1_2_16 struct {
	client *mongo.Client
	db     string
}

func (m *migration_1_2_16) Up(from migrate.Version) error {
	ctx := context.Background()
	idxDeployments := m.client.
		Database(m.db).
		Collection(CollectionDeployments).
		Indexes()

	_, err := idxDeployments.CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{
				Key:   StorageKeyDeploymentConstructorChecksum,
				Value: 1,
			},
		},
		Options: mopts.Index().
			SetName(IndexNameDeploymentConstructorChecksum).
			SetPartialFilterExpression(
				bson.D{
					{
						Key: StorageKeyDeploymentConstructorChecksum,
						Value: bson.D{
							{
								Key:   "$exists",
								Value: true,
							},
						},
					},
					{
						Key:   StorageKeyDeploymentActive,
						Value: true,
					},
				},
			).
			SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("mongo(1.2.16): failed to create index: %w", err)
	}
	return nil
}

func (m *migration_1_2_16) Version() migrate.Version {
	return migrate.MakeVersion(1, 2, 16)
}
