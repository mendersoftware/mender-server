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

import mender_client
from mender_client.api import (
    DeviceInventoryInternalAPIApi,
    DeviceInventoryManagementAPIApi,
    DeviceInventoryFiltersAndSearchManagementAPIApi,
)

from requests.packages.urllib3.exceptions import InsecureRequestWarning
from requests.utils import parse_header_links


requests.packages.urllib3.disable_warnings(InsecureRequestWarning)

DEFAULT_AUTH = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibWVuZGVyLnBsYW4iOiJlbnRlcnByaXNlIn0.s27fi93Qik81WyBmDB5APE0DfGko7Pq8BImbp33-gy4"


class Response:
    """Simple response object compatible with requests.Response interface"""
    def __init__(self, status_code, text=None, data=None):
        self.status_code = status_code
        self.text = text if text is not None else ""
        self.data = data


def default_auth(**kwargs):
    if not "Authorization" in kwargs:
        kwargs["Authorization"] = DEFAULT_AUTH

    if not kwargs["Authorization"].startswith("Bearer "):
        kwargs["Authorization"] = "Bearer " + kwargs["Authorization"]
    return kwargs
class ManagementClient:
    log = logging.getLogger("Client")

    def __init__(self):
        api_conf = mender_client.Configuration.get_default_copy()
        api_conf.access_token = DEFAULT_AUTH.replace("Bearer ", "")
        self.client = DeviceInventoryManagementAPIApi(mender_client.ApiClient(api_conf))

        self.group = mender_client.Group
        self.inventoryAttributeTag = mender_client.Tag

    def inventoryAttribute(self, name, value, scope, description=None):
        """Create an AttributeV1 with the value properly wrapped in AttributeValue."""
        wrapped_value = mender_client.AttributeValue(value)
        return mender_client.AttributeV1(name=name, value=wrapped_value, scope=scope, description=description)

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

    def getAllDevices(self, page=1, sort=None, has_group=None, JWT=DEFAULT_AUTH, per_page=500):
        if not JWT.startswith("Bearer "):
            JWT = "Bearer " + JWT

        self.client.api_client.configuration.access_token = JWT.replace("Bearer ", "")
        resp = self.client.list_device_inventories_with_http_info(
            page=page, sort=sort, has_group=has_group, per_page=per_page
        )
        r = resp.data
        headers = resp.headers or {}

        for i in parse_header_links(headers.get("link", "")):
            if i["rel"] == "next":
                page = int(
                    dict(urlparse.parse_qs(urlparse.urlsplit(i["url"]).query))["page"][
                        0
                    ]
                )
                return r + self.getAllDevices(page=page, sort=sort, per_page=per_page)
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
        kwargs = {"id": device_id, "tag": tags}
        if eTag is not None:
            kwargs["if_match"] = eTag
        r = self.client.add_tags(**kwargs)
        return r

    def setTagAttributes(self, device_id, tags, eTag=None, JWT=DEFAULT_AUTH):
        if not JWT.startswith("Bearer "):
            JWT = "Bearer " + JWT
        self.client.api_client.configuration.access_token = JWT.replace("Bearer ", "")
        kwargs = {"id": device_id, "tag": tags}
        if eTag is not None:
            kwargs["if_match"] = eTag
        r = self.client.assign_tags(**kwargs)
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
            r = self.client.assign_group(group=mender_client.Group(group=group), id=device)
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
        api_conf = mender_client.Configuration.get_default_copy()
        api_conf.access_token = DEFAULT_AUTH.replace("Bearer ", "")
        self.client = DeviceInventoryFiltersAndSearchManagementAPIApi(mender_client.ApiClient(api_conf))

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
        api_conf = mender_client.Configuration.get_default_copy()
        self.client = DeviceInventoryInternalAPIApi(mender_client.ApiClient(api_conf))

    def DeviceNew(self, **kwargs):
        return mender_client.DeviceNew(**kwargs)

    def Attribute(self, **kwargs):
        # Return a dict instead of the model to allow flexible value types
        # (the spec says string, but the backend accepts number, string, array, etc.)
        return kwargs

    def create_tenant(self, tenant_id):
        tenant = mender_client.TenantNew(tenant_id=tenant_id)
        r = self.client.inventory_internal_create_tenant_with_http_info(tenant_new=tenant)
        return Response(status_code=r.status_code, data=r.data)

    def create_device(self, device_id, attributes, description="test device"):
        # Convert attributes to the internal API format using mender_client.Attribute
        # The internal Attribute model has: name, description, value (AttributeValue)
        # The management AttributeV1 has: name, scope, description, value (AttributeV1Value), timestamp
        converted_attrs = []
        for attr in attributes:
            if hasattr(attr, 'value') and hasattr(attr.value, 'actual_instance'):
                # This is an AttributeV1 with AttributeV1Value - extract the raw value
                raw_value = attr.value.actual_instance
            elif hasattr(attr, 'to_dict'):
                attr_dict = attr.to_dict()
                raw_value = attr_dict.get("value")
            elif isinstance(attr, dict):
                raw_value = attr.get("value")
            else:
                raw_value = attr

            name = attr.name if hasattr(attr, 'name') else attr.get("name") if isinstance(attr, dict) else None
            desc = attr.description if hasattr(attr, 'description') else attr.get("description") if isinstance(attr, dict) else None

            # Create mender_client.Attribute with proper AttributeValue wrapper
            internal_attr = mender_client.Attribute(
                name=name,
                value=mender_client.AttributeValue(raw_value),
                description=desc
            )
            converted_attrs.append(internal_attr)

        device = self.DeviceNew(
            id=device_id, attributes=converted_attrs
        )
        r = self.client.initialize_device_with_http_info(tenant_id="", device_new=device)
        return Response(status_code=r.status_code, data=r.data)
