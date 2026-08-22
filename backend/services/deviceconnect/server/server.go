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
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"golang.org/x/sys/unix"

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/log"

	api "github.com/mendersoftware/mender-server/services/deviceconnect/api/http"
	"github.com/mendersoftware/mender-server/services/deviceconnect/app"
	"github.com/mendersoftware/mender-server/services/deviceconnect/client/nats"
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

	lim, err := dconfig.LoadReadiness(conf)
	if err != nil {
		return fmt.Errorf("failed to load memory limits: %w", err)
	}
	deviceConnectApp := app.New(
		dataStore, func(c *app.Config) {
			c.ReadinessLimits = lim
		},
	)

	gracefulShutdownTimeout := conf.GetDuration(dconfig.SettingGracefulShutdownTimeout)
	router, err := api.NewRouter(deviceConnectApp, natsClient, &api.RouterConfig{
		MaxRequestSize: config.Config.GetInt64(dconfig.SettingMaxRequestSize),
		MaxFileSize:    config.Config.GetInt64(dconfig.SettingMaxFileUploadSize),
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
	sigusr1 := make(chan os.Signal, 1)
	signal.Notify(quit, unix.SIGINT, unix.SIGTERM)
	signal.Notify(sigusr1, unix.SIGUSR1)
	var recvSignal os.Signal
	select {
	case recvSignal = <-sigusr1:
	case recvSignal = <-quit:
	}
	signal.Stop(quit) // Restore signal handlers for SIGTERM/SIGQUIT

	l.Infof("received signal %s, server shutting down", recvSignal)

	gracePeriod := time.Duration(0)
	if recvSignal == unix.SIGUSR1 {
		l.Info("received SIGUSR1, graceful shutdown")
		gracePeriod = gracefulShutdownTimeout
	}
	timeout := 5*time.Second + gracePeriod
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	errChan := make(chan error, 1)
	srv.RegisterOnShutdown(func() { errChan <- deviceConnectApp.Shutdown(ctx, gracePeriod) })
	err = srv.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("error shutting down server: %w", err)
	}
	select {
	case <-deviceConnectApp.Done():
		// Give the last websocket sessions a second to complete the handshake
		select {
		case err = <-errChan:
		case <-time.After(time.Second):
		case <-ctx.Done():
			err = fmt.Errorf("timeout waiting for server to shutdown: %w", err)
		}
	case <-ctx.Done():
		err = fmt.Errorf("timeout waiting for server to shutdown: %w", err)
	}
	return err
}
