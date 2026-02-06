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

import management_v1 as management_api
import internal_v1 as internal_api
from management_v1 import (
    ManagementAPIClient as GenManagementAPIClient,
    Configuration as mgmt_Configuration,
    ApiClient as mgmt_ApiClient,
)
from internal_v1 import (
    InternalAPIClient as GenInternalAPIClient,
    Configuration as intrnl_Configuration,
    ApiClient as intrnl_ApiClient,
)
from utils import generate_jwt


class ManagementAPIClient(GenManagementAPIClient):
    def __init__(self, tenant_id, subject="tester"):
        jwt = generate_jwt(tenant_id, subject, is_user=True)
        config = mgmt_Configuration.get_default_copy()
        config.access_token = jwt
        client = mgmt_ApiClient(configuration=config)
        super().__init__(api_client=client)

    def list_integrations(self, **kwargs):
        return self.io_t_manager_management_list_integrations(**kwargs)

    def register_integration(self, integration, **kwargs):
        return self.io_t_manager_management_register_integration(integration, **kwargs)

    def register_integration_with_http_info(self, integration, **kwargs):
        return self.io_t_manager_management_register_integration_with_http_info(integration, **kwargs)

    def remove_integration(self, id, **kwargs):
        return self.io_t_manager_management_remove_integration(id, **kwargs)

    def set_integration_credentials(self, id, credentials, **kwargs):
        return self.io_t_manager_management_set_integration_credentials(
            id, credentials, **kwargs
        )

    def unregister_device_integrations(self, device_id, **kwargs):
        return self.io_t_manager_management_unregister_device_integrations(
            device_id, **kwargs
        )

    def get_device_states(self, device_id, **kwargs):
        return self.io_t_manager_management_get_device_states(device_id, **kwargs)

    def replace_state(self, device_id, integration_id, **kwargs):
        return self.io_t_manager_management_replace_state(
            device_id, integration_id, **kwargs
        )

    def get_device_state(self, device_id, integration_id, **kwargs):
        return self.io_t_manager_management_get_device_state(
            device_id, integration_id, **kwargs
        )

    def list_events(self, **kwargs):
        return self.io_t_manager_management_list_events(**kwargs)


class InternalAPIClient(GenInternalAPIClient):
    def __init__(self):
        config = intrnl_Configuration.get_default_copy()
        client = intrnl_ApiClient(configuration=config)
        super().__init__(api_client=client)

    def check_health(self, **kwargs):
        return self.io_t_manager_internal_check_health(**kwargs)

    def check_liveliness(self, **kwargs):
        return self.io_t_manager_internal_check_liveliness(**kwargs)

    def provision_device(self, tenant_id, new_device, **kwargs):
        return self.io_t_manager_internal_provision_device(
            tenant_id, new_device, **kwargs
        )

    def delete_tenant(self, tenant_id, **kwargs):
        return self.io_t_manager_internal_delete_tenant(tenant_id, **kwargs)

    def decommission_device(self, tenant_id, device_id, **kwargs):
        return self.io_t_manager_internal_decommission_device(
            tenant_id, device_id, **kwargs
        )

    def update_device_statuses(self, tenant_id, status, request_body, **kwargs):
        return self.io_t_manager_internal_update_device_statuses(
            tenant_id, status, request_body, **kwargs
        )


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
