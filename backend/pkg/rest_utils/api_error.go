// Copyright 2023 Northern.tech AS
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	    http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.
package rest_utils

import (
	"encoding/json"
	"io"
)

// ApiError wraps errors returned by our APIs
type ApiError struct {
	Err   string `json:"error"`
	ReqId string `json:"request_id,omitempty"`
}

func (ae *ApiError) Error() string {
	return ae.Err
}

func IsApiError(e error) bool {
	_, ok := e.(*ApiError)
	return ok
}

func ParseApiError(source io.Reader) error {
	jd := json.NewDecoder(source)

	var aerr ApiError
	if err := jd.Decode(&aerr); err != nil {
		return err
	}

	return &aerr
}
