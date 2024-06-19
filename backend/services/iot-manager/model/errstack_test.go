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
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"
)

type testError struct{}

func (testError) Error() string {
	return "this is a test"
}

func TestErrorStack(t *testing.T) {
	var (
		errStack ErrorStack
		anErr    = errors.New("foo")
		valErr   validation.Error
	)
	errStack.Push(anErr).
		Push(validation.ErrEmpty)
	assert.ErrorIs(t, errStack, anErr)
	assert.ErrorAs(t, errStack, &valErr)
	var testErr testError
	assert.False(t, errors.Is(errStack, testErr))
	assert.False(t, errors.As(errStack, &testErr))
	errStack.Push(testErr)
	assert.ErrorIs(t, errStack, testErr)
	assert.ErrorContains(t, errStack, testErr.Error())
	assert.ErrorContains(t, errStack, anErr.Error())
	assert.ErrorContains(t, errStack, valErr.Error())
}
