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

import requests

from client import DeploymentsClient
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
                # this is a hack, since the swagger client is not prepared for the
                # specifications of API v2 in a separate file, and we are supposed
                # to move to openapi -- hence the fallback to requests.
                patch_release_url = (
                    "http://"
                    + pytest_config.getoption("host")
                    + f"/api/management/v2/deployments/deployments/releases/%s"
                )
                get_release_url = (
                    "http://"
                    + pytest_config.getoption("host")
                    + f"/api/management/v2/deployments/deployments/releases?name=%s"
                )
                get_types_url = (
                    "http://"
                    + pytest_config.getoption("host")
                    + f"/api/management/v2/deployments/releases/all/types"
                )
                release_name = "bar"
                for release_notes in [
                    "New Release security fixes 2023",
                    "New Release security fixes 2024",
                ]:
                    r = requests.patch(
                        patch_release_url % release_name,
                        verify=False,
                        headers={
                            "Authorization": "Bearer foo",
                            "Content-Type": "application/json",
                        },
                        data=json.dumps({"notes": release_notes}),
                    )
                    assert r.status_code == 204
                    r = requests.get(
                        get_release_url % release_name,
                        verify=False,
                        headers={"Authorization": "Bearer foo"},
                    )
                    releases = json.loads(r.text)
                    assert len(releases) > 0
                    assert releases[0]["notes"] == release_notes

                r = requests.get(
                    get_release_url % "foo",
                    verify=False,
                    headers={"Authorization": "Bearer foo"},
                )
                releases = json.loads(r.text)
                assert len(releases) > 0
                assert releases[0]["notes"] == ""

                r = requests.get(
                    get_types_url,
                    verify=False,
                    headers={"Authorization": "Bearer foo"},
                )
                types = json.loads(r.text)
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
                # this is a hack, since the swagger client is not prepared for the
                # specifications of API v2 in a separate file, and we are supposed
                # to move to openapi -- hence the fallback to requests.
                get_types_url = (
                    "http://"
                    + pytest_config.getoption("host")
                    + f"/api/management/v2/deployments/releases/all/types"
                )

                r = requests.get(
                    get_types_url,
                    verify=False,
                    headers={"Authorization": "Bearer foo"},
                )
                types = json.loads(r.text)
                assert len(types) > 0
                assert types == ["rootfs-image", "app", "single-file", "directory"]

    @pytest.mark.usefixtures("clean_db", "clean_minio")
    def test_get_all_releases_types(self, mongo, cli):
        with Lock(MONGO_LOCK_FILE) as l:
            cli.migrate()
            # this is a hack, since the swagger client is not prepared for the
            # specifications of API v2 in a separate file, and we are supposed
            # to move to openapi -- hence the fallback to requests.
            get_types_url = (
                "http://"
                + pytest_config.getoption("host")
                + f"/api/management/v2/deployments/releases/all/types"
            )

            r = requests.get(
                get_types_url, verify=False, headers={"Authorization": "Bearer foo"},
            )
            types = json.loads(r.text)
            assert types == []
