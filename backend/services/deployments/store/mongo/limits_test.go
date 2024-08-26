// Copyright 2019 Northern.tech AS
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
	"go.mongodb.org/mongo-driver/bson"

	"github.com/mendersoftware/mender-server/pkg/identity"
	ctxstore "github.com/mendersoftware/mender-server/pkg/store"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

// db and test management funcs
func getDb(ctx context.Context) *DataStoreMongo {
	db.Wipe()

	ds := NewDataStoreMongoWithClient(db.Client())

	return ds
}

func TestGetLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestGetLimit in short mode.")
	}

	lim1 := model.Limit{
		Name:  "foo",
		Value: 123,
	}
	lim2 := model.Limit{
		Name:  "bar",
		Value: 456,
	}
	lim3OtherTenant := model.Limit{
		Name:  "bar",
		Value: 920,
	}

	tenant := "foo"

	dbCtx := identity.WithContext(context.Background(), &identity.Identity{
		Tenant: tenant,
	})
	db := getDb(dbCtx)

	collection := db.client.Database(ctxstore.
		DbFromContext(dbCtx, DatabaseName)).Collection(CollectionLimits)
	_, err := collection.InsertMany(dbCtx,
		bson.A{lim1, lim2})
	assert.NoError(t, err)
	dbCtxOtherTenant := identity.WithContext(context.Background(), &identity.Identity{
		Tenant: "other-" + tenant,
	})
	collOtherTenant := db.client.Database(ctxstore.
		DbFromContext(dbCtxOtherTenant, DatabaseName)).
		Collection(CollectionLimits)
	_, err = collOtherTenant.InsertOne(dbCtxOtherTenant, lim3OtherTenant)
	assert.NoError(t, err)

	// check if value is fetched correctly
	lim, err := db.GetLimit(dbCtx, "foo")
	assert.NoError(t, err)
	assert.EqualValues(t, lim1, *lim)

	// try with something that does not exist
	lim, err = db.GetLimit(dbCtx, "nonexistent-foo")
	assert.EqualError(t, err, ErrLimitNotFound.Error())
	assert.Nil(t, lim)

	// switch tenants
	lim, err = db.GetLimit(dbCtxOtherTenant, "foo")
	assert.EqualError(t, err, ErrLimitNotFound.Error())

	lim, err = db.GetLimit(dbCtxOtherTenant, "bar")
	assert.NoError(t, err)
	assert.EqualValues(t, lim3OtherTenant, *lim)
}
