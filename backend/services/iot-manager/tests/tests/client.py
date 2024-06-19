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

import socket

import docker
import requests

from management_api.apis import ManagementAPIClient as GenManagementAPIClient
from management_api import (
    Configuration as mgmt_Configuration,
    ApiClient as mgmt_ApiClient,
)
from internal_api import (
    Configuration as intrnl_Configuration,
    ApiClient as intrnl_ApiClient,
)
from internal_api.apis import InternalAPIClient as GenInternalAPIClient
from utils import generate_jwt


class ManagementAPIClient(GenManagementAPIClient):
    def __init__(self, tenant_id, subject="tester"):
        jwt = generate_jwt(tenant_id, subject, is_user=True)
        config = mgmt_Configuration.get_default_copy()
        config.access_token = jwt
        client = mgmt_ApiClient(configuration=config)
        super().__init__(api_client=client)


class InternalAPIClient(GenInternalAPIClient):
    def __init__(self):
        config = intrnl_Configuration.get_default_copy()
        client = intrnl_ApiClient(configuration=config)
        super().__init__(api_client=client)


class CliIoTManager:
    def __init__(self, service="iot-manager"):
        self.docker = docker.from_env()
        _self = self.docker.containers.list(filters={"id": socket.gethostname()})[0]

        project = _self.labels.get("com.docker.compose.project")
        self.iot_manager = self.docker.containers.list(
            filters={
                "label": [
                    f"com.docker.compose.project={project}",
                    f"com.docker.compose.service={service}",
                ]
            },
            limit=1,
        )[0]

    def sync_devices(self, fail_early=False, batch_size=None, **kwargs):
        cmd = ["/usr/bin/iot-manager", "sync-devices"]
        if batch_size:
            cmd.append("--batch-size")
            cmd.append(str(batch_size))

        if fail_early:
            cmd.append("--fail-early")

        return self.iot_manager.exec_run(cmd, **kwargs)


class MMockAPIClient:
    def __init__(self, mmock_url: str):
        self.mmock_url = mmock_url.removesuffix("/")

    def reset(self):
        requests.get(self.mmock_url + "/api/request/reset")
        requests.get(self.mmock_url + "/api/scenarios/reset_all")

    @property
    def requests(self) -> list[dict]:
        rsp = requests.get(self.mmock_url + "/api/request/all")
        return rsp.json()

    @property
    def unmatched(self) -> list[dict]:
        rsp = requests.get(self.mmock_url + "/api/request/unmatched")
        return rsp.json()

    @property
    def matched(self) -> list[dict]:
        rsp = requests.get(self.mmock_url + "/api/request/matched")
        return rsp.json()

    def set_scenario(self, scenario: str, state: str):
        requests.put(self.mmock_url + f"/api/scenarios/set/{scenario}/{state}")
        requests.put(self.mmock_url + "/api/scenarios/unpause")
