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
package requestid

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// FromContext extracts current request Id from context.Context
func TestContext(t *testing.T) {

	assert.Equal(t, "", FromContext(context.Background()))
	assert.Equal(t, "foo",
		FromContext(WithContext(context.Background(), "foo")))

	// make sure that the helpers are using private types
	assert.Nil(t, WithContext(context.Background(), "foo").Value(0))
}
