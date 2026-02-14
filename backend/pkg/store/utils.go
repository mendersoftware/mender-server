// Copyright 2023 Northern.tech AS
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

package store

import (
	"context"
	"strings"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/mendersoftware/mender-server/pkg/identity"
	mdoc "github.com/mendersoftware/mender-server/pkg/mongo/doc"
)

const FieldTenantID = "tenant_id"

type TenantDbMatchFunc func(name string) bool

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

// DbFromContext generates database name using tenant field from identity extracted
// from context and original database name
func DbFromContext(ctx context.Context, origDbName string) string {
	identity := identity.FromContext(ctx)
	tenant := ""
	if identity != nil {
		tenant = identity.Tenant
	}

	return DbNameForTenant(tenant, origDbName)
}

// IsTenantDb returns a function of `TenantDbMatchFunc` that can be used for
// checking if database has a tenant DB name format
func IsTenantDb(baseDb string) TenantDbMatchFunc {
	prefix := baseDb + "-"
	return func(name string) bool {
		return strings.HasPrefix(name, prefix)
	}
}

// TenantFromDbName attempts to extract tenant ID from provided tenant DB name.
// Returns extracted tenant ID or an empty string.
func TenantFromDbName(dbName string, baseDb string) string {
	noBase := strings.TrimPrefix(dbName, baseDb+"-")
	if noBase == dbName {
		return ""
	}
	return noBase
}

// DbNameForTenant composes tenant's db name.
func DbNameForTenant(tenantId string, baseDb string) string {
	if tenantId == "" {
		return baseDb
	}
	return baseDb + "-" + tenantId
}
