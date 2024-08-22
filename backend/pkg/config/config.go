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
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (
	Config = viper.New()
)

func FromConfigFile(filePath string,
	defaults []Default,
	configValidators ...Validator) error {

	// Set default values for config
	SetDefaults(Config, defaults)

	// Find and read the config file
	if filePath != "" {
		Config.SetConfigFile(filePath)
		if err := Config.ReadInConfig(); err != nil {
			return errors.Wrap(err, "failed to read configuration")
		}
	}

	// Validate config
	if err := ValidateConfig(Config, configValidators...); err != nil {
		return errors.Wrap(err, "failed to validate configuration")
	}

	return nil
}

type Reader interface {
	Get(key string) interface{}
	GetBool(key string) bool
	GetFloat64(key string) float64
	GetInt(key string) int
	GetString(key string) string
	GetStringMap(key string) map[string]interface{}
	GetStringMapString(key string) map[string]string
	GetStringSlice(key string) []string
	GetTime(key string) time.Time
	GetDuration(key string) time.Duration
	IsSet(key string) bool
}

type Writer interface {
	SetDefault(key string, val interface{})
	Set(key string, val interface{})
}

type Handler interface {
	Reader
	Writer
}

type Default struct {
	Key   string
	Value interface{}
}

type Validator func(c Reader) error

// ValidateConfig validates conifg accroding to provided validators.
func ValidateConfig(c Reader, validators ...Validator) error {

	for _, validator := range validators {
		err := validator(c)
		if err != nil {
			return err
		}
	}

	return nil
}

func SetDefaults(c Writer, defaults []Default) {
	for _, def := range defaults {
		c.SetDefault(def.Key, def.Value)
	}
}
