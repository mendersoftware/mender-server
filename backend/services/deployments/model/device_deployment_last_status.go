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

package model

type DeviceDeploymentLastStatus struct {
	// Device id
	DeviceId string `json:"device_id" bson:"_id"`

	// Deployment id
	DeploymentId string `json:"deployment_id" bson:"deployment_id"`

	// Device Deployment id
	DeviceDeploymentId string `json:"device_deployment_id" bson:"device_deployment_id"`

	// Status
	// nolint:lll
	DeviceDeploymentStatus DeviceDeploymentStatus `json:"device_deployment_status" bson:"device_deployment_status"`

	// Tenant id
	TenantId string `json:"-" bson:"tenant_id"`
}

type DeviceDeploymentLastStatuses struct {
	// DeviceDeploymentLastStatuses array of last device deployments statuses
	// nolint:lll
	DeviceDeploymentLastStatuses []DeviceDeploymentLastStatus `json:"device_deployment_last_statuses" bson:"device_deployment_last_statuses"`
}

type DeviceDeploymentLastStatusReq struct {
	// Device ids
	DeviceIds []string `json:"device_ids"`
}
