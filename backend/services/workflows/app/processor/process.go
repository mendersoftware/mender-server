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

package processor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/workflows/model"
)

// this variable is set when running acceptance tests to disable ephemeral jobs
// without it, it is not possible to inspect the jobs collection to assert the
// values stored in the database
var NoEphemeralWorkflows = false

func (jp *JobProcessor) ProcessJob(ctx context.Context) error {
	l := log.FromContext(ctx)

	workflow, err := jp.store.GetWorkflowByName(ctx, jp.job.WorkflowName, jp.job.WorkflowVersion)
	if err != nil {
		l.Warnf("The workflow %q of job %s does not exist: %v",
			jp.job.WorkflowName, jp.job.ID, err)
		err := jp.store.UpdateJobStatus(ctx, jp.job, model.StatusFailure)
		if err != nil {
			return err
		}
		return nil
	}

	if !workflow.Ephemeral || NoEphemeralWorkflows {
		jp.job.Status = model.StatusPending
		_, err = jp.store.UpsertJob(ctx, jp.job)
		if err != nil {
			return errors.Wrap(err, "insert of the job failed")
		}
	}

	l.Infof("%s: started, %s", jp.job.ID, jp.job.WorkflowName)

	success := true
	for _, task := range workflow.Tasks {
		l.Infof("%s: started, %s task :%s", jp.job.ID, jp.job.WorkflowName, task.Name)
		var (
			result  *model.TaskResult
			attempt uint8 = 0
		)
		for attempt <= task.Retries {

			result, err = jp.processTask(task, workflow, l)
			if err != nil {
				_ = jp.store.UpdateJobStatus(ctx, jp.job, model.StatusFailure)
				return err
			}
			attempt++
			if result.Success {
				break
			}
			if task.RetryDelaySeconds > 0 {
				time.Sleep(time.Duration(task.RetryDelaySeconds) * time.Second)
			}
		}
		jp.job.Results = append(jp.job.Results, *result)
		if !workflow.Ephemeral || !result.Success || NoEphemeralWorkflows {
			err = jp.store.UpdateJobAddResult(ctx, jp.job, result)
			if err != nil {
				l.Errorf("Error uploading results: %s", err.Error())
			}
		}
		if !result.Success {
			success = false
			break
		}
	}

	var status int32
	if success {
		status = model.StatusDone
	} else {
		status = model.StatusFailure
	}
	if !workflow.Ephemeral || NoEphemeralWorkflows {
		newStatus := model.StatusToString(status)
		err = jp.store.UpdateJobStatus(ctx, jp.job, status)
		if err != nil {
			l.Warn(fmt.Sprintf("Unable to set job status to %s", newStatus))
			return err
		}
	}

	l.Infof("%s: done", jp.job.ID)
	return nil
}

func (jp *JobProcessor) processTask(task model.Task,
	workflow *model.Workflow, l *log.Logger) (*model.TaskResult, error) {

	var result *model.TaskResult
	var err error

	ps := NewJobStringProcessor(jp.job)
	if len(task.Requires) > 0 {
		for _, require := range task.Requires {
			require = ps.ProcessJobString(require)
			if require == "" {
				result := &model.TaskResult{
					Name:    task.Name,
					Type:    task.Type,
					Success: true,
					Skipped: true,
				}
				return result, nil
			}
		}
	}

	switch task.Type {
	case model.TaskTypeHTTP:
		var httpTask *model.HTTPTask = task.HTTP
		if httpTask == nil {
			return nil, fmt.Errorf(
				"Error: Task definition incompatible " +
					"with specified type (http)")
		}
		result, err = jp.ProcessHTTPTask(httpTask, ps, l)
	case model.TaskTypeCLI:
		var cliTask *model.CLITask = task.CLI
		if cliTask == nil {
			return nil, fmt.Errorf(
				"Error: Task definition incompatible " +
					"with specified type (cli)")
		}
		result, err = jp.ProcessCLITask(cliTask, ps)
	case model.TaskTypeNATS:
		var natsTask *model.NATSTask = task.NATS
		if natsTask == nil {
			return nil, fmt.Errorf(
				"Error: Task definition incompatible " +
					"with specified type (nats)")
		}
		result, err = jp.ProcessNATSTask(natsTask, ps)
	case model.TaskTypeSMTP:
		var smtpTask *model.SMTPTask = task.SMTP
		if smtpTask == nil {
			return nil, fmt.Errorf(
				"Error: Task definition incompatible " +
					"with specified type (smtp)")
		}
		l.Infof("processTask: calling smtp task: From: %s To: %s Subject: %s",
			ps.ProcessJobString(smtpTask.From),
			ps.ProcessJobString(strings.Join(smtpTask.To, ",")),
			ps.ProcessJobString(smtpTask.Subject),
		)
		result, err = jp.processSMTPTask(smtpTask, ps, l)
	default:
		result = nil
		err = fmt.Errorf("Unrecognized task type: %s", task.Type)
	}
	if result != nil {
		result.Name = task.Name
		result.Type = task.Type
	}
	return result, err
}
