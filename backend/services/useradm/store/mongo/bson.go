package mongo

import (
	"reflect"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/mendersoftware/mender-server/pkg/common/user"
	"github.com/mendersoftware/mender-server/pkg/mongo/v2/codec"
)

func newRegistry() *bson.Registry {
	registry := codec.NewRegistry()
	registry.RegisterTypeEncoder(tEmail, bson.ValueEncoderFunc(encodeEmail))
	return registry
}

var tEmail = reflect.TypeOf(user.Email(""))

func encodeEmail(ec bson.EncodeContext, w bson.ValueWriter, val reflect.Value) error {
	if !val.IsValid() || val.Type() != tEmail {
		return bson.ValueEncoderError{
			Name:     "EmailCodec",
			Types:    []reflect.Type{tEmail},
			Received: val,
		}
	}
	value := val.Interface().(user.Email)
	return w.WriteString(strings.ToLower(string(value)))
}
