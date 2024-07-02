// Copyright 2023 Northern.tech AS
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
package identity

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {

	identity := Identity{
		Subject: "foo",
		Tenant:  "bar",
	}
	assert.Empty(t, FromContext(context.Background()))
	assert.Equal(t, &identity, FromContext(WithContext(context.Background(), &identity)))

	ctx := WithContext(context.Background(), &identity)
	// trying to fetch with same value but different type should fail
	assert.Nil(t, ctx.Value(0))
}
