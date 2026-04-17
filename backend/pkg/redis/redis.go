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

// ClientFromConnectionString creates a new redis client (Cmdable) from the parameters in the
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
//
//nolint:lll
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
		redisurl.Host = net.JoinHostPort(redisurl.Scheme, redisurl.Opaque)
		redisurl.Scheme = "redis"
		redisurl.Opaque = ""
	}
	q := redisurl.Query()
	serverName := redisurl.Hostname()
	var found bool
	if redisurl.Scheme, found = strings.CutSuffix(redisurl.Scheme, "+srv"); found {
		var srv []*net.SRV
		serverName, srv, err = net.DefaultResolver.LookupSRV(
			ctx,
			redisurl.Scheme,
			"tcp",
			redisurl.Host,
		)
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

	}
	// Allow host to be a comma-separated list of hosts.
	addrs := strings.Split(redisurl.Host, ",")
	if len(addrs) > 1 {
		redisurl.Host = addrs[0]
		q := redisurl.Query()
		for _, addr := range addrs {
			q.Add("addr", addr)
		}
	}
	// To allow more flexibility for the srv record service
	// name we use "tls" query parameter to determine if we
	// should use TLS, otherwise we test if the service
	// name contains "rediss" before falling back to no TLS.
	var useTLS bool
	if redisurl.Scheme == "rediss" {
		useTLS = true
	} else {
		useTLS, _ = strconv.ParseBool(q.Get("tls"))
	}
	if useTLS {
		tlsOptions = &tls.Config{ServerName: serverName}
	}
	// Use cluster mode if `cluster` querystring is truthy or additional
	// addr parameters are supplied.
	var cluster bool
	if _, ok := q["cluster"]; ok {
		cluster, _ = strconv.ParseBool("cluster")
		delete(q, "cluster")
	} else {
		_, cluster = q["addr"]
	}
	redisurl.RawQuery = q.Encode()
	if cluster {
		var redisOpts *redis.ClusterOptions
		redisOpts, err = redis.ParseClusterURL(redisurl.String())
		if err == nil {
			if tlsOptions != nil {
				redisOpts.TLSConfig = tlsOptions
			}
			var retries int
			if redisOpts.MaxRetries < 0 {
				retries = 3 // Use same default as normal redis.Client
				redisOpts.MaxRetries = -1
			}
			rdb = redis.NewClusterClient(redisOpts)
			if retries > 0 {
				// Special retry hook for cluster retries
				rdb.(*redis.ClusterClient).AddHook(retryHook(retries))
			}
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

type retryHook int

func (retryHook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (retries retryHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		err := next(ctx, cmd)
		if err == nil {
			return err
		}
		for range retries {
			var netErr net.Error
			if redis.IsTryAgainError(err) || (errors.As(err, &netErr) && netErr.Timeout()) {
				err = next(ctx, cmd)
			} else {
				break
			}
		}
		return err
	}
}

func (retryHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
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
