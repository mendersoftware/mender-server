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

package model

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var (
	// Initialize validation rules once.
	lengthIn0To200  = validation.Length(0, 200)
	lengthIn1To4096 = validation.Length(1, 4096)

	lengthLessThan4096 = validation.Length(0, 4096)
)

type deviceDeploymentStatusValidator struct{}

func (deviceDeploymentStatusValidator) Validate(v interface{}) error {
	stat := v.(DeviceDeploymentStatus)
	_, err := stat.MarshalText()
	return err
}
