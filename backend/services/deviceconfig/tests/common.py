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

import mender_client
from mender_client.api import (
    DeviceConfigureDeviceAPIApi,
    DeviceConfigureInternalAPIApi,
    DeviceConfigureManagementAPIApi,
)


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

    @property
    def configuration(self):
        return self._client.api_client.configuration

    def call_api(self, *args, **kwargs):
        return self._client.api_client.call_api(*args, **kwargs)

    def get_device_configuration(self, device_id, **kwargs):
        return self._client.device_config_management_get_device_configuration(
            device_id, **kwargs
        )

    def get_device_configuration_with_http_info(self, device_id, **kwargs):
        return self._client.device_config_management_get_device_configuration_with_http_info(
            device_id, **kwargs
        )

    def set_device_configuration(self, device_id, **kwargs):
        return self._client.device_config_management_set_device_configuration(
            device_id, **kwargs
        )

    def set_device_configuration_with_http_info(self, device_id, **kwargs):
        return self._client.device_config_management_set_device_configuration_with_http_info(
            device_id, **kwargs
        )

    def deploy_device_configuration(self, device_id, **kwargs):
        return self._client.device_config_management_deploy_device_configuration(
            device_id, **kwargs
        )

    def deploy_device_configuration_with_http_info(self, device_id, **kwargs):
        return self._client.device_config_management_deploy_device_configuration_with_http_info(
            device_id, **kwargs
        )


class DeviceAPIClientWrapper:
    def __init__(self, client):
        self._client = client

    @property
    def configuration(self):
        return self._client.api_client.configuration

    def call_api(self, *args, **kwargs):
        return self._client.api_client.call_api(*args, **kwargs)

    def get_device_configuration(self, **kwargs):
        return self._client.device_config_get_device_configuration(**kwargs)

    def get_device_configuration_with_http_info(self, **kwargs):
        return self._client.device_config_get_device_configuration_with_http_info(**kwargs)

    def report_device_configuration(self, **kwargs):
        return self._client.device_config_report_device_configuration(**kwargs)

    def report_device_configuration_with_http_info(self, **kwargs):
        return self._client.device_config_report_device_configuration_with_http_info(**kwargs)


class InternalAPIClientWrapper:
    def __init__(self, client=None):
        if client is None:
            client = DeviceConfigureInternalAPIApi()
        self._client = client

    def provision_device_with_http_info(self, tenant_id, new_device, **kwargs):
        return self._client.device_config_internal_provision_device_with_http_info(
            tenant_id, provision_device=new_device, **kwargs
        )

    def decommission_device_with_http_info(self, tenant_id, device_id, **kwargs):
        return self._client.device_config_internal_decommission_device_with_http_info(
            tenant_id, device_id, **kwargs
        )


InternalAPIClient = InternalAPIClientWrapper


def management_api_with_params(user_id, plan=None, tenant_id=None):
    api_conf = mender_client.Configuration.get_default_copy()
    api_conf.access_token = make_user_token(user_id, plan, tenant_id)
    client = DeviceConfigureManagementAPIApi(mender_client.ApiClient(api_conf))
    return ManagementAPIClientWrapper(client)


def devices_api_with_params(device_id, plan=None, tenant_id=None):
    api_conf = mender_client.Configuration.get_default_copy()
    api_conf.access_token = make_device_token(device_id, plan, tenant_id)
    client = DeviceConfigureDeviceAPIApi(mender_client.ApiClient(api_conf))
    return DeviceAPIClientWrapper(client)


def management_api_set_config_raw(user_id, tenant_id, device_id, configuration):
    token = make_user_token(user_id, tenant_id=tenant_id)
    api_conf = mender_client.Configuration.get_default_copy()
    api_conf.access_token = token
    api_client = DeviceConfigureManagementAPIApi(mender_client.ApiClient(api_conf))
    url = f"{api_client.api_client.configuration.host}/api/management/v1/deviceconfig/configurations/device/{device_id}"
    return api_client.api_client.call_api(
        "PUT",
        url,
        header_params={
            "Content-Type": "application/json",
            "Authorization": f"Bearer {token}",
        },
        body=configuration,
    )


def devices_api_report_config_raw(device_id, tenant_id, configuration):
    token = make_device_token(device_id, tenant_id=tenant_id)
    api_conf = mender_client.Configuration.get_default_copy()
    api_conf.access_token = token
    api_client = DeviceConfigureDeviceAPIApi(mender_client.ApiClient(api_conf))
    url = f"{api_client.api_client.configuration.host}/api/devices/v1/deviceconfig/configuration"
    return api_client.api_client.call_api(
        "PUT",
        url,
        header_params={
            "Content-Type": "application/json",
            "Authorization": f"Bearer {token}",
        },
        body=configuration,
    )
