// Copyright 2026 Northern.tech AS
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
package migrate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"

	. "github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate"
	"github.com/mendersoftware/mender-server/pkg/store/v2"
)

func TestGetTenantDbs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestGetTenantDbs in short mode.")
	}

	baseDb := "servicename"
	testCases := map[string]struct {
		dbs []string
	}{
		"no tenant dbs": {
			dbs: []string{},
		},
		"1 tenant db": {
			dbs: []string{
				baseDb + "-tenant1",
			},
		},
		">1 tenant db": {
			dbs: []string{
				baseDb + "-tenant1",
				baseDb + "-tenant2",
				baseDb + "-tenant3",
			},
		},
	}
	for name := range testCases {
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {
			db.Wipe()
			client := db.Client()

			// dummy insert on test dbs to create them
			for _, dbname := range tc.dbs {
				_, err := client.Database(dbname).
					Collection("foo").
					InsertOne(db.CTX(), bson.M{"foo": "bar"})
				assert.NoError(t, err)
			}

			res, err := GetTenantDbs(db.CTX(), client, store.IsTenantDb(baseDb))
			assert.NoError(t, err)
			assert.Equal(t, tc.dbs, res)
		})
	}
}
