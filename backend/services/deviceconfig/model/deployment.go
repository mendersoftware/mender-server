// Copyright 2021 Northern.tech AS
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

import "github.com/google/uuid"

type DeployConfigurationRequest struct {
	// Retries represents the number of retries in case of deployment failures
	Retries uint `json:"retries"`

	// Optional update_control_map (Enterprise-only)
	UpdateControlMap map[string]interface{} `json:"update_control_map,omitempty"`
}

type DeployConfigurationResponse struct {
	DeploymentID uuid.UUID `json:"deployment_id"`
}
