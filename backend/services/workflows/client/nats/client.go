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
	"errors"
	"fmt"
	"time"

	natsio "github.com/nats-io/nats.go"

	"github.com/mendersoftware/mender-server/pkg/log"
)

const (
	// Set reconnect buffer size in bytes (10 MB)
	reconnectBufSize = 10 * 1024 * 1024
	// Set reconnect interval to 1 second
	reconnectWaitTime = 1 * time.Second
)

var (
	ErrIncompatibleConsumer = errors.New("nats: cannot subscribe to a pull consumer")
)

type UnsubscribeFunc func() error

// Client is the nats client
//
//go:generate ../../utils/mockgen.sh
type Client interface {
	Close()
	StreamName() string
	IsConnected() bool
	JetStreamCreateStream(streamName string) error
	GetConsumerConfig(name string) (*ConsumerConfig, error)
	CreateConsumer(name string, upsert bool, config ConsumerConfig) error
	JetStreamSubscribe(
		ctx context.Context,
		subj,
		durable string,
		q chan *natsio.Msg,
	) (UnsubscribeFunc, error)
	JetStreamPublish(string, []byte) error
}

// NewClient returns a new nats client
func NewClient(url string, streamName string, opts ...natsio.Option) (Client, error) {
	natsClient, err := natsio.Connect(url, opts...)
	if err != nil {
		return nil, err
	}
	js, err := natsClient.JetStream()
	if err != nil {
		return nil, err
	}
	return &client{
		nats:       natsClient,
		streamName: streamName,
		js:         js,
	}, nil
}

// NewClient returns a new nats client with default options
func NewClientWithDefaults(url string, streamName string) (Client, error) {
	ctx := context.Background()
	l := log.FromContext(ctx)

	natsClient, err := NewClient(url,
		streamName,
		func(o *natsio.Options) error {
			o.AllowReconnect = true
			o.MaxReconnect = -1
			o.ReconnectBufSize = reconnectBufSize
			o.ReconnectWait = reconnectWaitTime
			o.RetryOnFailedConnect = true
			o.ClosedCB = func(_ *natsio.Conn) {
				l.Info("nats client closed the connection")
			}
			o.DisconnectedErrCB = func(_ *natsio.Conn, e error) {
				if e != nil {
					l.Warnf("nats client disconnected, err: %v", e)
				}
			}
			o.ReconnectedCB = func(_ *natsio.Conn) {
				l.Warn("nats client reconnected")
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return natsClient, nil
}

type client struct {
	nats       *natsio.Conn
	js         natsio.JetStreamContext
	streamName string
}

// IsConnected returns true if the client is connected to nats
func (c *client) StreamName() string {
	return c.streamName
}

// Close closes the connection to nats
func (c *client) Close() {
	c.nats.Close()
}

// IsConnected returns true if the client is connected to nats
func (c *client) IsConnected() bool {
	return c.nats.IsConnected()
}

// JetStreamCreateStream creates a stream
func (c *client) JetStreamCreateStream(streamName string) error {
	stream, err := c.js.StreamInfo(streamName)
	if err != nil && err != natsio.ErrStreamNotFound {
		return err
	}
	if stream == nil {
		_, err = c.js.AddStream(&natsio.StreamConfig{
			Name:      streamName,
			NoAck:     false,
			MaxAge:    24 * time.Hour,
			Retention: natsio.WorkQueuePolicy,
			Storage:   natsio.FileStorage,
			Subjects:  []string{streamName + ".>"},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func noop() error {
	return nil
}

type ConsumerConfig struct {
	// Filter expression for which topics this consumer covers.
	Filter string
	// MaxPending messages in the work queue.
	// NOTE: This sets an upper limit on the horizontal scalability of the
	// service.
	MaxPending int
	// MaxDeliver sets the maximum amount of time the message will be
	// (re-) delivered.
	MaxDeliver int
	// AckWait sets the time to wait for message acknowledgement before
	// resending the message.
	AckWait time.Duration
}

func (cfg ConsumerConfig) Validate() error {
	if cfg.AckWait < time.Second {
		return fmt.Errorf(
			"invalid consumer configuration AckWait: %s < 1s",
			cfg.AckWait)
	}
	if cfg.MaxDeliver < 1 {
		return fmt.Errorf(
			"invalid consumer configuration MaxDeliver: %d < 1",
			cfg.MaxDeliver)
	}
	if cfg.MaxPending < 1 {
		return fmt.Errorf(
			"invalid consumer configuration MaxPending: %d < 1",
			cfg.MaxPending)
	}
	return nil
}

const consumerVersionString = "workflows/v1"

func (cfg ConsumerConfig) toNats(name string, deliverSubject string) *natsio.ConsumerConfig {
	if deliverSubject == "" {
		deliverSubject = natsio.NewInbox()
	}
	return &natsio.ConsumerConfig{
		Name:         name, // To preserve behavior of the internal library,
		Durable:      name, // the consumer-, durable- and delivery group name
		DeliverGroup: name, // are all set to the durable name.

		Description:    consumerVersionString,
		DeliverSubject: deliverSubject,

		FilterSubject: cfg.Filter,
		AckWait:       cfg.AckWait,
		MaxAckPending: cfg.MaxPending,
		MaxDeliver:    cfg.MaxDeliver,

		AckPolicy:     natsio.AckExplicitPolicy,
		DeliverPolicy: natsio.DeliverAllPolicy,
	}
}

func configFromNats(cfg natsio.ConsumerConfig) ConsumerConfig {
	return ConsumerConfig{
		Filter:     cfg.FilterSubject,
		MaxPending: cfg.MaxAckPending,
		MaxDeliver: cfg.MaxDeliver,
		AckWait:    cfg.AckWait,
	}
}

func (c *client) GetConsumerConfig(name string) (*ConsumerConfig, error) {
	consumerInfo, err := c.js.ConsumerInfo(c.streamName, name)
	if err != nil {
		return nil, err
	} else if consumerInfo == nil {
		return nil, fmt.Errorf("nats: nil consumer")
	}
	cfg := configFromNats(consumerInfo.Config)
	return &cfg, nil
}

func (c *client) CreateConsumer(name string, upsert bool, config ConsumerConfig) error {
	consumerInfo, err := c.js.ConsumerInfo(c.streamName, name)
	if errors.Is(err, natsio.ErrConsumerNotFound) {
		_, err = c.js.AddConsumer(c.streamName, config.toNats(name, ""))
		var apiErr *natsio.APIError
		if err == nil {
			return nil
		} else if errors.As(err, &apiErr) &&
			apiErr.ErrorCode == natsio.JSErrCodeConsumerAlreadyExists {
			// Race: consumer was just created between ConsumerInfo and AddConsumer
			consumerInfo, err = c.js.ConsumerInfo(c.streamName, name)
		}
	}
	if err != nil {
		return fmt.Errorf("nats: error getting consumer info: %w", err)
	}
	if upsert {
		if consumerInfo.Config.DeliverSubject == "" {
			return ErrIncompatibleConsumer
		}
		_, err = c.js.UpdateConsumer(
			c.streamName,
			config.toNats(name, consumerInfo.Config.DeliverSubject),
		)
		if err == nil {
			return nil
		}
	}
	return err
}

// JetStreamSubscribe subscribes to messages from the given subject with a durable subscriber
func (c *client) JetStreamSubscribe(
	ctx context.Context,
	subj, durable string,
	q chan *natsio.Msg,
) (UnsubscribeFunc, error) {
	sub, err := c.js.ChanQueueSubscribe(subj, durable, q,
		natsio.Bind(c.streamName, durable),
		natsio.ManualAck(),
		natsio.Context(ctx),
	)
	if err != nil {
		return noop, err
	}

	return sub.Unsubscribe, nil
}

// JetStreamPublish publishes a message to the given subject
func (c *client) JetStreamPublish(subj string, data []byte) error {
	_, err := c.js.Publish(subj, data)
	return err
}
