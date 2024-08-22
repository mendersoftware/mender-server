// Copyright 2020 Northern.tech AS
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

package mongo

import (
	"flag"
	"os"
	"testing"

	mtesting "github.com/mendersoftware/mender-server/pkg/mongo/testing"
)

var db mtesting.TestDBRunner

// Overwrites test execution and allows for test database setup
func TestMain(m *testing.M) {
	var status int
	if !flag.Parsed() {
		flag.Parse()
	}
	if !testing.Short() {
		status = mtesting.WithDB(func(dbtest mtesting.TestDBRunner) int {
			db = dbtest
			return m.Run()
		}, nil)
	} else {
		status = m.Run()
	}

	os.Exit(status)
}
