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

package nats

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/reporting/model"
)

const (
	// Set reconnect buffer size in bytes (10 MB)
	reconnectBufSize = 10 * 1024 * 1024
	// Set reconnect interval to 1 second
	reconnectWaitTime = 1 * time.Second
	// Set the number of redeliveries for a message
	maxRedeliverCount = 3
	// Set the number of inflight messages; setting it to 1 we explicitly
	// tell the NATS server that we want to process jobs serially, one by one
	maxAckPending = 10
	// Set the ACK wait
	ackWait = 30 * time.Second

	replicas = 2
)

var (
	ErrInconsistentConsumerConfig = errors.New(
		"consumer configuration is inconsistent: requires migration",
	)
)

type UnsubscribeFunc func() error

// Client is the nats client
//
//go:generate ../../../../utils/mockgen.sh
type Client interface {
	Close()
	IsConnected() bool
	JetStreamSubscribe(ctx context.Context, sub, dur string, q chan model.Job) error
	JetStreamPublish(string, []byte) error
	Migrate(ctx context.Context, sub, dur string, recreate bool) error
}

// NewClient returns a new nats client with default options
func NewClient(url string) (Client, error) {
	natsClient, err := nats.Connect(url,
		nats.ReconnectBufSize(reconnectBufSize),
		nats.ReconnectWait(reconnectWaitTime),
	)
	if err != nil {
		return nil, err
	}
	js, err := natsClient.JetStream()
	if err != nil {
		return nil, err
	}
	return &client{
		nats: natsClient,
		js:   js,
	}, nil
}

type client struct {
	nats *nats.Conn
	js   nats.JetStreamContext
}

// Close closes the connection to nats
func (c *client) Close() {
	c.nats.Close()
}

// IsConnected returns true if the client is connected to nats
func (c *client) IsConnected() bool {
	return c.nats.IsConnected()
}

func (c *client) Migrate(ctx context.Context, sub, dur string, recreate bool) error {
	cfg := &nats.ConsumerConfig{
		Name:          dur,
		Durable:       dur,
		Description:   "reporting/v2", // pull mode
		FilterSubject: sub,
		AckPolicy:     nats.AckExplicitPolicy,
		AckWait:       ackWait,
		MaxAckPending: maxAckPending,
		MaxDeliver:    maxRedeliverCount,
		Replicas:      replicas,
	}
	stream, err := c.js.StreamNameBySubject(sub)
	if err != nil {
		return err
	}
	info, err := c.js.ConsumerInfo(stream, dur)
	if err == nats.ErrConsumerNotFound {
		_, err = c.js.AddConsumer(stream, cfg)
		return err
	} else if err != nil {
		return err
	}

	if info.Config.Description != cfg.Description {
		if !recreate {
			return ErrInconsistentConsumerConfig
		}
		l := log.FromContext(ctx)
		l.Info("removing conflicting consumer configuration")
		err = c.js.DeleteConsumer(stream, dur)
		if err != nil {
			return err
		}
		l.Info("recreating consumer configuration")
		_, err = c.js.AddConsumer(stream, cfg)
	}
	return err
}

// JetStreamSubscribe subscribes to messages from the given subject with a durable subscriber
func (c *client) JetStreamSubscribe(
	ctx context.Context,
	subj, durable string,
	q chan model.Job,
) error {
	if q == nil {
		return errors.New("nats: nil subscription channel")
	}
	err := c.Migrate(ctx, subj, durable, false)
	if err != nil {
		return err
	}

	sub, err := c.js.PullSubscribe(subj, durable, nats.Context(ctx))
	if err != nil {
		if err == nats.ErrPullSubscribeToPushConsumer {
			return ErrInconsistentConsumerConfig
		}
		return err
	}
	go func() (err error) {
		l := log.FromContext(ctx)
		defer func() {
			_ = sub.Unsubscribe()
			if err != nil {
				l.Error(err)
			}
		}()
		opt := nats.Context(ctx)
		done := ctx.Done()
		var msgs []*nats.Msg
		for {
			msgs, err = sub.Fetch(1, opt)
			if err != nil {
				if err == context.DeadlineExceeded {
					continue
				}
				close(q)
				return err
			}
			for _, msg := range msgs {
				var job model.Job
				err = msg.Ack(opt)
				if err != nil {
					close(q)
					return err
				}
				err = json.Unmarshal(msg.Data, &job)
				if err != nil {
					close(q)
					return err
				}
				select {
				case q <- job:

				case <-done:
					close(q)
					return nil
				}
			}
		}
	}() //nolint: errcheck

	return nil
}

// JetStreamPublish publishes a message to the given subject
func (c *client) JetStreamPublish(subj string, data []byte) error {
	_, err := c.js.Publish(subj, data)
	return err
}
