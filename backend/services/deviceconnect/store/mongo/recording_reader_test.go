// Copyright 2021 Northern.tech AS
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
	"encoding/base64"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	mopts "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mendersoftware/mender-server/services/deviceconnect/model"
	"github.com/mendersoftware/mender-server/pkg/identity"
	mstore "github.com/mendersoftware/mender-server/pkg/store"
)

func TestReadRecording(t *testing.T) {
	testCases := []struct {
		Name string

		Ctx            context.Context
		SessionID      string
		RecordingData  []string
		ReadBufferSize int
		ExpectedString string

		Error error
	}{
		{
			Name: "ok",

			Ctx: identity.WithContext(
				context.Background(),
				&identity.Identity{
					Tenant: "000000000000000000000002",
				},
			),
			SessionID: "00000000-0000-0000-0000-000000000002",
			RecordingData: []string{
				"H4sIAKOjMmAAA1NOTc7IV8jP5gIiAIEuTnMMAAAA",
				"H4sIAFtvMmAAA1WPwWrDMAyGz8pTCHaeB4PtPlqPlTU7LOwcPFvNTG0r2G6gbz85hNH5IiOk7/90Fwo6WrwlyymRrQ/dnu2Z8skHAsDenKl9EfHbJ2n8GwZsb2IV2SnOfkKMxqexUqlqYmi13ACVsZbmapIlwE/9su+1ig7RBk9JaG6dvLccZy6krjGs9HKJkiQZFAAu1YfSHQ87/TFoECMz+9VDnE5+WiEF/sxkdzMrlBfKsFBynBtgfD0c9TDu3t6Hr16VH/P49IyCm29w6mo2CYnfqO3Edp0gK2fqfgGMUrrGRAEAAA==",
			},
			ReadBufferSize: 512,
			ExpectedString: "#echo ok\nok\n#ls deviceconnect/\nDockerfile\t\t Makefile   bin\t\t deviceconnect\t     go.mod.orig  main_test.go\ttests\nDockerfile.acceptance\t README.md  client\t docker-compose.yml  go.sum\t  model\t\tutils\nLICENSE\t\t\t api\t    config\t docs\t\t     go.sum.orig  server\tvendor\nLIC_FILES_CHKSUM.sha256  app\t    config.yaml  go.mod\t\t     main.go\t  store\n",
		},
		{
			Name: "ok short buffer",

			Ctx: identity.WithContext(
				context.Background(),
				&identity.Identity{
					Tenant: "000000000000000000000004",
				},
			),
			SessionID: "00000000-0000-0000-0000-000000000004",
			RecordingData: []string{
				"H4sIAKOjMmAAA1NOTc7IV8jP5gIiAIEuTnMMAAAA",
				"H4sIAFtvMmAAA1WPwWrDMAyGz8pTCHaeB4PtPlqPlTU7LOwcPFvNTG0r2G6gbz85hNH5IiOk7/90Fwo6WrwlyymRrQ/dnu2Z8skHAsDenKl9EfHbJ2n8GwZsb2IV2SnOfkKMxqexUqlqYmi13ACVsZbmapIlwE/9su+1ig7RBk9JaG6dvLccZy6krjGs9HKJkiQZFAAu1YfSHQ87/TFoECMz+9VDnE5+WiEF/sxkdzMrlBfKsFBynBtgfD0c9TDu3t6Hr16VH/P49IyCm29w6mo2CYnfqO3Edp0gK2fqfgGMUrrGRAEAAA==",
			},
			ReadBufferSize: 4,
			ExpectedString: "#echo ok\nok\n#ls deviceconnect/\nDockerfile\t\t Makefile   bin\t\t deviceconnect\t     go.mod.orig  main_test.go\ttests\nDockerfile.acceptance\t README.md  client\t docker-compose.yml  go.sum\t  model\t\tutils\nLICENSE\t\t\t api\t    config\t docs\t\t     go.sum.orig  server\tvendor\nLIC_FILES_CHKSUM.sha256  app\t    config.yaml  go.mod\t\t     main.go\t  store\n",
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			ds := &DataStoreMongo{client: db.Client()}
			defer ds.DropDatabase()

			database := db.Client().Database(mstore.DbNameForTenant(
				"000000000000000000000000", DbName,
			))
			collSess := database.Collection(RecordingsCollectionName)

			for i, _ := range tc.RecordingData {
				d, err := base64.StdEncoding.DecodeString(tc.RecordingData[i])
				assert.NoError(t, err)

				_, err = collSess.InsertOne(nil, &model.Recording{
					ID:        uuid.New(),
					SessionID: tc.SessionID,
					Recording: d,
					CreatedTs: time.Now().UTC(),
					ExpireTs:  time.Now().UTC(),
				})
				assert.NoError(t, err)
			}
			findOptions := mopts.Find()
			sortField := bson.M{
				"created_ts": 1,
			}
			findOptions.SetSort(sortField)
			c, err := collSess.Find(context.Background(),
				bson.M{
					dbFieldSessionID: tc.SessionID,
				},
				findOptions,
			)
			assert.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
			defer cancel()
			reader := NewRecordingReader(ctx, c)
			assert.NotNil(t, reader)
			buffer := make([]byte, tc.ReadBufferSize)
			s := ""
			for err != io.EOF {
				assert.NoError(t, err)
				n, err := reader.Read(buffer[:])
				if err == io.EOF {
					break
				}
				s += string(buffer[:n])
				//t.Logf("read: %s/%d", string(buffer[:n]), n)
			}
			//t.Logf("read(final): %s/%d", s, len(s))
			assert.Equal(t, tc.ExpectedString, s)
		})
	}
}
