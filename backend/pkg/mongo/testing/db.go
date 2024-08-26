// Copyright 2024 Northern.tech AS
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
	"os"

	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mendersoftware/mender-server/pkg/mongo/dbtest"
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
// reg is optional custom registry which can be set up for the test client.
func WithDB(f func(dbtest TestDBRunner) int, reg *bsoncodec.Registry) int {
	var runner TestDBRunner
	if url, ok := os.LookupEnv("TEST_MONGO_URL"); ok {
		clientOpts := options.Client().
			ApplyURI(url)
		if reg != nil {
			clientOpts.SetRegistry(reg)
		}
		client, err := mongo.Connect(context.Background(), clientOpts)
		if err != nil {
			panic(err)
		}
		runner = (*dbClientFromEnv)(client)
	} else {
		// Fallback to running mongod on host
		dbdir, _ := os.MkdirTemp("", "dbsetup-test")
		db := &dbtest.DBServer{}
		db.SetPath(dbdir)
		if reg != nil {
			db.SetRegistry(reg)
		}
		runner = db

		defer os.RemoveAll(dbdir)
		defer db.Stop()

	}

	return f(runner)
}
