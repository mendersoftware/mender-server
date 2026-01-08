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
	"context"
	"os/exec"

	"github.com/pkg/errors"
)

// Type of tasks
const (
	TaskTypeCLI  = "cli"
	TaskTypeHTTP = "http"
	TaskTypeNATS = "nats"
	TaskTypeSMTP = "smtp"
)

var (
	ErrCommandNotFound   = errors.New("command not found")
	ErrCommandNotAllowed = errors.New("command not allowed")

	allowedCommands = map[string]string{
		"create-artifact": "/usr/bin/create-artifact",
	}
)

// Task stores the definition of a task within a workflow
type Task struct {
	Name              string    `json:"name" bson:"name"`
	Type              string    `json:"type" bson:"type"`
	Retries           uint8     `json:"retries" bson:"retries"`
	RetryDelaySeconds uint8     `json:"retryDelaySeconds" bson:"retryDelaySeconds"`
	Requires          []string  `json:"requires,omitempty" bson:"requires,omitempty"`
	HTTP              *HTTPTask `json:"http,omitempty" bson:"http,omitempty"`
	CLI               *CLITask  `json:"cli,omitempty" bson:"cli,omitempty"`
	NATS              *NATSTask `json:"nats,omitempty" bson:"nats,omitempty"`
	SMTP              *SMTPTask `json:"smtp,omitempty" bson:"smtp,omitempty"`
}

// HTTPTask stores the parameters of the HTTP calls for a WorkflowTask
type HTTPTask struct {
	URI               string            `json:"uri" bson:"uri"`
	Method            string            `json:"method" bson:"method"`
	ContentType       string            `json:"contentType,omitempty" bson:"contentType"`
	Body              string            `json:"body,omitempty" bson:"body"`
	FormData          map[string]string `json:"formdata,omitempty" bson:"formdata"`
	JSON              interface{}       `json:"json,omitempty" bson:"json"`
	StatusCodes       []int             `json:"statusCodes,omitempty" bson:"statusCodes"`
	Headers           map[string]string `json:"headers" bson:"headers"`
	ConnectionTimeOut int               `json:"connectionTimeOut" bson:"connectionTimeOut"`
	ReadTimeOut       int               `json:"readTimeOut" bson:"readTimeOut"`
}

// CLITask stores the parameters of the CLI commands for a WorkflowTask
type CLITask struct {
	Command          []string `json:"command" bson:"command"`
	ExecutionTimeOut int      `json:"executionTimeOut" bson:"executionTimeOut"`
	// Override the NewCommand method. Meant to be use for testing
	commandCreator func(ctx context.Context, command []string) (*exec.Cmd, error)
}

func (c *CLITask) WithCustomCreator(
	creator func(context.Context, []string) (*exec.Cmd, error)) {
	c.commandCreator = creator
}

func (c *CLITask) NewCommand(ctx context.Context) (*exec.Cmd, error) {
	if len(c.Command) == 0 {
		return nil, ErrCommandNotFound
	}

	// check if there is a custom commandCreator
	if c.commandCreator != nil {
		return c.commandCreator(ctx, c.Command)
	}

	// default commandCreator
	executable, valid := allowedCommands[c.Command[0]]
	if !valid {
		return nil, ErrCommandNotAllowed
	}

	return exec.CommandContext(ctx, executable, c.Command[1:]...), nil

}

// NATSTask stores the parameters of the NATS parameters for a WorkflowTask
type NATSTask struct {
	Subject string      `json:"subject" bson:"subject"`
	Data    interface{} `json:"data" bson:"data"`
}

// SMTPTask stores the parameters of the SMTP messages for a WorkflowTask
type SMTPTask struct {
	From    string   `json:"from" bson:"from"`
	To      []string `json:"to" bson:"to"`
	Cc      []string `json:"cc" bson:"cc"`
	Bcc     []string `json:"bcc" bson:"bcc"`
	Subject string   `json:"subject" bson:"subject"`
	Body    string   `json:"body" bson:"body"`
	HTML    string   `json:"html" bson:"html"`
	Timeout int      `json:"timeout" bson:"timeout"`
}
