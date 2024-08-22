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
	"os"
	"time"

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"
	mstore "github.com/mendersoftware/mender-server/pkg/store"
	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"github.com/mendersoftware/mender-server/services/deployments/app"
	"github.com/mendersoftware/mender-server/services/deployments/client/workflows"
	dconfig "github.com/mendersoftware/mender-server/services/deployments/config"
	"github.com/mendersoftware/mender-server/services/deployments/store"
	"github.com/mendersoftware/mender-server/services/deployments/store/mongo"
)

const (
	deviceDeploymentsBatchSize = 512

	cliDefaultRateLimit = 50
)

func main() {
	doMain(os.Args)
}

func doMain(args []string) {

	var configPath string

	app := cli.NewApp()
	app.Usage = "Deployments Service"

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
			Name:  "propagate-reporting",
			Usage: "Trigger a reindex of all the device deployments in the reporting services ",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "tenant_id",
					Usage: "Tenant ID (optional) - propagate for just a single tenant.",
				},
				cli.UintFlag{
					Name:  "rate-limit",
					Usage: "`N`umber of reindexing batch requests per second",
					Value: cliDefaultRateLimit,
				},
				cli.BoolFlag{
					Name: "dry-run",
					Usage: "Do not perform any modifications," +
						" just scan and print devices.",
				},
			},

			Action: cmdPropagateReporting,
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

func cmdPropagateReporting(args *cli.Context) error {
	if config.Config.GetString(dconfig.SettingReportingAddr) == "" {
		return cli.NewExitError(errors.New("reporting address not configured"), 1)
	}
	c := config.Config
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Second*30,
	)
	defer cancel()
	dbClient, err := mongo.NewMongoClient(ctx, c)
	if err != nil {
		return err
	}
	defer func() {
		_ = dbClient.Disconnect(context.Background())
	}()

	db := mongo.NewDataStoreMongoWithClient(dbClient)

	wflows := workflows.NewClient()

	var requestPeriod time.Duration
	rateLimit := args.Uint("rate-limit")
	if rateLimit > 0 {
		requestPeriod = time.Second / time.Duration(args.Uint("rate-limit"))
	}

	err = propagateReporting(
		db,
		wflows,
		args.String("tenant_id"),
		requestPeriod,
		args.Bool("dry-run"),
	)
	if err != nil {
		return cli.NewExitError(err, 7)
	}
	return nil
}

func propagateReporting(
	db store.DataStore,
	wflows workflows.Client,
	tenant string,
	requestPeriod time.Duration,
	dryRun bool,
) error {
	l := log.NewEmpty()

	dbs, err := selectDbs(db, tenant)
	if err != nil {
		return errors.Wrap(err, "aborting")
	}

	var errReturned error
	for _, d := range dbs {
		err := tryPropagateReportingForDb(db, wflows, d, requestPeriod, dryRun)
		if err != nil {
			errReturned = err
			l.Errorf("giving up on DB %s due to fatal error: %s", d, err.Error())
			continue
		}
	}

	l.Info("all DBs processed, exiting.")
	return errReturned
}

func selectDbs(db store.DataStore, tenant string) ([]string, error) {
	l := log.NewEmpty()

	var dbs []string

	if tenant != "" {
		l.Infof("propagating deployments history for user-specified tenant %s", tenant)
		n := mstore.DbNameForTenant(tenant, mongo.DbName)
		dbs = []string{n}
	} else {
		l.Infof("propagating deployments history for all tenants")

		// infer if we're in ST or MT
		tdbs, err := db.GetTenantDbs()
		if err != nil {
			return nil, errors.Wrap(err, "failed to retrieve tenant DBs")
		}

		if len(tdbs) == 0 {
			l.Infof("no tenant DBs found - will try the default database %s", mongo.DbName)
			dbs = []string{mongo.DbName}
		} else {
			dbs = tdbs
		}
	}

	return dbs, nil
}

func tryPropagateReportingForDb(
	db store.DataStore,
	wflows workflows.Client,
	dbname string,
	requestPeriod time.Duration,
	dryRun bool,
) error {
	l := log.NewEmpty()

	l.Infof("propagating deployments data to reporting from DB: %s", dbname)

	tenant := mstore.TenantFromDbName(dbname, mongo.DbName)

	ctx := context.Background()
	if tenant != "" {
		ctx = identity.WithContext(ctx, &identity.Identity{
			Tenant: tenant,
		})
	}

	err := reindexDeploymentsReporting(ctx, db, wflows, tenant, requestPeriod, dryRun)
	if err != nil {
		l.Infof("Done with DB %s, but there were errors: %s.", dbname, err.Error())
	} else {
		l.Infof("Done with DB %s", dbname)
	}

	return err
}

func reindexDeploymentsReporting(
	ctx context.Context,
	db store.DataStore,
	wflows workflows.Client,
	tenant string,
	requestPeriod time.Duration,
	dryRun bool,
) error {
	var skip int

	done := ctx.Done()
	ticker := time.NewTicker(requestPeriod)
	defer ticker.Stop()
	skip = 0
	for {
		dd, err := db.GetDeviceDeployments(ctx, skip, deviceDeploymentsBatchSize, "", nil, true)
		if err != nil {
			return errors.Wrap(err, "failed to get device deployments")
		}

		if len(dd) < 1 {
			break
		}

		if !dryRun {
			deviceDeployments := make([]workflows.DeviceDeploymentShortInfo, len(dd))
			for i, d := range dd {
				deviceDeployments[i].ID = d.Id
				deviceDeployments[i].DeviceID = d.DeviceId
				deviceDeployments[i].DeploymentID = d.DeploymentId
			}
			err := wflows.StartReindexReportingDeploymentBatch(ctx, deviceDeployments)
			if err != nil {
				return err
			}
		}

		skip += deviceDeploymentsBatchSize
		if len(dd) < deviceDeploymentsBatchSize {
			break
		}
		select {
		case <-ticker.C:

		case <-done:
			return ctx.Err()
		}
	}
	return nil
}
