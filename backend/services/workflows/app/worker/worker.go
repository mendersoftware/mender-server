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

package worker

import (
	"context"
	"os"
	"os/signal"
	"time"

	natsio "github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/executor"
	"github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/workflows/client/nats"
	dconfig "github.com/mendersoftware/mender-server/services/workflows/config"
	"github.com/mendersoftware/mender-server/services/workflows/store"
)

// Workflows filters workflows executed by a worker
type Workflows struct {
	Included []string
	Excluded []string
}

// InitAndRun initializes the worker and runs it
func InitAndRun(
	conf config.Reader,
	workflows Workflows,
	dataStore store.DataStore,
	natsClient nats.Client,
) error {
	ctx, cancel := context.WithCancel(context.Background())
	// Calling cancel() before returning should shut down
	// all workers. However, the new driver is not
	// particularly good at listening to the context in the
	// current state, but it'll be forced to shut down
	// eventually.
	defer cancel()

	if conf.GetBool(dconfig.SettingDebugLog) {
		log.Setup(true)
	}
	l := log.FromContext(ctx)

	err := dataStore.LoadWorkflows(ctx, l)
	if err != nil {
		return errors.Wrap(err, "failed to load workflows")
	}

	streamName := config.Config.GetString(dconfig.SettingNatsStreamName)
	topic := config.Config.GetString(dconfig.SettingNatsSubscriberTopic)
	subject := streamName + "." + topic
	durableName := config.Config.GetString(dconfig.SettingNatsSubscriberDurable)
	concurrency := conf.GetInt(dconfig.SettingConcurrency)

	allowedBinaries := config.Config.GetStringSlice(dconfig.SettingWorkerAllowedBinaries)
	binExec := executor.New(allowedBinaries)

	cfg, err := natsClient.GetConsumerConfig(durableName)
	if err != nil {
		return err
	}
	notifyPeriod := cfg.AckWait * 8 / 10
	if notifyPeriod < time.Second {
		notifyPeriod = time.Second
	}

	jobChan := make(chan *natsio.Msg, concurrency)
	sub, err := natsClient.JetStreamSubscribe(
		ctx,
		subject,
		durableName,
		jobChan,
	)
	if err != nil {
		return errors.Wrap(err, "failed to subscribe to the nats JetStream")
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, unix.SIGINT, unix.SIGTERM)

	// Spawn worker pool
	wg := NewWorkGroup(jobChan, notifyPeriod, natsClient, dataStore, sub, binExec)
	for i := 0; i < concurrency; i++ {
		go wg.RunWorker(ctx)
	}

	// Run until a SIGTERM or SIGINT is received or one of the workers
	// stops unexpectedly
	select {
	case <-wg.FirstDone():
		l.Warnf("worker %d terminated, application is terminating", wg.TermID())
	case sig := <-quit:
		l.Warnf("received signal %s: terminating workers", sig)
	case <-ctx.Done():
		err = ctx.Err()
	}
	errSub := sub.Unsubscribe()
	if errSub != nil {
		l.Errorf("error unsubscribing from Jetstream: %s", errSub.Error())
	}
	// Notify workers that we're done
	close(jobChan)
	if err == nil {
		l.Infof("waiting up to %s for all workers to finish", cfg.AckWait)
		select {
		case <-wg.Done():
		case sig := <-quit:
			l.Warnf("received signal %s while waiting: aborting", sig)
		case <-time.After(cfg.AckWait):
		}
	}
	return err
}
