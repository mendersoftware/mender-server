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

package testing

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/mendersoftware/mender-server/pkg/mongo/dbtest"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestDBRunner exports selected calls of dbtest.DBServer API, just the ones
// that are useful in tests.
type TestDBRunner interface {
	Client() *mongo.Client
	Wipe()
	CTX() context.Context
}

// WithDB will set up a test DB instance and pass it to `f` callback as
// `dbtest`. Once `f()` is finished, the DB will be cleaned up. Value returned
// from `f()` is obtained as return status of a call to WithDB().
func WithDB(f func(dbtest TestDBRunner) int) int {
	var runner TestDBRunner
	if url, ok := os.LookupEnv("TEST_MONGO_URL"); ok {
		clientOpts := options.Client().
			ApplyURI(url)
		client, err := mongo.Connect(context.Background(), clientOpts)
		if err != nil {
			panic(err)
		}
		runner = (*dbClientFromEnv)(client)
	} else {
		// Fallback to running mongod on host
		dbdir, _ := ioutil.TempDir("", "dbsetup-test")
		db := &dbtest.DBServer{}
		db.SetPath(dbdir)
		runner = db

		defer os.RemoveAll(dbdir)
		defer db.Stop()

	}

	return f(runner)
}
