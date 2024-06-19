// Copyright 2020 Northern.tech AS
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

	"github.com/mendersoftware/mender-server/pkg/mongo/migrate"
)

type migration_1_0_1 struct {
	ms  *DataStoreMongo
	ctx context.Context
}

var attributesToIndex = []string{
	"identity-mac",
	"inventory-device_type",
	"system-group",
	"system-updated_ts",
}

func (m *migration_1_0_1) Up(from migrate.Version) error {
	for _, key := range attributesToIndex {
		_ = indexAttr(m.ms.client, m.ctx, key)
	}
	return nil
}

func (m *migration_1_0_1) Version() migrate.Version {
	return migrate.MakeVersion(1, 0, 1)
}
