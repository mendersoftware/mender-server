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

//go:build main
// +build main

package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"hash/crc64"
	"io"
	"os"
	"os/signal"
	"testing"

	"github.com/mendersoftware/mender-server/pkg/log"
)

func init() {
	// Make sure main does not exit before we have gathered coverage.
	log.Log.ExitFunc = func(int) {}
}

const (
	coverName = "coverage-acceptance"
	coverExt  = ".txt"
)

var stdout = os.Stdout

func TestMain(m *testing.M) {
	argHash := crc64.New(crc64.MakeTable(crc64.ECMA))
	for _, arg := range os.Args {
		_, _ = argHash.Write([]byte(arg))
	}
	var b [6]byte
	_, err := io.ReadFull(rand.Reader, b[:])
	if err != nil {
		panic(err)
	}
	// filename = "{coverName}@{hash(args)}-{48-bit rand}.txt"
	fileNameCover := fmt.Sprintf("%s@%s-%s%s",
		coverName,
		hex.EncodeToString(argHash.Sum(nil)),
		hex.EncodeToString(b[:]),
		coverExt,
	)

	// Override arguments passed to "testing" package
	os.Args = os.Args[:1]
	flag.Set("test.run", "TestRunMain")
	flag.Set("test.coverprofile", fileNameCover)

	// Run tests
	exitCode := m.Run()
	os.Stdout = stdout
	os.Exit(exitCode)
}

func TestRunMain(t *testing.T) {
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt)
	go func() {
		doMain(os.Args[:cap(os.Args)])
		stopChan <- os.Interrupt
	}()
	<-stopChan
	// Prevent the output from testing to hit stdout
	os.Stdout = os.Stderr
}
