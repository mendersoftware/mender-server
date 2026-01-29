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
	"crypto/tls"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/services/deployments/storage"
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

	u, _ := url.Parse("https://" + hostName)
	options := NewOptions().
		SetBucketName(bucketName).
		SetBufferSize(5*1024*1024).
		SetContentType("test").
		SetDefaultExpire(time.Minute).
		SetRegion(region).
		SetStaticCredentials(keyID, secret, token).
		SetURI(u).
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

// TestS3EndpointFormats validates that standard S3 and S3-compatible endpoint
// configurations generate correct URL formats.
//
// Standard S3 (no custom endpoint):
// - Virtual-hosted: https://bucket.s3.region.amazonaws.com/object
// - Path-style: https://s3.region.amazonaws.com/bucket/object
//
// S3-compatible (custom endpoint set):
// - Virtual-hosted: https://bucket.custom-endpoint.com/object
// - Path-style: https://custom-endpoint.com/bucket/object (not recommended)
//
// This test validates URL construction for different S3 configurations to
// ensure proper compatibility with AWS S3 and S3-compatible storage services.
func TestS3EndpointFormats(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		region       string
		bucket       string
		endpoint     string // Empty for standard S3
		pathStyle    bool
		expectedHost string
		expectedPath string
	}{
		{
			name:         "standard_s3/virtual-hosted/us-east-1",
			region:       "us-east-1",
			bucket:       "artifacts",
			endpoint:     "", // Standard S3
			pathStyle:    false,
			expectedHost: "artifacts.s3.us-east-1.amazonaws.com",
			expectedPath: "/",
		},
		{
			name:         "standard_s3/virtual-hosted/eu-west-1",
			region:       "eu-west-1",
			bucket:       "deployments",
			endpoint:     "", // Standard S3
			pathStyle:    false,
			expectedHost: "deployments.s3.eu-west-1.amazonaws.com",
			expectedPath: "/",
		},
		{
			name:         "standard_s3/path-style/us-east-1",
			region:       "us-east-1",
			bucket:       "artifacts",
			endpoint:     "", // Standard S3
			pathStyle:    true,
			expectedHost: "s3.us-east-1.amazonaws.com",
			expectedPath: "/artifacts",
		},
		{
			name:         "standard_s3/path-style/eu-west-1",
			region:       "eu-west-1",
			bucket:       "deployments",
			endpoint:     "", // Standard S3
			pathStyle:    true,
			expectedHost: "s3.eu-west-1.amazonaws.com",
			expectedPath: "/deployments",
		},
		{
			name:         "s3_compatible/virtual-hosted/minio",
			region:       "us-east-1",
			bucket:       "artifacts",
			endpoint:     "https://minio.example.com",
			pathStyle:    false,
			expectedHost: "artifacts.minio.example.com",
			expectedPath: "/",
		},
		{
			name:         "s3_compatible/virtual-hosted/custom",
			region:       "custom-region",
			bucket:       "test-bucket",
			endpoint:     "https://s3.custom-domain.io",
			pathStyle:    false,
			expectedHost: "test-bucket.s3.custom-domain.io",
			expectedPath: "/",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Track the last health check request
			var lastHealthCheckRequest *http.Request
			requestCount := 0

			// Mock S3 server for this specific test
			srv := httptest.NewTLSServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					requestCount++
					t.Logf("Request #%d - Host: %s, Path: %s",
						requestCount, r.Host, r.URL.Path)

					// Capture the last HEAD request (should be the health check)
					if r.Method == http.MethodHead && requestCount >= 2 {
						lastHealthCheckRequest = r
					}

					w.WriteHeader(http.StatusOK)
				},
			))
			defer srv.Close()

			srvURL, err := url.Parse(srv.URL)
			if err != nil {
				t.Fatal(err)
			}

			var d net.Dialer

			// Create custom transport that redirects to mock server
			customTransport := &http.Transport{
				DialContext: func(
					ctx context.Context,
					network, addr string,
				) (net.Conn, error) {
					return d.DialContext(ctx, network, srvURL.Host)
				},
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // Accept test server's self-signed cert
				},
			}

			options := NewOptions().
				SetBucketName(tc.bucket).
				SetRegion(tc.region).
				SetStaticCredentials("test-key", "test-secret", "").
				SetForcePathStyle(tc.pathStyle).
				SetTransport(customTransport)

			// Set custom endpoint only for S3-compatible tests
			if tc.endpoint != "" {
				u, _ := url.Parse(tc.endpoint)
				options.SetURI(u)
			}

			sss, err := New(context.Background(), options)
			if err != nil {
				t.Fatalf("Failed to create S3 client: %v", err)
			}

			err = sss.HealthCheck(context.Background())
			if err != nil {
				t.Fatalf("HealthCheck failed: %v", err)
			}

			// Validate endpoint format from the health check request
			if lastHealthCheckRequest == nil {
				t.Fatal("Health check request was not captured")
			}

			assert.Equal(t, tc.expectedHost, lastHealthCheckRequest.Host,
				"Host should match expected S3 endpoint format for %s", tc.name)
			assert.Equal(t, tc.expectedPath, lastHealthCheckRequest.URL.Path,
				"Path should match expected format for %s", tc.name)

			t.Logf("✓ Validated endpoint format: Host=%s, Path=%s",
				lastHealthCheckRequest.Host, lastHealthCheckRequest.URL.Path)
		})
	}
}

// TestS3CNAMEModeBehavior verifies CNAME mode behavior:
// 1. CNAME mode has NO effect on standard S3 (no custom endpoint)
// 2. CNAME mode DOES work with custom endpoints (OSS, MinIO, etc.)
func TestS3CNAMEModeBehavior(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		region       string
		bucket       string
		endpoint     string
		pathStyle    bool
		bucketURL    bool
		expectedHost string
		expectedPath string
	}{
		{
			name:         "standard_s3/cname_has_no_effect_without_custom_endpoint",
			region:       "us-east-1",
			bucket:       "test-bucket",
			endpoint:     "", // Standard S3
			pathStyle:    false,
			bucketURL:    false, // No effect without custom endpoint
			expectedHost: "test-bucket.s3.us-east-1.amazonaws.com",
			expectedPath: "/",
		},
		{
			name:         "custom_endpoint/cname_enabled",
			region:       "us-east-1",
			bucket:       "test-bucket",
			endpoint:     "https://artifacts.example.com", // OSS CNAME endpoint
			pathStyle:    false,
			bucketURL:    true, // Should activate CNAME addressing
			expectedHost: "artifacts.example.com",
			expectedPath: "/",
		},
		{
			name:         "custom_endpoint/cname_disabled",
			region:       "us-east-1",
			bucket:       "test-bucket",
			endpoint:     "https://minio.example.com",
			pathStyle:    false,
			bucketURL:    false, // Standard virtual-hosted addressing
			expectedHost: "test-bucket.minio.example.com",
			expectedPath: "/",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var lastHealthCheckRequest *http.Request
			requestCount := 0

			srv := httptest.NewTLSServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					requestCount++
					if r.Method == http.MethodHead && requestCount >= 2 {
						lastHealthCheckRequest = r
					}
					w.WriteHeader(http.StatusOK)
				},
			))
			defer srv.Close()

			srvURL, err := url.Parse(srv.URL)
			if err != nil {
				t.Fatal(err)
			}

			var d net.Dialer
			customTransport := &http.Transport{
				DialContext: func(
					ctx context.Context,
					network, addr string,
				) (net.Conn, error) {
					return d.DialContext(ctx, network, srvURL.Host)
				},
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // Accept test server's self-signed cert
				},
			}

			options := NewOptions().
				SetBucketName(tc.bucket).
				SetRegion(tc.region).
				SetStaticCredentials("test-key", "test-secret", "").
				SetForcePathStyle(tc.pathStyle).
				SetTransport(customTransport)

			if tc.endpoint != "" {
				u, _ := url.Parse(tc.endpoint)
				options.SetURI(u).SetLiteralBucketURI(tc.bucketURL)
			} else if tc.bucketURL {
				t.Fatalf("no endpoint specified for test with bucketURL")
			}

			sss, err := New(context.Background(), options)
			if err != nil {
				t.Fatalf("Failed to create S3 client: %v", err)
			}

			err = sss.HealthCheck(context.Background())
			if err != nil {
				t.Fatalf("HealthCheck failed: %v", err)
			}

			if lastHealthCheckRequest == nil {
				t.Fatal("Health check request was not captured")
			}

			// Validate CNAME mode behavior based on configuration
			assert.Equal(t, tc.expectedHost, lastHealthCheckRequest.Host,
				"Host should match expected format for %s", tc.name)
			assert.Equal(t, tc.expectedPath, lastHealthCheckRequest.URL.Path,
				"Path should match expected format for %s", tc.name)

			t.Logf("✓ Validated CNAME mode behavior: Host=%s, Path=%s",
				lastHealthCheckRequest.Host, lastHealthCheckRequest.URL.Path)
		})
	}
}

// TestHealthCheckAlibabaOSS tests health check against Alibaba Cloud OSS endpoints
// This test validates:
// 1. Correct OSS endpoint format (bucket.oss-region.aliyuncs.com) - virtual-hosted style only
// 2. AWS Signature V4 authentication (supported by OSS)
// 3. Proper request method (HEAD for HeadBucket)
// 4. Security token handling
// 5. OSS-specific error responses
//
// Note: Alibaba OSS only supports virtual-hosted style URLs for security reasons.
// Path-style URLs are NOT supported.
func TestHealthCheckAlibabaOSS(t *testing.T) {
	t.Parallel()

	type testCase struct {
		Name string

		// OSS Configuration
		Region        string
		BucketName    string
		UsePathStyle  bool
		AccessKey     string
		SecretKey     string
		SecurityToken string

		// Expected request validation
		ExpectedHost   string
		ExpectedPath   string
		ExpectedMethod string

		// Mock server response
		MockStatusCode int
		MockBody       string

		// Expected test result
		ExpectError    bool
		ErrorValidator func(t *testing.T, err error)
	}

	testCases := []testCase{
		{
			Name:           "ok/oss virtual-hosted style cn-hangzhou",
			Region:         "oss-cn-hangzhou",
			BucketName:     "mender-artifacts",
			UsePathStyle:   false,
			AccessKey:      "LTAI5tFakeAccessKey",
			SecretKey:      "FakeSecretKey123456",
			SecurityToken:  "CAIShwJ1q6Ft5B2yfSjIr5b1FakeToken",
			ExpectedHost:   "mender-artifacts.oss-cn-hangzhou.aliyuncs.com",
			ExpectedPath:   "/",
			ExpectedMethod: http.MethodHead,
			MockStatusCode: http.StatusOK,
			ExpectError:    false,
		},
		{
			Name:           "error/bucket not found",
			Region:         "oss-cn-beijing",
			BucketName:     "nonexistent-bucket",
			UsePathStyle:   false,
			AccessKey:      "LTAI5tFakeAccessKey",
			SecretKey:      "FakeSecretKey123456",
			SecurityToken:  "",
			ExpectedHost:   "nonexistent-bucket.oss-cn-beijing.aliyuncs.com",
			ExpectedPath:   "/",
			ExpectedMethod: http.MethodHead,
			MockStatusCode: http.StatusNotFound,
			MockBody: `<?xml version="1.0" encoding="UTF-8"?>
<Error>
  <Code>NoSuchBucket</Code>
  <Message>The specified bucket does not exist.</Message>
  <RequestId>5C3D9175B6FC201293AD****</RequestId>
  <HostId>test-bucket.oss-cn-beijing.aliyuncs.com</HostId>
  <BucketName>nonexistent-bucket</BucketName>
</Error>`,
			ExpectError: true,
		},
		{
			Name:           "error/access denied",
			Region:         "oss-cn-shanghai",
			BucketName:     "restricted-bucket",
			UsePathStyle:   false,
			AccessKey:      "LTAI5tInvalidKey",
			SecretKey:      "InvalidSecret",
			SecurityToken:  "",
			ExpectedHost:   "restricted-bucket.oss-cn-shanghai.aliyuncs.com",
			ExpectedPath:   "/",
			ExpectedMethod: http.MethodHead,
			MockStatusCode: http.StatusForbidden,
			MockBody: `<?xml version="1.0" encoding="UTF-8"?>
<Error>
  <Code>AccessDenied</Code>
  <Message>Access Denied</Message>
  <RequestId>5C3D8982A7C91C9C4B99****</RequestId>
  <HostId>restricted-bucket.oss-cn-shanghai.aliyuncs.com</HostId>
</Error>`,
			ExpectError: true,
		},
		{
			Name:           "ok/oss singapore region",
			Region:         "oss-ap-southeast-1",
			BucketName:     "singapore-bucket",
			UsePathStyle:   false,
			AccessKey:      "LTAI5tFakeAccessKey",
			SecretKey:      "FakeSecretKey123456",
			SecurityToken:  "",
			ExpectedHost:   "singapore-bucket.oss-ap-southeast-1.aliyuncs.com",
			ExpectedPath:   "/",
			ExpectedMethod: http.MethodHead,
			MockStatusCode: http.StatusOK,
			ExpectError:    false,
		},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			// Track requests received by mock server
			requestCount := 0
			var healthCheckRequest *http.Request

			// Create mock OSS server with TLS
			srv := httptest.NewTLSServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					requestCount++
					t.Logf("[Mock OSS] Request #%d received:", requestCount)
					t.Logf("  Method: %s", r.Method)
					t.Logf("  URL: %s", r.URL.String())
					t.Logf("  Host: %s", r.Host)
					t.Logf("  Path: %s", r.URL.Path)

					// During New() initialization, we expect:
					// 1. HEAD (HeadBucket) - check if bucket exists
					// 2. PUT (CreateBucket) - only if bucket doesn't exist
					// 3. HEAD (HeadBucket) - verify bucket exists after creation
					//
					// Then during HealthCheck():
					// 4. HEAD (HeadBucket) - the actual health check

					if r.Method == http.MethodHead {
						// This could be initialization check or health check
						if requestCount >= 2 {
							// This is likely the health check call
							healthCheckRequest = r

							// Validate request for health check
							assert.Equal(t, tc.ExpectedMethod, r.Method,
								"Request method should match OSS HeadBucket operation")

							// Validate Host header matches OSS endpoint format
							assert.Equal(t, tc.ExpectedHost, r.Host,
								"Host header should match OSS endpoint format")

							// Validate path
							assert.Equal(t, tc.ExpectedPath, r.URL.Path,
								"URL path should match expected OSS bucket path")

							// Validate AWS Signature V4 headers (OSS supports SigV4)
							authHeader := r.Header.Get("Authorization")
							assert.NotEmpty(t, authHeader, "Authorization header must be present")
							assert.True(t, strings.HasPrefix(authHeader, "AWS4-HMAC-SHA256"),
								"OSS should receive AWS Signature V4 authentication")

							// Validate X-Amz-Date header
							amzDate := r.Header.Get("X-Amz-Date")
							assert.NotEmpty(t, amzDate, "X-Amz-Date header must be present")
							_, err := time.Parse("20060102T150405Z", amzDate)
							assert.NoError(t, err, "X-Amz-Date must be in correct format")

							// Validate security token if provided
							if tc.SecurityToken != "" {
								token := r.Header.Get("X-Amz-Security-Token")
								assert.Equal(t, tc.SecurityToken, token,
									"Security token should be included in request")
							}

							// Validate Content-SHA256 header (required by OSS)
							contentSHA := r.Header.Get("X-Amz-Content-Sha256")
							assert.NotEmpty(t, contentSHA,
								"X-Amz-Content-Sha256 header required for OSS")

							// Validate credential contains correct region
							if strings.Contains(authHeader, "Credential=") {
								assert.Contains(t, authHeader, tc.Region,
									"Authorization credential should contain OSS region")
							}
						}

						// Respond to HEAD requests (both init and health check)
						w.WriteHeader(tc.MockStatusCode)
						if tc.MockBody != "" {
							w.Header().Set("Content-Type", "application/xml")
							_, _ = w.Write([]byte(tc.MockBody))
						}
					} else if r.Method == http.MethodPut {
						// CreateBucket during initialization
						w.WriteHeader(http.StatusOK)
					} else {
						t.Errorf("Unexpected HTTP method: %s", r.Method)
						w.WriteHeader(http.StatusBadRequest)
					}
				},
			))
			defer srv.Close()

			srvURL, err := url.Parse(srv.URL)
			require.NoError(t, err)

			// Create HTTP client that redirects to mock server
			var d net.Dialer
			httpTransport := &http.Transport{
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					t.Logf("[Dialer] Intercepting connection to %s", addr)
					t.Logf("[Dialer] Redirecting to mock OSS server: %s", srvURL.Host)
					return d.DialContext(ctx, network, srvURL.Host)
				},
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // Accept test server's self-signed cert
				},
			}

			// Configure S3 client for OSS
			ossEndpoint, _ := url.Parse(tc.Region + ".aliyuncs.com")
			options := NewOptions().
				SetBucketName(tc.BucketName).
				SetRegion(tc.Region).
				SetStaticCredentials(tc.AccessKey, tc.SecretKey, tc.SecurityToken).
				SetURI(ossEndpoint).
				SetForcePathStyle(tc.UsePathStyle).
				SetTransport(httpTransport)

			t.Logf("OSS Configuration:")
			t.Logf("  Endpoint: %s", ossEndpoint)
			t.Logf("  Region: %s", tc.Region)
			t.Logf("  Bucket: %s", tc.BucketName)
			t.Logf("  Path Style: %t", tc.UsePathStyle)
			t.Logf("  Expected Host: %s", tc.ExpectedHost)
			t.Logf("  Expected Path: %s", tc.ExpectedPath)

			// For error test cases, we expect failure during New() initialization
			// For success cases, we expect HealthCheck() to succeed
			if tc.ExpectError {
				// Create with short context timeout to avoid long waits
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
				defer cancel()

				_, err := New(ctx, options)
				assert.Error(t, err, "New() should fail during initialization for %s", tc.Name)
				if tc.ErrorValidator != nil {
					tc.ErrorValidator(t, err)
				}
				t.Logf("Total requests during failed initialization: %d", requestCount)
			} else {
				sss, err := New(context.Background(), options)
				require.NoError(t, err, "Failed to create S3 client for OSS")

				// Execute HealthCheck
				err = sss.HealthCheck(context.Background())
				assert.NoError(t, err, "HealthCheck should succeed for %s", tc.Name)

				// Ensure health check request was actually made
				assert.NotNil(t, healthCheckRequest,
					"Mock server should have received a health check request")
				t.Logf("Total requests received during test: %d", requestCount)
			}
		})
	}
}

// TestHealthCheckAlibabaOSSEndpointFormats validates that OSS virtual-hosted
// style endpoint configurations are handled correctly.
//
// Note: Alibaba OSS only supports virtual-hosted style URLs for security reasons.
// Path-style URLs are NOT supported and have been removed from these tests.
func TestHealthCheckAlibabaOSSEndpointFormats(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		region       string
		bucket       string
		pathStyle    bool
		expectedHost string
		expectedPath string
	}{
		{
			name:         "virtual-hosted/cn-hangzhou",
			region:       "oss-cn-hangzhou",
			bucket:       "my-bucket",
			pathStyle:    false,
			expectedHost: "my-bucket.oss-cn-hangzhou.aliyuncs.com",
			expectedPath: "/",
		},
		{
			name:         "virtual-hosted/us-east-1",
			region:       "oss-us-east-1",
			bucket:       "artifacts",
			pathStyle:    false,
			expectedHost: "artifacts.oss-us-east-1.aliyuncs.com",
			expectedPath: "/",
		},
		{
			name:         "virtual-hosted/eu-central-1",
			region:       "oss-eu-central-1",
			bucket:       "deployments",
			pathStyle:    false,
			expectedHost: "deployments.oss-eu-central-1.aliyuncs.com",
			expectedPath: "/",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Track the last health check request
			var lastHealthCheckRequest *http.Request
			requestCount := 0

			// Mock OSS server for this specific test
			srv := httptest.NewTLSServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					requestCount++
					t.Logf("Request #%d - Host: %s, Path: %s",
						requestCount, r.Host, r.URL.Path)

					// Capture the last HEAD request (should be the health check)
					if r.Method == http.MethodHead && requestCount >= 2 {
						lastHealthCheckRequest = r
					}

					w.WriteHeader(http.StatusOK)
				},
			))
			defer srv.Close()

			srvURL, err := url.Parse(srv.URL)
			require.NoError(t, err)

			var d net.Dialer

			// Create custom transport that redirects to mock server
			customTransport := &http.Transport{
				DialContext: func(
					ctx context.Context,
					network, addr string,
				) (net.Conn, error) {
					return d.DialContext(ctx, network, srvURL.Host)
				},
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // Accept test server's self-signed cert
				},
			}

			u, _ := url.Parse("https://" + tc.region + ".aliyuncs.com")
			options := NewOptions().
				SetBucketName(tc.bucket).
				SetRegion(tc.region).
				SetStaticCredentials("test", "secret", "").
				SetURI(u).
				SetForcePathStyle(tc.pathStyle).
				SetTransport(customTransport)

			sss, err := New(context.Background(), options)
			require.NoError(t, err)

			err = sss.HealthCheck(context.Background())
			require.NoError(t, err)

			// Validate endpoint format from the health check request
			require.NotNil(t, lastHealthCheckRequest,
				"Health check request should have been captured")
			assert.Equal(t, tc.expectedHost, lastHealthCheckRequest.Host,
				"Host should match OSS endpoint format for %s", tc.name)
			assert.Equal(t, tc.expectedPath, lastHealthCheckRequest.URL.Path,
				"Path should match expected format for %s", tc.name)

			t.Logf("✓ Validated endpoint format: Host=%s, Path=%s",
				lastHealthCheckRequest.Host, lastHealthCheckRequest.URL.Path)
		})
	}
}
