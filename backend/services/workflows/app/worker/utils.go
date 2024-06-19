// Copyright 2022 Northern.tech AS
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

package worker

import (
	"io/ioutil"
	"strings"

	"github.com/mendersoftware/mender-server/services/workflows/app/processor"
)

func processJobStringOrFile(data string, ps *processor.JobStringProcessor) (string, error) {
	data = ps.ProcessJobString(data)
	if strings.HasPrefix(data, "@") {
		filePath := data[1:]
		buffer, err := ioutil.ReadFile(filePath)
		if err != nil {
			return "", err
		}
		data = ps.ProcessJobString(string(buffer))
	}
	return data, nil
}
