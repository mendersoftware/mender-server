package mongo

import "go.mongodb.org/mongo-driver/v2/mongo/options"

var (
	defaultClientOptions = &options.BSONOptions{
		ObjectIDAsHexString: true,
		DefaultDocumentM:    true,
	}
)

func BaseClientOptions(uri string) *options.ClientOptions {
	return options.Client().
		ApplyURI(uri).
		SetBSONOptions(defaultClientOptions)
}
