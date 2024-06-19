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
package addons

const (
	MenderTroubleshoot = "troubleshoot"
	MenderConfigure    = "configure"
	MenderMonitor      = "monitor"
)

var (
	KnownAddons = []string{
		MenderTroubleshoot,
		MenderConfigure,
		MenderMonitor,
	}

	AllAddonsDisabled = []Addon{
		{
			Name:    MenderConfigure,
			Enabled: false,
		},
		{
			Name:    MenderTroubleshoot,
			Enabled: false,
		},
		{
			Name:    MenderMonitor,
			Enabled: false,
		},
	}
	AllAddonsEnabled = []Addon{
		{
			Name:    MenderConfigure,
			Enabled: true,
		},
		{
			Name:    MenderTroubleshoot,
			Enabled: true,
		},
		{
			Name:    MenderMonitor,
			Enabled: true,
		},
	}
	TrialAddons = AllAddonsEnabled
)

type Addon struct {
	Name    string `json:"name" bson:"name"`
	Enabled bool   `json:"enabled" bson:"enabled"`
}
