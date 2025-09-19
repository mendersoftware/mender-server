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

// Client is the nats client
//
//go:generate ../../../../utils/mockgen.sh
type Client interface {
	Publish(string, []byte) error
	ChanSubscribe(string, chan *natsio.Msg) (*natsio.Subscription, error)
}

// NewClient returns a new workflows client
func NewClient(url string, opts ...natsio.Option) (Client, error) {
	natsClient, err := natsio.Connect(url, opts...)
	if err != nil {
		return nil, err
	}
	return &client{
		nats: natsClient,
	}, nil
}

// NewClient returns a new nats client with default options
func NewClientWithDefaults(url string) (Client, error) {
	ctx := context.Background()
	l := log.FromContext(ctx)

	natsClient, err := NewClient(url,
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
	nats *natsio.Conn
}

func (c *client) Publish(subj string, data []byte) error {
	return c.nats.Publish(subj, data)
}

func (c *client) ChanSubscribe(subj string,
	channel chan *natsio.Msg) (*natsio.Subscription, error) {
	return c.nats.ChanSubscribe(subj, channel)
}
