// Copyright 2025 Northern.tech AS
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

package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

type testConfig struct {
	Foo string  `json:"foo"`
	Bar float64 `json:"bar"`
}

func TestUnmarshalSliceSetting(t *testing.T) {
	configData := []testConfig{
		{Foo: "first_foo", Bar: 9.1},
		{Foo: "second_foo", Bar: 8.2},
		{Foo: "third_foo", Bar: 7.3},
		{Foo: "fourth_foo", Bar: 6.4},
		{Foo: "fifth_foo", Bar: 5.5},
		{Foo: "sixth_foo", Bar: 4.6},
		{Foo: "seventh_foo", Bar: 3.7},
		{Foo: "eighth_foo", Bar: 2.8},
		{Foo: "ninth_foo", Bar: 1.9},
		{Foo: "tenth_foo", Bar: 0.0},
	}
	const configFileData = `
---
test:
- foo: first_foo
  bar: 9.1
- foo: second_foo
  bar: 8.2
- foo: third_foo
  bar: 7.3
- foo: fourth_foo
  bar: 6.4
- foo: fifth_foo
  bar: 5.5
- foo: sixth_foo
  bar: 4.6
- foo: seventh_foo
  bar: 3.7
- foo: eighth_foo
  bar: 2.8
- foo: ninth_foo
  bar: 1.9
- foo: tenth_foo
  bar: 0.0
`
	const configEnvDataProducts = `
{"foo": "first_foo",   "bar": 9.1}
{"foo": "second_foo",  "bar": 8.2}
{"foo": "third_foo",   "bar": 7.3}
{"foo": "fourth_foo",  "bar": 6.4}
{"foo": "fifth_foo",   "bar": 5.5}
{"foo": "sixth_foo",   "bar": 4.6}
{"foo": "seventh_foo", "bar": 3.7}
{"foo": "eighth_foo",  "bar": 2.8}
{"foo": "ninth_foo",   "bar": 1.9}
{"foo": "tenth_foo",   "bar": 0.0}
`
	evalAndCompare := func(cfg *viper.Viper) {
		var result []testConfig
		err := UnmarshalSliceSetting(cfg, "test", &result)
		if err != nil {
			t.Errorf("error loading offers: %s", err.Error())
			return
		}
		for i, item := range configData {
			if i >= len(result) {
				t.Errorf("missing item[%d]: %v", i, item)
				continue
			} else if !reflect.DeepEqual(item, result[i]) {
				t.Errorf("item[%d] does not match expectations: %v (actual) != %v (expected)",
					i, result[i], item)
			}
		}
	}

	t.Run("load from file", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "config1.yaml")
		err := os.WriteFile(configFile, []byte(configFileData), 0644)
		if err != nil {
			t.Error(err)
			return
		}
		cfg := viper.New()
		cfg.SetConfigFile(configFile)
		err = cfg.ReadInConfig()
		if err != nil {
			t.Errorf("unexpected error loading config file: %s", err)
			return
		}
		evalAndCompare(cfg)
	})
	t.Run("load from env", func(t *testing.T) {
		t.Setenv(
			"TEST",
			configEnvDataProducts,
		)
		cfg := viper.New()
		cfg.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		cfg.AutomaticEnv()
		evalAndCompare(cfg)
	})
}
