// Copyright 2022 Northern.tech AS
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

	"github.com/mendersoftware/mender-server/services/reporting/model"
)

// DataStore interface for DataStore services
//
//nolint:lll - skip line length check for interface declaration.
//go:generate ../x/mockgen.sh
type DataStore interface {
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
	DropDatabase(ctx context.Context) error
	Migrate(ctx context.Context, version string, automigrate bool) error
	MigrateLatest(ctx context.Context) error
	GetMapping(ctx context.Context, tenantID string) (*model.Mapping, error)
	UpdateAndGetMapping(ctx context.Context, tenantID string, inventory []string) (
		*model.Mapping, error)
}
