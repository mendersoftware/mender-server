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

package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/services/deployments/storage"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	))
	defer srv.Close()

	srvURL, err := url.Parse(srv.URL)
	if err != nil {
		panic(err)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			DialTLS: func(network, addr string) (net.Conn, error) {
				return net.Dial(network, srvURL.Host)
			},
		},
	}

	sss := SimpleStorageService{
		client: s3.New(s3.Options{
			Region:     "test",
			HTTPClient: httpClient,
			Credentials: StaticCredentials{
				Key:    "test",
				Secret: "secret",
				Token:  "token",
			},
		}),
		settings: storageSettings{BucketName: aws.String("test")},
	}

	err = sss.HealthCheck(context.Background())
	assert.NoError(t, err)
}

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (r RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return r(req)
}

func TestNewClient(t *testing.T) {
	// Test initializing a new client and that the pre-conditions are checked:
	// HeadBucket(404) -> CreateBucket(200) -> HeadBucket(404)
	const (
		bucketName     = "artifacts"
		hostName       = "testing.mender.io"
		bucketHostname = bucketName + "." + hostName
		region         = "poddlest"
		keyID          = "awskeyID"
		secret         = "secretkey"
		token          = "tokenMcTokenFace"
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	done := ctx.Done()

	chReq := make(chan *http.Request, 1)
	chRsp := make(chan *http.Response, 1)
	rt := RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		done := req.Context().Done()
		b, _ := httputil.DumpRequest(req, false)
		t.Log(string(b))
		assert.Equal(t, bucketHostname, req.URL.Host)

		// Check X-Amz-Date header
		amzTimeStr := req.Header.Get(paramAmzDate)
		amzTime, err := time.Parse(paramAmzDateFormat, amzTimeStr)
		if assert.NoError(t, err, "unexpected X-Amz-Date header value") {
			assert.WithinDuration(t, time.Now(), amzTime, time.Minute)
		}

		// Check X-Amz-Security-Token
		assert.Equal(t, token, req.Header.Get("X-Amz-Security-Token"))

		authz := req.Header.Get("Authorization")
		if assert.NotEmpty(t, authz) {
			assert.True(t,
				strings.HasPrefix(authz, "AWS4-HMAC-SHA256"),
				"unexpected Authorization header type")
			authz = strings.TrimPrefix(authz, "AWS4-HMAC-SHA256")
			idxDate := strings.IndexRune(amzTimeStr, 'T')
			if idxDate < 0 {
				idxDate = len(amzTimeStr)
			}
			expectedParams := map[string]struct{}{
				"Credential":    struct{}{},
				"Signature":     struct{}{},
				"SignedHeaders": struct{}{},
			}
			for _, param := range strings.Fields(authz) {
				keyValue := strings.SplitN(param, "=", 2)
				if len(keyValue) != 2 {
					continue
				}
				key, value := keyValue[0], keyValue[1]
				value = strings.TrimRight(value, ",")
				switch key {
				case "Credential":
					assert.Equal(t, fmt.Sprintf("%s/%s/%s/s3/aws4_request",
						keyID,
						amzTimeStr[:idxDate],
						region,
					), value, "Invalid Authorization parameter Credential")
				case "Signature":

				case "SignedHeaders":
					for _, hdr := range []string{"host", "x-amz-date", "x-amz-security-token"} {
						assert.Containsf(t,
							value,
							hdr,
							"SignedHeaders does not contain header %q",
							hdr)
					}
				default:
					continue
				}
				delete(expectedParams, key)
			}
			assert.Empty(t, expectedParams,
				"Some expected Authorization parameters was not present")
		}

		select {
		case chReq <- req:
		case <-done:
			return nil, errors.New("timeout")
		}

		var rsp *http.Response
		select {
		case rsp = <-chRsp:
		case <-done:
			return nil, errors.New("timeout")
		}

		if rsp == nil {
			err = errors.New("nil Response")
		}

		return rsp, err
	})

	options := NewOptions().
		SetBucketName(bucketName).
		SetBufferSize(5*1024*1024).
		SetContentType("test").
		SetDefaultExpire(time.Minute).
		SetRegion(region).
		SetStaticCredentials(keyID, secret, token).
		SetURI("https://" + hostName).
		SetForcePathStyle(false).
		SetUseAccelerate(false).
		SetUnsignedHeaders([]string{"Accept-Encoding"}).
		SetTransport(rt)
	t.Log(options.storageSettings)

	go func() {
		_, err := New(ctx, options) //nolint:errcheck
		assert.NoError(t, err)
		cancel()
	}()

	// HeadBucket
	select {
	case req := <-chReq:
		assert.Equal(t, http.MethodHead, req.Method)
		w := httptest.NewRecorder()
		w.WriteHeader(http.StatusNotFound)
		chRsp <- w.Result()

	case <-done:
		assert.FailNow(t, "timeout waiting for request")
	}

	// PutBucket
	select {
	case req := <-chReq:
		assert.Equal(t, http.MethodPut, req.Method)
		w := httptest.NewRecorder()
		w.WriteHeader(http.StatusOK)
		chRsp <- w.Result()

	case <-done:
		assert.FailNow(t, "timeout waiting for request")
	}

	// HeadBucket
	select {
	case req := <-chReq:
		assert.Equal(t, http.MethodHead, req.Method)
		w := httptest.NewRecorder()
		w.WriteHeader(http.StatusOK)
		chRsp <- w.Result()

	case <-done:
		assert.FailNow(t, "timeout waiting for request")
	}

}

func newTestServerAndClient(
	handler http.Handler,
	opts ...*Options,
) (storage.ObjectStorage, *httptest.Server) {
	initHandler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodHead, http.MethodPut:
				w.WriteHeader(http.StatusNoContent)
			default:
				w.WriteHeader(http.StatusOK)
			}
		},
	)
	srv := httptest.NewServer(initHandler)
	var d net.Dialer
	httpTransport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return d.DialContext(
				ctx,
				srv.Listener.Addr().Network(),
				srv.Listener.Addr().String(),
			)
		},
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return d.DialContext(
				ctx,
				srv.Listener.Addr().Network(),
				srv.Listener.Addr().String(),
			)
		},
	}

	opt := NewOptions().
		SetBucketName("bucket").
		SetRegion("region").
		SetStaticCredentials("test", "secret", "token")
	opts = append([]*Options{opt}, opts...)

	opt = NewOptions(opts...).
		SetTransport(httpTransport)

	sss, err := New(context.Background(), opt)
	if err != nil {
		panic(err)
	}
	srv.Config.Handler = handler
	return sss, srv
}

func TestGetObject(t *testing.T) {
	t.Parallel()

	type testCase struct {
		Name string

		CTX        context.Context
		ObjectPath string

		Handler func(t *testing.T) http.HandlerFunc
		Body    []byte
		Error   assert.ErrorAssertionFunc
	}

	testCases := []testCase{{
		Name: "ok",

		ObjectPath: "foo/bar",
		Handler: func(t *testing.T) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/foo/bar", r.URL.Path)
				assert.Equal(t, "bucket.s3.region.amazonaws.com", r.Host)

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("imagine artifacts"))
			}
		},
		Body: []byte("imagine artifacts"),
	}, {
		Name: "error/object not found",

		ObjectPath: "foo/bar",
		Handler: func(t *testing.T) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/foo/bar", r.URL.Path)
				assert.Equal(t, "bucket.s3.region.amazonaws.com", r.Host)

				w.WriteHeader(http.StatusNotFound)
			}
		},
		Error: func(t assert.TestingT, err error, _ ...interface{}) bool {
			return assert.ErrorIs(t, err, storage.ErrObjectNotFound)
		},
	}, {
		Name: "error/invalid settings from context",

		CTX: storage.SettingsWithContext(
			context.Background(),
			&model.StorageSettings{},
		),
		ObjectPath: "foo/bar",
		Handler: func(t *testing.T) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				assert.Fail(t, "the test was not supposed to make a request")
				w.WriteHeader(http.StatusInternalServerError)
			}
		},
		Error: func(t assert.TestingT, err error, _ ...interface{}) bool {
			var verr validation.Errors
			return assert.Error(t, err) &&
				assert.ErrorAs(t, err, &verr) &&
				assert.Contains(t, verr, "key") &&
				assert.Contains(t, verr, "secret") &&
				assert.Contains(t, verr, "bucket") &&
				assert.Contains(t, verr, "region")
		},
	}}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			s3c, srv := newTestServerAndClient(tc.Handler(t))
			defer srv.Close()
			var ctx context.Context
			if tc.CTX != nil {
				ctx = tc.CTX
			} else {
				ctx = context.Background()
			}
			obj, err := s3c.GetObject(ctx, tc.ObjectPath)
			if tc.Error != nil {
				tc.Error(t, err)
			} else if assert.NoError(t, err) {
				b, _ := io.ReadAll(obj)
				obj.Close()
				assert.Equal(t, tc.Body, b)
			}
		})
	}
}
