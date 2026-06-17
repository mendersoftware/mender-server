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
package mongo

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/mendersoftware/mender-server/pkg/api"
	"github.com/mendersoftware/mender-server/pkg/api/client"
	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"
	ctxstore "github.com/mendersoftware/mender-server/pkg/store"

	dconfig "github.com/mendersoftware/mender-server/services/deviceauth/config"
	"github.com/mendersoftware/mender-server/services/deviceauth/model"
)

type migration_1_7_0 struct {
	ms  *DataStoreMongo
	ctx context.Context
}

const (
	migrationContextTimeouts = 32
	devicesBatchSize         = 512
)

func (m *migration_1_7_0) updateDevicesStatus(ctx context.Context, status string) error {
	invAddr := config.Config.GetString(dconfig.SettingInventoryAddr)
	inventoryCfg, err := api.NewDefaultClientConfigurationFromURL(
		invAddr,
	)
	if err != nil {
		return err
	}

	inventoryCfg.HTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	c := client.NewAPIClient(inventoryCfg).DeviceInventoryInternalAPIAPI

	collectionDevices := m.ms.client.Database(ctxstore.DbFromContext(m.ctx, DbName)).
		Collection(DbDevicesColl)
	opts := options.Find()
	opts.SetNoCursorTimeout(true)
	cur, err := collectionDevices.Find(ctx, bson.M{"status": status}, opts)
	if err != nil {
		return err
	}
	id := identity.FromContext(m.ctx)
	var deviceUpdates []client.DeviceUpdate
	deviceUpdates = make([]client.DeviceUpdate, devicesBatchSize)

	i := 0
	for cur.Next(ctx) {
		var d model.Device
		err = cur.Decode(&d)
		if err != nil {
			return err
		}
		if i >= devicesBatchSize {
			//nolint:bodyclose
			_, err := c.UpdateStatusOfDevices(ctx, id.Tenant, status).
				DeviceUpdate(deviceUpdates).Execute()
			if err != nil {
				return err
			}

			deviceUpdates = make([]client.DeviceUpdate, devicesBatchSize)
			i = 0
		}
		deviceUpdates[i].Id = d.Id
		deviceUpdates[i].Revision = int32(d.Revision)
		i++
	}
	if i >= 1 {
		//nolint:bodyclose
		_, err := c.UpdateStatusOfDevices(ctx, id.Tenant, status).
			DeviceUpdate(deviceUpdates[:i]).Execute()
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *migration_1_7_0) Up(from migrate.Version) error {
	ctx, cancel := context.WithTimeout(context.Background(), migrationContextTimeouts*time.Second)
	defer cancel()

	err := m.updateDevicesStatus(ctx, "accepted")
	if err == nil {
		err = m.updateDevicesStatus(ctx, "pending")
	}
	if err == nil {
		err = m.updateDevicesStatus(ctx, "rejected")
	}
	if err == nil {
		err = m.updateDevicesStatus(ctx, "preauthorized")
	}

	return err
}

func (m *migration_1_7_0) Version() migrate.Version {
	return migrate.MakeVersion(1, 7, 0)
}
