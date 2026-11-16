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

package utils_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	. "github.com/mendersoftware/mender-server/pkg/mongo/v2/migrate/utils"
)

const testDbName = "migrate_utils_test"

type srcDoc struct {
	ID    string `bson:"_id"`
	Name  string `bson:"name"`
	Value int    `bson:"value"`
}

type dstDoc struct {
	ID      string `bson:"_id"`
	Name    string `bson:"name"`
	Doubled int    `bson:"doubled"`
}

func transformDoc(doc srcDoc) dstDoc {
	return dstDoc{
		ID:      doc.ID,
		Name:    doc.Name,
		Doubled: doc.Value * 2,
	}
}

func setupCollections(
	t *testing.T,
	srcDocs, dstDocs []any,
) (dst, src *mongo.Collection) {
	t.Helper()
	db.Wipe()
	database := db.Client().Database(testDbName)
	src = database.Collection("src")
	dst = database.Collection("dst")
	if len(srcDocs) > 0 {
		_, err := src.InsertMany(db.CTX(), srcDocs)
		require.NoError(t, err)
	}
	if len(dstDocs) > 0 {
		_, err := dst.InsertMany(db.CTX(), dstDocs)
		require.NoError(t, err)
	}
	return dst, src
}

func findAll[T any](t *testing.T, collection *mongo.Collection) []T {
	t.Helper()
	cur, err := collection.Find(db.CTX(), bson.D{},
		options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}))
	require.NoError(t, err)
	res := []T{}
	require.NoError(t, cur.All(db.CTX(), &res))
	return res
}

func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func TestCopyCollection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestCopyCollection in short mode.")
	}

	testCases := []struct {
		Name string

		CTX     context.Context
		SrcDocs []any
		DstDocs []any

		Expected []dstDoc
		Error    string
	}{{
		Name: "ok",

		SrcDocs: []any{
			srcDoc{ID: "1", Name: "one", Value: 1},
			srcDoc{ID: "2", Name: "two", Value: 2},
			srcDoc{ID: "3", Name: "three", Value: 3},
		},

		Expected: []dstDoc{
			{ID: "1", Name: "one", Doubled: 2},
			{ID: "2", Name: "two", Doubled: 4},
			{ID: "3", Name: "three", Doubled: 6},
		},
	}, {
		Name: "ok/empty source",

		Expected: []dstDoc{},
	}, {
		Name: "ok/appends to existing destination",

		SrcDocs: []any{
			srcDoc{ID: "2", Name: "two", Value: 2},
		},
		DstDocs: []any{
			dstDoc{ID: "1", Name: "one", Doubled: 2},
		},

		Expected: []dstDoc{
			{ID: "1", Name: "one", Doubled: 2},
			{ID: "2", Name: "two", Doubled: 4},
		},
	}, {
		Name: "error/find",

		CTX: canceledContext(),
		SrcDocs: []any{
			srcDoc{ID: "1", Name: "one", Value: 1},
		},

		Expected: []dstDoc{},
		Error:    "failed to copy documents",
	}, {
		Name: "error/decode",

		SrcDocs: []any{
			bson.D{
				{Key: "_id", Value: "1"},
				{Key: "name", Value: "one"},
				{Key: "value", Value: "not a number"},
			},
		},

		Expected: []dstDoc{},
		Error:    "failed to decode document",
	}, {
		Name: "error/insert duplicate key",

		SrcDocs: []any{
			srcDoc{ID: "1", Name: "one", Value: 1},
		},
		DstDocs: []any{
			dstDoc{ID: "1", Name: "existing", Doubled: 0},
		},

		Expected: []dstDoc{
			{ID: "1", Name: "existing", Doubled: 0},
		},
		Error: "failed to copy document",
	}}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			dst, src := setupCollections(t, tc.SrcDocs, tc.DstDocs)
			ctx := tc.CTX
			if ctx == nil {
				ctx = db.CTX()
			}

			err := CopyCollection(ctx, dst, src, transformDoc)
			if tc.Error != "" {
				assert.ErrorContains(t, err, tc.Error)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.Expected, findAll[dstDoc](t, dst))
		})
	}
}

func TestCopyUpsertCollection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestCopyUpsertCollection in short mode.")
	}

	dstQuery := func(doc srcDoc) any {
		return bson.D{{Key: "_id", Value: doc.ID}}
	}
	dstUpdate := func(doc srcDoc) any {
		return bson.D{{Key: "$set", Value: bson.D{
			{Key: "name", Value: doc.Name},
			{Key: "doubled", Value: doc.Value * 2},
		}}}
	}

	testCases := []struct {
		Name string

		CTX       context.Context
		SrcDocs   []any
		DstDocs   []any
		DstQuery  func(doc srcDoc) any
		DstUpdate func(doc srcDoc) any

		Expected []dstDoc
		Error    string
	}{{
		Name: "ok/insert into empty destination",

		SrcDocs: []any{
			srcDoc{ID: "1", Name: "one", Value: 1},
			srcDoc{ID: "2", Name: "two", Value: 2},
		},

		Expected: []dstDoc{
			{ID: "1", Name: "one", Doubled: 2},
			{ID: "2", Name: "two", Doubled: 4},
		},
	}, {
		Name: "ok/update existing and insert missing",

		SrcDocs: []any{
			srcDoc{ID: "1", Name: "one", Value: 1},
			srcDoc{ID: "2", Name: "two", Value: 2},
		},
		DstDocs: []any{
			dstDoc{ID: "1", Name: "stale", Doubled: 0},
			dstDoc{ID: "3", Name: "untouched", Doubled: 6},
		},

		Expected: []dstDoc{
			{ID: "1", Name: "one", Doubled: 2},
			{ID: "2", Name: "two", Doubled: 4},
			{ID: "3", Name: "untouched", Doubled: 6},
		},
	}, {
		Name: "ok/multiple source documents merge into one",

		SrcDocs: []any{
			srcDoc{ID: "1", Name: "one", Value: 1},
			srcDoc{ID: "2", Name: "two", Value: 2},
		},
		DstQuery: func(srcDoc) any {
			return bson.D{{Key: "_id", Value: "merged"}}
		},
		DstUpdate: func(doc srcDoc) any {
			return bson.D{{Key: "$inc", Value: bson.D{
				{Key: "doubled", Value: doc.Value * 2},
			}}}
		},

		Expected: []dstDoc{
			{ID: "merged", Doubled: 6},
		},
	}, {
		Name: "ok/empty source",

		DstDocs: []any{
			dstDoc{ID: "1", Name: "one", Doubled: 2},
		},

		Expected: []dstDoc{
			{ID: "1", Name: "one", Doubled: 2},
		},
	}, {
		Name: "error/find",

		CTX: canceledContext(),
		SrcDocs: []any{
			srcDoc{ID: "1", Name: "one", Value: 1},
		},

		Expected: []dstDoc{},
		Error:    "failed to copy documents",
	}, {
		Name: "error/decode",

		SrcDocs: []any{
			bson.D{
				{Key: "_id", Value: "1"},
				{Key: "name", Value: "one"},
				{Key: "value", Value: "not a number"},
			},
		},

		Expected: []dstDoc{},
		Error:    "failed to decode document",
	}, {
		Name: "error/invalid update document",

		SrcDocs: []any{
			srcDoc{ID: "1", Name: "one", Value: 1},
		},
		DstUpdate: func(doc srcDoc) any {
			// Update documents must only contain update operators
			return bson.D{{Key: "name", Value: doc.Name}}
		},

		Expected: []dstDoc{},
		Error:    "failed to copy document",
	}}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			dst, src := setupCollections(t, tc.SrcDocs, tc.DstDocs)
			ctx := tc.CTX
			if ctx == nil {
				ctx = db.CTX()
			}
			query := dstQuery
			update := dstUpdate
			if tc.DstQuery != nil {
				query = tc.DstQuery
			}
			if tc.DstUpdate != nil {
				update = tc.DstUpdate
			}

			err := CopyUpsertCollection(ctx, dst, src, query, update)
			if tc.Error != "" {
				assert.ErrorContains(t, err, tc.Error)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.Expected, findAll[dstDoc](t, dst))
		})
	}
}
