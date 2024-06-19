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

package model

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

type DeploymentsSearchParams struct {
	Page             int                          `json:"page"`
	PerPage          int                          `json:"per_page"`
	Filters          []DeploymentsFilterPredicate `json:"filters"`
	Sort             []DeploymentsSortCriteria    `json:"sort"`
	Attributes       []DeploymentsSelectAttribute `json:"attributes"`
	DeviceIDs        []string                     `json:"device_ids"`
	DeploymentIDs    []string                     `json:"deployment_ids"`
	DeploymentGroups []string                     `json:"-"`
	TenantID         string                       `json:"-"`
}

type DeploymentsFilterPredicate struct {
	Attribute string      `json:"attribute" bson:"attribute"`
	Type      string      `json:"type" bson:"type"`
	Value     interface{} `json:"value" bson:"value"`
}

type DeploymentsSortCriteria struct {
	Attribute string `json:"attribute"`
	Order     string `json:"order"`
}

type DeploymentsSelectAttribute struct {
	Attribute string `json:"attribute" bson:"attribute"`
}

func (sp DeploymentsSearchParams) Validate() error {
	for _, f := range sp.Filters {
		err := f.Validate()
		if err != nil {
			return err
		}
	}

	for _, s := range sp.Sort {
		err := validation.ValidateStruct(&s,
			validation.Field(&s.Attribute, validation.Required),
			validation.Field(&s.Order,
				validation.Required, validation.In(validSortOrders...),
			),
		)
		if err != nil {
			return err
		}
	}

	for _, s := range sp.Attributes {
		err := validation.ValidateStruct(&s,
			validation.Field(&s.Attribute, validation.Required))
		if err != nil {
			return err
		}
	}
	return nil
}

func (f DeploymentsFilterPredicate) Validate() error {
	return validation.ValidateStruct(&f,
		validation.Field(&f.Attribute, validation.Required),
		validation.Field(&f.Type, validation.Required, validation.In(validSelectors...)),
		validation.Field(&f.Value, validation.NotNil))
}

// ValueType returns actual type info of the value:
// type, is_array, err
func (f DeploymentsFilterPredicate) ValueType() (Type, bool, error) {
	isArr := false
	typ := TypeStr

	switch f.Value.(type) {
	case bool:
		typ = TypeBool
	case float64:
		typ = TypeNum
	case string:
		break
	case []string:
		isArr = true
	case []interface{}:
		isArr = true
		ival := f.Value.([]interface{})
		switch ival[0].(type) {
		case bool:
			typ = TypeBool
		case float64:
			typ = TypeNum
		case string:
			break
		default:
			return 0, false, errors.New(
				fmt.Sprintf("unknown attribute value type: %v %T",
					ival[0], ival[0]),
			)
		}
	default:
		return 0, false, errors.New(
			fmt.Sprintf("unknown attribute value type: %v %T",
				f.Value, f.Value),
		)

	}

	return typ, isArr, nil
}
