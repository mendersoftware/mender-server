#!/usr/bin/python
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
import json
import logging
import os.path
import socket

import docker
import pytest  # noqa
import requests

from bson.objectid import ObjectId
from base64 import b64encode
from datetime import datetime, timedelta
import hmac
import uuid

import devices_api
import internal_api
import management_api
from management_api import models as management_models


def generate_jwt(tenant_id: str = "", subject: str = "", is_user: bool = True) -> str:
    if len(subject) == 0:
        subject = str(uuid.uuid4())

    hdr = {
        "alg": "HS256",
        "typ": "JWT",
    }
    hdr64 = (
        b64encode(json.dumps(hdr).encode(), altchars=b"-_").decode("ascii").rstrip("=")
    )

    claims = {
        "sub": subject,
        "exp": (datetime.utcnow() + timedelta(hours=1)).isoformat("T"),
        "mender.user": is_user,
        "mender.device": not is_user,
        "mender.tenant": tenant_id,
    }
    if is_user:
        claims["mender.user"] = True
    else:
        claims["mender.device"] = True

    claims64 = (
        b64encode(json.dumps(claims).encode(), altchars=b"-_")
        .decode("ascii")
        .rstrip("=")
    )

    jwt = hdr64 + "." + claims64
    sign = hmac.new(b"secretJWTkey", msg=jwt.encode(), digestmod="sha256")
    sign64 = b64encode(sign.digest(), altchars=b"-_").decode("ascii").rstrip("=")
    return jwt + "." + sign64


class BaseApiClient:
    def __init__(self, hostname):
        self.api_url = "http://%s/api/management/v1/devauth/" % hostname

    def make_api_url(self, path):
        return os.path.join(
            self.api_url, path if not path.startswith("/") else path[1:]
        )


class BaseDevicesApiClient(BaseApiClient):
    def __init__(self, hostname):
        self.api_url = "http://%s/api/devices/v1/authentication/" % hostname

    @property
    def auth_requests_url(self):
        """Provides device identity as a string"""
        return self.make_api_url("/auth_requests")


class SwaggerApiClient(BaseApiClient):
    config = {
        "also_return_response": True,
        "validate_responses": True,
        "validate_requests": False,
        "validate_swagger_spec": True,
        "use_models": True,
    }

    log = logging.getLogger("client.SwaggerApiClient")

    def __init__(self, hostname, swagger_spec):
        self.spec = swagger_spec
        self.config = management_api.Configuration()
        api_conf = management_api.Configuration.get_default_copy()
        device_id = str(uuid.uuid4())
        tenant_id = ""  # str(ObjectId())
        user_id = str(uuid.uuid4())
        api_conf.access_token = generate_jwt(tenant_id, user_id, is_user=True)
        self.client = management_api.ManagementAPIClient(
            management_api.ApiClient(configuration=api_conf)
        )
        api_conf = internal_api.Configuration.get_default_copy()
        self.clientInternal = internal_api.InternalAPIClient(
            internal_api.ApiClient(configuration=api_conf)
        )
        super().__init__(hostname)


class InternalClient(SwaggerApiClient):
    def __init__(self, hostname, swagger_spec):
        super().__init__(hostname, swagger_spec)
        self.api_url = "http://%s/api/internal/v1/devauth/" % hostname

    log = logging.getLogger("client.InternalClient")

    spec_option = "spec"

    def get_max_devices_limit(self, tenant_id):
        return self.clientInternal.get_device_limit(tenant_id=tenant_id)

    def put_max_devices_limit(self, tenant_id, limit, client_side_validation=True):
        if client_side_validation:
            l = internal_api.Limit(limit=limit)
            return self.clientInternal.update_device_limit(tenant_id=tenant_id, limit=l)
        else:
            api_client = self.clientInternal.api_client
            url = f"{api_client.configuration.host}/tenant/{tenant_id}/limits/max_devices"
            return api_client.call_api(
                "PUT",
                url,
                header_params={"Content-Type": "application/json"},
                body={"limit": limit},
            )

    def create_tenant(self, tenant_id):
        return self.clientInternal.create_tenant(new_tenant={"tenant_id": tenant_id})

    def delete_device(self, device_id, tenant_id="", headers={}):
        return self.clientInternal.delete_device(tid=tenant_id, did=device_id)

    def verify_jwt(self, authorization):
        return self.clientInternal.verify_jwt(authorization=authorization)


class SimpleInternalClient(InternalClient):
    """Internal API client. Cannot be used as pytest base class"""

    log = logging.getLogger("client.SimpleInternalClient")

    def __init__(self, hostname, swagger_spec):
        super().__init__(hostname, swagger_spec)


class ManagementClient(SwaggerApiClient):
    def __init__(self, hostname, swagger_spec):
        super().__init__(hostname, swagger_spec)
        self.api_url = "http://%s/api/management/v2/devauth/" % hostname

    log = logging.getLogger("client.ManagementClient")

    spec_option = "management_spec"

    def accept_device(self, devid, aid, **kwargs):
        return self.put_device_status(devid, aid, "accepted", **kwargs)

    def reject_device(self, devid, aid, **kwargs):
        return self.put_device_status(devid, aid, "rejected", **kwargs)

    def put_device_status(self, devid, aid, status, **kwargs):
        st = management_models.Status(status=status)
        return self.client.set_authentication_status(
            id=devid, aid=aid, status=st, **kwargs
        )

    def decommission_device(
        self, devid, headers={}, x_men_request_id="", authorization=""
    ):
        return self.client.decommission_device(
            id=devid,
            x_men_request_id=x_men_request_id,
        )

    def delete_authset(self, devid, aid, **kwargs):
        return self.client.remove_authentication(id=devid, aid=aid)

    def count_devices(self, status=None, **kwargs):
        count = self.client.count_devices(status=status, **kwargs)
        return count.count

    def make_auth(self, tenant_token):
        return {"Authorization": "Bearer " + tenant_token}


class SimpleManagementClient(ManagementClient):
    """Management API client. Cannot be used as pytest base class"""

    log = logging.getLogger("client.SimpleManagementClient")

    def __init__(self, hostname, swagger_spec):
        super().__init__(hostname, swagger_spec)

    def list_devices(self, **kwargs):
        return self.client.list_devices(**kwargs)

    def get_device_limit(self, **kwargs):
        return self.client.get_device_limit(**kwargs)

    def get_device(self, **kwargs):
        return self.client.get_device(**kwargs)

    def get_single_device(self, **kwargs):
        page = 1
        per_page = 100

        devs = self.list_devices(page=page, per_page=per_page, **kwargs)
        return devs[0]

    def find_device_by_identity(self, identity, **kwargs):
        page = 1
        per_page = 100
        self.log.debug("find device with identity: %s", identity)

        while True:
            self.log.debug("trying page %d", page)
            devs = self.list_devices(page=page, per_page=per_page, **kwargs)
            for dev in devs:
                if (
                    json.dumps({"mac": dev.identity_data.mac}, separators=(",", ":"))
                    == identity
                ):
                    # found
                    return dev
            # try another page
            if len(devs) < per_page:
                break
            page += 1

        return None

    def delete_token(self, **kwargs):
        return self.client.revoke_api_token(**kwargs)


class CliClient:
    exec_path = "/usr/bin/deviceauth"

    def __init__(self, service="deviceauth"):
        self.docker = docker.from_env()
        _self = self.docker.containers.list(filters={"id": socket.gethostname()})[0]

        project = _self.labels.get("com.docker.compose.project")
        self.device_auth = self.docker.containers.list(
            filters={
                "label": [
                    f"com.docker.compose.project={project}",
                    f"com.docker.compose.service={service}",
                ]
            },
            limit=1,
        )[0]

    def __call__(self, *args, **kwargs):
        cmd = [self.exec_path] + list(args)
        code, (stdout, stderr) = self.device_auth.exec_run(cmd, demux=True)
        return code, stdout, stderr

    def migrate(self, tenant=None, **kwargs):
        cmd = [self.exec_path, "migrate"]

        if tenant is not None:
            cmd.extend(["--tenant", tenant])

        code, (stdout, stderr) = self.device_auth.exec_run(cmd, demux=True, **kwargs)
        return code, stdout, stderr

    def check_device_limits(self, threshold=90.0, **kwargs):
        cmd = [self.exec_path, "check-device-limits", "--threshold", "%.2f" % threshold]

        code, (stdout, stderr) = self.device_auth.exec_run(cmd, demux=True, **kwargs)
        return code, stdout, stderr

    def list_tenants(self, tenant=None, **kwargs):
        cmd = [self.exec_path, "migrate", "--list-tenants"]

        code, (stdout, stderr) = self.device_auth.exec_run(cmd, demux=True, **kwargs)
        return code, stdout, stderr
