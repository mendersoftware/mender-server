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
import re
from contextlib import contextmanager
from datetime import datetime, timedelta, timezone

from websocket import create_connection

import devices_v1 as devices_api
import internal_v1 as internal_api
import management_v1 as management_api


@contextmanager
def ws_session(url, **kwargs):
    conn = create_connection(url, **kwargs)
    yield conn
    conn.close()


class Device:
    def __init__(self, device_id=None, plan=None, tenant_id=None):
        if device_id is None:
            device_id = str(uuid.uuid4())
        self.id = device_id
        self.tenant_id = tenant_id
        if tenant_id is None:
            tenant_id = ""
        self.plan = plan

        client = internal_api.InternalAPIClient()
        r = client.device_connect_internal_provision_device_with_http_info(
            tenant_id=tenant_id,
            provision_device=internal_api.ProvisionDevice(device_id=device_id),
        )
        assert r.status_code == 201

    def connect(self):
        return ws_session(
            devices_api.Configuration.get_default_copy().host.replace(
                "http://", "ws://"
            )
            + "/api/devices/v1/deviceconnect/connect",
            cookie="JWT=%s" % self.jwt,
        )

    @property
    def jwt(self):
        claims = {
            "jti": str(uuid.uuid4()),
            "sub": self.id,
            "exp": int((datetime.now(tz=timezone.utc) + timedelta(days=7)).timestamp()),
            "mender.device": True,
        }
        if self.tenant_id is not None:
            claims["mender.tenant"] = self.tenant_id

        if self.plan is not None:
            claims["mender.plan"] = self.plan

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

    @property
    def api(self):
        # Setup device api with token
        api_conf = devices_api.Configuration.get_default_copy()
        api_conf.access_token = self.jwt
        return devices_api.DeviceAPIClient(devices_api.ApiClient(api_conf))


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


class ManagementAPIClientWrapper:
    def __init__(self, client):
        self._client = client

    def get_device(self, id, **kwargs):
        return self._client.device_connect_management_get_device(id, **kwargs)

    def check_update(self, id, **kwargs):
        return self._client.device_connect_management_check_update(id, **kwargs)

    def connect(self, id, **kwargs):
        return self._client.device_connect_management_connect(id, **kwargs)

    def download(self, id, path, **kwargs):
        return self._client.device_connect_management_download(id, path, **kwargs)

    def send_inventory(self, id, **kwargs):
        return self._client.device_connect_management_send_inventory(id, **kwargs)

    def playback(self, session_id, **kwargs):
        return self._client.device_connect_management_playback(session_id, **kwargs)

    def upload(self, id, **kwargs):
        return self._client.device_connect_management_upload(id, **kwargs)


def management_api_with_params(user_id, plan=None, tenant_id=None):
    api_conf = management_api.Configuration.get_default_copy()
    api_conf.access_token = make_user_token(user_id, plan, tenant_id)
    client = management_api.ManagementAPIClient(management_api.ApiClient(api_conf))
    return ManagementAPIClientWrapper(client)


def management_api_connect(
    device_id: str,
    user_id: str = None,
    tenant_id: str = None,
    plan: str = None,
    api_conf: management_api.Configuration = None,
    **sess_args,
):
    if api_conf is None:
        api_conf = management_api.Configuration.get_default_copy()
    jwt = make_user_token(user_id=user_id, tenant_id=tenant_id, plan=plan)
    url = (
        re.sub(r"^http(s?://.+$)", r"ws\1", api_conf.host).rstrip("/")
        + f"/api/management/v1/deviceconnect/devices/{device_id}/connect"
    )
    return ws_session(url, cookie=f"JWT={jwt}", **sess_args)
