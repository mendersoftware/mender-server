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
package config

import (
	"errors"
	"testing"
	"time"
)

type MockConfigReader struct{}

func (m *MockConfigReader) Get(key string) interface{}                      { return nil }
func (m *MockConfigReader) GetBool(key string) bool                         { return true }
func (m *MockConfigReader) GetFloat64(key string) float64                   { return 1.1 }
func (m *MockConfigReader) GetInt(key string) int                           { return 1 }
func (m *MockConfigReader) GetString(key string) string                     { return "some string" }
func (m *MockConfigReader) GetStringMap(key string) map[string]interface{}  { return nil }
func (m *MockConfigReader) GetStringMapString(key string) map[string]string { return nil }
func (m *MockConfigReader) GetStringSlice(key string) []string              { return []string{} }
func (m *MockConfigReader) GetTime(key string) time.Time                    { return time.Now() }
func (m *MockConfigReader) GetDuration(key string) time.Duration            { return time.Second }
func (m *MockConfigReader) IsSet(key string) bool                           { return true }

type MockConfigWriter struct {
	vals map[string]interface{}
}

func (m *MockConfigWriter) SetDefault(key string, val interface{}) {
	m.vals[key] = val
}

func (m *MockConfigWriter) Set(key string, val interface{}) {
	m.vals[key] = val
}

func NewMockWriter() *MockConfigWriter {
	return &MockConfigWriter{
		make(map[string]interface{}),
	}
}

func TestValidateConfig(t *testing.T) {

	err := errors.New("test error")

	testList := []struct {
		out        error
		c          Reader
		validators []Validator
	}{
		{nil, &MockConfigReader{}, []Validator{}},
		{err, &MockConfigReader{}, []Validator{func(c Reader) error { return err }}},
	}

	for _, test := range testList {
		if ValidateConfig(test.c, test.validators...) != test.out {
			t.FailNow()
		}
	}
}

func TestSetDefaultConfigs(t *testing.T) {
	defaults := []Default{
		{"foo", "bar"},
		{"baz", 1},
	}

	c := NewMockWriter()

	SetDefaults(c, defaults)

	val_foo, ok := c.vals["foo"]
	if ok != true || val_foo != "bar" {
		t.FailNow()
	}

	val_baz, ok := c.vals["baz"]
	if ok != true || val_baz != 1 {
		t.FailNow()
	}
}

func TestFromConfigFile(t *testing.T) {
	if err := FromConfigFile("testdata/config-empty.yaml", []Default{}); err != nil {
		t.Fatal(err)
	}

	err := FromConfigFile("", nil)
	if err != nil {
		t.Fatal(err)
	}

	err = FromConfigFile("non-existing-file.yaml", nil)
	if err == nil {
		t.FailNow()
	}
}
