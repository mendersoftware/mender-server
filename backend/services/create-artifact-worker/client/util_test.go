// Copyright 2020 Northern.tech AS
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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoin(t *testing.T) {
	s, err := join("http://server",
		"/api/internal/v1/deployments/tenants/{id}/artifacts",
		map[string]string{"id": "foo"})
	assert.NoError(t, err)
	assert.Equal(t,
		"http://server/api/internal/v1/deployments/tenants/foo/artifacts",
		s)

	s, err = join("http://server/",
		"/api/internal/v1/deployments/tenants/{id}/artifacts",
		map[string]string{"id": "foo"})
	assert.NoError(t, err)
	assert.Equal(t,
		"http://server/api/internal/v1/deployments/tenants/foo/artifacts",
		s)

	s, err = join("https://server:9000",
		"/api/internal/v1/deployments/tenants/{id}/artifacts",
		map[string]string{"id": "foo"})
	assert.NoError(t, err)
	assert.Equal(t,
		"https://server:9000/api/internal/v1/deployments/tenants/foo/artifacts",
		s)
}
