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

package client

import (
	"fmt"
)

type HTTPError interface {
	error
	Code() int
}

func NewHTTPError(code int) HTTPError {
	return httpError{
		code: code,
	}
}

func WrapHTTPError(cause error, code int) HTTPError {
	return httpErrorWithCause{
		httpError: httpError{
			code: code,
		},
		cause: cause,
	}
}

type httpError struct {
	code int
}

func (err httpError) Code() int {
	return err.code
}

func (err httpError) Error() string {
	errMsg := fmt.Sprintf("client: unexpected status code from API: %d", err.code)
	return errMsg
}

type httpErrorWithCause struct {
	httpError
	cause error
}

func (err httpErrorWithCause) Error() string {
	errMsg := fmt.Sprintf("client: unexpected status code from API: %d", err.code)
	if err.cause != nil {
		errMsg += ": " + err.cause.Error()
	}
	return errMsg
}

func (err httpErrorWithCause) Unwrap() error {
	return err.cause
}

func (err httpErrorWithCause) Cause() error {
	return err.cause
}
