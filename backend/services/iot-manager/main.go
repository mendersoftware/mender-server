// Copyright 2024 Northern.tech AS
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

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/mendersoftware/mender-server/services/iot-manager/app"
	"github.com/mendersoftware/mender-server/services/iot-manager/client/devauth"
	"github.com/mendersoftware/mender-server/services/iot-manager/client/iotcore"
	"github.com/mendersoftware/mender-server/services/iot-manager/client/iothub"
	"github.com/mendersoftware/mender-server/services/iot-manager/client/workflows"
	"github.com/mendersoftware/mender-server/services/iot-manager/cmd"
	dconfig "github.com/mendersoftware/mender-server/services/iot-manager/config"
	"github.com/mendersoftware/mender-server/services/iot-manager/crypto"
	"github.com/mendersoftware/mender-server/services/iot-manager/model"
	"github.com/mendersoftware/mender-server/services/iot-manager/server"
	store "github.com/mendersoftware/mender-server/services/iot-manager/store/mongo"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/log"
)

func main() {
	doMain(os.Args)
}

func doMain(args []string) {
	var configPath string

	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "config",
				Usage: "Configuration `FILE`. " +
					"Supports JSON, TOML, YAML and HCL " +
					"formatted configs.",
				Value:       "/etc/iot-manager/config.yaml",
				Destination: &configPath,
			},
			&cli.StringFlag{
				Name:  "log-level",
				Usage: "Log `LEVEL` to emit to standard error.",
				Value: "info",
			},
		},
		Commands: []cli.Command{
			{
				Name:   "server",
				Usage:  "Run the HTTP API server",
				Action: cmdServer,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "automigrate",
						Usage: "Run database migrations before starting.",
					},
				},
			},
			{
				Name:   "migrate",
				Usage:  "Run the migrations",
				Action: cmdMigrate,
			},
			{
				Name:   "re-encrypt",
				Usage:  "Re-encrypt the secrets using the (new) encryption key",
				Action: cmdReencrypt,
			},
			{
				Name:   "sync-devices",
				Usage:  "Synchronize device state across IoT platforms.",
				Action: cmdSync,
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:  "batch-size",
						Usage: "Maximum number of devices to sync in a batch.",
						Value: 50,
					},
					&cli.BoolFlag{
						Name:  "fail-early",
						Usage: "Do not ignore non-fatal errors.",
					},
				},
			},
		},
	}
	app.Usage = "IoT Manager"
	app.Action = cmdServer

	app.Before = func(args *cli.Context) error {
		lvl, err := logrus.ParseLevel(args.String("log-level"))
		if err != nil {
			return err
		}
		log.Log.Level = lvl

		err = config.FromConfigFile(configPath, dconfig.Defaults)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("error loading configuration: %s", err),
				1)
		}

		// Enable setting config values by environment variables
		config.Config.SetEnvPrefix("IOT_MANAGER")
		config.Config.AutomaticEnv()
		config.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

		// Set encryption keys
		err = crypto.SetAESEncryptionKey(
			config.Config.GetString(dconfig.SettingAESEncryptionKey),
		)
		if err == nil {
			err = crypto.SetAESEncryptionFallbackKey(
				config.Config.GetString(dconfig.SettingAESEncryptionFallbackKey),
			)
		}
		model.SetTrustedHostnames(
			config.Config.GetStringSlice(dconfig.SettingDomainWhitelist),
		)

		store.SetEventExpiration(
			config.Config.GetInt64(dconfig.SettingEventExpirationTimeout),
		)

		return err
	}

	err := app.Run(args)
	if err != nil {
		logrus.Fatal(err)
	}
}

func cmdServer(args *cli.Context) error {
	mgoConfig := store.NewConfig().SetAutomigrate(args.Bool("automigrate"))
	dataStore, err := store.SetupDataStore(mgoConfig)
	if err != nil {
		return err
	}
	defer dataStore.Close()
	return server.InitAndRun(config.Config, dataStore)
}

func cmdMigrate(args *cli.Context) error {
	mgoConfig := store.NewConfig().SetAutomigrate(true)
	dataStore, err := store.SetupDataStore(mgoConfig)
	if err != nil {
		return err
	}
	return dataStore.Close()
}

func cmdReencrypt(args *cli.Context) error {
	mgoConfig := store.NewConfig().SetAutomigrate(args.Bool("automigrate"))
	dataStore, err := store.SetupDataStore(mgoConfig)
	if err != nil {
		return err
	}
	defer dataStore.Close()
	return cmd.Reencrypt(dataStore)
}

func cmdSync(args *cli.Context) error {
	if bs := args.Int("batch-size"); bs <= 0 {
		return cli.NewExitError(
			"invalid flag 'batch-size': must be a positive integer", 1,
		)
	} else if bs > 500 {
		// This is the max page size from deviceauth
		return cli.NewExitError(
			"invalid flag 'batch-size': must be less than 500", 1,
		)
	}
	httpClient := new(http.Client)
	ctx := context.Background()

	wf := workflows.NewClient(
		config.Config.GetString(dconfig.SettingWorkflowsURL),
		workflows.NewOptions().SetClient(httpClient),
	)
	hub := iothub.NewClient(iothub.NewOptions().SetClient(httpClient))
	core := iotcore.NewClient()
	mgoConfig := store.NewConfig()
	devauth, err := devauth.NewClient(devauth.Config{
		Client:         httpClient,
		DevauthAddress: config.Config.GetString(dconfig.SettingDeviceauthURL),
	})
	if err != nil {
		return err
	}

	ds, err := store.SetupDataStore(mgoConfig)
	if err != nil {
		return err
	}
	defer ds.Close()
	app := app.New(ds, wf, devauth).WithIoTHub(hub).WithIoTCore(core)
	app = app.WithWebhooksTimeout(config.Config.GetUint(dconfig.SettingWebhooksTimeoutSeconds))
	return app.SyncDevices(ctx, args.Int("batch-size"), args.Bool("fail-early"))
}
