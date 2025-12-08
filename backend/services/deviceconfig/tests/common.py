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

import devices_api
import management_api


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


def management_api_with_params(user_id, plan=None, tenant_id=None):
    api_conf = management_api.Configuration.get_default_copy()
    api_conf.access_token = make_user_token(user_id, plan, tenant_id)
    return management_api.ManagementAPIClient(management_api.ApiClient(api_conf))


def devices_api_with_params(device_id, plan=None, tenant_id=None):
    api_conf = devices_api.Configuration.get_default_copy()
    api_conf.access_token = make_device_token(device_id, plan, tenant_id)
    return devices_api.DeviceAPIClient(devices_api.ApiClient(api_conf))


def management_api_set_config_raw(user_id, tenant_id, device_id, configuration):
    token = make_user_token(user_id, tenant_id=tenant_id)
    api_conf = management_api.Configuration.get_default_copy()
    api_conf.access_token = token
    api_client = management_api.ApiClient(api_conf)
    url = f"{api_client.configuration.host}/configurations/device/{device_id}"
    return api_client.call_api(
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
    api_conf = devices_api.Configuration.get_default_copy()
    api_conf.access_token = token
    api_client = devices_api.ApiClient(api_conf)
    url = f"{api_client.configuration.host}/configuration"
    return api_client.call_api(
        "PUT",
        url,
        header_params={
            "Content-Type": "application/json",
            "Authorization": f"Bearer {token}",
        },
        body=configuration,
    )
