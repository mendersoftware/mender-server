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

import "github.com/mendersoftware/mender-server/services/reporting/model"

func groupJobsIntoTenantActionIDs(jobs []model.Job) TenantActionIDs {
	tenantsActionIDs := make(TenantActionIDs)
	for _, job := range jobs {
		if _, ok := tenantsActionIDs[job.TenantID]; !ok {
			tenantsActionIDs[job.TenantID] = make(ActionIDs)
		}
		if _, ok := tenantsActionIDs[job.TenantID][job.Action]; !ok {
			tenantsActionIDs[job.TenantID][job.Action] = make(IDs)
		}
		var ID string
		if job.Action == model.ActionReindex {
			ID = job.DeviceID
		} else if job.Action == model.ActionReindexDeployment {
			ID = job.ID
		}
		if _, ok := tenantsActionIDs[job.TenantID][job.Action][ID]; !ok {
			tenantsActionIDs[job.TenantID][job.Action][ID] = true
		}
	}
	return tenantsActionIDs
}
