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

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sys/unix"

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/config/ratelimits"
	"github.com/mendersoftware/mender-server/pkg/log"
	"github.com/mendersoftware/mender-server/pkg/rate"
	"github.com/mendersoftware/mender-server/pkg/redis"

	api_http "github.com/mendersoftware/mender-server/services/deviceauth/api/http"
	"github.com/mendersoftware/mender-server/services/deviceauth/cache"
	"github.com/mendersoftware/mender-server/services/deviceauth/client/orchestrator"
	dconfig "github.com/mendersoftware/mender-server/services/deviceauth/config"
	"github.com/mendersoftware/mender-server/services/deviceauth/devauth"
	"github.com/mendersoftware/mender-server/services/deviceauth/jwt"
	"github.com/mendersoftware/mender-server/services/deviceauth/store/mongo"
)

func RunServer(c config.Reader) error {
	l := log.New(log.Ctx{})

	db, err := mongo.NewDataStoreMongo(
		mongo.DataStoreMongoConfig{
			ConnectionString: c.GetString(dconfig.SettingDb),

			SSL:           c.GetBool(dconfig.SettingDbSSL),
			SSLSkipVerify: c.GetBool(dconfig.SettingDbSSLSkipVerify),

			Username: c.GetString(dconfig.SettingDbUsername),
			Password: c.GetString(dconfig.SettingDbPassword),
		})
	if err != nil {
		return errors.Wrap(err, "database connection failed")
	}

	jwtHandler, err := jwt.NewJWTHandler(
		c.GetString(dconfig.SettingServerPrivKeyPath),
	)
	var jwtFallbackHandler jwt.Handler
	fallback := c.GetString(dconfig.SettingServerFallbackPrivKeyPath)
	if err == nil && fallback != "" {
		jwtFallbackHandler, err = jwt.NewJWTHandler(
			fallback,
		)
	}
	if err != nil {
		return err
	}

	orchClientConf := orchestrator.Config{
		OrchestratorAddr: c.GetString(dconfig.SettingOrchestratorAddr),
		Timeout:          time.Duration(30) * time.Second,
	}

	devauth := devauth.NewDevAuth(db,
		orchestrator.NewClient(orchClientConf),
		jwtHandler,
		devauth.Config{
			Issuer:         c.GetString(dconfig.SettingJWTIssuer),
			ExpirationTime: int64(c.GetInt(dconfig.SettingJWTExpirationTimeout)),
			InventoryAddr:  config.Config.GetString(dconfig.SettingInventoryAddr),

			EnableReporting: config.Config.GetBool(dconfig.SettingEnableReporting),
		})

	if jwtFallbackHandler != nil {
		devauth = devauth.WithJWTFallbackHandler(jwtFallbackHandler)
	}

	var apiOptions []api_http.Option

	cacheConnStr := c.GetString(dconfig.SettingRedisConnectionString)
	if cacheConnStr == "" {
		// for backward compatibility check old redis_addr setting
		cacheConnStr = c.GetString(dconfig.SettingRedisAddr)
	}
	if cacheConnStr != "" {
		srvCache, rateLimits, err := setupRedis(c, cacheConnStr)
		if err != nil {
			return err
		}
		devauth = devauth.WithCache(srvCache)
		if rateLimits != nil {
			apiOptions = append(apiOptions,
				api_http.ConfigAuthVerifyRatelimits(rateLimits.
					WithRewriteRequests(true).
					MiddlewareGin),
			)
		}
	}
	apiOptions = append(apiOptions, api_http.SetMaxRequestSize(
		int64(c.GetInt(dconfig.SettingMaxRequestSize)),
	))
	apiHandler := api_http.NewRouter(devauth, db, apiOptions...)

	addr := c.GetString(dconfig.SettingListen)
	l.Printf("listening on %s", addr)

	srv := &http.Server{
		Addr:    addr,
		Handler: apiHandler,
	}

	errChan := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, unix.SIGINT, unix.SIGTERM)
	select {
	case sig := <-quit:
		l.Infof("received signal %s: terminating", sig)
	case err = <-errChan:
		l.Errorf("server terminated unexpectedly: %s", err.Error())
		return err
	}

	l.Info("server shutdown")
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxWithTimeout); err != nil {
		l.Error("error when shutting down the server ", err)
	}
	return nil
}

func setupRedis(c config.Reader, connStr string) (cache.Cache, *rate.HTTPLimiter, error) {
	l := log.NewEmpty()
	l.Infof("setting up redis cache")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	redisClient, err := redis.ClientFromConnectionString(ctx, connStr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize redis client: %w", err)
	}

	redisKeyPrefix := c.GetString(dconfig.SettingRedisKeyPrefix)
	cache := cache.NewRedisCache(
		redisClient,
		redisKeyPrefix,
		c.GetInt(dconfig.SettingRedisLimitsExpSec),
	)

	rateLimiter, err := ratelimits.SetupRedisRateLimits(
		redisClient, c.GetString(dconfig.SettingRedisKeyPrefix), c,
	)
	if err != nil {
		var configDisabled *ratelimits.ConfigDisabledError
		if errors.As(err, &configDisabled) {
			return cache, nil, nil
		}
		return nil, nil, fmt.Errorf("error configuring rate limits: %w", err)
	}
	return cache, rateLimiter, nil
}
