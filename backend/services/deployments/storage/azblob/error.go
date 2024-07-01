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

package azblob

import "errors"

type OpError struct {
	Op      string
	Message string
	Reason  error
}

func (err OpError) Error() string {
	errStr := "azblob"
	if err.Op != "" {
		errStr += " " + err.Op
	}
	if err.Message != "" {
		errStr += ": " + err.Message
	}
	if err.Reason != nil {
		errStr += ": " + err.Reason.Error()
	}
	return errStr
}

func (err OpError) Unwrap() error {
	return err.Reason
}

const (
	OpHealthCheck   = "HealthCheck"
	OpGetObject     = "GetObject"
	OpPutObject     = "PutObject"
	OpDeleteObject  = "DeleteObject"
	OpStatObject    = "StatObject"
	OpGetRequest    = "GetRequest"
	OpDeleteRequest = "DeleteRequest"
	OpPutRequest    = "PutRequest"
)

var (
	ErrStorageSettings = errors.New("storage settings invalid")
	ErrEmptyClient     = errors.New("storage client not configured")
)
