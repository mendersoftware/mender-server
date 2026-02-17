package mongo

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
