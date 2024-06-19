// Copyright 2021 Northern.tech AS
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

package mongo

import "go.mongodb.org/mongo-driver/mongo"

const (
	ErrCodeDuplicateKey = 11000
)

// IsDuplicateKeyErr checks the errors and inspects if (one of) the error(s)
// is duplicate key a error.
func IsDuplicateKeyErr(err error) bool {
	switch t := err.(type) {
	case mongo.BulkWriteError:
		if t.Code == ErrCodeDuplicateKey {
			return true
		}
	case mongo.CommandError:
		if t.Code == ErrCodeDuplicateKey {
			return true
		}
	case mongo.WriteError:
		if t.Code == ErrCodeDuplicateKey {
			return true
		}
	case mongo.WriteErrors:
		for _, e := range t {
			if e.Code == ErrCodeDuplicateKey {
				return true
			}
		}
	case mongo.WriteException:
		return IsDuplicateKeyErr(t.WriteErrors)
	}
	return false
}
