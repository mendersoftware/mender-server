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

package model

import (
	"time"

	"github.com/pkg/errors"
)

// Status
const (
	StatusDone = int32(iota)
	StatusPending
	StatusProcessing
	StatusFailure
)

// ErrMsgMissingParamF is the error message for missing input parameters
const ErrMsgMissingParamF = "Missing input parameters: %s"

var (
	// ErrInvalidStatus is the error for invalid status
	ErrInvalidStatus = errors.New("Invalid status")
)

// Job defines the execution job a workflow
type Job struct {
	// Id is the ID of the job
	ID string `json:"id" bson:"_id"`

	// WorkflowName contains the name of the workflow
	WorkflowName string `json:"workflowName" bson:"workflow_name"`

	// InputParameters contains the name of the workflow
	InputParameters InputParameters `json:"inputParameters" bson:"input_parameters"`

	// Enumerated status of the Job and string field used for unmarshalling
	Status       int32  `json:"-" bson:"status"`
	StatusString string `json:"status,omitempty" bson:"-"`

	// Results produced by a finished job. If status is not "done" this
	// field will always be nil.
	Results []TaskResult `json:"results" bson:"results,omitempty"`

	// insert time
	InsertTime time.Time `json:"insert_time" bson:"insert_time,omitempty"`

	//workflow version
	WorkflowVersion string `json:"version" bson:"version,omitempty"`
}

// InputParameter defines the input parameter of a job
type InputParameter struct {
	// Name of the parameter
	Name string `json:"name" bson:"name"`

	// Value of the input parameter
	Value string `json:"value" bson:"value"`

	// Raw value of the input parameter
	Raw interface{} `json:"raw,omitempty" bson:"raw"`
}

type InputParameters []InputParameter

func (param InputParameters) Map() map[string]interface{} {
	var ret = map[string]interface{}{}
	for _, val := range param {
		ret[val.Name] = val.Value
	}
	return ret
}

// TaskResult contains the result of the execution of a task
type TaskResult struct {
	Name        string                 `json:"name" bson:"name"`
	Type        string                 `json:"type" bson:"type"`
	Success     bool                   `json:"success" bson:"success"`
	Skipped     bool                   `json:"skipped" bson:"skipped"`
	CLI         *TaskResultCLI         `json:"cli,omitempty" bson:"cli,omitempty"`
	HTTPRequest *TaskResultHTTPRequest `json:"httpRequest,omitempty" bson:"httpRequest,omitempty"`
	//nolint:lll
	HTTPResponse *TaskResultHTTPResponse `json:"httpResponse,omitempty" bson:"httpResponse,omitempty"`
	NATS         *TaskResultNATS         `json:"nats,omitempty" bson:"nats,omitempty"`
	SMTP         *TaskResultSMTP         `json:"smtp,omitempty" bson:"smtp,omitempty"`
}

// TaskResultCLI contains the CLI command, the output and the exit status
type TaskResultCLI struct {
	Command  []string `json:"command" bson:"command"`
	Output   string   `json:"output" bson:"output"`
	Error    string   `json:"error" bson:"error"`
	ExitCode int      `json:"exitCode" bson:"exitCode"`
}

type TaskResultNATS struct {
	Error string `json:"error" bson:"error"`
}

// TaskResultHTTPRequest contains the request
type TaskResultHTTPRequest struct {
	URI     string   `json:"uri" bson:"uri"`
	Method  string   `json:"method" bson:"method"`
	Body    string   `json:"body" bson:"body"`
	Headers []string `json:"headers" bson:"headers"`
}

// TaskResultHTTPResponse contains the response
type TaskResultHTTPResponse struct {
	StatusCode int    `json:"statusCode" bson:"status_code"`
	Body       string `json:"body" bson:"body"`
}

// TaskResultSMTP contains the SMTP message, the output and the exit status
type TaskResultSMTP struct {
	Sender     string   `json:"sender" bson:"sender"`
	Recipients []string `json:"recipients" bson:"recipients"`
	Message    string   `json:"message" bson:"message"`
	Error      string   `json:"error" bson:"error"`
}

// Validate job against workflow. Check that all required parameters are present.
func (job *Job) Validate(workflow *Workflow) error {
	var missing []string
	inputParameters := make(map[string]string)
	for _, param := range job.InputParameters {
		inputParameters[param.Name] = param.Value
	}
	for _, key := range workflow.InputParameters {
		if _, ok := inputParameters[key]; !ok {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return errors.Errorf(ErrMsgMissingParamF, missing)
	}
	return nil
}

func (job *Job) PrepareForJSONMarshalling() {
	job.StatusString = StatusToString(job.Status)
	for i := range job.InputParameters {
		job.InputParameters[i].Raw = nil
	}
}

// StatusToString returns the job's status as a string
func StatusToString(status int32) string {
	var ret string
	switch status {
	case StatusPending:
		ret = "pending"
	case StatusProcessing:
		ret = "processing"
	case StatusDone:
		ret = "done"
	case StatusFailure:
		ret = "failed"
	default:
		ret = "unknown"
	}
	return ret
}
