// Copyright 2019 Northern.tech AS
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

const (
	LimitStorage = "storage"
)

var (
	ValidLimits = []string{LimitStorage}
)

type Limit struct {
	Name  string `bson:"_id"`
	Value uint64 `bson:"value" json:"value"`
}

func (l Limit) IsLess(what uint64) bool {
	return what < l.Value
}

func IsValidLimit(name string) bool {
	for _, n := range ValidLimits {
		if name == n {
			return true
		}
	}
	return false
}
