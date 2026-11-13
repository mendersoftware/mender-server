package mongo

import (
	"fmt"
	"maps"
	"reflect"
	"slices"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

type Depend struct {
	Key   string `bson:"k"`
	Value string `bson:"v"`
}

type Provide = Depend

var (
	tDepends       = reflect.TypeOf([]Depend{})
	tProvides      = tDepends
	tMap           = reflect.TypeOf(map[string]any(nil))
	tMapString     = reflect.TypeOf(map[string]string(nil))
	tModelDepends  = reflect.TypeOf(model.Depends{})
	tModelProvides = reflect.TypeOf(model.Provides(nil))
	tReleaseKind   = reflect.TypeOf(model.ReleaseKind(""))
	tString        = reflect.TypeOf("")
)

func newRegistry() *bson.Registry {
	reg := bson.NewMgoRegistry()
	reg.RegisterTypeEncoder(tReleaseKind, bson.ValueEncoderFunc(
		func(ec bson.EncodeContext, vw bson.ValueWriter, v reflect.Value) error {
			enc, err := ec.LookupEncoder(tString)
			if err != nil {
				return err
			}
			if v.Len() == 0 {
				v = reflect.ValueOf(model.ReleaseKindRelease)
			}
			return enc.EncodeValue(ec, vw, v)
		},
	))
	reg.RegisterTypeEncoder(tModelProvides, bson.ValueEncoderFunc(
		func(ec bson.EncodeContext, vw bson.ValueWriter, v reflect.Value) error {
			provides := v.Interface().(model.Provides)
			doc := make([]Provide, 0, len(provides))
			for _, key := range slices.Sorted(maps.Keys(provides)) {
				doc = append(doc, Provide{Key: key, Value: provides[key]})
			}
			rVal := reflect.ValueOf(doc)
			enc, err := reg.LookupEncoder(tProvides)
			if err != nil {
				return err
			}
			return enc.EncodeValue(ec, vw, rVal)
		},
	))
	reg.RegisterTypeDecoder(tModelProvides, bson.ValueDecoderFunc(
		func(dc bson.DecodeContext, vr bson.ValueReader, v reflect.Value) error {
			switch typ := vr.Type(); typ {
			case bson.TypeNull:
				v.Set(reflect.ValueOf(model.Provides{}))
				return vr.ReadNull()
			case bson.TypeArray:
				var provides model.Provides
				dec, err := reg.LookupDecoder(tProvides)
				if err != nil {
					return err
				}
				var res []Provide
				err = dec.DecodeValue(dc, vr, reflect.ValueOf(&res).Elem())
				if err != nil {
					return err
				}
				provides = make(model.Provides, len(res))
				for _, elem := range res {
					provides[elem.Key] = elem.Value
				}
				v.Set(reflect.ValueOf(provides))
				return nil

			case bson.TypeEmbeddedDocument:
				// Backward compatibility with old schema
				dec, err := reg.LookupDecoder(tMapString)
				if err != nil {
					return err
				}
				return dec.DecodeValue(dc, vr, v)
			default:
				return fmt.Errorf("invalid data type %s for DependsDecode", typ)
			}
		},
	))
	reg.RegisterTypeEncoder(tModelDepends, bson.ValueEncoderFunc(
		func(ec bson.EncodeContext, vw bson.ValueWriter, v reflect.Value) error {
			deps := v.Interface().(model.Depends)
			depends := make([]Depend, 0, len(deps))
			// Sort keys to get a deterministic encoding
			for _, key := range slices.Sorted(maps.Keys(deps)) {
				for _, value := range slices.Sorted(slices.Values(deps[key])) {
					depends = append(depends, Depend{Key: key, Value: value})
				}
			}
			rVal := reflect.ValueOf(depends)
			enc, err := reg.LookupEncoder(tDepends)
			if err != nil {
				return err
			}
			return enc.EncodeValue(ec, vw, rVal)
		},
	))
	reg.RegisterTypeDecoder(tModelDepends, bson.ValueDecoderFunc(
		func(dc bson.DecodeContext, vr bson.ValueReader, v reflect.Value) error {
			var depends model.Depends
			switch typ := vr.Type(); typ {
			case bson.TypeArray:
				dec, err := reg.LookupDecoder(tDepends)
				if err != nil {
					return err
				}
				var res []Depend
				err = dec.DecodeValue(dc, vr, reflect.ValueOf(&res).Elem())
				if err != nil {
					return err
				}
				depends = make(model.Depends)
				for _, elem := range res {
					depends.Add(elem.Key, elem.Value)
				}

			case bson.TypeEmbeddedDocument:
				// Backward compatibility with old schema
				dec, err := reg.LookupDecoder(tMap)
				if err != nil {
					return err
				}
				var legacyDepends map[string]any
				err = dec.DecodeValue(dc, vr, reflect.ValueOf(&legacyDepends).Elem())
				if err == nil {
					depends, err = model.DependsFromMenderArtifact(legacyDepends)
				}
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("invalid data type %s for DependsDecode", typ)
			}
			v.Set(reflect.ValueOf(depends))
			return nil
		},
	))
	return reg
}
