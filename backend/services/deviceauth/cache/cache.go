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

// Token Management
//
// Tokens are expected at:
// `tenant:<tid>:version<tenant_key_version>:device:<did>:tok: <token>`
//
// Cache invalidation.
// We achive cache invalidation by incrementing tenant key version.
// Each tenant related key in the cache has to contain tenant key version.
// This way, by incrementing tenant key version, we invalidate all tenant
// related keys.

package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/deviceauth/model"
)

const (
	IdTypeDevice = "device"
	IdTypeUser   = "user"
	// expiration of the device check in time - one week
	CheckInTimeExpiration = time.Duration(time.Hour * 24 * 7)
)

var (
	ErrNoPositiveInteger = errors.New("must be a positive integer")
	ErrNegativeInteger   = errors.New("cannot be a negative integer")
)

//go:generate ../../../utils/mockgen.sh
type Cache interface {
	// Throttle retrieves a cached token.
	// These ops are bundled because the implementation will pipeline them for a single network
	// roundtrip for max performance.
	// Returns:
	// - the token (if any)
	Throttle(
		ctx context.Context,
		rawToken string,
		tid,
		id,
		idtype,
		url,
		action string,
	) (string, error)

	// CacheToken caches the token under designated key, with expiration
	CacheToken(ctx context.Context, tid, id, idtype, token string, expire time.Duration) error

	// DeleteToken deletes the token for 'id'
	DeleteToken(ctx context.Context, tid, id, idtype string) error

	// GetLimit gets a limit from cache (see store.Datastore.GetLimit)
	GetLimit(ctx context.Context, name string) (*model.Limit, error)
	// SetLimit writes a limit to cache (see store.Datastore.SetLimit)
	SetLimit(ctx context.Context, limit *model.Limit) error
	// DeleteLimit evicts the limit with the given name from cache
	DeleteLimit(ctx context.Context, name string) error

	// CacheCheckInTime caches the last device check in time
	CacheCheckInTime(ctx context.Context, t *time.Time, tid, id string) error

	// GetCheckInTime gets the last device check in time from cache
	GetCheckInTime(ctx context.Context, tid, id string) (*time.Time, error)

	// GetCheckInTimes gets the last device check in time from cache
	// for each device with id from the list of ids
	GetCheckInTimes(ctx context.Context, tid string, ids []string) ([]*time.Time, error)

	// SuspendTenant increment tenant key version
	// tenant key is used in all cache keys, this way, when we increment the key version,
	// all the keys are no longer accessible - in other words, be incrementing tenant key version
	// we invalidate all tenant keys
	SuspendTenant(ctx context.Context, tid string) error
}

type RedisCache struct {
	c               redis.Cmdable
	prefix          string
	LimitsExpireSec int
	DefaultExpire   time.Duration
}

func NewRedisCache(
	redisClient redis.Cmdable,
	prefix string,
	limitsExpireSec int,
) *RedisCache {
	return &RedisCache{
		c:               redisClient,
		LimitsExpireSec: limitsExpireSec,
		prefix:          prefix,
		DefaultExpire:   time.Hour * 3,
	}
}

func (rl *RedisCache) keyLimit(tenantID, name string) string {
	if tenantID == "" {
		tenantID = "default"
	}
	return fmt.Sprintf("%s:tenant:%s:limit:%s", rl.prefix, tenantID, name)
}

func (rl *RedisCache) GetLimit(ctx context.Context, name string) (*model.Limit, error) {
	var tenantID string
	id := identity.FromContext(ctx)
	if id != nil {
		tenantID = id.Tenant
	}
	value, err := rl.c.Get(ctx, rl.keyLimit(tenantID, name)).Uint64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	return &model.Limit{
		TenantID: tenantID,
		Value:    value,
		Name:     name,
	}, nil
}

func (rl *RedisCache) SetLimit(ctx context.Context, limit *model.Limit) error {
	if limit == nil {
		return nil
	}
	var tenantID string
	id := identity.FromContext(ctx)
	if id != nil {
		tenantID = id.Tenant
	}
	key := rl.keyLimit(tenantID, limit.Name)
	return rl.c.SetEx(ctx, key, limit.Value, rl.DefaultExpire).Err()
}

func (rl *RedisCache) DeleteLimit(ctx context.Context, name string) error {
	var tenantID string
	id := identity.FromContext(ctx)
	if id != nil {
		tenantID = id.Tenant
	}
	key := rl.keyLimit(tenantID, name)
	return rl.c.Del(ctx, key).Err()
}

func (rl *RedisCache) Throttle(
	ctx context.Context,
	rawToken string,
	tid,
	id,
	idtype,
	url,
	action string,
) (string, error) {
	var tokenGet *redis.StringCmd

	pipe := rl.c.Pipeline()

	version, err := rl.getTenantKeyVersion(ctx, tid)
	if err != nil {
		return "", err
	}

	// queue token fetching
	// for piped execution
	tokenGet = rl.pipeToken(ctx, pipe, tid, id, idtype, version)

	_, err = pipe.Exec(ctx)
	if err != nil && !isErrRedisNil(err) {
		return "", err
	}

	// collect quota/burst control and token fetch results
	tok, err := rl.checkToken(tokenGet, rawToken)
	if err != nil {
		return "", err
	}

	return tok, nil
}

func (rl *RedisCache) pipeToken(
	ctx context.Context,
	pipe redis.Pipeliner,
	tid,
	id,
	idtype string,
	version int,
) *redis.StringCmd {
	key := rl.KeyToken(tid, id, idtype, version)
	return pipe.Get(ctx, key)
}

func (rl *RedisCache) checkToken(cmd *redis.StringCmd, raw string) (string, error) {
	err := cmd.Err()
	if err != nil {
		if isErrRedisNil(err) {
			return "", nil
		}
		return "", err
	}

	token := cmd.Val()
	if token == raw {
		return token, nil
	} else {
		// must be a stale token - we don't want to use it
		// let it expire in the background
		return "", nil
	}
}

func (rl *RedisCache) CacheToken(
	ctx context.Context,
	tid,
	id,
	idtype,
	token string,
	expire time.Duration,
) error {
	version, err := rl.getTenantKeyVersion(ctx, tid)
	if err != nil {
		return err
	}
	res := rl.c.Set(ctx, rl.KeyToken(tid, id, idtype, version),
		token,
		expire)
	return res.Err()
}

func (rl *RedisCache) DeleteToken(ctx context.Context, tid, id, idtype string) error {
	version, err := rl.getTenantKeyVersion(ctx, tid)
	if err != nil {
		return err
	}
	res := rl.c.Del(ctx, rl.KeyToken(tid, id, idtype, version))
	return res.Err()
}

func (rl *RedisCache) KeyToken(tid, id, idtype string, version int) string {
	return fmt.Sprintf(
		"%s:tenant:%s:version:%d:%s:%s:tok",
		rl.prefix, tid, version, idtype, id)
}

func (rl *RedisCache) KeyCheckInTime(tid, id, idtype string, version int) string {
	return fmt.Sprintf(
		"%s:tenant:%s:version:%d:%s:%s:checkInTime",
		rl.prefix, tid, version, idtype, id)
}

func (rl *RedisCache) KeyTenantVersion(tid string) string {
	return fmt.Sprintf("%s:tenant:%s:version", rl.prefix, tid)
}

// isErrRedisNil checks for a very common non-error, "redis: nil",
// which just means the key was not found, and is normal
// it's routinely returned e.g. from GET, or pipelines containing it
func isErrRedisNil(e error) bool {
	return errors.Is(e, redis.Nil)
}

func (rl *RedisCache) CacheCheckInTime(
	ctx context.Context,
	t *time.Time,
	tid,
	id string,
) error {
	tj, err := json.Marshal(t)
	if err != nil {
		return err
	}

	version, err := rl.getTenantKeyVersion(ctx, tid)
	if err != nil {
		return err
	}

	res := rl.c.Set(
		ctx,
		rl.KeyCheckInTime(tid, id, IdTypeDevice, version),
		tj,
		CheckInTimeExpiration,
	)

	return res.Err()
}

func (rl *RedisCache) GetCheckInTime(
	ctx context.Context,
	tid,
	id string,
) (*time.Time, error) {
	version, err := rl.getTenantKeyVersion(ctx, tid)
	if err != nil {
		return nil, err
	}

	res := rl.c.Get(ctx, rl.KeyCheckInTime(tid, id, IdTypeDevice, version))

	if res.Err() != nil {
		if isErrRedisNil(res.Err()) {
			return nil, nil
		}
		return nil, res.Err()
	}

	var checkInTime time.Time

	err = json.Unmarshal([]byte(res.Val()), &checkInTime)
	if err != nil {
		return nil, err
	}

	return &checkInTime, nil
}

func (rl *RedisCache) GetCheckInTimes(
	ctx context.Context,
	tid string,
	ids []string,
) ([]*time.Time, error) {
	l := log.FromContext(ctx)

	version, err := rl.getTenantKeyVersion(ctx, tid)
	if err != nil {
		return nil, err
	}
	checkInTimes := make([]*time.Time, len(ids))
	if _, ok := rl.c.(*redis.ClusterClient); ok {
		pipe := rl.c.Pipeline()
		for _, id := range ids {
			pipe.Get(ctx, rl.KeyCheckInTime(tid, id, IdTypeDevice, version))
		}
		results, err := pipe.Exec(ctx)
		if err != nil && !errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("failed to fetch check in times: %w", err)
		}
		for i, result := range results {
			cmd, ok := result.(*redis.StringCmd)
			if !ok {
				continue // should never happen
			}
			b, err := cmd.Bytes()
			if err != nil {
				if errors.Is(err, redis.Nil) {
					continue
				} else {
					l.Errorf("failed to get device: %s", err.Error())
				}
			} else {
				checkInTime := new(time.Time)
				err = json.Unmarshal(b, checkInTime)
				if err != nil {
					l.Errorf("failed to deserialize check in time: %s", err.Error())
				} else {
					checkInTimes[i] = checkInTime
				}

			}
		}
	} else {
		keys := make([]string, len(ids))
		for i, id := range ids {
			keys[i] = rl.KeyCheckInTime(tid, id, IdTypeDevice, version)
		}
		res := rl.c.MGet(ctx, keys...)

		for i, v := range res.Val() {
			if v != nil {
				b, ok := v.(string)
				if !ok {
					continue
				}
				var checkInTime time.Time
				err := json.Unmarshal([]byte(b), &checkInTime)
				if err != nil {
					l.Errorf("failed to unmarshal check-in time: %s", err.Error())
					continue
				}
				checkInTimes[i] = &checkInTime
			}
		}
	}

	return checkInTimes, nil
}

func (rl *RedisCache) SuspendTenant(
	ctx context.Context,
	tid string,
) error {
	res := rl.c.Incr(ctx, rl.KeyTenantVersion(tid))
	return res.Err()
}

func (rl *RedisCache) getTenantKeyVersion(ctx context.Context, tid string) (int, error) {
	res := rl.c.Get(ctx, rl.KeyTenantVersion(tid))
	if res.Err() != nil {
		if isErrRedisNil(res.Err()) {
			return 0, nil
		}
		return 0, res.Err()
	}

	var version int

	err := json.Unmarshal([]byte(res.Val()), &version)
	if err != nil {
		return 0, err
	}

	return version, nil
}
