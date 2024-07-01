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

import "time"

func UnixMilis() int64 {
	return time.Now().UnixNano() / 1000000
}

type Clock interface {
	Now() time.Time
	Forward(secs int64)
}

type clock struct{}

func NewClock() *clock {
	return &clock{}
}

func (c *clock) Now() time.Time {
	return time.Now()
}

func (c *clock) Forward(secs int64) {
	//noop
}

type mockClock struct {
	unixNow int64
}

func NewMockClock(unixNow int64) *mockClock {
	return &mockClock{
		unixNow: unixNow,
	}
}

func (mc *mockClock) Now() time.Time {
	return time.Unix(mc.unixNow, 0)
}

func (mc *mockClock) Forward(secs int64) {
	mc.unixNow += secs
}
