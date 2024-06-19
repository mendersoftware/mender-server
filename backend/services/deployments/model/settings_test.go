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
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStorageSettingsDeserialize(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		Raw      string
		Expected StorageSettings

		Error error
	}{{
		Name: "ok/s3",

		Raw: `
{
  "type": "s3",
  "key": "not_so_secret_key_id",
  "secret": "super_secret",
  "bucket": "bucketMcBucketFace",
  "region": "wrld-east-west-1"
}
`,
		Expected: StorageSettings{
			Type:   StorageTypeS3,
			Bucket: "bucketMcBucketFace",
			Key:    "not_so_secret_key_id",
			Secret: "super_secret",
			Region: "wrld-east-west-1",
		},
	}, {
		Name: "ok/default type s3",

		Raw: `
{
  "key": "not_so_secret_key_id",
  "secret": "super_secret",
  "bucket": "bucketMcBucketFace",
  "region": "wrld-east-west-1"
}
`,
		Expected: StorageSettings{
			Type:   StorageTypeS3,
			Bucket: "bucketMcBucketFace",
			Key:    "not_so_secret_key_id",
			Secret: "super_secret",
			Region: "wrld-east-west-1",
		},
	}, {
		Name: "ok/azure",

		Raw: `
{
  "type": "azure",
  "account_name": "AccountName",
  "account_key": "AccountKey",
  "container_name": "containerMcBucketFace"
}
`,
		Expected: StorageSettings{
			Type:   StorageTypeAzure,
			Bucket: "containerMcBucketFace",
			Key:    "AccountName",
			Secret: "AccountKey",
		},
	}, {
		Name: "error/malformed data",

		Raw:   `<insert xss snippet>`,
		Error: &json.SyntaxError{},
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			actual, err := ParseStorageSettingsRequest(strings.NewReader(tc.Raw))
			if tc.Error != nil {
				assert.ErrorAs(t, err, &tc.Error)
			} else if assert.NoError(t, err) && assert.NotNil(t, actual) {
				assert.Equal(t, tc.Expected, *actual)
			}
		})
	}
}
