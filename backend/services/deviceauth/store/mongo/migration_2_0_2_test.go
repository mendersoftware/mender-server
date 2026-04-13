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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"
	"github.com/mendersoftware/mender-server/services/deviceauth/model"
)

func TestMigration_2_0_2(t *testing.T) {
	db.Wipe()
	client := db.Client()
	ds := NewDataStoreMongoWithClient(client)

	ctx1 := identity.WithContext(context.Background(), &identity.Identity{
		Tenant: bson.NewObjectID().Hex(),
	})

	err := ds.PutLimit(ctx1, model.Limit{Name: model.LimitMaxDevicesCount, Value: 0}) // Should change
	require.NoError(t, err, "unexpected error when creating limits")

	ctx2 := identity.WithContext(context.Background(), &identity.Identity{
		Tenant: bson.NewObjectID().Hex(),
	})

	err = ds.PutLimit(ctx2, model.Limit{Name: model.LimitMaxDevicesCount, Value: 123}) // Should not change
	require.NoError(t, err, "unexpected error when creating limits")

	migration := migration_2_0_2{
		ds:  ds,
		ctx: context.Background(),
	}

	err = migration.Up(migrate.MakeVersion(2, 0, 1))
	require.NoError(t, err, "unexpected error when running migration")

	lim, err := ds.GetLimit(ctx1, model.LimitMaxDevicesCount)
	require.NoError(t, err, "unexpected error when reading limit")
	assert.Equal(t, int64(-1), lim.Value)

	lim, err = ds.GetLimit(ctx2, model.LimitMaxDevicesCount)
	require.NoError(t, err, "unexpected error when reading limit")
	assert.Equal(t, int64(123), lim.Value)
}
