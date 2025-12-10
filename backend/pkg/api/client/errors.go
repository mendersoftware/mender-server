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
	"errors"
	"fmt"
	"io"
	"net/http"

	rest "github.com/mendersoftware/mender-server/pkg/rest.utils"
)

// HTTPError represents an HTTP error response from an API.
// It uses rest.Error to reuse the standard API error structure.
type HTTPError struct {
	// APIError contains the parsed error body from the API response.
	APIError *rest.Error
	// StatusCode is the HTTP status code.
	StatusCode int
	// Status is the HTTP status text (e.g., "404 Not Found").
	Status string
}

// Error implements the error interface.
func (e *HTTPError) Error() string {
	if e.APIError != nil && e.APIError.Err != "" {
		return fmt.Sprintf("%s: %s", e.Status, e.APIError.Err)
	}
	return e.Status
}

// Message returns the error message from the API response body.
func (e *HTTPError) Message() string {
	if e.APIError != nil {
		return e.APIError.Err
	}
	return ""
}

// RequestID returns the request ID from the API response body, if available.
func (e *HTTPError) RequestID() string {
	if e.APIError != nil {
		return e.APIError.RequestID
	}
	return ""
}

// Code returns the HTTP status code.
func (e *HTTPError) Code() int {
	return e.StatusCode
}

// IsHTTPError checks if the error is an HTTPError.
func IsHTTPError(err error) bool {
	var httpErr *HTTPError
	return errors.As(err, &httpErr)
}

// GetHTTPError extracts an HTTPError from an error, if present.
func GetHTTPError(err error) *HTTPError {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr
	}
	return nil
}

// IsNotFound returns true if the error represents a 404 Not Found response.
func IsNotFound(err error) bool {
	if httpErr := GetHTTPError(err); httpErr != nil {
		return httpErr.StatusCode == http.StatusNotFound
	}
	return false
}

// IsConflict returns true if the error represents a 409 Conflict response.
func IsConflict(err error) bool {
	if httpErr := GetHTTPError(err); httpErr != nil {
		return httpErr.StatusCode == http.StatusConflict
	}
	return false
}

// IsUnauthorized returns true if the error represents a 401 Unauthorized response.
func IsUnauthorized(err error) bool {
	if httpErr := GetHTTPError(err); httpErr != nil {
		return httpErr.StatusCode == http.StatusUnauthorized
	}
	return false
}

// IsForbidden returns true if the error represents a 403 Forbidden response.
func IsForbidden(err error) bool {
	if httpErr := GetHTTPError(err); httpErr != nil {
		return httpErr.StatusCode == http.StatusForbidden
	}
	return false
}

// NewHTTPError creates a new HTTPError from an HTTP response.
// It attempts to parse the response body as a rest.Error.
func NewHTTPError(resp *http.Response) *HTTPError {
	httpErr := &HTTPError{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
	}

	// Try to parse the body as a rest.Error
	if resp.Body != nil {
		if apiErr := rest.ParseApiError(resp.Body); apiErr != nil {
			if restErr, ok := apiErr.(*rest.Error); ok {
				httpErr.APIError = restErr
			}
		}
	}

	return httpErr
}

// NewHTTPErrorFromReader creates a new HTTPError from an HTTP response,
// reading the error body from the provided reader.
func NewHTTPErrorFromReader(resp *http.Response, body io.Reader) *HTTPError {
	httpErr := &HTTPError{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
	}

	// Try to parse the body as a rest.Error
	if body != nil {
		if apiErr := rest.ParseApiError(body); apiErr != nil {
			if restErr, ok := apiErr.(*rest.Error); ok {
				httpErr.APIError = restErr
			}
		}
	}

	return httpErr
}
