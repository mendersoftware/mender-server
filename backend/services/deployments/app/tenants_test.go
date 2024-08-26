// Copyright 2023 Northern.tech AS
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

package app

import (
	"context"
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	fs_mocks "github.com/mendersoftware/mender-server/services/deployments/storage/mocks"

	mstore "github.com/mendersoftware/mender-server/services/deployments/store/mocks"
)

func TestProvisionTenant(t *testing.T) {
	testCases := []struct {
		id string

		storeErr error

		err error
	}{
		{
			id: "foo",
		},
		{
			id:       "foo",
			storeErr: errors.New("connection failed"),
			err:      errors.New("failed to provision tenant: connection failed"),
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			db := mstore.DataStore{}
			db.On("ProvisionTenant",
				mock.MatchedBy(
					func(_ context.Context) bool {
						return true
					}),
				tc.id).Return(tc.storeErr)

			fs := &fs_mocks.ObjectStorage{}

			d := NewDeployments(&db, fs, 0, false)

			ctx := context.Background()

			err := d.ProvisionTenant(ctx, tc.id)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
