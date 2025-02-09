// Copyright 2022 Northern.tech AS
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

package client

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"syscall"
	"time"

	"github.com/mendersoftware/mender-server/pkg/netutils"

	"github.com/mendersoftware/mender-server/services/iot-manager/model"
)

const (
	ParamAlgorithmType = "X-Men-Algorithm"
	ParamSignature     = "X-Men-Signature"

	HdrKeyContentType    = "Content-Type"
	AlgorithmTypeHMAC256 = "MEN-HMAC-SHA256-Payload"
)

func New(ipFilter netutils.IPFilter) *http.Client {
	return &http.Client{
		Transport: NewTransport(ipFilter),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func ipFilterControl(ipFilter netutils.IPFilter) func(
	network, address string, _ syscall.RawConn,
) error {
	return func(network, address string, _ syscall.RawConn) error {
		ipAddr, _, err := net.SplitHostPort(address)
		if err != nil {
			ipAddr = address
		}
		ip := net.ParseIP(ipAddr)
		if ip == nil {
			return &net.ParseError{
				Type: "IP address",
				Text: address,
			}
		} else if !ipFilter.IsAllowed(ip) {
			return net.InvalidAddrError("destination address is in reserved address range")
		}
		return nil
	}
}

func NewTransport(ipFilter netutils.IPFilter) http.RoundTripper {
	dialer := &net.Dialer{
		Control: ipFilterControl(ipFilter),
	}
	tlsDialer := &tls.Dialer{
		NetDialer: dialer,
	}
	return &http.Transport{
		Proxy:                 nil,
		DialContext:           dialer.DialContext,
		DialTLSContext:        tlsDialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

// NewSignedRequest appends header X-Men-Signature with value:
// HMAC256(Request.Body, secret)
func NewSignedRequest(
	ctx context.Context,
	secret []byte,
	method string,
	url string,
	body []byte,
) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(ParamAlgorithmType, AlgorithmTypeHMAC256)

	sign := hmac.New(sha256.New, secret)
	_, _ = sign.Write(body) // Writer cannot error

	req.Header.Set(ParamSignature, hex.EncodeToString(sign.Sum(nil)))

	return req, nil
}

func NewWebhookRequest(
	ctx context.Context,
	creds *model.Credentials,
	event model.WebhookEvent,
) (*http.Request, error) {
	err := creds.Validate()
	if err != nil {
		return nil, err
	} else if creds.Type != model.CredentialTypeHTTP {
		return nil, errors.New("invalid credentials for webhooks")
	}

	b, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	var req *http.Request
	if creds.HTTP.Secret != nil {
		req, err = NewSignedRequest(
			ctx,
			[]byte(*creds.HTTP.Secret),
			http.MethodPost,
			creds.HTTP.URL,
			b,
		)
	} else {
		req, err = http.NewRequestWithContext(
			ctx, http.MethodPost,
			creds.HTTP.URL, bytes.NewReader(b),
		)
	}
	if err == nil {
		req.Header.Set(HdrKeyContentType, "application/json")
	}
	return req, err
}
