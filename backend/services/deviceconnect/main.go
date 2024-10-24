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
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli"

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/version"

	dconfig "github.com/mendersoftware/mender-server/services/deviceconnect/config"
	"github.com/mendersoftware/mender-server/services/deviceconnect/server"
	store "github.com/mendersoftware/mender-server/services/deviceconnect/store/mongo"
)

var appVersion = version.Get()

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
		},
		Version: appVersion.Version,
	}
	app.Usage = "Device Connect"
	app.Action = cmdServer

	app.Before = func(args *cli.Context) error {
		err := config.FromConfigFile(configPath, dconfig.Defaults)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("error loading configuration: %s", err),
				1)
		}

		// Enable setting config values by environment variables
		config.Config.SetEnvPrefix("DEVICECONNECT")
		config.Config.AutomaticEnv()
		config.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

		return nil
	}

	err := app.Run(args)
	if err != nil {
		log.Fatal(err)
	}
}

func cmdServer(args *cli.Context) error {
	dataStore, err := store.SetupDataStore(args.Bool("automigrate"))
	if err != nil {
		return err
	}
	defer dataStore.Close()
	return server.InitAndRun(config.Config, dataStore)
}

func cmdMigrate(args *cli.Context) error {
	_, err := store.SetupDataStore(true)
	if err != nil {
		return err
	}
	return nil
}
