// Copyright 2026 Northern.tech AS
//
// Licensed under the Apache License, Version 2.0 (see LICENSE).

package health

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/urfave/cli"
)

func runAction(t *testing.T, url string) int {
	t.Helper()
	code := 0
	action := makeAction(
		func(*cli.Context) string { return url },
		func(c int) { code = c },
	)
	ctx := cli.NewContext(cli.NewApp(), flag.NewFlagSet("test", 0), nil)
	if err := action(ctx); err != nil {
		t.Fatalf("action returned error: %v", err)
	}
	return code
}

func TestCommandAction_Healthy(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	if code := runAction(t, ts.URL); code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

func TestCommandAction_Unhealthy(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	if code := runAction(t, ts.URL); code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}
