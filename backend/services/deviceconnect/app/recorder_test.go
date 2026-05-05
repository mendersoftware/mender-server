// Copyright 2021 Northern.tech AS
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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	store_mocks "github.com/mendersoftware/mender-server/services/deviceconnect/store/mocks"
)

func TestNewRecorder(t *testing.T) {
	sessionID := "sessionID"
	r := NewRecorder(sessionID, nil)
	assert.NotNil(t, r)
}

func TestRecorderWrite(t *testing.T) {
	ctx := context.Background()
	sessionID := "sessionID"

	testCases := []struct {
		Name                       string
		DbGetSessionRecordingError error
		Data                       []byte
	}{
		{
			Name: "ok",
			Data: []byte("some data"),
		},
		{
			Name:                       "error from the store",
			Data:                       []byte("some data"),
			DbGetSessionRecordingError: errors.New("some error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			store := &store_mocks.DataStore{}

			r := NewRecorder(sessionID, store)
			assert.NotNil(t, r)
			require.IsType(t, &recorder{}, r)
			r.(*recorder).init(func(ctx context.Context, b []byte) error {
				return tc.DbGetSessionRecordingError
			}, 5)

			err := r.Record(ctx, tc.Data)

			if tc.DbGetSessionRecordingError == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
