package mongo

import (
	"reflect"
	"strings"

	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"

	"github.com/mendersoftware/mender-server/pkg/mongo/codec"
	"github.com/mendersoftware/mender-server/services/useradm/model"
)

func newRegistry() *bsoncodec.Registry {
	registry := codec.NewRegistry()
	registry.RegisterTypeEncoder(tEmail, bsoncodec.ValueEncoderFunc(encodeEmail))
	return registry
}

var tEmail = reflect.TypeOf(model.Email(""))

func encodeEmail(ec bsoncodec.EncodeContext, w bsonrw.ValueWriter, val reflect.Value) error {
	if !val.IsValid() || val.Type() != tEmail {
		return bsoncodec.ValueEncoderError{
			Name:     "EmailCodec",
			Types:    []reflect.Type{tEmail},
			Received: val,
		}
	}
	value := val.Interface().(model.Email)
	return w.WriteString(strings.ToLower(string(value)))
}
