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
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/version"

	"github.com/mendersoftware/mender-server/services/workflows/app/server"
	"github.com/mendersoftware/mender-server/services/workflows/app/worker"
	"github.com/mendersoftware/mender-server/services/workflows/client/nats"
	dconfig "github.com/mendersoftware/mender-server/services/workflows/config"
	"github.com/mendersoftware/mender-server/services/workflows/model"
	store "github.com/mendersoftware/mender-server/services/workflows/store/mongo"
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
				Usage: "Configuration `FILE`." +
					" Supports JSON, TOML, YAML and HCL formatted configs.",
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
				Name:   "worker",
				Usage:  "Run the worker process",
				Action: cmdWorker,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "automigrate",
						Usage: "Run database migrations before starting.",
					},
					&cli.StringFlag{
						Name:  "workflows",
						Usage: "Comma-separated list of workflows executed by this worker",
					},
					&cli.StringFlag{
						Name:  "excluded-workflows",
						Usage: "Comma-separated list of workflows NOT executed by this worker",
					},
				},
			},
			{
				Name:   "migrate",
				Usage:  "Run the migrations",
				Action: cmdMigrate,
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:   "skip-nats",
						Usage:  "Skip migrating the NATS Jetstream configuration",
						EnvVar: "WORKFLOWS_MIGRATION_SKIP_NATS",
					},
					cli.BoolFlag{
						Name:   "skip-database",
						Usage:  "Skip migrating the database",
						EnvVar: "WORKFLOWS_MIGRATION_SKIP_DATABASE",
					},
					cli.BoolFlag{
						Name:   "nats-force",
						Usage:  "Force the the consumers to migrate from push to pull mode",
						EnvVar: "WORKFLOWS_MIGRATION_FORCE_NATS",
					},
				},
			},
			{
				Name:   "list-jobs",
				Usage:  "List jobs",
				Action: cmdListJobs,
				Flags: []cli.Flag{
					cli.Int64Flag{
						Name:  "page",
						Usage: "page number to show",
					},
					cli.Int64Flag{
						Name:  "perPage",
						Usage: "number of results per page",
					},
				},
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
	app.Usage = "Workflows"
	app.Action = cmdServer

	app.Before = func(args *cli.Context) error {
		err := config.FromConfigFile(configPath, dconfig.Defaults)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("error loading configuration: %s", err),
				1)
		}

		// Enable setting config values by environment variables
		config.Config.SetEnvPrefix("WORKFLOWS")
		config.Config.AutomaticEnv()
		config.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

		return nil
	}

	err := app.Run(args)
	if err != nil {
		log.Fatal(err)
	}
}

func getNatsClient() (nats.Client, error) {
	natsURI := config.Config.GetString(dconfig.SettingNatsURI)
	streamName := config.Config.GetString(dconfig.SettingNatsStreamName)
	nats, err := nats.NewClientWithDefaults(natsURI, streamName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to nats")
	}
	return nats, err
}

func initJetstream(nc nats.Client, producer, upsert bool) (err error) {
	durableName := config.Config.GetString(dconfig.SettingNatsSubscriberDurable)
	replicas := config.Config.GetInt(dconfig.SettingNatsJetStreamReplicas)
	if producer {
		err = nc.JetStreamCreateStream(nc.StreamName(),
			nats.SetReplicas(replicas))
		if err != nil {
			return err
		}
	} else {
		var cfg nats.ConsumerConfig
		cfg, err = dconfig.GetNatsConsumerConfig(config.Config)
		if err != nil {
			return err
		}
		err = nc.CreateConsumer(durableName, upsert, cfg)
	}
	return err
}

func cmdServer(args *cli.Context) error {
	dataStore, err := store.SetupDataStore(args.Bool("automigrate"))
	if err != nil {
		return err
	}
	defer dataStore.Close()

	nats, err := getNatsClient()
	if err != nil {
		return err
	}
	defer nats.Close()

	if err = initJetstream(nats, true, args.Bool("automigrate")); err != nil {
		return errors.WithMessage(err, "failed to apply Jetstream migrations")
	}

	return server.InitAndRun(config.Config, dataStore, nats)
}

func cmdWorker(args *cli.Context) error {
	dataStore, err := store.SetupDataStore(args.Bool("automigrate"))
	if err != nil {
		return err
	}
	defer dataStore.Close()

	nats, err := getNatsClient()
	if err != nil {
		return err
	}
	defer nats.Close()

	if err = initJetstream(nats, false, args.Bool("automigrate")); err != nil {
		return errors.WithMessage(err, "failed to apply Jetstream consumer migrations")
	}

	var included, excluded []string

	includedWorkflows := args.String("workflows")
	if includedWorkflows != "" {
		included = strings.Split(includedWorkflows, ",")
	}

	excludedWorkflows := args.String("excluded-workflows")
	if excludedWorkflows != "" {
		excluded = strings.Split(excludedWorkflows, ",")
	}

	workflows := worker.Workflows{
		Included: included,
		Excluded: excluded,
	}
	return worker.InitAndRun(config.Config, workflows, dataStore, nats)
}

func cmdMigrate(args *cli.Context) error {
	var err error
	if !args.Bool("skip-database") {
		_, err = store.SetupDataStore(true)
		if err != nil {
			return err
		}
	}
	if !args.Bool("skip-nats") {
		var nc nats.Client
		nc, err = getNatsClient()
		if err != nil {
			return err
		}
		if args.Bool("nats-force") {
			durableName := config.Config.GetString(dconfig.SettingNatsSubscriberDurable)
			// delete the consumer if it is push based
			err := nc.DeleteConsumerByMode(durableName, nats.PushMode)
			if err != nil {
				return err
			}
		}
		if err = initJetstream(nc, true, true); err == nil {
			err = initJetstream(nc, false, true)
		}

	}
	return err
}

func cmdListJobs(args *cli.Context) error {
	dataStore, err := store.SetupDataStore(false)
	if err != nil {
		return err
	}
	defer dataStore.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var page int64
	var perPage int64
	page = args.Int64("page")
	perPage = args.Int64("perPage")

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 75
	}
	jobs, count, _ := dataStore.GetAllJobs(ctx, page, perPage)
	fmt.Printf("all jobs: %d; page: %d/%d perPage:%d\n%29s %24s %10s %s\n",
		count, page, count/perPage, perPage, "insert time", "id", "status", "workflow")
	for _, j := range jobs {
		format := "Mon, 2 Jan 2006 15:04:05 MST"
		fmt.Printf(
			"%29s %24s %10s %s\n",
			j.InsertTime.Format(format),
			j.ID, model.StatusToString(j.Status),
			j.WorkflowName,
		)
	}
	fmt.Printf("all jobs: %d; page: %d/%d\n", count, page, count/perPage)

	return nil
}
