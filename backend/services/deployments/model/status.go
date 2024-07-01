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

import (
	"encoding/json"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

var (
	ErrBadStatus = errors.New("unknown status value")
)

type StatusReport struct {
	Status   DeviceDeploymentStatus `json:"status"`
	SubState string                 `json:"substate"`
}

func (s StatusReport) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.SubState, lengthIn0To200),
		validation.Field(&s.Status, validation.In(
			DeviceDeploymentStatusDownloading,
			DeviceDeploymentStatusInstalling,
			DeviceDeploymentStatusRebooting,
			DeviceDeploymentStatusSuccess,
			DeviceDeploymentStatusFailure,
			DeviceDeploymentStatusAlreadyInst,
			DeviceDeploymentStatusPauseBeforeInstall,
			DeviceDeploymentStatusPauseBeforeCommit,
			DeviceDeploymentStatusPauseBeforeReboot,
		)),
	)
}

func (s *StatusReport) UnmarshalJSON(raw []byte) error {
	type statusReport StatusReport
	err := json.Unmarshal(raw, (*statusReport)(s))
	if err != nil {
		return err
	}

	if err := s.Validate(); err != nil {
		return err
	}

	return nil
}
