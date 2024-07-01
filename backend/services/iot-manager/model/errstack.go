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

package model

import (
	"errors"
	"strings"
)

type ErrorStack []error

func (stack ErrorStack) Is(target error) bool {
	for _, err := range stack {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

func (stack ErrorStack) As(target interface{}) bool {
	for _, err := range stack {
		if errors.As(err, target) {
			return true
		}
	}
	return false
}

func (stack *ErrorStack) Push(err error) *ErrorStack {
	*stack = append(*stack, err)
	return stack
}

func (stack ErrorStack) Error() string {
	var (
		strBldr strings.Builder
		lim     = len(stack) - 1
	)
	for i, err := range stack {
		strBldr.WriteString(err.Error())
		if i < lim {
			strBldr.WriteString("; ")
		}
	}
	return strBldr.String()
}
