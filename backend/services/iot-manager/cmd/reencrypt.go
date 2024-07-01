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

	"github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/iot-manager/model"
	"github.com/mendersoftware/mender-server/services/iot-manager/store"
)

const defaultLimit = int64(100)

func Reencrypt(dataStore store.DataStore) error {
	ctx := context.Background()
	l := log.FromContext(ctx)

	skip := int64(0)
	for {
		filter := model.IntegrationFilter{
			Skip:  skip,
			Limit: defaultLimit,
		}
		integrations, err := dataStore.GetIntegrations(ctx, filter)
		if err != nil {
			return err
		} else if len(integrations) == 0 {
			break
		}
		for _, integration := range integrations {
			if integration.Credentials.Type == model.CredentialTypeSAS {
				l.Infof("Re-encrypting credentials for integration %s", integration.ID)
				err = dataStore.SetIntegrationCredentials(ctx,
					integration.ID, integration.Credentials)
				if err != nil {
					return err
				}
			}
		}
		skip += defaultLimit
	}
	return nil
}
