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
package mongo

import (
	"os"
	"testing"

	mtesting "github.com/mendersoftware/mender-server/pkg/mongo/v2/testing"
)

var db mtesting.TestDBRunner

// Overwrites test execution and allows for test database setup
func TestMain(m *testing.M) {

	status := mtesting.WithDB(func(d mtesting.TestDBRunner) int {
		db = d
		defer db.Client().Disconnect(db.CTX())
		return m.Run()
	}, nil)

	os.Exit(status)
}
