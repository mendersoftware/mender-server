// Copyright 2026 Northern.tech AS
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

package http

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var ErrMissingPathParam = errors.New("[bug] path parameter not found")

type pathParamError struct {
	param string
	err   error
}

func (err pathParamError) Unwrap() error {
	return err.err
}

func (err pathParamError) Error() string {
	return fmt.Sprintf("invalid path parameter %s: %s", err.param, err.err)
}

func PathParam(c *gin.Context, name string, rulez ...validation.Rule) (string, error) {
	value, found := c.Params.Get(name)
	if !found {
		return "", ErrMissingPathParam
	}
	err := validation.Validate(value, rulez...)
	if err != nil {
		return "", pathParamError{
			param: name,
			err:   err,
		}
	}
	return value, nil
}
