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

package rest

import (
	"encoding/json"
	"errors"
	"io"
)

type Error struct {
	Err       string `json:"error"`
	RequestID string `json:"request_id,omitempty"`
}

func (err Error) Error() string {
	return err.Err
}

func IsApiError(e error) bool {
	var target *Error
	return errors.As(e, &target)
}

func ParseApiError(source io.Reader) error {
	jd := json.NewDecoder(source)

	var aerr Error
	if err := jd.Decode(&aerr); err != nil {
		return err
	}

	return &aerr
}
