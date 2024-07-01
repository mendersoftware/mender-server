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
package migrate

import (
	"fmt"

	"github.com/pkg/errors"
)

type Version struct {
	Major uint `bson:"major"`
	Minor uint `bson:"minor"`
	Patch uint `bson:"patch"`
}

func NewVersion(s string) (*Version, error) {
	var maj, min, patch uint

	n, err := fmt.Sscanf(s, "%d.%d.%d", &maj, &min, &patch)

	if err != nil {
		return nil, errors.Wrap(err, "failed to parse Version")
	} else if n != 3 {
		return nil, errors.New("invalid semver format")
	}

	return &Version{Major: maj, Minor: min, Patch: patch}, nil
}

func MakeVersion(maj uint, min uint, patch uint) Version {
	return Version{Major: maj, Minor: min, Patch: patch}
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func VersionIsLess(left Version, right Version) bool {
	if left.Major < right.Major {
		return true
	} else if left.Major == right.Major {
		if left.Minor < right.Minor {
			return true
		} else if left.Minor == right.Minor {
			if left.Patch < right.Patch {
				return true
			}
		}
	}
	return false
}
