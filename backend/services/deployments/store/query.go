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
	"errors"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

type ListQuery struct {
	Skip         int
	Limit        int
	DeploymentID string
	Status       *string
}

func (l ListQuery) Validate() error {
	if l.Limit <= 0 {
		return errors.New("limit: must be a positive integer")
	}
	if l.DeploymentID == "" {
		return errors.New("deployment_id: cannot be blank")
	}
	if l.Status != nil {
		if *l.Status == model.DeviceDeploymentStatusPauseStr ||
			*l.Status == model.DeviceDeploymentStatusActiveStr ||
			*l.Status == model.DeviceDeploymentStatusFinishedStr {
			return nil
		}
		stat := model.NewStatus(*l.Status)
		if stat == model.DeviceDeploymentStatusNull {
			return errors.New("status: must be a valid value")
		}
	}
	return nil
}
