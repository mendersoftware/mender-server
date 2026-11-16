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

package mongo

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var updateGolden = flag.Bool("update-golden", false, "update golden files")

const (
	migration_2_0_0_input  = "testdata/migration_2_0_0/input"
	migration_2_0_0_golden = "testdata/migration_2_0_0/golden"
)

// loadFixtures restores a database snapshot from dir. Each subdirectory of dir
// is a database name and each <collection>.json file inside it contains one
// canonical Extended JSON document per line, as produced by:
//
//	mongoexport --jsonFormat=canonical --db <db> --collection <collection>
func loadFixtures(t *testing.T, ctx context.Context, client *mongo.Client, dir string) {
	t.Helper()
	dbDirs, err := os.ReadDir(dir)
	require.NoError(t, err)
	for _, dbDir := range dbDirs {
		if !dbDir.IsDir() {
			continue
		}
		files, err := filepath.Glob(filepath.Join(dir, dbDir.Name(), "*.json"))
		require.NoError(t, err)
		for _, file := range files {
			coll := strings.TrimSuffix(filepath.Base(file), ".json")
			docs := readExtJSONLines(t, file)
			if len(docs) == 0 {
				continue
			}
			_, err := client.Database(dbDir.Name()).
				Collection(coll).
				InsertMany(ctx, docs)
			require.NoErrorf(t, err, "failed to load fixture %s", file)
		}
	}
}

func readExtJSONLines(t *testing.T, file string) []any {
	t.Helper()
	f, err := os.Open(file)
	require.NoError(t, err)
	defer f.Close()

	var docs []any
	scanner := bufio.NewScanner(f)
	// Documents can be up to 16MiB
	scanner.Buffer(nil, 16*1024*1024+1)
	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var doc bson.D
		err := bson.UnmarshalExtJSON(line, true, &doc)
		require.NoErrorf(t, err, "%s:%d: invalid canonical Extended JSON", file, lineNo)
		docs = append(docs, doc)
	}
	require.NoError(t, scanner.Err())
	return docs
}

// dumpCollection returns a deterministic canonical Extended JSON
// representation of all documents in the collection. Generated ObjectIDs are
// replaced with a placeholder, embedded document keys are sorted (Go maps are
// encoded in random order), and documents are sorted by their serialized form
// so that the output is independent of insertion order.
func dumpCollection(t *testing.T, ctx context.Context, coll *mongo.Collection) string {
	t.Helper()
	cur, err := coll.Find(ctx, bson.D{})
	require.NoError(t, err)
	docs := allOrdered(t, ctx, cur)

	serialized := make([]string, 0, len(docs))
	for _, doc := range docs {
		for i := range doc {
			if _, ok := doc[i].Value.(bson.ObjectID); ok && doc[i].Key == "_id" {
				doc[i].Value = "<ObjectID>"
			} else {
				doc[i].Value = sortEmbeddedKeys(doc[i].Value)
			}
		}
		b, err := bson.MarshalExtJSONIndent(doc, true, false, "", "  ")
		require.NoError(t, err)
		serialized = append(serialized, string(b))
	}
	slices.Sort(serialized)
	return strings.Join(serialized, "\n") + "\n"
}

// sortEmbeddedKeys recursively sorts the keys of embedded documents in v.
// Array element order is preserved.
func sortEmbeddedKeys(v any) any {
	switch val := v.(type) {
	case bson.D:
		for i := range val {
			val[i].Value = sortEmbeddedKeys(val[i].Value)
		}
		slices.SortFunc(val, func(a, b bson.E) int {
			return strings.Compare(a.Key, b.Key)
		})
		return val
	case bson.A:
		for i := range val {
			val[i] = sortEmbeddedKeys(val[i])
		}
		return val
	default:
		return v
	}
}

// allOrdered decodes all documents from the cursor preserving field order of
// embedded documents (the client registry decodes them into maps).
func allOrdered(t *testing.T, ctx context.Context, cur *mongo.Cursor) []bson.D {
	t.Helper()
	var raws []bson.Raw
	require.NoError(t, cur.All(ctx, &raws))
	docs := make([]bson.D, len(raws))
	for i, raw := range raws {
		require.NoError(t, bson.Unmarshal(raw, &docs[i]))
	}
	return docs
}

func assertGolden(t *testing.T, file, actual string) {
	t.Helper()
	if *updateGolden {
		require.NoError(t, os.MkdirAll(filepath.Dir(file), 0o755))
		require.NoError(t, os.WriteFile(file, []byte(actual), 0o644))
		return
	}
	expected, err := os.ReadFile(file)
	require.NoErrorf(t, err, "missing golden file (run with -update to create it)")
	assert.Equal(t, string(expected), actual, "golden file %s mismatch", file)
}

func TestMigration_2_0_0(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping migration 2.0.0 test in short mode")
	}
	db.Wipe()
	ctx := context.Background()
	client := db.Client()

	loadFixtures(t, ctx, client, migration_2_0_0_input)

	ds := NewDataStoreMongoWithClient(client)
	err := ds.Migrate(ctx, DbVersion, true)
	require.NoError(t, err)

	for _, coll := range []string{
		CollectionDeployments,
		CollectionDevices,
		CollectionReleases,
		CollectionSettings,
		CollectionUploadIntents,
	} {
		t.Run(coll, func(t *testing.T) {
			actual := dumpCollection(t, ctx, client.Database(DefaultDbName).Collection(coll))
			assertGolden(t, filepath.Join(migration_2_0_0_golden, coll+".json"), actual)
		})
	}

	t.Run("indexes", func(t *testing.T) {
		var specs []string
		for _, coll := range []string{
			CollectionDeployments,
			CollectionDevices,
			CollectionReleases,
			CollectionUploadIntents,
		} {
			cur, err := client.Database(DefaultDbName).
				Collection(coll).
				Indexes().
				List(ctx)
			require.NoError(t, err)
			for _, idx := range allOrdered(t, ctx, cur) {
				// Drop server-dependent fields
				idx = slices.DeleteFunc(idx, func(e bson.E) bool { return e.Key == "v" })
				b, err := bson.MarshalExtJSON(idx, false, false)
				require.NoError(t, err)
				specs = append(specs, coll+": "+string(b))
			}
		}
		slices.Sort(specs)
		assertGolden(t,
			filepath.Join(migration_2_0_0_golden, "indexes.txt"),
			strings.Join(specs, "\n")+"\n")
	})
}
