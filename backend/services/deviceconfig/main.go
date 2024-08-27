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

package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/urfave/cli"

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"

	. "github.com/mendersoftware/mender-server/services/deviceconfig/config"
	"github.com/mendersoftware/mender-server/services/deviceconfig/server"
	"github.com/mendersoftware/mender-server/services/deviceconfig/store"
	"github.com/mendersoftware/mender-server/services/deviceconfig/store/mongo"
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
				Destination: &configPath,
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
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "tenant-id",
						Usage: "If an `ID` is provided, the migrations " +
							"will only apply to the specified tenant.",
					},
					&cli.StringFlag{
						Name:  "db-version",
						Value: mongo.DbVersion,
						Usage: "Target `VERSION` for the migration.",
					},
				},
			},
		},
	}
	app.Usage = "Device Configure"
	app.Version = "1.0.0"
	app.Action = cmdServer

	app.Before = func(args *cli.Context) error {
		err := config.FromConfigFile(configPath, Defaults)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("error loading configuration: %s", err),
				1)
		}

		// Enable setting config values by environment variables
		config.Config.SetEnvPrefix("DEVICECONFIG")
		config.Config.AutomaticEnv()
		config.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

		log.Setup(config.Config.GetBool(SettingDebugLog))

		return nil
	}

	err := app.Run(args)
	if err != nil {
		log.Log.Fatal(err)
	}
}

func initStoreFromConfig() (store.DataStore, error) {
	mgoURL, err := url.Parse(config.Config.GetString(SettingMongo))
	if err != nil {
		return nil, err
	}

	storeConfig := mongo.MongoStoreConfig{
		MongoURL: mgoURL,
		Username: config.Config.GetString(SettingDbUsername),
		Password: config.Config.GetString(SettingDbPassword),
		DbName:   mongo.DbName,
	}

	if config.Config.GetBool(SettingDbSSLSkipVerify) {
		storeConfig.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	return mongo.NewMongoStore(context.Background(), storeConfig)
}

func cmdServer(args *cli.Context) error {
	ctx := context.Background()
	ds, err := initStoreFromConfig()
	if err != nil {
		return err
	}
	defer ds.Close(ctx)
	err = ds.Migrate(ctx, mongo.DbVersion, args.Bool("automigrate"))
	if err != nil {
		return err
	}
	return server.InitAndRun(ds)
}

func cmdMigrate(args *cli.Context) error {
	ctx := context.Background()
	version := args.String("db-version")
	tenantID := args.String("tenant-id")
	if tenantID != "" {
		ctx = identity.WithContext(ctx, &identity.Identity{
			Tenant: tenantID,
		})
	}
	if version == "" {
		version = mongo.DbVersion
	}

	ds, err := initStoreFromConfig()
	if err != nil {
		return err
	}
	defer ds.Close(ctx)

	return ds.Migrate(ctx, version, true)
}
