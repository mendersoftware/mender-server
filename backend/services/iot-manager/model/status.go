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

import "github.com/pkg/errors"

type Status string

const (
	StatusAccepted       Status = "accepted"
	StatusNoAuth         Status = "noauth"
	StatusPending        Status = "pending"
	StatusPreauthorized  Status = "preauthorized"
	StatusRejected       Status = "rejected"
	StatusDecommissioned Status = "decommissioned"
)

func (stat Status) Validate() error {
	switch stat {
	case StatusAccepted, StatusPending, StatusRejected,
		StatusNoAuth, StatusPreauthorized, StatusDecommissioned:
		return nil
	default:
		return errors.Errorf("invalid status '%s'", stat)
	}
}
