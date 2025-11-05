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
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/pkg/errors"
)

type Duration time.Duration

func (duration *Duration) UnmarshalText(b []byte) error {
	d, err := time.ParseDuration(string(b))
	if err != nil {
		return err
	}
	*duration = Duration(d)
	return nil
}

func mapstructureDecode[T any](value any, target *[]T) error {
	decoder, err := mapstructure.NewDecoder(
		&mapstructure.DecoderConfig{
			TagName: "json",
			Result:  target,
			Squash:  true,
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				mapstructure.TextUnmarshallerHookFunc(),
				mapstructure.StringToTimeDurationHookFunc(),
			),
		},
	)
	if err != nil {
		return err
	}
	err = decoder.Decode(value)
	return err
}

// UnmarshalSliceSetting will unmarshal an array of objects into the result T
// using either newline separated json objects for strings (from env) or
// an array of objects (using json struct tag) for any configuration source
// that is able to express a slice of objects.
func UnmarshalSliceSetting[T any](c Reader, path string, result *[]T) error {
	if result == nil {
		return errors.New("cannot unmarshal setting to nil result")
	}
	value := c.Get(path)
	var err error
	switch cfg := value.(type) {
	case string:
		decoder := json.NewDecoder(strings.NewReader(cfg))
		for {
			var elem T
			err = decoder.Decode(&elem)
			if err != nil {
				break
			}
			*result = append(*result, elem)
		}
		if errors.Is(err, io.EOF) {
			err = nil
		}
	case []any:
		err = mapstructureDecode(cfg, result)
	case nil:
		// pass (empty config)
	default:
		// Try to handle slice of structs.
		typ := reflect.TypeOf(value)
		if typ.Kind() == reflect.Slice && typ.Elem().Kind() == reflect.Struct {
			err = mapstructureDecode(cfg, result)
		} else {
			err = fmt.Errorf("invalid config type %T", cfg)
		}
	}
	return err
}
