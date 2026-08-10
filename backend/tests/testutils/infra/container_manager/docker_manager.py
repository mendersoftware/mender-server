# Copyright 2022 Northern.tech AS
#
#    Licensed under the Apache License, Version 2.0 (the "License");
#    you may not use this file except in compliance with the License.
#    You may obtain a copy of the License at
#
#        http://www.apache.org/licenses/LICENSE-2.0
#
#    Unless required by applicable law or agreed to in writing, software
#    distributed under the License is distributed on an "AS IS" BASIS,
#    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#    See the License for the specific language governing permissions and
#    limitations under the License.
import logging
import subprocess

from .base import BaseContainerManagerNamespace

logger = logging.getLogger(__name__)


class DockerNamespace(BaseContainerManagerNamespace):
    def __init__(self, name):
        BaseContainerManagerNamespace.__init__(self, name)

    def setup(self):
        pass

    def teardown(self):
        pass

    def execute(self, container_id, cmd):
        cmd = ["docker", "exec", "{}".format(container_id)] + cmd
        ret = subprocess.check_output(cmd).decode("utf-8").strip()
        return ret

    def cmd(self, container_id, docker_cmd, cmd=[]):
        cmd = ["docker", docker_cmd] + [str(container_id)] + cmd
        ret = subprocess.run(
            cmd, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE
        )
        return ret.stdout.decode("utf-8").strip()

    def download(self, container_id, source, destination):
        return self._cp(f"{container_id}:{source}", destination)

    def upload(self, container_id, source, destination):
        return self._cp(source, f"{container_id}:{destination}")

    def _cp(self, source, destination):
        cmd = ["docker", "cp", source, destination]
        ret = subprocess.check_output(cmd).decode("utf-8").strip()
        return ret

    def getid(self, service):
        """Container id of `service` within this namespace.

        Matched on the compose labels rather than by grepping `docker ps` output.
        A substring match also hits any project whose name merely *contains* this
        one -- "mender" against "mender_enterprise", or a primary namespace
        against its own failover backend -- and returns several ids on one line.
        """
        cmd = (
            "docker ps -q "
            "--filter label=com.docker.compose.project={project} "
            "--filter label=com.docker.compose.service={service}"
        ).format(project=self.name, service=service)
        ids = subprocess.check_output(cmd, shell=True).decode("utf-8").split()

        if not ids:
            raise RuntimeError(
                "no container for service %r in project %r" % (service, self.name)
            )
        if len(ids) > 1:
            logger.debug(
                "%d containers for service %r in project %r, using the first",
                len(ids),
                service,
                self.name,
            )
        return ids[0]
