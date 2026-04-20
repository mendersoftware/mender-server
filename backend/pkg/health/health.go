// Copyright 2026 Northern.tech AS
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

// Package health provides a tiny self-probe helper that each Mender backend
// service exposes via a "healthcheck" CLI subcommand. The subcommand is
// intended to be invoked as the image-level Docker HEALTHCHECK from
// scratch-based container images, which lack wget/curl/sh.
package health

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// DefaultTimeout is the per-probe client-side timeout. Docker handles
// retries via its own HEALTHCHECK retries: knob, so one short attempt is
// enough here.
const DefaultTimeout = 2 * time.Second

// Probe issues a single GET against url and returns nil if the response
// status is 2xx. Any transport error, timeout, or non-2xx status produces
// a non-nil error whose message identifies the failure.
func Probe(ctx context.Context, url string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("healthcheck %s: build request: %w", url, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("healthcheck %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("healthcheck %s: unexpected status %d", url, resp.StatusCode)
	}
	return nil
}
