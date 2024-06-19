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

//go:build acceptance
// +build acceptance

package main

import (
	"bufio"
	"encoding/hex"
	"flag"
	"fmt"
	"hash/crc64"
	"io"
	"os"
	"os/signal"
	"testing"
)

const (
	coverName = "coverage-acceptance"
	coverExt  = ".txt"
)

func TestMain(m *testing.M) {
	argHash := crc64.New(crc64.MakeTable(crc64.ECMA))
	for _, arg := range os.Args {
		_, _ = argHash.Write([]byte(arg))
	}
	fileNameCoverRun := fmt.Sprintf("%s@%s%s",
		coverName, hex.EncodeToString(argHash.Sum(nil)), coverExt)
	fileNameCoverAll := coverName + coverExt
	os.Args = os.Args[:1]
	flag.Set("test.run", "TestRunMain")
	flag.Set("test.coverprofile", fileNameCoverRun)
	exitCode := m.Run()
	if exitCode > 0 {
		os.Exit(exitCode)
	}
	// TODO: Should protect 'coverage-acceptance.txt' with a filelock
	err := aggregateCoverage(fileNameCoverRun, fileNameCoverAll)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func TestRunMain(t *testing.T) {
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)
	go func() {
		doMain(os.Args[:cap(os.Args)])
		signal.Stop(stopChan)
		close(stopChan)
	}()
	<-stopChan
}

func aggregateCoverage(src, dst string) (err error) {
	fdRun, err := os.OpenFile(src, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open file '%s': %s\n", src, err)
		os.Exit(1)
	}
	defer fdRun.Close()

	var rSrc io.Reader = fdRun

	fd, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if os.IsExist(err) {
		fd, err = os.OpenFile(dst, os.O_RDWR, 0644)
		if err != nil {
			return err
		}
		defer fd.Close()
		_, err := fd.Seek(0, io.SeekEnd)
		if err != nil {
			return err
		}
		rdr := bufio.NewReader(fdRun)
		// Skip first line
		_, err = rdr.ReadSlice('\n')
		if err != nil {
			return err
		}
		rSrc = rdr
	} else if err != nil {
		return err
	} else {
		// Coverage file created
		defer fd.Close()
	}
	_, err = io.Copy(fd, rSrc)
	return err
}
