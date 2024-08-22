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
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/log"
	mlog "github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/reporting/app/indexer"
	"github.com/mendersoftware/mender-server/services/reporting/app/server"
	"github.com/mendersoftware/mender-server/services/reporting/client/nats"
	dconfig "github.com/mendersoftware/mender-server/services/reporting/config"
	"github.com/mendersoftware/mender-server/services/reporting/store"
	"github.com/mendersoftware/mender-server/services/reporting/store/mongo"
	"github.com/mendersoftware/mender-server/services/reporting/store/opensearch"
)

const (
	opensearchMaxWaitingTime      = 300
	opensearchRetryDelayInSeconds = 1
)

func main() {
	os.Exit(doMain(os.Args))
}

func doMain(args []string) int {
	var configPath string

	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "config",
				Usage: "Configuration `FILE`. " +
					"Supports JSON, TOML, YAML and HCL formatted configs.",
				Value:       "config.yaml",
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
				Name:   "indexer",
				Usage:  "Run the indexer process",
				Action: cmdIndexer,
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
		},
	}
	app.Usage = "Reporting"
	app.Action = cmdServer

	app.Before = func(args *cli.Context) error {
		err := config.FromConfigFile(configPath, dconfig.Defaults)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("error loading configuration: %s", err),
				1)
		}

		// Enable setting config values by environment variables
		config.Config.SetEnvPrefix("REPORTING")
		config.Config.AutomaticEnv()
		config.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

		// setup logging
		mlog.Setup(config.Config.GetBool(dconfig.SettingDebugLog))

		return nil
	}

	err := app.Run(args)
	if err != nil {
		mlog.NewEmpty().Fatal(err)
		return 1
	}
	return 0
}

func cmdServer(args *cli.Context) error {
	store, err := getStore(args)
	if err != nil {
		return err
	}
	ctx := context.Background()
	ds, err := getDatastore(args)
	if err != nil {
		return err
	}
	defer ds.Close(ctx)
	if args.Bool("automigrate") {
		nats, err := getNatsClient()
		if err != nil {
			return err
		}
		ctx := context.Background()
		err = migrate(ctx, store, ds, nats)
		nats.Close()
		if err != nil {
			return err
		}
	}
	return server.InitAndRun(config.Config, store, ds)
}

func getNatsClient() (nats.Client, error) {
	natsURI := config.Config.GetString(dconfig.SettingNatsURI)
	nats, err := nats.NewClient(natsURI)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to nats")
	}
	return nats, nil
}

func cmdIndexer(args *cli.Context) error {
	store, err := getStore(args)
	if err != nil {
		return err
	}
	nats, err := getNatsClient()
	if err != nil {
		return err
	}
	defer nats.Close()
	ctx := context.Background()
	ds, err := getDatastore(args)
	if err != nil {
		return err
	}
	defer ds.Close(ctx)
	if args.Bool("automigrate") {
		err = migrate(ctx, store, ds, nats)
		if err != nil {
			return err
		}

	}
	return indexer.InitAndRun(config.Config, store, ds, nats)
}

func cmdMigrate(args *cli.Context) error {
	ctx := context.Background()
	store, err := getStore(args)
	if err != nil {
		return err
	}
	ds, err := getDatastore(args)
	if err != nil {
		return err
	}
	nats, err := getNatsClient()
	if err != nil {
		return err
	}
	return migrate(ctx, store, ds, nats)
}

func migrate(ctx context.Context, store store.Store, ds store.DataStore, nats nats.Client) error {
	err := store.Migrate(ctx)
	if err != nil {
		return err
	}
	err = ds.Migrate(ctx, mongo.DbVersion, true)
	if err != nil {
		return err
	}
	dur := config.Config.GetString(dconfig.SettingNatsSubscriberDurable)
	stream := config.Config.GetString(dconfig.SettingNatsStreamName)
	topic := config.Config.GetString(dconfig.SettingNatsSubscriberTopic)
	sub := stream + "." + topic
	return nats.Migrate(ctx, sub, dur, true)
}

func getStore(args *cli.Context) (store.Store, error) {
	addresses := config.Config.GetStringSlice(dconfig.SettingOpenSearchAddresses)
	devicesIndexName := config.Config.GetString(dconfig.SettingOpenSearchDevicesIndexName)
	devicesIndexShards := config.Config.GetInt(dconfig.SettingOpenSearchDevicesIndexShards)
	devicesIndexReplicas := config.Config.GetInt(
		dconfig.SettingOpenSearchDevicesIndexReplicas)
	deploymentsIndexName := config.Config.GetString(dconfig.SettingOpenSearchDeploymentsIndexName)
	deploymentsIndexShards := config.Config.GetInt(dconfig.SettingOpenSearchDeploymentsIndexShards)
	deploymentsIndexReplicas := config.Config.GetInt(
		dconfig.SettingOpenSearchDeploymentsIndexReplicas)
	store, err := opensearch.NewStore(
		opensearch.WithServerAddresses(addresses),
		opensearch.WithDevicesIndexName(devicesIndexName),
		opensearch.WithDevicesIndexShards(devicesIndexShards),
		opensearch.WithDevicesIndexReplicas(devicesIndexReplicas),
		opensearch.WithDeploymentsIndexName(deploymentsIndexName),
		opensearch.WithDeploymentsIndexShards(deploymentsIndexShards),
		opensearch.WithDeploymentsIndexReplicas(deploymentsIndexReplicas),
	)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	l := log.FromContext(ctx)
	for i := 0; i < opensearchMaxWaitingTime; i++ {
		err = store.Ping(ctx)
		if err == nil {
			break
		}
		l.Warn(err)
		time.Sleep(opensearchRetryDelayInSeconds * time.Second)
	}
	if err != nil {
		l.Error(err)
		return nil, err
	}
	l.Info("successfully connected to OpenSearch")
	return store, nil
}

func getDatastore(args *cli.Context) (store.DataStore, error) {
	mgoURL, err := url.Parse(config.Config.GetString(dconfig.SettingMongo))
	if err != nil {
		return nil, err
	}

	storeConfig := mongo.MongoStoreConfig{
		MongoURL:      mgoURL,
		SSL:           config.Config.GetBool(dconfig.SettingDbSSL),
		SSLSkipVerify: config.Config.GetBool(dconfig.SettingDbSSLSkipVerify),
		Username:      config.Config.GetString(dconfig.SettingDbUsername),
		Password:      config.Config.GetString(dconfig.SettingDbPassword),
		DbName:        mongo.DbName,
	}

	return mongo.NewMongoStore(context.Background(), storeConfig)
}
