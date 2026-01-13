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
	"bytes"
	"context"
	"os/exec"
	"time"

	"github.com/mendersoftware/mender-server/services/workflows/model"
)

const (
	// MaxExecutionTime is the maximum number of seconds a task can run
	MaxExecutionTime int = 3600 * 4
)

func (jp *JobProcessor) ProcessCLITask(
	cliTask *model.CLITask,
	ps *JobStringProcessor,
) (*model.TaskResult, error) {
	commands := make([]string, 0, 10)
	for _, command := range cliTask.Command {
		command := ps.ProcessJobString(command)
		commands = append(commands, command)
	}

	ctx := context.Background()
	timeout := cliTask.ExecutionTimeOut
	if timeout <= 0 {
		timeout = MaxExecutionTime
	}
	ctxWithOptionalTimeOut, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	var result *model.TaskResult = &model.TaskResult{
		CLI: &model.TaskResultCLI{
			Command: commands,
		},
	}

	cmd := exec.CommandContext(ctxWithOptionalTimeOut, commands[0], commands[1:]...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		result.Success = false
		if exiterr, ok := err.(*exec.ExitError); ok {
			result.CLI.ExitCode = exiterr.ExitCode()
		} else {
			result.CLI.ExitCode = -1
		}
	} else {
		result.Success = true
		result.CLI.ExitCode = 0
	}
	result.CLI.Output = stdout.String()
	result.CLI.Error = stderr.String()

	return result, nil
}
