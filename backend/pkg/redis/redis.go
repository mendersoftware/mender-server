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
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client = redis.Cmdable

const (
	CacheInvalidValue           = ""
	CacheInvalidValueExpiration = time.Second * 10 // 10s
)

var (
	ErrCacheInvalid = errors.New("cache invalidated")
)

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

func IsUnavailableErr(err error) bool {
	if err == nil {
		return false
	}

	checkers := []func(error) bool{
		redis.IsClusterDownError, // The cluster is down
		redis.IsLoadingError,     // Redis is still loading the dataset
		redis.IsMasterDownError,  // The master node is down
		redis.IsTryAgainError,    // The operation should be retried
	}

	for _, checker := range checkers {
		if checker(err) {
			return true
		}
	}
	var netErr net.Error
	return errors.As(err, &netErr)
}

// Invalidates a cache entry by swapping it with a temporary short-lived empty value.
// When a function receives this temporary value, it should count it as a cache miss,
// but also stop the caching of the value fetched DB value.
func InvalidateCache(ctx context.Context, c Client, key string) error {
	if c == nil {
		return errors.New("redis client is nil")
	}
	expiration := CacheInvalidValueExpiration
	deadline, ok := ctx.Deadline()
	if ok {
		expiration = time.Until(deadline)
	}
	res := c.Set(ctx, key,
		CacheInvalidValue, expiration)
	return res.Err()
}

// Retrieves a cache entry. If the invalidated value is found, returns `ErrCacheInvalid`
// instead of the raw result to prevent parsing errors.
func GetCache(ctx context.Context, c Client, key string) *redis.StringCmd {
	if c == nil {
		cmd := redis.NewStringCmd(ctx)
		cmd.SetErr(errors.New("redis client is nil"))
		return cmd
	}
	cmd := c.Get(ctx, key)
	if cmd.Val() == CacheInvalidValue && !IsErrRedisNil(cmd.Err()) {
		cmd.SetErr(ErrCacheInvalid)
	}

	return cmd
}

func IsErrCacheInvalid(err error) bool {
	return errors.Is(err, ErrCacheInvalid)
}

func IsErrRedisNil(e error) bool {
	return errors.Is(e, redis.Nil)
}
