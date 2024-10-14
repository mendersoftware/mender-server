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

package indexer

import (
	"context"

	"github.com/mendersoftware/mender-server/services/reporting/client/deployments"
	"github.com/mendersoftware/mender-server/services/reporting/client/deviceauth"
	"github.com/mendersoftware/mender-server/services/reporting/client/inventory"
	"github.com/mendersoftware/mender-server/services/reporting/client/nats"
	"github.com/mendersoftware/mender-server/services/reporting/mapping"
	"github.com/mendersoftware/mender-server/services/reporting/model"
	"github.com/mendersoftware/mender-server/services/reporting/store"
)

//go:generate ../../../../utils/mockgen.sh
type Indexer interface {
	GetJobs(ctx context.Context, jobs chan model.Job) error
	ProcessJobs(ctx context.Context, jobs []model.Job)
}

type indexer struct {
	store      store.Store
	mapper     mapping.Mapper
	nats       nats.Client
	devClient  deviceauth.Client
	invClient  inventory.Client
	deplClient deployments.Client
}

func NewIndexer(
	store store.Store,
	ds store.DataStore,
	nats nats.Client,
	devClient deviceauth.Client,
	invClient inventory.Client,
	deplClient deployments.Client,
) Indexer {
	mapper := mapping.NewMapper(ds)
	return &indexer{
		store:      store,
		mapper:     mapper,
		nats:       nats,
		devClient:  devClient,
		invClient:  invClient,
		deplClient: deplClient,
	}
}
