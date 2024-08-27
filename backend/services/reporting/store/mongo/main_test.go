// Copyright 2022 Northern.tech AS
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
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	mopts "go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
)

var (
	client   *mongo.Client
	mongoURL string
)

func TestMain(m *testing.M) {
	if url, ok := os.LookupEnv("TEST_MONGO_URL"); ok {
		var err error
		mongoURL = url
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		client, err = mongo.Connect(ctx, mopts.Client().ApplyURI(url))
		cancel()
		if err != nil {
			log.Printf("failed to connect to MongoDB: %s", err.Error())
			os.Exit(1)
		}
	} else {
		log.Println("TEST_MONGO_URL must be set to run unit tests")
		os.Exit(1)
	}
	os.Exit(m.Run())
}

var dbNameReplacer = strings.NewReplacer(
	`/`, ``, `\`, ``, `.`, ``, ` `, ``,
	`"`, ``, `$`, ``, `*`, ``, `<`, ``,
	`>`, ``, `:`, ``, `|`, ``, `?`, ``,
)

// legalizeDbName ensures the database name does not contain illegal characters
// and that the length does not exceed the maximum 64 characters.
func legalizeDbName(testName string) string {
	dbName := dbNameReplacer.Replace(testName)
	if len(dbName) >= 64 {
		dbName = dbName[len(dbName)-64:]
	}
	return dbName
}

// GetTestDataStore creates a new DataStoreMongo with the database name
// set to the test name (is safe to call inside subtests, but be aware that
// t.Name() is different from inside and outside of t.Run scope).
// Make sure you always defer DataStore.DropDatabase inside tests to free
// up storage.
func GetTestDataStore(t *testing.T) *MongoStore {
	dbName := legalizeDbName(t.Name())
	return &MongoStore{
		client: client,
		config: MongoStoreConfig{
			DbName: dbName,
		},
	}
}

// GetTestDatabase as function above returns the test-local database.
func GetTestDatabase(ctx context.Context, t *testing.T) *mongo.Database {
	return client.Database(DbName)
}
