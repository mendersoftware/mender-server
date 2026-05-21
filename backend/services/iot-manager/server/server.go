// Copyright 2024 Northern.tech AS
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

	oas "github.com/mendersoftware/mender-server/pkg/api"
	openapi "github.com/mendersoftware/mender-server/pkg/api/client"
	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/log"

	api "github.com/mendersoftware/mender-server/services/iot-manager/api/http"
	"github.com/mendersoftware/mender-server/services/iot-manager/app"
	"github.com/mendersoftware/mender-server/services/iot-manager/client/iotcore"
	"github.com/mendersoftware/mender-server/services/iot-manager/client/iothub"
	dconfig "github.com/mendersoftware/mender-server/services/iot-manager/config"
	"github.com/mendersoftware/mender-server/services/iot-manager/store"
)

// InitAndRun initializes the server and runs it
func InitAndRun(conf config.Reader, dataStore store.DataStore) error {
	ctx := context.Background()
	httpClient := new(http.Client)
	wfCfg, err := oas.NewDefaultClientConfigurationFromURL(
		conf.GetString(dconfig.SettingWorkflowsURL),
	)
	if err != nil {
		return fmt.Errorf("failed to initialize workflows client: %w", err)
	}
	wf := openapi.NewAPIClient(wfCfg).WorkflowsOtherAPI
	hub := iothub.NewClient(iothub.NewOptions().SetClient(httpClient))
	core := iotcore.NewClient()

	log.Setup(conf.GetBool(dconfig.SettingDebugLog))
	l := log.FromContext(ctx)

	daCfg, err := oas.NewDefaultClientConfigurationFromURL(
		conf.GetString(dconfig.SettingDeviceauthURL),
	)
	if err != nil {
		return fmt.Errorf("failed to initialize deviceauth client: %w", err)
	}
	da := openapi.NewAPIClient(daCfg).DeviceAuthenticationInternalAPIAPI

	azureIotManagerApp := app.New(dataStore, wf, da).WithIoTHub(hub).WithIoTCore(core)
	azureIotManagerApp = azureIotManagerApp.
		WithWebhooksTimeout(config.Config.GetUint(dconfig.SettingWebhooksTimeoutSeconds))

	router := api.NewRouter(azureIotManagerApp,
		api.NewConfig().
			SetClient(httpClient).
			SetMaxRequestSize(int64(conf.GetInt(dconfig.SettingMaxRequestSize))),
	)

	var listen = conf.GetString(dconfig.SettingListen)
	srv := &http.Server{
		Addr:    listen,
		Handler: router,
	}

	l.Info("IoT Manager service starting up")
	l.Infof("listening on %s", listen)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, unix.SIGINT, unix.SIGTERM)
	<-quit

	l.Info("server shutdown")

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxWithTimeout); err != nil {
		l.Errorf("error when shutting down the server: %s", err.Error())
		return err
	}

	l.Info("server exiting")
	return nil
}
