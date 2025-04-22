package redis

import (
	"os"
	"testing"
)

const EnvRedisURL = "TEST_REDIS_URL"

var RedisURL = os.Getenv(EnvRedisURL)

func requireRedis(t *testing.T) {
	if RedisURL == "" {
		t.Skipf("skipping test %q due to missing redis URL, "+
			"use environment variable %q to run test",
			t.Name(), EnvRedisURL)
	}
}
