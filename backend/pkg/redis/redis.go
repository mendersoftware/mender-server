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

package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

type Client = redis.Cmdable

// nolint:lll
// NewClient creates a new redis client (Cmdable) from the parameters in the
// connectionString URL format:
// Standalone mode:
// (redis|rediss|unix)://[<user>:<password>@](<host>|<socket path>)[:<port>[/<db_number>]][?option=value]
// Cluster mode:
// (redis|rediss|unix)[+srv]://[<user>:<password>@]<host1>[,<host2>[,...]][:<port>][?option=value]
//
// The following query parameters are also available:
// client_name         string
// conn_max_idle_time  duration
// conn_max_lifetime   duration
// dial_timeout        duration
// max_idle_conns      int
// max_retries         int
// max_retry_backoff   duration
// min_idle_conns      int
// min_retry_backoff   duration
// pool_fifo           bool
// pool_size           int
// pool_timeout        duration
// protocol            int
// read_timeout        duration
// tls                 bool
// write_timeout       duration
func ClientFromConnectionString(
	ctx context.Context,
	connectionString string,
) (redis.Cmdable, error) {
	var (
		redisurl   *url.URL
		tlsOptions *tls.Config
		rdb        redis.Cmdable
	)
	redisurl, err := url.Parse(connectionString)
	if err != nil {
		return nil, err
	}
	// in case connection string was provided in form of host:port
	// add scheme and parse again
	if redisurl.Host == "" {
		redisurl, err = url.Parse("redis://" + connectionString)
		if err != nil {
			return nil, err
		}
	}
	q := redisurl.Query()
	scheme := redisurl.Scheme
	cname := redisurl.Hostname()
	if strings.HasSuffix(scheme, "+srv") {
		scheme = strings.TrimSuffix(redisurl.Scheme, "+srv")
		var srv []*net.SRV
		cname, srv, err = net.DefaultResolver.LookupSRV(ctx, scheme, "tcp", redisurl.Host)
		if err != nil {
			return nil, err
		}
		addrs := make([]string, 0, len(srv))
		for i := range srv {
			if srv[i] == nil {
				continue
			}
			host := strings.TrimSuffix(srv[i].Target, ".")
			addrs = append(addrs, fmt.Sprintf("%s:%d", host, srv[i].Port))
		}
		redisurl.Host = strings.Join(addrs, ",")
		// cleanup the scheme with one known to Redis
		// to avoid: invalid URL scheme: tcp-redis+srv
		redisurl.Scheme = "redis"

	} else if scheme == "" {
		redisurl.Scheme = "redis"
	}
	// To allow more flexibility for the srv record service
	// name we use "tls" query parameter to determine if we
	// should use TLS, otherwise we test if the service
	// name contains "rediss" before falling back to no TLS.
	var useTLS bool
	if scheme == "rediss" {
		useTLS = true
	} else {
		useTLS, _ = strconv.ParseBool(q.Get("tls"))
	}
	if useTLS {
		tlsOptions = &tls.Config{ServerName: cname}
	}
	// Allow host to be a comma-separated list of hosts.
	if idx := strings.LastIndexByte(redisurl.Host, ','); idx > 0 {
		nodeAddrs := strings.Split(redisurl.Host[:idx], ",")
		for i := range nodeAddrs {
			const redisPort = ":6379"
			idx := strings.LastIndex(nodeAddrs[i], ":")
			if idx < 0 {
				nodeAddrs[i] = nodeAddrs[i] + redisPort
			}
		}
		q["addr"] = nodeAddrs
		redisurl.RawQuery = q.Encode()
		redisurl.Host = redisurl.Host[idx+1:]
	}
	var cluster bool
	if _, ok := q["addr"]; ok {
		cluster = true
	}
	if cluster {
		var redisOpts *redis.ClusterOptions
		redisOpts, err = redis.ParseClusterURL(redisurl.String())
		if err == nil {
			if tlsOptions != nil {
				redisOpts.TLSConfig = tlsOptions
			}
			rdb = redis.NewClusterClient(redisOpts)
		}
	} else {
		var redisOpts *redis.Options
		redisOpts, err = redis.ParseURL(redisurl.String())
		if err == nil {
			rdb = redis.NewClient(redisOpts)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("redis: invalid connection string: %w", err)
	}
	_, err = rdb.
		Ping(ctx).
		Result()
	return rdb, err
}
