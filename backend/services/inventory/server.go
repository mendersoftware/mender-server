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
package main

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/log"

	api_http "github.com/mendersoftware/mender-server/services/inventory/api/http"
	"github.com/mendersoftware/mender-server/services/inventory/client/devicemonitor"
	"github.com/mendersoftware/mender-server/services/inventory/client/workflows"
	"github.com/mendersoftware/mender-server/services/inventory/config"
	inventory "github.com/mendersoftware/mender-server/services/inventory/inv"
	"github.com/mendersoftware/mender-server/services/inventory/store/mongo"
)

func RunServer(c config.Reader) error {

	l := log.New(log.Ctx{})

	db, err := mongo.NewDataStoreMongo(makeDataStoreConfig())
	if err != nil {
		return errors.Wrap(err, "database connection failed")
	}

	limitAttributes := c.GetInt(SettingLimitAttributes)
	limitTags := c.GetInt(SettingLimitTags)

	inv := inventory.NewInventory(db).WithLimits(limitAttributes, limitTags)

	devicemonitorAddr := c.GetString(SettingDevicemonitorAddr)
	if devicemonitorAddr != "" {
		c := devicemonitor.NewClient(devicemonitorAddr)
		inv = inv.WithDevicemonitor(c)
	}

	if inv, err = maybeWithInventory(inv, c); err != nil {
		return err
	}

	invapi := api_http.NewInventoryApiHandlers(inv)
	handler, err := invapi.Build()
	if err != nil {
		return errors.Wrap(err, "inventory API handlers setup failed")
	}
	addr := c.GetString(SettingListen)
	l.Printf("listening on %s", addr)

	return http.ListenAndServe(addr, handler)
}

func maybeWithInventory(
	inv inventory.InventoryApp,
	c config.Reader,
) (inventory.InventoryApp, error) {
	if reporting := c.GetBool(SettingEnableReporting); reporting {
		orchestrator := c.GetString(SettingOrchestratorAddr)
		if orchestrator == "" {
			return inv, errors.New("reporting integration needs orchestrator address")
		}

		c := workflows.NewClient(orchestrator)
		inv = inv.WithReporting(c)
	}
	return inv, nil
}
