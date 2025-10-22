# Copyright 2021 Northern.tech AS
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

import base64
import json
import uuid
from datetime import datetime, timedelta, timezone

import devices_v1 as devices_api
import internal_v1 as internal_api
import management_v1 as management_api


def make_user_token(user_id=None, plan=None, tenant_id=None):
    if user_id is None:
        user_id = str(uuid.uuid4())
    claims = {
        "jti": str(uuid.uuid4()),
        "sub": user_id,
        "exp": int((datetime.now(tz=timezone.utc) + timedelta(days=7)).timestamp()),
        "mender.user": True,
    }
    if tenant_id is not None:
        claims["mender.tenant"] = tenant_id
    if plan is not None:
        claims["mender.plan"] = plan

    return ".".join(
        [
            base64.urlsafe_b64encode(b'{"alg":"RS256","typ":"JWT"}')
            .decode("ascii")
            .strip("="),
            base64.urlsafe_b64encode(json.dumps(claims).encode())
            .decode("ascii")
            .strip("="),
            base64.urlsafe_b64encode(b"Signature").decode("ascii").strip("="),
        ]
    )


def make_device_token(device_id=None, plan=None, tenant_id=None):
    if device_id is None:
        device_id = str(uuid.uuid4())
    claims = {
        "jti": str(uuid.uuid4()),
        "sub": device_id,
        "exp": int((datetime.now(tz=timezone.utc) + timedelta(days=7)).timestamp()),
        "mender.device": True,
    }
    if tenant_id is not None:
        claims["mender.tenant"] = tenant_id
    if plan is not None:
        claims["mender.plan"] = plan

    return ".".join(
        [
            base64.urlsafe_b64encode(b'{"alg":"RS256","typ":"JWT"}')
            .decode("ascii")
            .strip("="),
            base64.urlsafe_b64encode(json.dumps(claims).encode())
            .decode("ascii")
            .strip("="),
            base64.urlsafe_b64encode(b"Signature").decode("ascii").strip("="),
        ]
    )


class ManagementAPIClientWrapper:
    def __init__(self, client):
        self._client = client

    def get_device_configuration(self, device_id, **kwargs):
        return self._client.device_config_management_get_device_configuration(device_id, **kwargs)

    def set_device_configuration(self, device_id, **kwargs):
        return self._client.device_config_management_set_device_configuration(device_id, **kwargs)

    def deploy_device_configuration(self, device_id, **kwargs):
        return self._client.device_config_management_deploy_device_configuration(device_id, **kwargs)


class DeviceAPIClientWrapper:
    def __init__(self, client):
        self._client = client

    def get_device_configuration(self, **kwargs):
        return self._client.device_config_get_device_configuration(**kwargs)

    def report_device_configuration(self, **kwargs):
        return self._client.device_config_report_device_configuration(**kwargs)


class InternalAPIClientWrapper:
    def __init__(self, client=None):
        if client is None:
            client = internal_api.InternalAPIClient()
        self._client = client

    def provision_device_with_http_info(self, **kwargs):
        return self._client.device_config_internal_provision_device_with_http_info(**kwargs)

    def decommission_device_with_http_info(self, **kwargs):
        return self._client.device_config_internal_decommission_device_with_http_info(**kwargs)


InternalAPIClient = InternalAPIClientWrapper


def management_api_with_params(user_id, plan=None, tenant_id=None):
    api_conf = management_api.Configuration.get_default_copy()
    api_conf.access_token = make_user_token(user_id, plan, tenant_id)
    client = management_api.ManagementAPIClient(management_api.ApiClient(api_conf))
    return ManagementAPIClientWrapper(client)


def devices_api_with_params(device_id, plan=None, tenant_id=None):
    api_conf = devices_api.Configuration.get_default_copy()
    api_conf.access_token = make_device_token(device_id, plan, tenant_id)
    client = devices_api.DeviceAPIClient(devices_api.ApiClient(api_conf))
    return DeviceAPIClientWrapper(client)
