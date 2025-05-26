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

import json
import os
import socket

from base64 import urlsafe_b64encode as b64encode
from datetime import datetime, timedelta
from typing import Union
from uuid import uuid4

import docker

from cryptography.hazmat.primitives import hashes, hmac

import openapi_client as oas


def generate_access_token(
    subject: str = "00000000-0000-0000-0000-000000000000",
    tenant_id: Union[str, None] = None,
    is_device: bool = False,
    key: Union[bytes, None] = None,
) -> str:
    """
    Generate a JSON Web Token (JWT) for testing purposes within Mender Server.

    Parameters:
    -----------
    subject : str, optional
        The subject claim/identifier of the JWT. Defaults to "00000000-0000-0000-0000-000000000000".
    tenant_id : str, optional
        Tenant identifier for Mender tenant claim in the JWT. Defaults to "000000000000000000000000".
    is_device : bool, optional
        Boolean flag indicating if the token is for a device (True) or a user (False). Defaults to False.
    key : bytes or None, optional
        Secret key used to sign the JWT. If not provided, a random key will be generated.

    Returns:
    -------
    str
        A signed JSON Web Token string containing various claims, including:
        subject, tenant_id, expiration time (exp), etc.
    """
    if key is None:
        key = os.urandom(32)
    h = hmac.HMAC(key, hashes.SHA256())
    iat = datetime.now()
    exp = iat + timedelta(days=7)
    hdr = {
        "alg": "HS256",
        "typ": "JWT",
    }
    claims = {
        "jti": str(uuid4()),
        "sub": subject,
        "exp": int(exp.timestamp()),
        "nbf": int(iat.timestamp()),
        "iat": int(iat.timestamp()),
        "mender.plan": "enterprise",
        "mender.trial": False,
        "mender.addons": [
            {"name": "troubleshoot", "enabled": True},
            {"name": "configure", "enabled": True},
            {"name": "monitor", "enabled": True},
        ],
    }
    if tenant_id is not None:
        claims["mender.tenant"] = tenant_id

    if is_device:
        claims["mender.device"] = True
    else:
        claims["mender.user"] = True
    jwt = f"{b64encode(json.dumps(hdr).encode()).decode('ascii')}.{b64encode(json.dumps(claims).encode()).decode('ascii')}"
    h.update(jwt.encode())
    jwt_signed = f"{jwt}.{b64encode(h.finalize()).decode('ascii')}".replace("=", "")
    return jwt_signed


def make_authenticated_client(
    tenant_id=None, is_device=False, subject="00000000-0000-0000-0000-000000000000"
) -> oas.ApiClient:
    cfg = oas.Configuration.get_default_copy()
    cfg.access_token = generate_access_token(
        tenant_id=tenant_id, subject=subject, is_device=is_device
    )
    client = oas.ApiClient(cfg)
    return client


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
