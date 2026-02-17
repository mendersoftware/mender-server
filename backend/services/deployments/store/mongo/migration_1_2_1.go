// Copyright 2023 Northern.tech AS
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

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"
)

var (
	IndexDeploymentArtifactName_0_0_0 = "deploymentconstructor." +
		"name_text_deploymentconstructor.artifactname_text"
)

type migration_1_2_1 struct {
	client *mongo.Client
	db     string
}

// Up drops old index with extremely long name
func (m *migration_1_2_1) Up(from migrate.Version) error {
	ctx := context.Background()
	collDpl := m.client.Database(m.db).Collection(CollectionDeployments)
	indexView := collDpl.Indexes()

	err := indexView.DropOne(ctx, IndexDeploymentArtifactName_0_0_0)
	if err != nil {
		// Supress NamespaceNotFound and IndexNotFound errors
		// - index is missing
		if except, ok := err.(mongo.CommandError); ok {
			if except.Code != errorCodeNamespaceNotFound &&
				except.Code != errorCodeIndexNotFound {
				return err
			}
		} else {
			return err
		}
	}

	// create the 'short' index
	storage := NewDataStoreMongoWithClient(m.client)
	retErr := storage.EnsureIndexes(m.db,
		CollectionDeployments, StorageIndexes)
	return retErr
}

func (m *migration_1_2_1) Version() migrate.Version {
	return migrate.MakeVersion(1, 2, 1)
}
