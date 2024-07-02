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
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	mopts "go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
)

// TestDB is a stripped down TestDBRunner interface.
type TestDB interface {
	NewClient(ctx context.Context) *mongo.Client
	URL() string
}

var (
	db     TestDB
	client *mongo.Client
)

func TestMain(m *testing.M) {
	status := func() int {
		name, err := ioutil.TempDir("", "mongod-test")
		if err != nil {
			panic(err)
		}
		instance := NewMongoTestInstance(name)
		db = instance
		defer instance.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		client = db.NewClient(ctx)
		ret := m.Run()
		return ret
	}()
	os.Exit(status)
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

type MongoTestInstance struct {
	Process  *exec.Cmd
	HostAddr string
	DbPath   string
	ShutDown chan struct{}
}

func NewMongoTestInstance(path string) *MongoTestInstance {
	db := new(MongoTestInstance)
	db.ShutDown = make(chan struct{})
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic("unable to listen on a local address: " + err.Error())
	}
	addr := l.Addr().(*net.TCPAddr)
	l.Close()
	db.HostAddr = addr.String()

	args := []string{
		"--dbpath", path,
		"--bind_ip", "127.0.0.1",
		"--port", strconv.Itoa(addr.Port),
		"--nojournal",
	}
	var stdout, stderr bytes.Buffer
	db.Process = exec.Command("mongod", args...)
	db.Process.Stdout = &stdout
	db.Process.Stderr = &stderr
	err = db.Process.Start()
	if err != nil {
		// print error to facilitate troubleshooting as the panic will be caught in a panic handler
		fmt.Fprintf(os.Stderr, "mongod failed to start: %v\n", err)
		panic(err)
	}
	go func() {
		err := db.Process.Wait()
		select {
		case <-db.ShutDown:
		default:
			if err != nil {
				fmt.Fprintf(os.Stderr, "!!! mongod process died uenxpectedly:\n")
				fmt.Fprintf(os.Stderr, "!!! stdout:\n%s\n", stdout.String())
				fmt.Fprintf(os.Stderr, "!!! stderr:\n%s\n", stderr.String())
				panic(err)
			} else {
				fmt.Fprintf(os.Stderr, "!!! mongod process died uenxpectedly:\n")
				fmt.Fprintf(os.Stderr, "!!! stdout:\n%s\n", stdout.String())
				fmt.Fprintf(os.Stderr, "!!! stderr:\n%s\n", stderr.String())
				panic("mongod process died unexpectedly")
			}
		}
	}()
	return db
}

func (db *MongoTestInstance) Stop() {
	close(db.ShutDown)
	db.Process.Process.Signal(os.Interrupt)
	for i := 0; i < 10; i++ {
		if db.Process.ProcessState != nil {
			return
		}
		time.Sleep(time.Millisecond * 500)
	}
	db.Process.Process.Kill()
}

func (db *MongoTestInstance) NewClient(ctx context.Context) *mongo.Client {
	client, err := mongo.Connect(ctx, mopts.Client().ApplyURI(db.URL()))
	if err != nil {
		panic(err)
	}
	return client
}

func (db *MongoTestInstance) URL() string {
	return "mongodb://" + db.HostAddr
}
