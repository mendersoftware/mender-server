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
package plan

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsMinimum(t *testing.T) {
	var rc bool

	rc = IsHigherOrEqual("undefinedplan", "undefinedplantoo")
	assert.False(t, rc, "two undefined plans: return false")

	rc = IsHigherOrEqual("undefinedplan", PlanOpenSource)
	assert.False(t, rc, "first argument undefined plan: return false")

	rc = IsHigherOrEqual(PlanOpenSource, "undefinedplantoo")
	assert.False(t, rc, "second argument undefined plan: return false")

	rc = IsHigherOrEqual(PlanOpenSource, PlanOpenSource)
	assert.True(t, rc, "os plan is higher or equal os plan")

	rc = IsHigherOrEqual(PlanProfessional, PlanProfessional)
	assert.True(t, rc, "pro plan is higher or equal pro plan")

	rc = IsHigherOrEqual(PlanEnterprise, PlanEnterprise)
	assert.True(t, rc, "ent plan is higher or equal ent plan")

	rc = IsHigherOrEqual(PlanProfessional, PlanEnterprise)
	assert.False(t, rc, "pro plan is not higher or equal ent plan")

	rc = IsHigherOrEqual(PlanOpenSource, PlanEnterprise)
	assert.False(t, rc, "os plan is not higher or equal ent plan")

	rc = IsHigherOrEqual(PlanProfessional, PlanOpenSource)
	assert.True(t, rc, "pro plan is higher or equal os plan")

	rc = IsHigherOrEqual(PlanEnterprise, PlanOpenSource)
	assert.True(t, rc, "ent plan is higher or equal os plan")

	rc = IsHigherOrEqual(PlanEnterprise, PlanProfessional)
	assert.True(t, rc, "ent plan is higher or equal pro plan")

}
