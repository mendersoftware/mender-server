package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/mendersoftware/mender-server/pkg/identity"

	mdoc "github.com/mendersoftware/mender-server/pkg/mongo/v2/doc"
)

const FieldTenantID = "tenant_id"

// WithTenantID adds the tenant_id field to a bson document using the value extracted
// from the identity of the context
func WithTenantID(ctx context.Context, doc interface{}) bson.D {
	var (
		tenantID string
		res      bson.D
	)

	identity := identity.FromContext(ctx)
	if identity != nil {
		tenantID = identity.Tenant
	}
	tenantElem := bson.E{Key: FieldTenantID, Value: tenantID}

	switch v := doc.(type) {
	case map[string]interface{}:
		res = make(bson.D, 0, len(v)+1)
		for k, v := range v {
			res = append(res, bson.E{Key: k, Value: v})
		}
	case bson.M:
		res = make(bson.D, 0, len(v)+1)
		for k, v := range v {
			res = append(res, bson.E{Key: k, Value: v})
		}
	case bson.D:
		res = make(bson.D, len(v), len(v)+1)
		copy(res, v)

	case bson.Marshaler:
		b, err := v.MarshalBSON()
		if err != nil {
			return nil
		}
		err = bson.Unmarshal(b, &res)
		if err != nil {
			return nil
		}
	default:
		return mdoc.DocumentFromStruct(v, tenantElem)
	}
	res = append(res, tenantElem)

	return res
}

// ArrayWithTenantID adds the tenant_id field to an array of bson documents
// using the value extracted from the identity of the context
func ArrayWithTenantID(ctx context.Context, doc bson.A) bson.A {
	res := bson.A{}
	for _, item := range doc {
		res = append(res, WithTenantID(ctx, item))
	}
	return res
}
