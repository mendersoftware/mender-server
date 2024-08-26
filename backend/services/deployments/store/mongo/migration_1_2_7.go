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
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
)

type migration_1_2_7 struct {
	client *mongo.Client
	db     string
}

func (m *migration_1_2_7) Up(from migrate.Version) error {
	storage := NewDataStoreMongoWithClient(m.client)
	return storage.EnsureIndexes(m.db, CollectionImages,
		IndexImageMetaDescriptionModel,
		IndexImageMetaArtifactDeviceTypeCompatibleModel,
	)
}

func (m *migration_1_2_7) Version() migrate.Version {
	return migrate.MakeVersion(1, 2, 7)
}
