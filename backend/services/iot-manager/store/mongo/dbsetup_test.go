// Copyright 2024 Northern.tech AS
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
	"context"
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	mgopts "go.mongodb.org/mongo-driver/mongo/options"

	mtesting "github.com/mendersoftware/mender-server/pkg/mongo/testing"
)

var db mtesting.TestDBRunner

type externalDBRunner mongo.Client

func (ext *externalDBRunner) Client() *mongo.Client { return (*mongo.Client)(ext) }

func (ext *externalDBRunner) CTX() context.Context { return context.Background() }

func (ext *externalDBRunner) Wipe() {
	client := (*mongo.Client)(ext)
	err := client.Database(DbName).Drop(ext.CTX())
	if err != nil {
		panic(err)
	}
}

// Overwrites test execution and allows for test database setup
func TestMain(m *testing.M) {
	var status int
	if url := os.Getenv("MONGO_URL"); url != "" {
		client, err := mongo.Connect(
			context.Background(),
			mgopts.Client().ApplyURI(url).SetRegistry(newRegistry()),
		)
		if err != nil {
			panic(err)

		}
		db = (*externalDBRunner)(client)
		status = m.Run()
	} else {
		status = mtesting.WithDB(func(d mtesting.TestDBRunner) int {
			db = d
			defer func() {
				err := db.Client().Disconnect(db.CTX())
				if err != nil {
					panic(err)
				}
			}()
			return m.Run()
		},
			newRegistry(),
		)
	}
	os.Exit(status)
}
