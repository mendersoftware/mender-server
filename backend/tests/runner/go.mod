module github.com/mendersoftware/mender-server/tests/runner

go 1.25.7

require (
	github.com/google/uuid v1.6.0
	github.com/mendersoftware/mender-server v0.0.0
	github.com/stretchr/testify v1.11.1
	go.mongodb.org/mongo-driver v1.17.9
	gopkg.in/validator.v2 v2.0.1
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/mendersoftware/mender-server => ../..

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/klauspost/compress v1.18.3 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/text v0.33.0 // indirect
)
