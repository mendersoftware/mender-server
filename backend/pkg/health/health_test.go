// Copyright 2026 Northern.tech AS
//
// Licensed under the Apache License, Version 2.0 (see LICENSE).

package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProbe_2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	if err := Probe(context.Background(), ts.URL, DefaultTimeout); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestProbe_5xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	if err := Probe(context.Background(), ts.URL, DefaultTimeout); err == nil {
		t.Fatalf("expected error for 500, got nil")
	}
}

func TestProbe_ConnectionRefused(t *testing.T) {
	// Bind and close to claim a guaranteed-closed port.
	ts := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	url := ts.URL
	ts.Close()

	if err := Probe(context.Background(), url, DefaultTimeout); err == nil {
		t.Fatalf("expected error for refused connection, got nil")
	}
}

func TestProbe_Timeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	if err := Probe(context.Background(), ts.URL, 10*time.Millisecond); err == nil {
		t.Fatalf("expected timeout error, got nil")
	}
}
