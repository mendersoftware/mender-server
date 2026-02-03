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

package s3

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mendersoftware/mender-server/services/deployments/storage"
)

const (
	testTimeout       = 5 * time.Second
	testCNAMEEndpoint = "artifacts.example.cn"
	testBucketName    = "test-bucket-china"
	testRegion        = "oss-cn-shanghai"
	testTenantID      = "aabbccdd11223344556677aa"
	testObjectID      = "ffffffff-eeee-dddd-cccc-999999999999"
	testAccessKey     = "LTAI5tFakeAccessKey"
	testSecretKey     = "FakeSecretKey123456"
)

// TestOSSCNAMEAddressing tests Alibaba OSS CNAME (custom domain) addressing mode.
//
// OSS CNAME Requirements (for Mainland China deployments):
// - Endpoint: Custom domain (e.g., https://artifacts.example.cn)
// - Bucket name: Must NOT appear in the URL
// - URL format: https://artifacts.example.cn/object-path
//
// This is DIFFERENT from standard S3:
// - S3 virtual-hosted: https://bucket.s3.region.amazonaws.com/object-path
// - S3 path-style: https://s3.region.amazonaws.com/bucket/object-path
//
// IMPORTANT: Starting March 20, 2025, CNAME is REQUIRED for new OSS users in
// Chinese mainland regions to perform data API operations. Default public
// endpoints are restricted.
//
// References:
// - CNAME Documentation: https://www.alibabacloud.com/help/en/oss/user-guide/access-buckets-via-custom-domain-names
// - Error Codes: https://www.alibabacloud.com/help/en/oss/support/overview-14/
// - Policy Change: https://www.alibabacloud.com/help/en/doc-detail/195675.html
//
// NOTE: These tests will FAIL with the current implementation.
// They will PASS once OSS CNAME addressing is properly implemented.
func TestOSSCNAMEAddressing(t *testing.T) {
	t.Parallel()

	objectPath := testTenantID + "/" + testObjectID

	type testCase struct {
		Name         string
		ObjectPath   string
		Operation    func(t *testing.T, ctx context.Context, client storage.ObjectStorage, objectPath string) error
		ValidateReqs func(t *testing.T, requests []*http.Request, objectPath string)
	}

	testCases := []testCase{
		{
			Name:       "initialize_with_CNAME",
			ObjectPath: objectPath,
			Operation: func(t *testing.T, ctx context.Context, client storage.ObjectStorage, _ string) error {
				// Initialization already happened, test HealthCheck
				return client.HealthCheck(ctx)
			},
			ValidateReqs: func(t *testing.T, requests []*http.Request, _ string) {
				require.NotEmpty(t, requests, "Should have made initialization requests")

				for _, req := range requests {
					assert.Equal(t, testCNAMEEndpoint, req.Host,
						"Host should be CNAME endpoint (no bucket subdomain)")
					assert.NotContains(t, req.Host, testBucketName,
						"Bucket name should not appear in Host")
					assert.NotContains(t, req.URL.Path, "/"+testBucketName+"/",
						"Bucket name should not appear in URL path")
				}
			},
		},
		{
			Name:       "get_object_with_CNAME",
			ObjectPath: objectPath,
			Operation: func(t *testing.T, ctx context.Context, client storage.ObjectStorage, objectPath string) error {
				obj, err := client.GetObject(ctx, objectPath)
				if err != nil {
					return err
				}
				defer obj.Close()
				_, _ = io.ReadAll(obj)
				return nil
			},
			ValidateReqs: func(t *testing.T, requests []*http.Request, objectPath string) {
				var getRequest *http.Request
				for _, req := range requests {
					if req.Method == http.MethodGet {
						getRequest = req
						break
					}
				}
				require.NotNil(t, getRequest, "Should have found GET request")

				assert.Equal(t, testCNAMEEndpoint, getRequest.Host,
					"GetObject Host must be CNAME endpoint without bucket subdomain")
				assert.Equal(t, "/"+objectPath, getRequest.URL.Path,
					"GetObject path must be /object-path (no bucket in path)")
			},
		},
		{
			Name:       "put_object_with_CNAME",
			ObjectPath: objectPath,
			Operation: func(t *testing.T, ctx context.Context, client storage.ObjectStorage, objectPath string) error {
				testData := "test artifact data"
				return client.PutObject(ctx, objectPath, strings.NewReader(testData))
			},
			ValidateReqs: func(t *testing.T, requests []*http.Request, objectPath string) {
				var putRequest *http.Request
				for _, req := range requests {
					if req.Method == http.MethodPut {
						putRequest = req
						break
					}
				}
				require.NotNil(t, putRequest, "Should have found PUT request")

				assert.Equal(t, testCNAMEEndpoint, putRequest.Host,
					"PutObject Host must be CNAME endpoint without bucket subdomain")
				assert.Equal(t, "/"+objectPath, putRequest.URL.Path,
					"PutObject path must be /object-path (no bucket in path)")
			},
		},
		{
			Name:       "presigned_URL_with_CNAME",
			ObjectPath: objectPath,
			Operation: func(t *testing.T, ctx context.Context, client storage.ObjectStorage, objectPath string) error {
				link, err := client.GetRequest(ctx, objectPath, "", 15*time.Minute, true)
				if err != nil {
					return err
				}

				presignedURL, err := url.Parse(link.Uri)
				require.NoError(t, err, "Should be able to parse presigned URL")

				assert.Equal(t, testCNAMEEndpoint, presignedURL.Host,
					"Presigned URL Host must be CNAME endpoint without bucket subdomain")
				assert.Equal(t, "/"+objectPath, presignedURL.Path,
					"Presigned URL path must be /object-path (no bucket in path)")
				assert.NotEmpty(t, presignedURL.Query().Get("X-Amz-Algorithm"),
					"Presigned URL must have signature")

				return nil
			},
			ValidateReqs: func(t *testing.T, requests []*http.Request, _ string) {
				// Presigned URLs don't make HTTP requests during generation
				// Validation is done in the Operation function
			},
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			// Track requests for this specific subtest
			var requests []*http.Request

			// Create isolated mock server for this subtest
			srv := httptest.NewServer(createCNAMEHandler(t, &requests))
			defer srv.Close()

			// Create isolated HTTP transport for this subtest
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

			// Configure S3 client for OSS CNAME
			options := NewOptions().
				SetBucketName(testBucketName).
				SetRegion(testRegion).
				SetStaticCredentials(testAccessKey, testSecretKey, "").
				SetURI("https://" + testCNAMEEndpoint).
				SetExternalURI("https://" + testCNAMEEndpoint).
				SetForcePathStyle(false).
				SetTransport(httpTransport)

			ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
			defer cancel()

			// Initialize client
			client, err := New(ctx, options)
			require.NoError(t, err, "Should be able to initialize S3 client with CNAME endpoint")

			// Clear initialization requests before running the actual operation
			requests = nil

			// Run the test operation
			err = tc.Operation(t, ctx, client, tc.ObjectPath)
			require.NoError(t, err, "Operation %s should succeed with CNAME", tc.Name)

			// Validate the requests
			tc.ValidateReqs(t, requests, tc.ObjectPath)
		})
	}
}

// createCNAMEHandler creates an HTTP handler that enforces OSS CNAME requirements.
// It simulates the behavior of a real Alibaba OSS endpoint with CNAME addressing.
//
// Mock Behavior vs Real OSS:
//
// 1. Bucket as subdomain (e.g., bucket.custom-domain.com):
//    Mock: Returns 403 SecondLevelDomainForbidden
//    Real: DNS would likely fail to resolve (CNAME only maps custom-domain.com, not bucket.custom-domain.com)
//    Test Purpose: Ensures client never generates bucket.cname-domain URLs
//
// 2. Bucket in URL path (e.g., custom-domain.com/bucket/object):
//    Mock: Returns 403 InvalidBucketName
//    Real: OSS would likely return object not found or similar error
//    Test Purpose: Ensures bucket name doesn't appear in URL path
//
// 3. Public OSS endpoint (e.g., *.aliyuncs.com):
//    Mock: Returns 403 PublicEndpointForbidden
//    Real: Exact behavior - OSS returns this error for China mainland regions (as of March 2025)
//    Reference: https://www.alibabacloud.com/help/en/oss/user-guide/access-buckets-via-custom-domain-names
//
// 4. Correct CNAME format (e.g., custom-domain.com/object-path):
//    Mock: Returns 200 OK
//    Real: OSS accepts and routes to mapped bucket
//    Reference: URL format is https://YourDomainName/ObjectName (no bucket in URL)
//
// The mock is intentionally STRICT to catch any incorrect URL generation patterns.
func createCNAMEHandler(t *testing.T, requests *[]*http.Request) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		*requests = append(*requests, r)

		// REJECT: Bucket as subdomain (virtual-hosted style with CNAME)
		// e.g., test-bucket-china.artifacts.example.cn/object
		// Real behavior: DNS resolution would fail
		// Mock behavior: Return error to catch this pattern in tests
		if strings.HasPrefix(r.Host, testBucketName+".") {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<Error>
  <Code>SecondLevelDomainForbidden</Code>
  <Message>Please use virtual hosted style to access.</Message>
</Error>`))
			return
		}

		// REJECT: Bucket in URL path (path-style)
		// e.g., artifacts.example.cn/test-bucket-china/object
		// Real behavior: OSS would not find object or return error
		// Reference: Bucket name must not appear in CNAME URLs
		if strings.HasPrefix(r.URL.Path, "/"+testBucketName+"/") ||
			r.URL.Path == "/"+testBucketName {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<Error>
  <Code>InvalidBucketName</Code>
  <Message>The bucket name is not allowed to appear in the CNAME URL path.</Message>
</Error>`))
			return
		}

		// REJECT: Public OSS endpoint instead of CNAME
		// e.g., test-bucket-china.oss-cn-shanghai.aliyuncs.com
		// Real behavior: Returns PublicEndpointForbidden (HTTP 400, EC 0048-00000401)
		// Reference: https://www.alibabacloud.com/help/en/oss/user-guide/access-buckets-via-custom-domain-names
		// Required for China mainland regions as of March 20, 2025
		if strings.Contains(r.Host, "aliyuncs.com") {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<Error>
  <Code>PublicEndpointForbidden</Code>
  <Message>Not allowed using the OSS public endpoint, please use CNAME instead.</Message>
</Error>`))
			return
		}

		// ACCEPT: Correct CNAME format
		// e.g., artifacts.example.cn/tenant-id/object-id
		// Real behavior: OSS accepts and routes request to bucket mapped to this CNAME
		// URL format: https://YourDomainName/ObjectName (bucket name not in URL)
		if r.Host == testCNAMEEndpoint {
			switch r.Method {
			case http.MethodHead:
				w.WriteHeader(http.StatusOK)
			case http.MethodGet:
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("mock object data"))
			case http.MethodPut:
				w.WriteHeader(http.StatusOK)
			default:
				w.WriteHeader(http.StatusOK)
			}
			return
		}

		// Unknown format
		w.WriteHeader(http.StatusBadRequest)
	}
}
