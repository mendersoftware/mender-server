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
import os
import socket

from urllib import parse as urlparse

import docker
import pytest  # noqa
import requests

from bravado.client import SwaggerClient, RequestsClient
from bravado.swagger_model import load_file
from requests.packages.urllib3.exceptions import InsecureRequestWarning
from requests.utils import parse_header_links


requests.packages.urllib3.disable_warnings(InsecureRequestWarning)


class ManagementClient:
    log = logging.getLogger("Client")

    def __init__(self, host, spec, api):
        config = {
            "also_return_response": True,
            "validate_responses": False,
            "validate_requests": False,
            "validate_swagger_spec": True,
            "use_models": True,
        }

        http_client = RequestsClient()
        http_client.session.verify = False

        self.client = SwaggerClient.from_spec(
            load_file(spec), config=config, http_client=http_client,
        )
        self.client.swagger_spec.api_url = "http://%s/api/%s" % (host, api)

        self.group = self.client.get_model("Group")
        self.inventoryAttribute = self.client.get_model("Attribute")
        self.inventoryAttributeTag = self.client.get_model("Tag")

    def deleteAllGroups(self):
        groups = self.client.Management_API.List_Groups().result()[0]
        for g in groups:
            for d in self.getGroupDevices(g):
                self.deleteDeviceInGroup(g, d)

    def getAllDevices(self, page=1, sort=None, has_group=None, JWT="foo.bar.baz"):
        if not JWT.startswith("Bearer "):
            JWT = "Bearer " + JWT
        r, h = self.client.Management_API.List_Device_Inventories(
            page=page, sort=sort, has_group=has_group, Authorization=JWT
        ).result()
        for i in parse_header_links(h.headers["link"]):
            if i["rel"] == "next":
                page = int(
                    dict(urlparse.parse_qs(urlparse.urlsplit(i["url"]).query))["page"][
                        0
                    ]
                )
                return r + self.getAllDevices(page=page, sort=sort)
        else:
            return r

    def getDevice(self, device_id, Authorization="foo"):
        if not Authorization.startswith("Bearer "):
            Authorization = "Bearer " + Authorization
        r, _ = self.client.Management_API.Get_Device_Inventory(
            id=device_id, Authorization=Authorization
        ).result()
        return r

    def updateTagAttributes(self, device_id, tags, eTag=None, JWT="foo.bar.baz"):
        r, _ = self.client.Management_API.Add_Tags(
            id=device_id, If_Match=eTag, tags=tags, Authorization=JWT,
        ).result()
        return r

    def setTagAttributes(self, device_id, tags, eTag=None, JWT="foo.bar.baz"):
        r, _ = self.client.Management_API.Assign_Tags(
            id=device_id, If_Match=eTag, tags=tags, Authorization=JWT,
        ).result()
        return r

    def getAllGroups(self):
        r, _ = self.client.Management_API.List_Groups().result()
        return r

    def getGroupDevices(self, group, expected_error=False):
        try:
            r = self.client.Management_API.Get_Devices_in_Group(
                name=group, Authorization=False
            ).result()
        except Exception as e:
            if expected_error:
                return []
            else:
                pytest.fail()

        else:
            return r[0]

    def deleteDeviceInGroup(self, group, device, expected_error=False):
        try:
            r = self.client.Management_API.Clear_Group(
                id=device, name=group, Authorization=False
            ).result()
        except Exception:
            if expected_error:
                return []
            else:
                pytest.fail()

        else:
            return r

    def addDeviceToGroup(self, group, device, expected_error=False, JWT="foo.bar.baz"):
        if not JWT.startswith("Bearer "):
            JWT = "Bearer " + JWT
        try:
            r = self.client.Management_API.Assign_Group(
                group=group, id=device, Authorization=JWT
            ).result()
        except Exception:
            if expected_error:
                return []
            else:
                pytest.fail()

        else:
            return r


class ManagementClientV2:
    log = logging.getLogger("Client")

    def __init__(self, host, spec, api):
        config = {
            "also_return_response": True,
            "validate_responses": True,
            "validate_requests": True,
            "validate_swagger_spec": True,
            "use_models": True,
        }

        http_client = RequestsClient()
        http_client.session.verify = False

        self.client = SwaggerClient.from_spec(
            load_file(spec), config=config, http_client=http_client,
        )
        self.client.swagger_spec.api_url = (
            "http://%s/api/management/v2/inventory/" % host
        )

    def getFiltersAttributes(self):
        r, _ = self.client.Management_API.Get_filterable_attributes().result()
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


class ApiClient:

    log = logging.getLogger("client.ApiClient")

    def make_api_url(self, path):
        return os.path.join(
            self.api_url, path if not path.startswith("/") else path[1:]
        )

    def setup_swagger(self, host, spec):
        config = {
            "also_return_response": True,
            "validate_responses": True,
            "validate_requests": False,
            "validate_swagger_spec": False,
            "use_models": True,
        }

        self.http_client = RequestsClient()
        self.http_client.session.verify = False

        self.client = SwaggerClient.from_spec(
            load_file(spec), config=config, http_client=self.http_client
        )
        self.client.swagger_spec.api_url = self.api_url

    def __init__(self, host, spec):
        self.setup_swagger(host, spec)


class InternalApiClient(ApiClient):
    log = logging.getLogger("client.InternalClient")

    def __init__(self, host, spec):
        self.api_url = "http://%s/api/internal/v1/inventory/" % host
        super().__init__(host, spec)

    def verify(self, token, uri="/api/management/1.0/auth/verify", method="POST"):
        if not token.startswith("Bearer "):
            token = "Bearer " + token
        return self.client.auth.post_auth_verify(
            **{
                "Authorization": token,
                "X-Forwarded-Uri": uri,
                "X-Forwarded-Method": method,
            }
        ).result()

    def DeviceNew(self, **kwargs):
        return self.client.get_model("DeviceNew")(**kwargs)

    def Attribute(self, **kwargs):
        return self.client.get_model("Attribute")(**kwargs)

    def create_tenant(self, tenant_id):
        return self.client.Internal_API.Create_Tenant(
            tenant={"tenant_id": tenant_id,}
        ).result()

    def create_device(self, device_id, attributes, description="test device"):
        device = self.DeviceNew(
            id=device_id, description=description, attributes=attributes
        )

        return self.client.Internal_API.Initialize_Device(
            tenant_id="", device=device
        ).result()
