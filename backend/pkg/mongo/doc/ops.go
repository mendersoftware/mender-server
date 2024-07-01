// Copyright 2023 Northern.tech AS
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

package doc

import (
	"fmt"
	"reflect"
	"sort"

	"go.mongodb.org/mongo-driver/bson"
)

const (
	// PermutationThreshold limits the complexity of Unwound maps to the
	// given array length.
	PermutationThreshold = 1024
)

var (
	// ErrPermutations is returned if the map contains arrays with a combined
	// number of permutations than the threshold.
	ErrPermutations = fmt.Errorf(
		"object complexity (array permutations) exceeded threshold")
)

// item is an internally used struct to organize the output in UnwindMap
type item struct {
	Key   string
	Value reflect.Value
}

// UnwindMap takes a string map and perform mongodb $unwind aggregation operation
// on all array entries.
// Don't return a map - but a sorted list of tuples; mongo indices are sensitive
// to order.
// Example:
//
//	{"foo": ["1", "2"], "bar": "3"} becomes
//	[ {{"foo", "1", "bar", "3"}}, {{"foo": "2", "bar": "3"}} ]
func UnwindMap(
	in interface{},
) ([]bson.D, error) {
	// returned bloated map
	var unwoundMap []bson.D
	// permutations and cumulative sum of permutations
	var permutations int64 = 1

	inVal := reflect.ValueOf(in)
	if inVal.Kind() != reflect.Map {
		return nil, fmt.Errorf(
			"invalid argument type: %s != map[string]interface{}",
			inVal.Kind().String())
	}
	mapLen := inVal.Len()
	orderedMap := make([]item, mapLen)

	// Compute number of permutations and initialize ordered map.
	var i int
	keys := inVal.MapKeys()
	for _, key := range keys {
		var keyStr string
		var ok bool
		value := inVal.MapIndex(key)
		if key.Kind() == reflect.Interface {
			key = key.Elem()
		}
		if value.Kind() == reflect.Interface {
			value = value.Elem()
		}
		if keyStr, ok = key.Interface().(string); !ok {
			return nil, fmt.Errorf("invalid argument type: "+
				"%s != map[string]interface{}", inVal.Kind())
		}

		switch value.Kind() {
		case reflect.Slice, reflect.Array:
			orderedMap[i] = item{
				Key: keyStr,
			}
			prevPermutations := permutations
			permutations = permutations * int64(value.Len())

			// Check with constraint or overflow
			if permutations > PermutationThreshold ||
				permutations < prevPermutations {
				return nil, ErrPermutations
			}

		case reflect.String:
			orderedMap[i] = item{
				Key: keyStr,
			}
		default:
			return nil, fmt.Errorf(
				"cannot unwind entry %s of type: %s",
				key, value.Kind().String())
		}
		orderedMap[i].Value = value
		i++
	}

	sort.Slice(orderedMap, func(i, j int) bool {
		return orderedMap[i].Key < orderedMap[j].Key
	})

	// Allocate returned map array
	unwoundMap = make([]bson.D, permutations)

	// Fill in the map entries
	for k := int64(0); k < permutations; k++ {
		var tmpPerm int64 = 1
		unwoundMap[k] = make(bson.D, mapLen)
		for i, item := range orderedMap {
			unwoundMap[k][i].Key = item.Key
			itemKind := item.Value.Kind()
			if itemKind == reflect.Slice ||
				itemKind == reflect.Array {

				// Compute index
				// - think of it as counting using array
				//   length as base
				itemLen := item.Value.Len()
				idx := (k / tmpPerm) % int64(itemLen)

				unwoundMap[k][i].Value = item.Value.
					Index(int(idx)).Interface()

				// Update cumulative permutations
				tmpPerm *= int64(itemLen)
			} else {
				unwoundMap[k][i].Value = item.Value.Interface()
			}
		}
	}
	return unwoundMap, nil
}
