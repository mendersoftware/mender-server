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
package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/mender-server/pkg/identity"
)

func TestDbFromContextEmptyContext(t *testing.T) {
	db := DbFromContext(context.Background(), "foo")
	assert.Equal(t, db, "foo")
}

func TestDbFromContextNoTenant(t *testing.T) {
	ctx := context.Background()
	id := identity.Identity{
		Subject: "subject",
	}
	db := DbFromContext(identity.WithContext(ctx, &id), "foo")
	assert.Equal(t, db, "foo")
}

func TestDbFromContext(t *testing.T) {
	ctx := context.Background()
	id := identity.Identity{
		Subject: "subject",
		Tenant:  "bar",
	}
	db := DbFromContext(identity.WithContext(ctx, &id), "foo")
	assert.Equal(t, db, "foo")
}

func TestIsTenantDb(t *testing.T) {
	matcher := IsTenantDb("servicedb")

	assert.True(t, matcher("servicedb-tenant1"))
	assert.False(t, matcher("servicedb"))
	assert.False(t, matcher("servicedbtenant1"))

}

func TestTenantFromDbName(t *testing.T) {
	assert.Equal(t, "tenant1", TenantFromDbName("ser-vice_dev-adm-tenant1", "ser-vice_dev-adm"))
	assert.Equal(t, "", TenantFromDbName("-tenant1", "service_devadm"))
	assert.Equal(t, "", TenantFromDbName("service_devadm", "service_devadm"))
	assert.Equal(t, "198273913adsjhakdh",
		TenantFromDbName("123__--afff-198273913adsjhakdh", "123__--afff"))
}

func TestDbNameForTenant(t *testing.T) {
	assert.Equal(t, "basedb", DbNameForTenant("tenant1", "basedb"))
	assert.Equal(t, "basedb", DbNameForTenant("", "basedb"))
}
