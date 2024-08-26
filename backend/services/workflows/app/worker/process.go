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
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/mendersoftware/mender-server/pkg/log"

	"github.com/mendersoftware/mender-server/services/workflows/app/processor"
	"github.com/mendersoftware/mender-server/services/workflows/client/nats"
	"github.com/mendersoftware/mender-server/services/workflows/model"
	"github.com/mendersoftware/mender-server/services/workflows/store"
)

// this variable is set when running acceptance tests to disable ephemeral jobs
// without it, it is not possible to inspect the jobs collection to assert the
// values stored in the database
var NoEphemeralWorkflows = false

func processJob(ctx context.Context, job *model.Job,
	dataStore store.DataStore, nats nats.Client) error {
	l := log.FromContext(ctx)

	workflow, err := dataStore.GetWorkflowByName(ctx, job.WorkflowName, job.WorkflowVersion)
	if err != nil {
		l.Warnf("The workflow %q of job %s does not exist: %v",
			job.WorkflowName, job.ID, err)
		err := dataStore.UpdateJobStatus(ctx, job, model.StatusFailure)
		if err != nil {
			return err
		}
		return nil
	}

	if !workflow.Ephemeral || NoEphemeralWorkflows {
		job.Status = model.StatusPending
		_, err = dataStore.UpsertJob(ctx, job)
		if err != nil {
			return errors.Wrap(err, "insert of the job failed")
		}
	}

	l.Infof("%s: started, %s", job.ID, job.WorkflowName)

	success := true
	for _, task := range workflow.Tasks {
		l.Infof("%s: started, %s task :%s", job.ID, job.WorkflowName, task.Name)
		var (
			result  *model.TaskResult
			attempt uint8 = 0
		)
		for attempt <= task.Retries {
			result, err = processTask(task, job, workflow, nats, l)
			if err != nil {
				_ = dataStore.UpdateJobStatus(ctx, job, model.StatusFailure)
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
		job.Results = append(job.Results, *result)
		if !workflow.Ephemeral || !result.Success || NoEphemeralWorkflows {
			err = dataStore.UpdateJobAddResult(ctx, job, result)
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
		err = dataStore.UpdateJobStatus(ctx, job, status)
		if err != nil {
			l.Warn(fmt.Sprintf("Unable to set job status to %s", newStatus))
			return err
		}
	}

	l.Infof("%s: done", job.ID)
	return nil
}

func processTask(task model.Task, job *model.Job,
	workflow *model.Workflow, nats nats.Client, l *log.Logger) (*model.TaskResult, error) {

	var result *model.TaskResult
	var err error

	ps := processor.NewJobStringProcessor(job)
	jp := processor.NewJobProcessor(job)
	if len(task.Requires) > 0 {
		for _, require := range task.Requires {
			require = ps.ProcessJobString(require)
			require = ps.MaybeExecuteGoTemplate(require)
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
		result, err = processHTTPTask(httpTask, ps, jp, l)
	case model.TaskTypeCLI:
		var cliTask *model.CLITask = task.CLI
		if cliTask == nil {
			return nil, fmt.Errorf(
				"Error: Task definition incompatible " +
					"with specified type (cli)")
		}
		result, err = processCLITask(cliTask, ps, jp)
	case model.TaskTypeNATS:
		var natsTask *model.NATSTask = task.NATS
		if natsTask == nil {
			return nil, fmt.Errorf(
				"Error: Task definition incompatible " +
					"with specified type (nats)")
		}
		result, err = processNATSTask(natsTask, ps, jp, nats)
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
		result, err = processSMTPTask(smtpTask, ps, jp, l)
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
