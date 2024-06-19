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

	"github.com/mendersoftware/mender-server/services/deployments/model"
	fs_mocks "github.com/mendersoftware/mender-server/services/deployments/storage/mocks"
	"github.com/mendersoftware/mender-server/services/deployments/store/mocks"
	"github.com/mendersoftware/mender-server/services/deployments/store/mongo"
)

func TestGetLimit(t *testing.T) {
	testCases := []struct {
		name string

		getLimit *model.Limit
		getErr   error

		expected *model.Limit
		err      error
	}{
		{
			name: "foo",
			getLimit: &model.Limit{
				Name:  "foo",
				Value: 123,
			},
			expected: &model.Limit{
				Name:  "foo",
				Value: 123,
			},
		},
		{
			name:   "not-found",
			getErr: mongo.ErrLimitNotFound,
			expected: &model.Limit{
				Name:  "not-found",
				Value: 0,
			},
		},
		{
			name:   "not-found",
			getErr: errors.New("error"),
			err:    errors.New("failed to obtain limit from storage: error"),
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			db := mocks.DataStore{}
			db.On("GetLimit",
				mock.MatchedBy(
					func(_ context.Context) bool {
						return true
					}),
				tc.name).Return(tc.getLimit, tc.getErr)

			fs := &fs_mocks.ObjectStorage{}

			d := NewDeployments(&db, fs, 0, false)

			ctx := context.Background()
			lim, err := d.GetLimit(ctx, tc.name)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}
			if tc.expected != nil {
				assert.Equal(t, tc.expected, lim)
			}

			db.AssertExpectations(t)
		})
	}
}
