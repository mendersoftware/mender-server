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

	api "github.com/mendersoftware/mender-server/services/deviceconfig/api/http"
	"github.com/mendersoftware/mender-server/services/deviceconfig/app"
	"github.com/mendersoftware/mender-server/services/deviceconfig/client/workflows"
	. "github.com/mendersoftware/mender-server/services/deviceconfig/config"
	"github.com/mendersoftware/mender-server/services/deviceconfig/store"
)

// InitAndRun initializes the server and runs it
func InitAndRun(dataStore store.DataStore) error {
	ctx := context.Background()

	l := log.FromContext(ctx)
	wflows := workflows.NewClient(
		config.Config.GetString(SettingWorkflowsURL),
	)
	appl := app.New(
		dataStore, wflows, app.Config{
			HaveAuditLogs: config.Config.GetBool(SettingEnableAudit),
		},
	)

	router := api.NewRouter(appl)

	var listen = config.Config.GetString(SettingListen)
	srv := &http.Server{
		Addr:    listen,
		Handler: router,
	}

	go func() {
		l.Infof("Server listening for connections on \"%s\"", listen)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, unix.SIGINT, unix.SIGTERM)
	<-quit

	l.Info("Server shutting down")

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxWithTimeout); err != nil {
		l.Fatal("error when shutting down the server ", err)
	}

	l.Info("Server exited")
	return nil
}
