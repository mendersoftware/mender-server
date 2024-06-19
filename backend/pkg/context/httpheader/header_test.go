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
package httpheader

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHttpHeader(t *testing.T) {

	ctxb := context.Background()
	ctx := WithContext(ctxb, nil, "foo")
	assert.Equal(t, ctxb, ctx)

	ctx = WithContext(ctxb, http.Header{}, "foo")
	assert.Equal(t, ctxb, ctx)

	ctx = WithContext(ctxb,
		http.Header{
			"Authorization":     []string{"foo"},
			"X-Mender-Identity": []string{"barbar"},
		},
		"Authorization", "Foobar", "X-Mender-Identity")
	assert.NotNil(t, ctx)
	assert.NotEqual(t, ctxb, ctx)

	assert.Equal(t, "foo", FromContext(ctx, "Authorization"))
	assert.Equal(t, "", FromContext(ctx, "Foobar"))
	assert.Equal(t, "barbar", FromContext(ctx, "X-Mender-Identity"))

	assert.Nil(t, ctx.Value("X-Mender-Identity"))
}
