// Copyright 2026 Northern.tech AS
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

package cli

import (
	"fmt"
	"os/exec"
	"strings"
)

const DefaultNamespace = "backend-tests"

// ExecInContainer runs a command inside a Docker container by service name.
// namespace is the Docker Compose project name (e.g. "backend-tests").
func ExecInContainer(namespace, service string, args ...string) (string, error) {
	if namespace == "" {
		namespace = DefaultNamespace
	}
	container := fmt.Sprintf("%s-%s-1", namespace, service)

	cmdArgs := append([]string{"exec", container}, args...)
	cmd := exec.Command("docker", cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("docker exec %s failed: %w: %s", service, err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}
