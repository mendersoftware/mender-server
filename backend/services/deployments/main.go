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

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli"

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/log"
	mstore "github.com/mendersoftware/mender-server/pkg/store"
	"github.com/mendersoftware/mender-server/pkg/version"

	"github.com/mendersoftware/mender-server/services/deployments/app"
	dconfig "github.com/mendersoftware/mender-server/services/deployments/config"
	"github.com/mendersoftware/mender-server/services/deployments/store/mongo"
)

var appVersion = version.Get()

func main() {
	doMain(os.Args)
}

func doMain(args []string) {

	var configPath string

	app := cli.NewApp()
	app.Usage = "Deployments Service"
	app.Version = appVersion.Version

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "config",
			Usage: "Configuration `FILE`." +
				" Supports JSON, TOML, YAML and HCL formatted configs.",
			Destination: &configPath,
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
			Usage: "Run migrations and exit",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "tenant",
					Usage: "Tenant ID (optional).",
				},
			},

			Action: cmdMigrate,
		},
		{
			Name:  "storage-daemon",
			Usage: "Start storage daemon cleaning up expired objects from storage",
			Flags: []cli.Flag{
				cli.DurationFlag{
					Name: "interval",
					Usage: "Time interval to run cleanup routine; " +
						"a value of 0 runs the daemon for one " +
						"iteration and terminates (cron mode).",
					Value: 0,
				},
				cli.DurationFlag{
					Name: "time-jitter",
					Usage: "The time jitter added for expired links. " +
						"Links must be expired for `DURATION` " +
						"to be removed.",
					Value: time.Second * 3,
				},
			},
			Action: cmdStorageDaemon,
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

	app.Action = cmdServer
	app.Before = func(args *cli.Context) error {
		if err := dconfig.Setup(configPath); err != nil {
			return cli.NewExitError(err.Error(), 1)
		}

		return nil
	}

	err := app.Run(args)
	if err != nil {
		log.NewEmpty().Fatal(err.Error())
	}
}

func cmdServer(args *cli.Context) error {
	devSetup := args.GlobalBool("dev")

	l := log.New(log.Ctx{})

	if devSetup {
		l.Infof("setting up development configuration")
		config.Config.Set(dconfig.SettingMiddleware, dconfig.EnvDev)
	}

	l.Print("Deployments Service starting up")
	err := migrate("", args.Bool("automigrate"))
	if err != nil {
		return err
	}

	setupContext, cancel := context.WithTimeout(
		context.Background(),
		time.Second*30,
	)
	err = RunServer(setupContext)
	cancel()
	if err != nil {
		return cli.NewExitError(err.Error(), 4)
	}

	return nil
}

func cmdMigrate(args *cli.Context) error {
	tenant := args.String("tenant")
	return migrate(tenant, true)
}

func migrate(tenant string, automigrate bool) error {
	ctx := context.Background()

	dbClient, err := mongo.NewMongoClient(ctx, config.Config)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("failed to connect to db: %v", err),
			3)
	}
	defer func() {
		_ = dbClient.Disconnect(ctx)
	}()

	dbVersion := mongo.DbVersion
	if !automigrate {
		dbVersion = mongo.DbMinimumVersion
	}

	if tenant != "" {
		db := mstore.DbNameForTenant(tenant, mongo.DbName)
		err = mongo.MigrateSingle(ctx, db, dbVersion, dbClient, automigrate)
	} else {
		err = mongo.Migrate(ctx, dbVersion, dbClient, automigrate)
	}
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("failed to run migrations: %v", err),
			3)
	}

	return nil
}

func cmdStorageDaemon(args *cli.Context) error {
	ctx := context.Background()
	objectStorage, err := SetupObjectStorage(ctx)
	if err != nil {
		return err
	}
	mgo, err := mongo.NewMongoClient(ctx, config.Config)
	if err != nil {
		return err
	}
	database := mongo.NewDataStoreMongoWithClient(mgo)
	app := app.NewDeployments(database, objectStorage, 0, false)
	return app.CleanupExpiredUploads(
		ctx,
		args.Duration("interval"),
		args.Duration("time-jitter"),
	)
}
