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

//go:generate ../../../utils/mockgen.sh
type Store interface {
	BulkIndexDeployments(ctx context.Context, deployments []*model.Deployment) error
	BulkIndexDevices(ctx context.Context, devices, removedDevices []*model.Device) error
	GetDevicesIndex(tid string) string
	GetDevicesRoutingKey(tid string) string
	GetDevicesIndexMapping(ctx context.Context, tid string) (map[string]interface{}, error)
	GetDeploymentsIndex(tid string) string
	GetDeploymentsRoutingKey(tid string) string
	GetDeploymentsIndexMapping(ctx context.Context, tid string) (map[string]interface{}, error)
	Migrate(ctx context.Context) error
	AggregateDevices(ctx context.Context, query model.Query) (model.M, error)
	AggregateDeployments(ctx context.Context, query model.Query) (model.M, error)
	SearchDevices(ctx context.Context, query model.Query) (model.M, error)
	SearchDeployments(ctx context.Context, query model.Query) (model.M, error)
	Ping(ctx context.Context) error
}
