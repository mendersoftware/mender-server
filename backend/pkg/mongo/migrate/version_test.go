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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewVersion(t *testing.T) {
	testCases := map[string]struct {
		input string

		output Version
		err    string
	}{
		"1.0.0": {
			input:  "1.0.0",
			output: Version{Major: 1, Minor: 0, Patch: 0},
			err:    "",
		},
		"1.0": {
			input:  "1.0",
			output: Version{Major: 1, Minor: 0, Patch: 0},
			err:    "failed to parse Version: unexpected EOF",
		},
		"1_0_0": {
			input:  "1_0_0",
			output: Version{Major: 1, Minor: 0, Patch: 0},
			err:    "failed to parse Version: input does not match format",
		},
	}
	for name, tc := range testCases {
		t.Logf("test case: %s", name)
		ver, err := NewVersion(tc.input)
		if tc.err != "" {
			assert.EqualError(t, err, tc.err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tc.output, *ver)
		}
	}
}

func TestVersionString(t *testing.T) {
	testCases := map[string]struct {
		input Version

		output string
	}{
		"1.0.0": {
			input:  Version{Major: 1, Minor: 0, Patch: 0},
			output: "1.0.0",
		},
		"4.2.5": {
			input:  Version{Major: 4, Minor: 2, Patch: 5},
			output: "4.2.5",
		},
	}
	for name, tc := range testCases {
		t.Logf("test case: %s", name)
		str := tc.input.String()
		assert.Equal(t, tc.output, str)
	}
}

func TestVersionIsLess(t *testing.T) {
	assert.True(t, VersionIsLess(Version{0, 0, 0}, Version{1, 1, 0}))
	assert.True(t, VersionIsLess(Version{1, 0, 0}, Version{1, 1, 0}))
	assert.True(t, VersionIsLess(Version{1, 0, 0}, Version{1, 0, 1}))
	assert.False(t, VersionIsLess(Version{1, 0, 0}, Version{0, 1, 0}))
	assert.False(t, VersionIsLess(Version{1, 1, 0}, Version{1, 0, 1}))
	assert.False(t, VersionIsLess(Version{1, 1, 0}, Version{1, 1, 0}))
}
