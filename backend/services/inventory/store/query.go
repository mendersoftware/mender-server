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
package store

import "time"

type ComparisonOperator int

const (
	Eq ComparisonOperator = 1 << iota
)

type Filter struct {
	AttrName   string
	AttrScope  string
	Value      string
	ValueFloat *float64
	ValueTime  *time.Time
	Operator   ComparisonOperator
}

type Sort struct {
	AttrName  string
	AttrScope string
	Ascending bool
}

type ListQuery struct {
	Skip      int
	Limit     int
	Filters   []Filter
	Sort      *Sort
	HasGroup  *bool
	GroupName string
}
