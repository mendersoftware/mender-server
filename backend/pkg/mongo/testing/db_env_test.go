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
	"os"
	"testing"
	"time"
)

func TestRunnerFromEnv(t *testing.T) {
	if _, ok := os.LookupEnv("TEST_MONGO_URL"); !ok {
		t.Skip("Test requires TEST_MONGO_URL to be set")
	}

	_ = WithDB(func(runner TestDBRunner) int {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		err := runner.Client().
			Ping(ctx, nil)
		if err != nil {
			t.Errorf("failed to ping MongoDB server: %s", err)
			t.Fail()
		} else {
			t.Log("YAY!")
		}
		return 0
	})
}
