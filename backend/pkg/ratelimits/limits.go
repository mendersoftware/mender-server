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
package ratelimits

import (
	"encoding/json"
)

// ApiBurst defines minimal interval (seconds) between subsequent calls to action on uri
type ApiBurst struct {
	Action         string `json:"action" bson:"action"`
	Uri            string `json:"uri" bson:"uri"`
	MinIntervalSec int    `json:"min_interval_sec" bson:"min_interval_sec"`
}

// ApiQuota defines the ratelimit quota as max number of calls in a given time interval (in seconds)
// MaxCalls == 0 means 'no quota limit'
type ApiQuota struct {
	MaxCalls    int `json:"max_calls" bson:"max_calls"`
	IntervalSec int `json:"interval_sec" bson:"interval_sec"`
}

// ApiLimits combines burst limits and usage quota limits
type ApiLimits struct {
	ApiBursts []ApiBurst `json:"bursts" bson:"bursts"`
	ApiQuota  ApiQuota   `json:"quota" bson:"quota"`
}

// MarshalJSON makes sure even defaut nil ApiLimits.ApiBursts are actually empty lists
func (al ApiLimits) MarshalJSON() ([]byte, error) {
	if al.ApiBursts == nil {
		al.ApiBursts = make([]ApiBurst, 0)
	}

	type Copy ApiLimits
	copy := struct {
		Copy
	}{
		Copy: (Copy)(al),
	}

	return json.Marshal(&copy)
}
