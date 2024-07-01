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

import hmac
import json
import re
import uuid

from base64 import b64encode
from datetime import datetime, timedelta


def compare_expectations(expected, actual):
    if isinstance(expected, re.Pattern):
        assert isinstance(actual, str)
        assert bool(
            expected.match(actual)
        ), f"regular expression did not match value: {expected.pattern} != '{actual}'"
    else:
        if isinstance(expected, list):
            assert isinstance(actual, list)
            assert len(expected) == len(actual)
            for i in range(len(expected)):
                compare_expectations(expected[i], actual[i])
        elif isinstance(expected, dict):
            assert isinstance(actual, dict)
            for key, expected_value in expected.items():
                assert key in actual
                compare_expectations(expected_value, actual[key])
        else:
            assert expected == actual, f"{expected} != {actual}"
    return True


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
