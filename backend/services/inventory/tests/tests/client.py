# Copyright 2023 Northern.tech AS
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
import socket

from urllib import parse as urlparse

import docker
import pytest  # noqa
import requests

import internal_v1
import management_v1
import management_v2

from requests.packages.urllib3.exceptions import InsecureRequestWarning
from requests.utils import parse_header_links


requests.packages.urllib3.disable_warnings(InsecureRequestWarning)

DEFAULT_AUTH = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibWVuZGVyLnBsYW4iOiJlbnRlcnByaXNlIn0.s27fi93Qik81WyBmDB5APE0DfGko7Pq8BImbp33-gy4"

def default_auth(**kwargs):
    if not "Authorization" in kwargs:
        kwargs["Authorization"] = DEFAULT_AUTH

    if not kwargs["Authorization"].startswith("Bearer "):
        kwargs["Authorization"] = "Bearer " + kwargs["Authorization"]
    return kwargs
class ManagementClient:
    log = logging.getLogger("Client")

    def __init__(self):
        api_conf = management_v1.Configuration.get_default_copy()
        api_conf.access_token = DEFAULT_AUTH.replace("Bearer ", "")
        self.client = management_v1.ManagementAPIClient(management_v1.ApiClient(api_conf))

        self.group = management_v1.Group
        self.inventoryAttribute = management_v1.AttributeV1
        self.inventoryAttributeTag = management_v1.Tag

    def deleteAllGroups(self, **kwargs):
        if "Authorization" not in kwargs:
            kwargs["Authorization"] = DEFAULT_AUTH
        if not kwargs["Authorization"].startswith("Bearer "):
            kwargs["Authorization"] = "Bearer " + kwargs["Authorization"]

        self.client.api_client.configuration.access_token = kwargs["Authorization"].replace("Bearer ", "")
        groups = self.client.list_groups()
        for g in groups:
            for d in self.getGroupDevices(g):
                self.deleteDeviceInGroup(g, d)

    def getAllDevices(self, page=1, sort=None, has_group=None, JWT=DEFAULT_AUTH):
        if not JWT.startswith("Bearer "):
            JWT = "Bearer " + JWT

        self.client.api_client.configuration.access_token = JWT.replace("Bearer ", "")
        r, status, headers = self.client.list_device_inventories_with_http_info(
            page=page, sort=sort, has_group=has_group
        )

        for i in parse_header_links(headers.get("link", "")):
            if i["rel"] == "next":
                page = int(
                    dict(urlparse.parse_qs(urlparse.urlsplit(i["url"]).query))["page"][
                        0
                    ]
                )
                return r + self.getAllDevices(page=page, sort=sort)
        else:
            return r

    def getDevice(self, device_id, Authorization=DEFAULT_AUTH):
        if not Authorization.startswith("Bearer "):
            Authorization = "Bearer " + Authorization
        self.client.api_client.configuration.access_token = Authorization.replace("Bearer ", "")
        r = self.client.get_device_inventory(id=device_id)
        return r

    def updateTagAttributes(self, device_id, tags, eTag=None, JWT=DEFAULT_AUTH):
        if not JWT.startswith("Bearer "):
            JWT = "Bearer " + JWT
        self.client.api_client.configuration.access_token = JWT.replace("Bearer ", "")
        r = self.client.add_tags(id=device_id, if_match=eTag, tag=tags)
        return r

    def setTagAttributes(self, device_id, tags, eTag=None, JWT=DEFAULT_AUTH):
        if not JWT.startswith("Bearer "):
            JWT = "Bearer " + JWT
        self.client.api_client.configuration.access_token = JWT.replace("Bearer ", "")
        r = self.client.assign_tags(id=device_id, if_match=eTag, tag=tags)
        return r

    def getAllGroups(self, **kwargs):
        if "Authorization" not in kwargs:
            kwargs["Authorization"] = DEFAULT_AUTH
        if not kwargs["Authorization"].startswith("Bearer "):
            kwargs["Authorization"] = "Bearer " + kwargs["Authorization"]
        self.client.api_client.configuration.access_token = kwargs["Authorization"].replace("Bearer ", "")
        r = self.client.list_groups()
        return r

    def getGroupDevices(self, group, expected_error=False, **kwargs):
        try:
            if "Authorization" not in kwargs:
                kwargs["Authorization"] = DEFAULT_AUTH
            if not kwargs["Authorization"].startswith("Bearer "):
                kwargs["Authorization"] = "Bearer " + kwargs["Authorization"]

            self.client.api_client.configuration.access_token = kwargs["Authorization"].replace("Bearer ", "")
            r = self.client.get_devices_in_group(name=group)
        except Exception as e:
            if expected_error:
                return []
            else:
                pytest.fail()
        else:
            return r

    def deleteDeviceInGroup(self, group, device, expected_error=False, **kwargs):
        try:
            if "Authorization" not in kwargs:
                kwargs["Authorization"] = DEFAULT_AUTH
            if not kwargs["Authorization"].startswith("Bearer "):
                kwargs["Authorization"] = "Bearer " + kwargs["Authorization"]
            self.client.api_client.configuration.access_token = kwargs["Authorization"].replace("Bearer ", "")
            r = self.client.clear_group(id=device, name=group)
        except Exception:
            if expected_error:
                return []
            else:
                pytest.fail()
        else:
            return r

    def addDeviceToGroup(self, group, device, expected_error=False, JWT=DEFAULT_AUTH):
        if not JWT.startswith("Bearer "):
            JWT = "Bearer " + JWT
        try:
            self.client.api_client.configuration.access_token = JWT.replace("Bearer ", "")
            r = self.client.assign_group(group=management_v1.Group(group), id=device)
        except Exception:
            if expected_error:
                return []
            else:
                pytest.fail()
        else:
            return r


class ManagementClientV2:
    log = logging.getLogger("Client")

    def __init__(self):
        api_conf = management_v2.Configuration.get_default_copy()
        api_conf.access_token = DEFAULT_AUTH.replace("Bearer ", "")
        self.client = management_v2.ManagementAPIClient(management_v2.ApiClient(api_conf))

    def getFiltersAttributes(self, **kwargs):
        if "Authorization" not in kwargs:
            kwargs["Authorization"] = DEFAULT_AUTH
        if not kwargs["Authorization"].startswith("Bearer "):
            kwargs["Authorization"] = "Bearer " + kwargs["Authorization"]
        self.client.api_client.configuration.access_token = kwargs["Authorization"].replace("Bearer ", "")
        r = self.client.get_filterable_attributes()
        return r


class CliClient:
    exec_path = "/usr/bin/inventory"

    def __init__(self, service="inventory"):
        self.docker = docker.from_env()
        # HACK: Find docker-compose project by identifying the container
        # we're running inside. The hostname equals the container id.
        _self = self.docker.containers.list(filters={"id": socket.gethostname()})[0]

        project = _self.labels.get("com.docker.compose.project")
        self.container = self.docker.containers.list(
            filters={
                "label": [
                    f"com.docker.compose.project={project}",
                    f"com.docker.compose.service={service}",
                ]
            },
            limit=1,
        )[0]

    def migrate(self, tenant_id=None, **kwargs):
        cmd = [self.exec_path, "migrate"]

        if tenant_id is not None:
            cmd.extend(["--tenant", tenant_id])

        code, (stdout, stderr) = self.container.exec_run(cmd, demux=True, **kwargs)
        return code, stdout, stderr


class InternalApiClient:
    log = logging.getLogger("client.InternalClient")

    def __init__(self):
        api_conf = internal_v1.Configuration.get_default_copy()
        self.client = internal_v1.InternalAPIClient(internal_v1.ApiClient(api_conf))

    def DeviceNew(self, **kwargs):
        return internal_v1.DeviceNew(**kwargs)

    def Attribute(self, **kwargs):
        return internal_v1.Attribute(**kwargs)

    def create_tenant(self, tenant_id):
        tenant = internal_v1.TenantNew(tenant_id=tenant_id)
        return self.client.inventory_internal_create_tenant(tenant_new=tenant)

    def create_device(self, device_id, attributes, description="test device"):
        device = self.DeviceNew(
            id=device_id, attributes=attributes
        )
        return self.client.initialize_device(tenant_id="", device_new=device)
