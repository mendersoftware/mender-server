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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJsonSort(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		what string
		out  string
		err  error
	}{
		{
			what: "{\"mac\": \"de:ad:be:ef\"}",
			out:  "{\"mac\":\"de:ad:be:ef\"}",
		},
		{
			what: "{\"mac\":\"de:ad:be:ef\"}",
			out:  "{\"mac\":\"de:ad:be:ef\"}",
		},
		{
			what: "  {\"mac\":  \"de:ad:be:ef\" }",
			out:  "{\"mac\":\"de:ad:be:ef\"}",
		},
		{
			what: "{\"sn\":\"00001\",\"mac\": \"de:ad:be:ef\"}",
			out:  "{\"mac\":\"de:ad:be:ef\",\"sn\":\"00001\"}",
		},
		{
			what: "{\"sn\": \"00001\", \"mac\":\"de:ad:be:ef\"}",
			out:  "{\"mac\":\"de:ad:be:ef\",\"sn\":\"00001\"}",
		},
		{
			what: "{\"sn\":\"00001\",\"mac\": \"de:ad:be:ef\", \"attribute_foo\":\"foo\"}",

			out: "{\"attribute_foo\":\"foo\",\"mac\":\"de:ad:be:ef\",\"sn\":\"00001\"}",
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(fmt.Sprintf("tc %d", i), func(t *testing.T) {
			t.Parallel()

			out, err := JsonSort(tc.what)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.out, out)
			}

		})
	}

}
