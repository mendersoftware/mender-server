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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli"

	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/version"

	"github.com/mendersoftware/mender-server/services/inventory/config"
	"github.com/mendersoftware/mender-server/services/inventory/store/mongo"
)

var appVersion = version.Get()

func main() {
	doMain(os.Args)
}

const maintenanceDescription = `Run migrations in maintenance mode.
   WARNING: All external endpoints modifying state in the database must be
            temporarily disabled while maintenance is in progress:
       - PUT    /api/management/v1/inventory/devices/{id}/group
       - DELETE /api/management/v1/inventory/devices/{id}
       - DELETE /api/management/v1/inventory/devices/{id}/group/{name}
       - PATCH  /api/devices/v1/inventory/devices/attributes`

func doMain(args []string) {
	var configPath string
	var debug bool

	app := cli.NewApp()
	app.Usage = "Device Authentication Service"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "config",
			Usage: "Configuration `FILE`." +
				" Supports JSON, TOML, YAML and HCL formatted configs.",
			Destination: &configPath,
		},
		cli.BoolFlag{
			Name:  "dev",
			Usage: "Use development setup",
		},
		cli.BoolFlag{
			Name:        "debug",
			Usage:       "Enable debug logging",
			Destination: &debug,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "server",
			Usage: "Run the service as a server",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "automigrate",
					Usage: "Run database migrations before starting.",
				},
			},

			Action: cmdServer,
		},
		{
			Name:  "migrate",
			Usage: "Run migrations",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "tenant",
					Usage: "Takes ID of specific tenant to migrate.",
				},
			},

			Action: cmdMigrate,
		},
		{
			Name:        "maintenance",
			Description: maintenanceDescription,
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name: "tenant, t",
					Usage: "Takes ID of specific " +
						"tenant(s) to migrate. " +
						"Flag can be provided " +
						"multiple times.",
				},
				cli.StringFlag{
					Name:  "version",
					Usage: "Target version to migrate",
					Value: mongo.DbVersion,
				},
			},

			Action: cmdMaintenence,
		},
		{
			Name:  "version",
			Usage: "Show version information",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "output",
					Usage: "Output format <json|text>",
					Value: "text",
				},
			},
			Action: func(args *cli.Context) error {
				switch strings.ToLower(args.String("output")) {
				case "text":
					fmt.Print(appVersion)
				case "json":
					_ = json.NewEncoder(os.Stdout).Encode(appVersion)
				default:
					return fmt.Errorf("Unknown output format %q", args.String("output"))
				}
				return nil
			},
		},
	}

	app.Version = appVersion.Version
	app.Action = cmdServer
	app.Before = func(args *cli.Context) error {
		log.Setup(debug)

		err := config.FromConfigFile(configPath, configDefaults)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("error loading configuration: %s", err),
				1)
		}

		// Enable setting conig values by environment variables
		config.Config.SetEnvPrefix("INVENTORY")
		config.Config.AutomaticEnv()

		return nil
	}

	_ = app.Run(args)
}

func makeDataStoreConfig() mongo.DataStoreMongoConfig {
	return mongo.DataStoreMongoConfig{
		ConnectionString: config.Config.GetString(SettingDb),

		SSL:           config.Config.GetBool(SettingDbSSL),
		SSLSkipVerify: config.Config.GetBool(SettingDbSSLSkipVerify),

		Username: config.Config.GetString(SettingDbUsername),
		Password: config.Config.GetString(SettingDbPassword),
	}

}

func cmdServer(args *cli.Context) error {
	l := log.New(log.Ctx{})

	db, err := mongo.NewDataStoreMongo(makeDataStoreConfig())
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("failed to connect to db: %v", err),
			3)
	}

	if args.Bool("automigrate") {
		db = db.WithAutomigrate()
	}

	ctx := context.Background()
	err = db.Migrate(ctx, mongo.DbVersion)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("failed to run migrations: %v", err),
			3)
	}

	l.Print("Inventory Service starting up")

	err = RunServer(config.Config)
	if err != nil {
		return cli.NewExitError(err.Error(), 4)
	}

	return nil
}

func cmdMigrate(args *cli.Context) error {
	tenantId := args.String("tenant")

	l := log.New(log.Ctx{})

	l.Print("Inventory Service starting up")

	if tenantId != "" {
		l.Printf("migrating tenant %v", tenantId)
	} else {
		l.Printf("migrating all the tenants")
	}

	db, err := mongo.NewDataStoreMongo(makeDataStoreConfig())

	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("failed to connect to db: %v", err),
			3)
	}

	// we want to apply migrations
	db = db.WithAutomigrate()

	ctx := context.Background()

	if tenantId != "" {
		err = db.MigrateTenant(ctx, mongo.DbVersion, tenantId)
	} else {
		err = db.Migrate(ctx, mongo.DbVersion)
	}
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("failed to run migrations: %v", err),
			3)
	}

	return nil
}

func cmdMaintenence(args *cli.Context) error {
	tenantIDs := args.StringSlice("tenant")
	version := args.String("version")

	l := log.New(log.Ctx{})

	if len(tenantIDs) > 0 {
		l.Infof("performing maintenence for tenants: %v", tenantIDs)
	} else {
		l.Info("performing maintenance for all the tenants")
	}
	db, err := mongo.NewDataStoreMongo(makeDataStoreConfig())

	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("failed to connect to db: %v", err),
			3)
	}

	// we want to apply migrations
	db = db.WithAutomigrate()

	ctx := context.Background()

	err = db.Maintenance(ctx, version, tenantIDs...)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("failed to run migrations: %v", err),
			3)
	}

	return nil
}
