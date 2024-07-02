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
package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTruncateToDay(t *testing.T) {
	now := time.Now()
	nowTruncated := TruncateToDay(now)

	assert.Equal(t, nowTruncated.Hour(), 0)
	assert.Equal(t, nowTruncated.Minute(), 0)
	assert.Equal(t, nowTruncated.Second(), 0)

	assert.Equal(t, now.Year(), nowTruncated.Year())
	assert.Equal(t, now.Month(), nowTruncated.Month())
	assert.Equal(t, now.Day(), nowTruncated.Day())
}
