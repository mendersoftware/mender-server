package mongo

import (
	"fmt"
	"reflect"
	"slices"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/mendersoftware/mender-server/services/deployments/model"
)

type Depend struct {
	Key   string `bson:"k"`
	Value string `bson:"v"`
}

var tModelDepends = reflect.TypeOf(model.Depends{})
var tMap = reflect.TypeOf(map[string]any{})
var tDepends = reflect.TypeOf([]Depend{})

func newRegistry() *bson.Registry {
	reg := bson.NewMgoRegistry()
	reg.RegisterTypeEncoder(tModelDepends, bson.ValueEncoderFunc(
		func(ec bson.EncodeContext, vw bson.ValueWriter, v reflect.Value) error {
			deps := v.Interface().(model.Depends)
			rVal := reflect.ValueOf(slices.Collect(func(yield func(Depend) bool) {
				for key, values := range deps {
					for _, value := range values {
						if !yield(Depend{
							Key:   key,
							Value: value,
						}) {
							return
						}
					}
				}
			}),
			)
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
