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
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

const AttributesMaxLength = 100

var (
	lengthLessThan4096 = validation.Length(0, 4096)

	validateAttributeValue = validation.By(func(value interface{}) error {
		switch value.(type) {
		case string:
			return nil

		default:
			// NOTE: we will support more types in the future
			return errors.Errorf("invalid type: %T", value)
		}
	})

	validateAttributesLength = validation.Length(
		0, AttributesMaxLength,
	).Error(fmt.Sprintf(
		"too many configuration attributes, maximum is %d",
		AttributesMaxLength,
	))
)
