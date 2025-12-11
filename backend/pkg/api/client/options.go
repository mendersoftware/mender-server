// Copyright 2025 Northern.tech AS
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	    http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.

package client

import (
	"net/http"
	"time"
)

const (
	// DefaultTimeout is the default HTTP client timeout.
	DefaultTimeout = 10 * time.Second
)

// Config holds HTTP client configuration options.
type Config struct {
	// Timeout specifies the timeout for HTTP requests.
	Timeout time.Duration
	// Transport is the underlying RoundTripper for the HTTP client.
	// If nil, a MenderTransport wrapping http.DefaultTransport is used.
	Transport http.RoundTripper
}

// Option is a function that configures the HTTP client.
type Option func(*Config)

// defaultConfig returns the default configuration.
func defaultConfig() Config {
	return Config{
		Timeout:   DefaultTimeout,
		Transport: nil,
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithTransport sets a custom RoundTripper for the HTTP client.
// Note: If you set a custom transport, it will NOT be wrapped with MenderTransport.
// If you need Mender header propagation, wrap your transport with MenderTransport first.
func WithTransport(transport http.RoundTripper) Option {
	return func(c *Config) {
		c.Transport = transport
	}
}

// NewHTTPClient creates a new http.Client configured with MenderTransport
// for automatic header propagation from context.
func NewHTTPClient(opts ...Option) *http.Client {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	transport := cfg.Transport
	if transport == nil {
		transport = &MenderTransport{Base: http.DefaultTransport}
	}

	return &http.Client{
		Timeout:   cfg.Timeout,
		Transport: transport,
	}
}
