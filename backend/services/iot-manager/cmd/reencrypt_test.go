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

package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/mendersoftware/mender-server/services/iot-manager/crypto"
	"github.com/mendersoftware/mender-server/services/iot-manager/model"
	"github.com/mendersoftware/mender-server/services/iot-manager/store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReencrypt(t *testing.T) {
	store := &mocks.DataStore{}
	defer store.AssertExpectations(t)

	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest"))
	integrations := []model.Integration{
		{
			ID:       integrationID,
			Provider: model.ProviderIoTHub,
			Credentials: model.Credentials{
				Type: model.CredentialTypeSAS,
				ConnectionString: &model.ConnectionString{
					HostName: "localhost:8080",
					Key:      crypto.String("not-so-secret-key"),
					Name:     "foobar",
				},
			},
		},
	}
	store.On("GetIntegrations",
		mock.MatchedBy(func(_ context.Context) bool {
			return true
		}),
		model.IntegrationFilter{
			Skip:  int64(0),
			Limit: defaultLimit,
		},
	).Return(integrations, nil).Once()

	store.On("SetIntegrationCredentials",
		mock.MatchedBy(func(_ context.Context) bool {
			return true
		}),
		integrationID,
		integrations[0].Credentials,
	).Return(nil).Once()

	store.On("GetIntegrations",
		mock.MatchedBy(func(_ context.Context) bool {
			return true
		}),
		model.IntegrationFilter{
			Skip:  int64(0) + defaultLimit,
			Limit: defaultLimit,
		},
	).Return([]model.Integration{}, nil).Once()

	err := Reencrypt(store)
	assert.NoError(t, err)
}

func TestReencryptErrorGetIntegrations(t *testing.T) {
	store := &mocks.DataStore{}
	defer store.AssertExpectations(t)

	errStore := errors.New("error")
	store.On("GetIntegrations",
		mock.MatchedBy(func(_ context.Context) bool {
			return true
		}),
		model.IntegrationFilter{
			Skip:  int64(0),
			Limit: defaultLimit,
		},
	).Return(nil, errStore).Once()

	err := Reencrypt(store)
	assert.Error(t, err)
	assert.EqualError(t, err, errStore.Error())
}

func TestReencryptErrorSetIntegrationCredentials(t *testing.T) {
	store := &mocks.DataStore{}
	defer store.AssertExpectations(t)

	errStore := errors.New("error")

	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest"))
	integrations := []model.Integration{
		{
			ID:       integrationID,
			Provider: model.ProviderIoTHub,
			Credentials: model.Credentials{
				Type: model.CredentialTypeSAS,
				ConnectionString: &model.ConnectionString{
					HostName: "localhost:8080",
					Key:      crypto.String("not-so-secret-key"),
					Name:     "foobar",
				},
			},
		},
	}
	store.On("GetIntegrations",
		mock.MatchedBy(func(_ context.Context) bool {
			return true
		}),
		model.IntegrationFilter{
			Skip:  int64(0),
			Limit: defaultLimit,
		},
	).Return(integrations, nil).Once()

	store.On("SetIntegrationCredentials",
		mock.MatchedBy(func(_ context.Context) bool {
			return true
		}),
		integrationID,
		integrations[0].Credentials,
	).Return(errStore).Once()

	err := Reencrypt(store)
	assert.Error(t, err)
	assert.EqualError(t, err, errStore.Error())
}
