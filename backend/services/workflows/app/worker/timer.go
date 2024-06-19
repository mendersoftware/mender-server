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

package worker

import "time"

// reusableTimer allows using the time.After interface without reallocations.
type reusableTimer time.Timer

func newStoppedTimer() *reusableTimer {
	t := (*reusableTimer)(time.NewTimer(0))
	t.Stop()
	return t
}

// Stop ensures that the timer is stopped and the channel is cleared.
func (t *reusableTimer) Stop() {
	if (*time.Timer)(t).Stop() {
		select {
		case <-t.C:
		default:
		}
	}
}

func (t *reusableTimer) Reset(d time.Duration) {
	t.Stop()
	(*time.Timer)(t).Reset(d)
}

func (t *reusableTimer) After(d time.Duration) <-chan time.Time {
	t.Reset(d)
	return t.C
}
