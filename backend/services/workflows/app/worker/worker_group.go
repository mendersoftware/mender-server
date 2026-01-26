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
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	natsio "github.com/nats-io/nats.go"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/executor"
	"github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/workflows/app/processor"
	"github.com/mendersoftware/mender-server/services/workflows/client/nats"
	"github.com/mendersoftware/mender-server/services/workflows/model"
	"github.com/mendersoftware/mender-server/services/workflows/store"
)

type workerGroup struct {
	mu        sync.Mutex    // Mutex to protect the shared firstDone channel
	firstDone chan struct{} // First worker finished
	done      chan struct{} // All workers finished
	termID    int32         // ID of the first worker to finish

	workerCount int32

	input        <-chan *natsio.Msg
	notifyPeriod time.Duration

	sub     *natsio.Subscription
	client  nats.Client
	store   store.DataStore
	binExec executor.BinaryExecutor
}

func NewWorkGroup(
	input <-chan *natsio.Msg,
	notifyPeriod time.Duration,
	nc nats.Client,
	ds store.DataStore,
	sub *natsio.Subscription,
	be executor.BinaryExecutor,
) *workerGroup {
	return &workerGroup{
		done:      make(chan struct{}),
		firstDone: make(chan struct{}),

		input:        input,
		notifyPeriod: notifyPeriod,
		client:       nc,
		store:        ds,
		sub:          sub,
		binExec:      be,
	}
}

// Done returns a channel (barrier) that is closed when the last worker
// has exited.
func (w *workerGroup) Done() <-chan struct{} {
	return w.done
}

// FirstDone returns a channel (barrier) that is closed when the first
// worker has exited.
func (w *workerGroup) FirstDone() <-chan struct{} {
	return w.firstDone
}

// TermID is the ID of the first worker that quit.
func (w *workerGroup) TermID() int32 {
	return w.termID
}

func (w *workerGroup) RunWorker(ctx context.Context) {
	id := atomic.AddInt32(&w.workerCount, 1)
	l := log.FromContext(ctx)
	l = l.F(log.Ctx{"worker_id": id})
	ctx = log.WithContext(ctx, l)

	sidecarChan := make(chan *natsio.Msg, 1)
	sidecarDone := make(chan struct{})
	defer func() {
		l.Info("worker shutting down")
		close(sidecarChan)

		w.mu.Lock()
		remaining := atomic.AddInt32(&w.workerCount, -1)
		// Is this the last worker to quit?
		if remaining <= 0 {
			select {
			case <-w.done:

			default:
				close(w.done)
			}
		} else {
			// Is this the first worker to quit?
			select {
			case <-w.firstDone:

			default:
				w.termID = id
				close(w.firstDone)
			}
		}
		w.mu.Unlock()
	}()
	l.Info("worker starting up")
	// workerSidecar is responsible for notifying the broker about slow workflows
	go w.workerSidecar(ctx, sidecarChan, sidecarDone)
	switch w.sub.Type() {
	case natsio.PullSubscription:
		w.pullWorkerMain(ctx, sidecarChan, sidecarDone)
	case natsio.ChanSubscription:
		w.workerMain(ctx, sidecarChan, sidecarDone)
	default:
		l.Error("unsupported consumer type")
	}
}

func (w *workerGroup) pullWorkerMain(
	ctx context.Context,
	sidecarChan chan *natsio.Msg,
	sidecarDone chan struct{},
) {
	l := log.FromContext(ctx)
	timeoutTimer := newStoppedTimer()
	for {
		var msg *natsio.Msg
		select {
		case <-ctx.Done():
			return
		case <-sidecarDone:
			return
		default:
			msgs, err := w.sub.Fetch(1, natsio.MaxWait(1*time.Second))
			if err != nil {
				if errors.Is(err, natsio.ErrTimeout) {
					continue
				}
				l.Errorf("failed to fetch message: %s", err)
				return
			}
			if len(msgs) == 0 {
				continue
			}
			msg = msgs[0]
		}

		stop := w.doWokerJob(ctx, sidecarChan, sidecarDone, timeoutTimer, msg)
		if stop {
			return
		}
	}
}

func (w *workerGroup) workerMain(
	ctx context.Context,
	sidecarChan chan *natsio.Msg,
	sidecarDone chan struct{},
) {
	ctxDone := ctx.Done()
	timeoutTimer := newStoppedTimer()
	for {
		var (
			msg    *natsio.Msg
			isOpen bool
		)
		select {
		case msg, isOpen = <-w.input:
			if !isOpen {
				return
			}

		case <-sidecarDone:
			return
		case <-ctxDone:
			return
		}
		stop := w.doWokerJob(ctx, sidecarChan, sidecarDone, timeoutTimer, msg)
		if stop {
			return
		}
	}
}

func (w *workerGroup) doWokerJob(
	ctx context.Context,
	sidecarChan chan *natsio.Msg,
	sidecarDone chan struct{},
	timeoutTimer *reusableTimer,
	msg *natsio.Msg) (stop bool) {
	l := log.FromContext(ctx)
	ctxDone := ctx.Done()
	// Notify the sidecar routine about the new message
	select {
	case sidecarChan <- msg:

	case <-timeoutTimer.After(w.notifyPeriod / 8):
		l.Warn("timeout notifying sidecar routine about message")

	case <-sidecarDone:
		return true
	case <-ctxDone:
		return true
	}

	job := &model.Job{}
	err := json.Unmarshal(msg.Data, job)
	if err != nil {
		l.Error(errors.Wrap(err, "failed to unmarshall message"))
		if err := msg.Term(); err != nil {
			l.Error(errors.Wrap(err, "failed to term the message"))
		}
		return false
	}
	// process the job
	l.Infof("processing job %s workflow %s", job.ID, job.WorkflowName)
	jp := processor.NewJobProcessor(job, w.store, w.client, w.binExec)
	err = jp.ProcessJob(ctx)
	if err != nil {
		l.Errorf("error processing job: %s", err.Error())
	} else {
		l.Infof("finished job %s workflow %s", job.ID, job.WorkflowName)
	}
	// stop the in progress ticker and ack the message
	select {
	case sidecarChan <- nil:

	case <-timeoutTimer.After(w.notifyPeriod):
		l.Errorf("timeout notifying sidecar about job completion")

	case <-ctxDone:
		return true
	case <-sidecarDone:
		return true
	}
	// Release message
	if err := msg.AckSync(); err != nil {
		l.Error(errors.Wrap(err, "failed to ack the message"))
	}
	return false
}

// workerSidecar helps notifying the NATS server about slow workflows.
// When workerMain picks up a new task, this routine is woken up and starts
// a timer that sends an "IN PROGRESS" package back to the broker if the worker
// takes too long.
func (w *workerGroup) workerSidecar(
	ctx context.Context,
	msgIn <-chan *natsio.Msg,
	done chan<- struct{},
) {
	var (
		isOpen        bool
		msgInProgress *natsio.Msg
		ctxDone       = ctx.Done()
		l             = log.FromContext(ctx)
	)
	defer close(done)

	t := newStoppedTimer()
	for {
		select {
		case <-t.C:
			ctx, cancel := context.WithTimeout(ctx, w.notifyPeriod)
			err := msgInProgress.InProgress(natsio.Context(ctx))
			cancel()
			if err != nil {
				l.Errorf("error notifying broker about message in progress: %s", err)
				// If the +WPI message fails, let's not try again, but
				// wait for the next message.
			} else {
				t.Reset(w.notifyPeriod)
			}
		case msgInProgress, isOpen = <-msgIn:
			if !isOpen {
				return
			} else if msgInProgress == nil {
				t.Stop()
			} else {
				t.Reset(w.notifyPeriod)
			}
		case <-ctxDone:
			return
		}
	}
}
