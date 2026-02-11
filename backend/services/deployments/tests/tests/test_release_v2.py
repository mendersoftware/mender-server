#!/usr/bin/python
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

import pytest

from client import DeploymentsClient
from client import management_v2_client
from common import (
    artifacts_added_from_data,
    artifacts_update_module_added_from_data,
    clean_db,
    clean_minio,
    s3_bucket,
    mongo,
    cli,
    Lock,
    MONGO_LOCK_FILE,
)

from config import pytest_config
import json

import mender_client
from bson.objectid import ObjectId
from base64 import b64encode
from datetime import datetime, timedelta
import hmac
import uuid


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


class TestRelease:
    d = DeploymentsClient()

    @pytest.mark.usefixtures("clean_minio")
    def test_get_all_releases(self, mongo, cli):
        with Lock(MONGO_LOCK_FILE) as l:
            cli.migrate()
            with artifacts_added_from_data(
                [
                    ("foo", "device-type-1"),
                    ("foo", "device-type-2"),
                    ("bar", "device-type-2"),
                ]
            ):
                release_name = "bar"
                for release_notes in [
                    "New Release security fixes 2023",
                    "New Release security fixes 2024",
                ]:
                    management_v2_client(
                        jwt=self.d.get_jwt()
                    ).update_release_information(
                        release_name=release_name,
                        release_update=mender_client.ReleaseUpdate(notes=release_notes),
                    )
                    release = management_v2_client(
                        jwt=self.d.get_jwt()
                    ).get_release_with_given_name(release_name=release_name)
                    assert release.notes == release_notes

                release = management_v2_client(
                    jwt=self.d.get_jwt()
                ).get_release_with_given_name(release_name="foo")
                assert release.notes == ""

                types = management_v2_client(jwt=self.d.get_jwt()).list_release_types()
                assert len(types) == 1
                assert types[0] == "rootfs-image"

    @pytest.mark.usefixtures("clean_minio")
    def test_get_all_releases_types(self, mongo, cli):
        with Lock(MONGO_LOCK_FILE) as l:
            cli.migrate()
            with artifacts_update_module_added_from_data(
                [
                    ("foo", "device-type-1", "app"),
                    ("foo", "device-type-2", "single-file"),
                    ("bar", "device-type-2", "directory"),
                ]
            ):
                types = management_v2_client(jwt=self.d.get_jwt()).list_release_types()
                assert len(types) > 0
                assert types == ["rootfs-image", "app", "single-file", "directory"]

    @pytest.mark.usefixtures("clean_db", "clean_minio")
    def test_get_all_releases_types_empty(self, mongo, cli):
        with Lock(MONGO_LOCK_FILE) as l:
            cli.migrate()
            types = management_v2_client(jwt=self.d.get_jwt()).list_release_types()
            assert len(types) == 0
