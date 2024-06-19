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
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type dbClientFromEnv mongo.Client

// CTX is required for testing.DBTestRunner
func (self *dbClientFromEnv) CTX() context.Context {
	return context.TODO()
}

func (self *dbClientFromEnv) Client() *mongo.Client {
	return (*mongo.Client)(self)
}

func (self *dbClientFromEnv) Wipe() {
	client := self.Client()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	names, err := client.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		panic(err)
	}
	for _, name := range names {
		switch name {
		case "admin", "local", "config":
		default:
			err = client.
				Database(name).
				Drop(ctx)
			if err != nil {
				panic(err)
			}
		}
	}
}
