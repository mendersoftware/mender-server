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
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/mendersoftware/mender-server/pkg/config"

	dconfig "github.com/mendersoftware/mender-server/services/workflows/config"
)

var (
	testDataStore *DataStoreMongo
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Short() {
		setupDatabase()
		defer testDataStore.Close()
	}
	result := m.Run()
	if !testing.Short() {
		teardownDatabase()
	}
	os.Exit(result)
}

func setupDatabase() {
	testingDbName := dconfig.SettingDbNameDefault + "_tests"

	config.SetDefaults(config.Config, dconfig.Defaults)
	config.Config.Set(dconfig.SettingDbName, testingDbName)

	config.Config.SetEnvPrefix("WORKFLOWS")
	config.Config.AutomaticEnv()
	config.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	dataStore, err := SetupDataStore(true)
	if err != nil {
		fmt.Println(fmt.Sprintf("error connecting to database: %s", err))
		os.Exit(1)
	}

	testDataStore = dataStore
}

func (db *DataStoreMongo) dropDatabase() error {
	ctx := context.Background()
	err := db.client.Database(db.dbName).Drop(ctx)
	return err
}

func teardownDatabase() {
	testDataStore.dropDatabase()
}
