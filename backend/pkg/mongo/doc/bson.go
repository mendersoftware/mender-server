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
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

// MarshallBSONOrDocumentFromStruct marshals a structure to BSON if it implements
// the bson.Marshaler interface, otherwise invokes DocumentFromStruct on it.
func MarshallBSONOrDocumentFromStruct(
	sct interface{},
	appendElements ...bson.E,
) (doc bson.D) {
	if marshaller, ok := sct.(bson.Marshaler); ok {
		var doc bson.D
		if data, err := marshaller.MarshalBSON(); err == nil {
			_ = bson.Unmarshal(data, &doc)
		}
		return doc
	}
	return DocumentFromStruct(sct, appendElements...)
}

// DocumentFromStruct creates a bson document from a struct in the order of the
// underlying data structure. Additional fields can be appended to the struct
// with the appendElements, these fields will be added at the end of the
// document.
func DocumentFromStruct(
	sct interface{},
	appendElements ...bson.E,
) (doc bson.D) {
	s := reflect.ValueOf(sct)
	defer func() {
		if r := recover(); r != nil {
			doc = nil
		}
	}()

	s = dereferenceValue(s)
	if s.Kind() != reflect.Struct {
		return nil
	}

	numAppends := len(appendElements)
	numFields := s.NumField()
	doc = make(bson.D, 0, numFields+numAppends)
	fields := s.Type()
	for i := 0; i < numFields; i++ {
		field := fields.Field(i)
		value := s.Field(i)
		key, valFace, set := valueFromStructField(field, value)
		if key == "inline" {
			doc = append(doc, DocumentFromStruct(valFace)...)
			continue
		}
		if set {
			doc = append(doc, bson.E{Key: key, Value: valFace})
		}
	}
	for i := 0; i < numAppends; i++ {
		doc = append(doc, appendElements[i])
	}
	return doc
}

func dereferenceValue(val reflect.Value) reflect.Value {
	const maxDereference = 4
	for i := 0; i < maxDereference; i++ {
		switch val.Kind() {
		case reflect.Ptr:
			val = val.Elem()
		case reflect.Interface:
			val = val.Elem()
		}
	}
	return val
}

type FlattenOptions struct {
	// Transform provides an option for transforming key/value pairs in
	// the flattened array. The input contains the values that would
	// otherwise be added to the document. This can be useful for
	// transforming query containing arrays to add an $in operator.
	Transform func(key string, elem interface{}) (string, interface{})
}

func NewFlattenOptions() *FlattenOptions {
	return &FlattenOptions{}
}

func (opts *FlattenOptions) SetTransform(
	transform func(key string, elem interface{}) (string, interface{}),
) *FlattenOptions {
	opts.Transform = transform
	return opts
}

func mergeFlattenOptions(opts []*FlattenOptions) *FlattenOptions {
	var ret = &FlattenOptions{}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if opt.Transform != nil {
			ret.Transform = opt.Transform
		}
	}
	return ret
}

// FlattenDocument consumes a struct or map and returns a flattened BSON
// document (bson.D) where embedded structs and maps are flattened into the
// root level of the map and field-names are concatenated by dots ('.').
// This function respects the "omitempty" bson tags such that it's possible
// to create queries from structs; the resulting document respects the struct
// declaration order with a depth first expansion of embedded fields. Please
// note, however, that there are no ordering guarantees on map-types.
//
//	type Foo struct {
//	    Bar: map[string]string{
//	        "baz": "foo"
//	    },  `bson:"bar"`
//	}
//
// Becomes:
//
//	bson.D{
//	  {Key: "bar.baz", Value: "foo"}
//	}
func FlattenDocument(
	mapping interface{}, options ...*FlattenOptions,
) (doc bson.D, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = v
			case string:
				err = errors.New(v)

			default:
				panic(v)
			}
		}
	}()
	opts := mergeFlattenOptions(options)

	s := reflect.ValueOf(mapping)
	s = dereferenceValue(s)

	switch s.Kind() {
	case reflect.Struct:
		return flattenStruct(s, "", opts), nil
	case reflect.Map:
		return flattenMap(s, "", opts), nil
	}
	return nil, errors.Errorf(
		"[programming error] invalid argument type %s, "+
			"expected struct or map-like type",
		s.Kind(),
	)
}

func valueFromStructField(
	key reflect.StructField,
	value reflect.Value,
) (string, interface{}, bool) {
	if rune(key.Name[0]) >= 'a' && rune(key.Name[0]) <= 'z' {
		// unexported field
		return "", nil, false
	}
	tag := key.Tag.Get("bson")
	if tag == "" {
		tag = strings.ToLower(key.Name)
	}
	tags := strings.Split(tag, ",")
	name := tags[0]
	if name == "" {
		name = strings.ToLower(key.Name)
	}
	for _, t := range tags {
		if t == "omitempty" && value.IsZero() {
			return "", nil, false
		}
	}
	return name, value.Interface(), true
}

func flattenStruct(
	sct reflect.Value,
	prefix string,
	options *FlattenOptions,
) (doc bson.D) {
	doc = bson.D{}
	sType := sct.Type()
	numSFields := sct.NumField()
	for i := 0; i < numSFields; i++ {
		sVal := sct.Field(i)
		sKey := sType.Field(i)

		sVal = dereferenceValue(sVal)
		fieldName, val, set := valueFromStructField(sKey, sVal)
		if !set {
			continue
		}
		if len(prefix) > 0 {
			fieldName = prefix + "." + fieldName
		}
		switch sVal.Kind() {
		case reflect.Struct:
			ret := flattenStruct(sVal, fieldName, options)
			if ret != nil {
				doc = append(doc, ret...)
			}
		case reflect.Map:
			ret := flattenMap(sVal, fieldName, options)
			if ret != nil {
				doc = append(doc, ret...)
			}
		default:
			if options.Transform != nil {
				fieldName, val = options.Transform(fieldName, val)
			}
			doc = append(doc, bson.E{
				Key:   fieldName,
				Value: val,
			})
		}
	}
	return doc
}

func flattenMap(
	m reflect.Value, prefix string, options *FlattenOptions,
) (doc bson.D) {
	rKeys := m.MapKeys()
	for _, rKey := range rKeys {
		// NOTE: Will panic if map keys are not string!
		key := rKey.String()
		fieldName := key
		if prefix != "" {
			fieldName = prefix + "." + fieldName
		}
		rVal := m.MapIndex(rKey)
		rVal = dereferenceValue(rVal)
		switch rVal.Kind() {
		case reflect.Struct:
			ret := flattenStruct(rVal, fieldName, options)
			if ret != nil {
				doc = append(ret, doc...)
			}
		case reflect.Map:
			ret := flattenMap(rVal, fieldName, options)
			if ret != nil {
				doc = append(ret, doc...)
			}
		default:
			val := rVal.Interface()
			if options.Transform != nil {
				fieldName, val = options.Transform(fieldName, val)
			}
			doc = append(doc, bson.E{Key: fieldName, Value: val})
		}
	}
	return doc
}
