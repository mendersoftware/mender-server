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
	"encoding/json"

	"github.com/mendersoftware/mender-server/services/workflows/app/processor"
	"github.com/mendersoftware/mender-server/services/workflows/client/nats"
	"github.com/mendersoftware/mender-server/services/workflows/model"
)

func processNATSTask(
	natsTask *model.NATSTask,
	ps *processor.JobStringProcessor,
	jp *processor.JobProcessor,
	nats nats.Client,
) (*model.TaskResult, error) {

	var result *model.TaskResult = &model.TaskResult{
		NATS: &model.TaskResultNATS{},
	}

	dataJSON := jp.ProcessJSON(natsTask.Data, ps)
	dataJSONBytes, err := json.Marshal(dataJSON)
	if err == nil {
		subject := nats.StreamName() + "." + natsTask.Subject
		err = nats.JetStreamPublish(subject, dataJSONBytes)
	}
	result.Success = err == nil
	if err != nil {
		result.NATS.Error = err.Error()
	}

	return result, nil
}
