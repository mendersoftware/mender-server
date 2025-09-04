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

package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

const (
	limitsExpSec = 1800
	cachePrefix  = "deviceauth:v1"
)

func newRedisClient(t *testing.T) (*miniredis.Miniredis, redis.Cmdable) {
	r := miniredis.NewMiniRedis()
	err := r.Start()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	t.Cleanup(r.Close)
	client := redis.NewClient(&redis.Options{
		Addr: r.Addr(),
	})
	return r, client
}

func TestRedisCacheThrottleToken(t *testing.T) {
	r, client := newRedisClient(t)

	ctx := context.TODO()
	rcache := NewRedisCache(client, cachePrefix, limitsExpSec)

	// token not found
	tok, err := rcache.Throttle(ctx,
		"tokenstring",
		"tenant-foo",
		"device-bar",
		IdTypeDevice,
		"/some/url",
		"GET")

	assert.NoError(t, err)
	assert.Equal(t, "", tok)

	//insert token
	r.Set(rcache.KeyToken("tenant-foo", "device-bar", IdTypeDevice, 0), "tokenstring")
	r.SetTTL(rcache.KeyToken("tenant-foo", "device-bar", IdTypeDevice, 0), time.Duration(10*time.Second))

	tok, err = rcache.Throttle(ctx,
		"tokenstring",
		"tenant-foo",
		"device-bar",
		IdTypeDevice,
		"/some/url",
		"GET")

	assert.NoError(t, err)
	assert.Equal(t, "tokenstring", tok)

	// wait, but before token expiration - token still found
	r.FastForward(time.Duration(5 * time.Second))
	tok, err = rcache.Throttle(ctx,
		"tokenstring",
		"tenant-foo",
		"device-bar",
		IdTypeDevice,
		"/some/url",
		"GET")

	assert.NoError(t, err)
	assert.Equal(t, "tokenstring", tok)

	// token not found past expiration
	r.FastForward(time.Duration(6 * time.Second))
	tok, err = rcache.Throttle(ctx,
		"tokenstring",
		"tenant-foo",
		"device-bar",
		IdTypeDevice,
		"/some/url",
		"GET")

	assert.NoError(t, err)
	assert.Equal(t, "", tok)

	// for some reason, the cache finds a valid token with different contents
	// and defensively rejects it
	r.Set(rcache.KeyToken("tenant-foo", "device-bar", IdTypeDevice, 0), "unknown")
	tok, err = rcache.Throttle(ctx,
		"tokenstring",
		"tenant-foo",
		"device-bar",
		IdTypeDevice,
		"/some/url",
		"GET")

	assert.NoError(t, err)
	assert.Equal(t, "", tok)

	// again insert token with Cache method
	rcache.CacheToken(ctx,
		"tenant-foo",
		"device-bar",
		IdTypeDevice,
		"tokenstr",
		time.Duration(10*time.Second))

	tok, err = rcache.Throttle(ctx,
		"tokenstr",
		"tenant-foo",
		"device-bar",
		IdTypeDevice,
		"/some/url",
		"GET")

	assert.NoError(t, err)
	assert.Equal(t, "tokenstr", tok)

	r.SetTTL(rcache.KeyToken("tenant-foo", "device-bar", IdTypeDevice, 0), time.Duration(10*time.Second))
	r.FastForward(time.Duration(11 * time.Second))

	tok, err = rcache.Throttle(ctx,
		"tokenstr",
		"tenant-foo",
		"device-bar",
		IdTypeDevice,
		"/some/url",
		"GET")

	assert.NoError(t, err)
	assert.Equal(t, "", tok)
}

func TestRedisCacheTokenDelete(t *testing.T) {
	ctx := context.TODO()

	_, client := newRedisClient(t)

	rcache := NewRedisCache(client, cachePrefix, limitsExpSec)

	// cache 2 tokens, remove first one, other one should still be available
	rcache.CacheToken(ctx,
		"tenant-foo",
		"device-1",
		IdTypeDevice,
		"tokenstr-1",
		time.Duration(10*time.Second))

	rcache.CacheToken(ctx,
		"tenant-foo",
		"device-2",
		IdTypeDevice,
		"tokenstr-2",
		time.Duration(10*time.Second))

	err := rcache.DeleteToken(ctx, "tenant-foo", "device-1", IdTypeDevice)
	assert.NoError(t, err)

	tok1, err := rcache.Throttle(ctx,
		"tokenstr-1",
		"tenant-foo",
		"device-1",
		IdTypeDevice,
		"/some/url",
		"GET")
	assert.NoError(t, err)
	assert.Equal(t, "", tok1)

	tok2, err := rcache.Throttle(ctx,
		"tokenstr-2",
		"tenant-foo",
		"device-2",
		IdTypeDevice,
		"/some/url",
		"GET")

	assert.NoError(t, err)
	assert.Equal(t, "tokenstr-2", tok2)

	// second delete (no token) doesn't trigger an error
	err = rcache.DeleteToken(ctx, "tenant-foo", "device-1", IdTypeDevice)
	assert.NoError(t, err)
}

func TestRedisCacheGetSetCheckInTime(t *testing.T) {
	_, client := newRedisClient(t)

	ctx := context.TODO()

	rcache := NewRedisCache(client, cachePrefix, limitsExpSec)

	res, err := rcache.GetCheckInTime(ctx, "tenant-foo", "device-bar")

	assert.Nil(t, res)
	assert.NoError(t, err)

	checkInTime := time.Now()
	err = rcache.CacheCheckInTime(ctx, &checkInTime, "tenant-foo", "device-bar")
	assert.NoError(t, err)

	res, err = rcache.GetCheckInTime(ctx, "tenant-foo", "device-bar")
	assert.NoError(t, err)
	assert.WithinDuration(t, checkInTime, *res, time.Second)

	times, err := rcache.GetCheckInTimes(ctx, "tenant-foo", []string{"device-bar"})
	assert.NoError(t, err)
	assert.Len(t, times, 1)
	assert.WithinDuration(t, checkInTime, *times[0], time.Second)
}
