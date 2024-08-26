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

package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"golang.org/x/sys/unix"

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/log"

	api "github.com/mendersoftware/mender-server/services/deviceconnect/api/http"
	"github.com/mendersoftware/mender-server/services/deviceconnect/app"
	"github.com/mendersoftware/mender-server/services/deviceconnect/client/inventory"
	"github.com/mendersoftware/mender-server/services/deviceconnect/client/nats"
	"github.com/mendersoftware/mender-server/services/deviceconnect/client/workflows"
	dconfig "github.com/mendersoftware/mender-server/services/deviceconnect/config"
	"github.com/mendersoftware/mender-server/services/deviceconnect/store"
)

// InitAndRun initializes the server and runs it
func InitAndRun(conf config.Reader, dataStore store.DataStore) error {
	ctx := context.Background()

	log.Setup(conf.GetBool(dconfig.SettingDebugLog))
	l := log.FromContext(ctx)

	allowedOrigin := conf.GetStringSlice(dconfig.SettingWSAllowedOrigins)
	if allowedOrigin != nil {
		api.SetAcceptedOrigins(allowedOrigin)
	}

	natsClient, err := nats.NewClientWithDefaults(
		config.Config.GetString(dconfig.SettingNatsURI),
	)
	if err != nil {
		return err
	}
	inventory := inventory.NewClient(
		config.Config.GetString(dconfig.SettingInventoryURI),
		config.Config.GetInt(dconfig.SettingInventoryTimeout),
	)
	wflows := workflows.NewClient(
		config.Config.GetString(dconfig.SettingWorkflowsURL),
	)
	deviceConnectApp := app.New(
		dataStore, inventory,
		wflows, app.Config{
			HaveAuditLogs: conf.GetBool(dconfig.SettingEnableAuditLogs),
		},
	)

	gracefulShutdownTimeout := conf.GetDuration(dconfig.SettingGracefulShutdownTimeout)
	router, err := api.NewRouter(deviceConnectApp, natsClient, &api.RouterConfig{
		GracefulShutdownTimeout: gracefulShutdownTimeout,
	})
	if err != nil {
		l.Fatal(err)
	}

	var listen = conf.GetString(dconfig.SettingListen)
	srv := &http.Server{
		Addr:    listen,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, unix.SIGINT, unix.SIGTERM, unix.SIGUSR1)
	recvSignal := <-quit

	l.Info("server shutdown")

	if recvSignal == unix.SIGUSR1 {
		l.Info("received SIGUSR1, graceful shutdown")
		srv.RegisterOnShutdown(func() {
			deviceConnectApp.Shutdown(gracefulShutdownTimeout)
		})
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxWithTimeout); err != nil {
		l.Fatal("error when shutting down the server ", err)
	}
	l.Info("server exited")

	if recvSignal == unix.SIGUSR1 {
		deviceConnectApp.ShutdownDone()
		l.Info("graceful shutdown completed")
	}

	return nil
}
