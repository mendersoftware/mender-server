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
	"testing"

	"github.com/mendersoftware/mender-server/services/reporting/model"
	"github.com/stretchr/testify/assert"
)

func TestGroupJobsIntoTenantActionIDs(t *testing.T) {
	jobs := []model.Job{
		{
			Action:   model.ActionReindex,
			TenantID: "t1",
			DeviceID: "d1",
			Service:  model.ServiceInventory,
		},
		{
			Action:   model.ActionReindex,
			TenantID: "t1",
			DeviceID: "d1",
			Service:  model.ServiceDeviceauth,
		},
		{
			Action:   model.ActionReindex,
			TenantID: "t1",
			DeviceID: "d2",
			Service:  model.ServiceInventory,
		},
		{
			Action:   model.ActionReindex,
			TenantID: "t2",
			DeviceID: "d1",
			Service:  model.ServiceInventory,
		},
		{
			Action:   model.ActionReindexDeployment,
			TenantID: "t2",
			ID:       "d1",
			Service:  model.ServiceInventory,
		},
		{
			Action:   model.ActionReindexDeployment,
			TenantID: "t2",
			ID:       "d2",
			Service:  model.ServiceInventory,
		},
	}

	tenantActionIDs := groupJobsIntoTenantActionIDs(jobs)
	expected := TenantActionIDs{
		"t1": ActionIDs{
			model.ActionReindex: {
				"d1": true,
				"d2": true,
			},
		},
		"t2": ActionIDs{
			model.ActionReindex: {
				"d1": true,
			},
			model.ActionReindexDeployment: {
				"d1": true,
				"d2": true,
			},
		},
	}

	assert.Equal(t, expected, tenantActionIDs)
}
