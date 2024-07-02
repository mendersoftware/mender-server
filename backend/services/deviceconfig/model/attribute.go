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
)

type Attribute struct {
	Key   string      `json:"key" bson:"key"`
	Value interface{} `json:"value" bson:"value"`
}

func (attr Attribute) Validate() error {
	return validation.ValidateStruct(&attr,
		validation.Field(&attr.Key,
			validation.Required,
			lengthLessThan4096,
		),
		validation.Field(&attr.Value,
			validateAttributeValue,
			lengthLessThan4096,
		),
	)
}

type Attributes []Attribute

func (a Attributes) Validate() error {
	return validation.Validate([]Attribute(a), validateAttributesLength)
}

func map2Attributes(configurationMap map[string]interface{}) Attributes {
	attributes := make(Attributes, len(configurationMap))
	i := 0
	for k, v := range configurationMap {
		attributes[i] = Attribute{
			Key:   k,
			Value: v,
		}
		i++
	}

	return attributes
}

func (a *Attributes) UnmarshalJSON(b []byte) error {
	var m map[string]interface{}

	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}

	*a = map2Attributes(m)

	return nil
}

func attributes2Map(attributes []Attribute) map[string]interface{} {
	configurationMap := make(map[string]interface{}, len(attributes))
	for _, a := range attributes {
		configurationMap[a.Key] = a.Value.(string)
	}

	return configurationMap
}

func (a Attributes) MarshalJSON() ([]byte, error) {
	return json.Marshal(attributes2Map(a))
}
