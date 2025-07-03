// Copyright 2022 Northern.tech AS
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

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/version"

	. "github.com/mendersoftware/mender-server/services/useradm/config"
	"github.com/mendersoftware/mender-server/services/useradm/model"
	"github.com/mendersoftware/mender-server/services/useradm/store/mongo"
)

var appVersion = version.Get()

func main() {
	if err := doMain(os.Args); err != nil {
		os.Exit(1)
	}
}

func doMain(args []string) error {
	var debug bool
	var configPath string

	app := cli.NewApp()
	app.Usage = "user administration service"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "config",
			Usage: "Configuration `FILE`." +
				" Supports JSON, TOML, YAML and HCL formatted configs.",
			Destination: &configPath,
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

			Action: runServer,
		},
		{
			Name:  "create-user",
			Usage: "Create user",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "username",
					Usage: "Name of created user, must be an email address",
				},
				cli.StringFlag{
					Name:  "password",
					Usage: "User's password, leave empty to have it read from stdin",
				},
				cli.StringFlag{
					Name:  "user-id",
					Usage: "User ID, if already generated/available (optional).",
				},
				cli.StringFlag{
					Name:  "tenant-id",
					Usage: "Tenant ID, if running a multitenant setup (optional).",
				},
			},
			Action: runCreateUser,
		},
		{
			Name:  "set-password",
			Usage: "Set password",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "username",
					Usage: "Name of created user, must be an email address",
				},
				cli.StringFlag{
					Name:  "password",
					Usage: "User's password, leave empty to have it read from stdin",
				},
				cli.StringFlag{
					Name:  "tenant-id",
					Usage: "Tenant ID, if running a multitenant setup (optional).",
				},
			},
			Action: runSetPassword,
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

			Action: runMigrate,
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
	app.Action = runServer
	app.Before = func(args *cli.Context) error {
		log.Setup(debug)

		err := config.FromConfigFile(configPath, ConfigDefaults)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("error loading configuration: %s", err),
				1)
		}

		// Enable setting conig values by environment variables
		config.Config.SetEnvPrefix("USERADM")
		config.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
		config.Config.AutomaticEnv()

		if config.Config.IsSet(SettingPlanDefinitions) {
			log.NewEmpty().Info(
				"loading plan definitions from configuration",
			)
			err := model.LoadPlans(
				config.Config.GetString(SettingPlanDefinitions),
			)
			if err != nil {
				return fmt.Errorf(
					"failed to load Plans from config: %w",
					err,
				)
			}
		}

		return nil
	}
	err := app.Run(args)
	if err != nil {
		log.NewEmpty().Fatal(err)
	}
	return err
}

func runServer(args *cli.Context) error {
	l := log.New(log.Ctx{})

	l.Print("User Administration Service starting up")

	ctx := context.Background()

	db, err := mongo.NewDataStoreMongo(dataStoreMongoConfigFromAppConfig(config.Config))
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("failed to connect to db: %v", err),
			2)
	}

	if args.Bool("automigrate") {
		db = db.WithAutomigrate()
	}

	if config.Config.Get(SettingTenantAdmAddr) != "" {
		db = db.WithMultitenant()
	}
	err = db.Migrate(ctx, mongo.DbVersion)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("failed to run migrations: %v", err),
			3)
	}

	err = RunServer(config.Config)
	if err != nil {
		return cli.NewExitError(err.Error(), 4)
	}
	return nil
}

func runCreateUser(args *cli.Context) error {
	email := model.Email(strings.ToLower(args.String("username")))
	err := commandCreateUser(
		config.Config,
		email,
		args.String("password"),
		args.String("user-id"),
		args.String("tenant-id"),
	)
	if err != nil {
		return cli.NewExitError(err.Error(), 5)
	}
	return nil
}

func runMigrate(args *cli.Context) error {
	err := commandMigrate(config.Config, args.String("tenant"))
	if err != nil {
		return cli.NewExitError(err.Error(), 6)
	}
	return nil
}

func runSetPassword(args *cli.Context) error {
	email := model.Email(strings.ToLower(args.String("username")))
	err := commandSetPassword(
		config.Config, email,
		args.String("password"), args.String("tenant-id"),
	)
	if err != nil {
		return cli.NewExitError(err.Error(), 7)
	}
	return nil
}
