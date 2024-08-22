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
	"github.com/stretchr/testify/mock"

	store_mocks "github.com/mendersoftware/mender-server/services/deviceconnect/store/mocks"
)

func TestNewRecorder(t *testing.T) {
	ctx := context.Background()
	sessionID := "sessionID"
	r := NewRecorder(ctx, sessionID, nil)
	assert.NotNil(t, r)
	assert.Equal(t, r.ctx, ctx)
	assert.Equal(t, r.sessionID, sessionID)
	assert.Equal(t, r.store, nil)
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
			DbGetSessionRecordingError: errors.New("some error"),
		},
	}

	for _, tc := range testCases {
		store := &store_mocks.DataStore{}
		store.On("InsertSessionRecording",
			mock.MatchedBy(func(ctx context.Context) bool {
				return true
			}),
			sessionID,
			mock.AnythingOfType("[]uint8"),
		).Return(tc.DbGetSessionRecordingError)

		r := NewRecorder(ctx, sessionID, store)
		assert.NotNil(t, r)

		n, err := r.Write(tc.Data)

		if tc.DbGetSessionRecordingError == nil {
			assert.NoError(t, err)
			assert.Equal(t, len(tc.Data), n)
		} else {
			assert.Error(t, err)
			assert.Equal(t, -1, n)
		}
	}
}
