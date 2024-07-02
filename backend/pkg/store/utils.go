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
	"strings"

	"github.com/mendersoftware/mender-server/pkg/identity"
)

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

type TenantDbMatchFunc func(name string) bool

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
