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

import hmac
import json
import os
import pytest
import re
import uuid

from base64 import b64encode
from datetime import datetime, timedelta
from typing import Union
from urllib.parse import urljoin

import requests

from internal_api.models import Device, DeviceAttribute
from management_api.models import Deployment

type_decoder = {
    str: "str",
    int: "num",
    float: "num",
    bool: "bool",
}


def index_deployment(tenant_id: str, deployment: Deployment):
    requests.post(
        "http://mender-workflows-server:8080/api/v1/workflow/reindex_reporting_deployment",
        json={
            "action": "reindex",
            "request_id": "req",
            "tenant_id": tenant_id,
            "id": deployment.id,
            "device_id": deployment.device_id,
            "deployment_id": deployment.deployment_id,
            "service": "inventory",
        },
    )


def index_device(tenant_id: str, device: Device):
    requests.post(
        "http://mender-workflows-server:8080/api/v1/workflow/reindex_reporting",
        json={
            "action": "reindex",
            "request_id": "req",
            "tenant_id": tenant_id,
            "device_id": device.id,
            "service": "inventory",
        },
    )


def attributes_to_document(attrs: list[DeviceAttribute]) -> dict[str, object]:
    doc = {}
    if attrs is not None:
        for attr in attrs:
            typ = type_decoder[type(attr.value)]
            doc[f"{attr.scope}_{attr.name}_{typ}"] = [attr.value]
    return doc


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
