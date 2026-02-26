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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"
	mstore "github.com/mendersoftware/mender-server/pkg/store"
)

func TestMigration_1_2_16(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMigration_1_2_16 in short mode.")
	}

	db.Wipe()
	c := db.Client()

	ctx := context.TODO()

	//store := NewDataStoreMongoWithClient(c)
	database := c.Database(mstore.DbFromContext(ctx, DatabaseName))
	collRel := database.Collection(CollectionDeployments)

	// apply migration (1.2.16)
	mnew := &migration_1_2_16{
		client: c,
		db:     DbName,
	}
	err := mnew.Up(migrate.MakeVersion(1, 2, 16))
	assert.NoError(t, err)

	indices := collRel.Indexes()
	exists, err := hasIndex(ctx, IndexNameDeploymentConstructorChecksum, indices)
	assert.NoError(t, err)
	assert.True(t, exists, "index "+IndexNameReleaseTags+" must exist in 1.2.16")
}
