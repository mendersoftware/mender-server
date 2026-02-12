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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/mendersoftware/mender-server/pkg/identity"
)

type SampleObject struct {
	Attribute string `json:"attribute" bson:"attribute"`
}

type SampleMarshalerObject struct {
	Attribute string
}

func (s SampleMarshalerObject) MarshalBSON() ([]byte, error) {
	m := map[string]string{}
	m["attribute"] = s.Attribute
	return bson.Marshal(m)
}

type SampleBadMarshalerObject struct{ Foo bool }

func (SampleBadMarshalerObject) MarshalBSON() ([]byte, error) {
	return nil, errors.New("dunno")
}

type SampleBadMarshalerObject2 struct{ Foo bool }

func (SampleBadMarshalerObject2) MarshalBSON() ([]byte, error) {
	return []byte("this is an invalid BSON type"), nil
}

func TestWithTenantID(t *testing.T) {
	ctx := context.Background()

	sample := &SampleObject{Attribute: "value"}
	sample2 := SampleMarshalerObject{Attribute: "val"}
	sampleBad := SampleBadMarshalerObject{}
	sampleBad2 := SampleBadMarshalerObject2{}

	// without tenant ID
	res := WithTenantID(ctx, map[string]interface{}{"key": "value"})
	assert.Equal(t, bson.D{{Key: "key", Value: "value"}, {Key: FieldTenantID, Value: ""}}, res)

	res = WithTenantID(ctx, bson.M{"key": "value"})
	assert.Equal(t, bson.D{{Key: "key", Value: "value"}, {Key: FieldTenantID, Value: ""}}, res)

	res = WithTenantID(ctx, bson.D{{Key: "key", Value: "value"}})
	assert.Equal(t, bson.D{{Key: "key", Value: "value"}, {Key: FieldTenantID, Value: ""}}, res)

	res = WithTenantID(ctx, sample)
	assert.Equal(t, bson.D{{Key: "attribute", Value: "value"}, {Key: FieldTenantID, Value: ""}}, res)

	res = WithTenantID(ctx, sample2)
	assert.Equal(t, bson.D{{Key: "attribute", Value: "val"}, {Key: FieldTenantID, Value: ""}}, res)

	res = WithTenantID(ctx, sampleBad)
	assert.Nil(t, res)

	res = WithTenantID(ctx, sampleBad2)
	assert.Nil(t, res)

	res = WithTenantID(ctx, "dummy-value")
	assert.Nil(t, res)

	// with tenant ID
	const tenantID = "bar"
	id := &identity.Identity{
		Subject: "subject",
		Tenant:  tenantID,
	}
	ctx = identity.WithContext(ctx, id)

	res = WithTenantID(ctx, map[string]interface{}{"key": "value"})
	assert.Equal(t, bson.D{{Key: "key", Value: "value"}, {Key: FieldTenantID, Value: tenantID}}, res)

	res = WithTenantID(ctx, bson.M{"key": "value"})
	assert.Equal(t, bson.D{{Key: "key", Value: "value"}, {Key: FieldTenantID, Value: tenantID}}, res)

	res = WithTenantID(ctx, bson.D{{Key: "key", Value: "value"}})
	assert.Equal(t, bson.D{{Key: "key", Value: "value"}, {Key: FieldTenantID, Value: tenantID}}, res)

	res = WithTenantID(ctx, sample)
	assert.Equal(t, bson.D{{Key: "attribute", Value: "value"}, {Key: FieldTenantID, Value: tenantID}}, res)

	res = WithTenantID(ctx, "dummy-value")
	assert.Nil(t, res)
}

func TestArrayWithTenantID(t *testing.T) {
	ctx := context.Background()

	// without tenant ID
	res := ArrayWithTenantID(ctx, bson.A{bson.M{"key": "value"}})
	assert.Equal(t, bson.A{bson.D{{Key: "key", Value: "value"}, {Key: FieldTenantID, Value: ""}}}, res)

	// with tenant ID
	const tenantID = "bar"
	id := &identity.Identity{
		Subject: "subject",
		Tenant:  tenantID,
	}
	ctx = identity.WithContext(ctx, id)

	res = ArrayWithTenantID(ctx, bson.A{bson.M{"key": "value"}})
	assert.Equal(t, bson.A{bson.D{{Key: "key", Value: "value"}, {Key: FieldTenantID, Value: tenantID}}}, res)
}

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
	assert.Equal(t, db, "foo-bar")
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
	assert.Equal(t, "basedb-tenant1", DbNameForTenant("tenant1", "basedb"))
	assert.Equal(t, "basedb", DbNameForTenant("", "basedb"))
}
