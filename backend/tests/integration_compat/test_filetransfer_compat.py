# Copyright 2026 Northern.tech AS
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
"""
Compatibility tests for the filetransfer protocol: the deviceconnect service
negotiates the filetransfer protocol version with mender-connect and must keep
supporting clients which only speak the original (v1) protocol, with the end to
end acknowledgements.
"""
import hashlib
import io
import os
import random
import string
import urllib.parse

import docker
import pytest
import redo

from testutils.api.client import ApiClient
from testutils.infra.cli import CliUseradm, CliDeviceauth
import testutils.api.deviceauth as deviceauth
import testutils.api.deviceconnect as deviceconnect
import testutils.api.useradm as useradm

from testutils.common import create_user

COMPOSE_PROJECT = "backend-tests"

# Compose service (see docker/docker-compose.backend-tests-compat.yml) running
# a mender-connect release that predates the simplified filetransfer protocol.
LEGACY_CLIENT_SERVICE = "mender-client-6.0.0"
LEGACY_CONNECT_VERSION = "3.0.0"

# With redo's exponential backoff (x1.5 from 1s) this waits ~75s in total,
# enough to cover the client's 30s authentication retry interval.
MAX_TRIES = 10


def random_filename():
    return "".join(random.choices(string.ascii_lowercase, k=8))


def find_client_container(service):
    client = docker.from_env()
    containers = client.containers.list(
        filters={
            "label": [
                f"com.docker.compose.project={COMPOSE_PROJECT}",
                f"com.docker.compose.service={service}",
            ]
        }
    )
    assert len(containers) == 1, f"expected exactly one container for {service}"
    return containers[0]


def accept_device_by_hostname(utoken, hostname):
    """Accept the pending device with the given identity hostname and
    return its device ID."""
    devauthm = ApiClient(deviceauth.URL_MGMT)
    for _ in redo.retrier(attempts=MAX_TRIES, sleeptime=1):
        r = devauthm.with_auth(utoken).call(
            "GET",
            deviceauth.URL_MGMT_DEVICES,
            qs_params={"page": 1, "per_page": 500, "status": "pending"},
        )
        assert r.status_code == 200
        device = next(
            (
                dev
                for dev in r.json()
                if dev["identity_data"].get("hostname") == hostname
            ),
            None,
        )
        if device is not None:
            break
    else:
        pytest.fail(f"timeout: waiting for device {hostname} to become pending")

    r = devauthm.with_auth(utoken).call(
        "PUT",
        deviceauth.URL_AUTHSET_STATUS,
        deviceauth.req_status("accepted"),
        path_params={"did": device["id"], "aid": device["auth_sets"][0]["id"]},
    )
    assert r.status_code == 204
    return device["id"]


def wait_for_connect(utoken, devid):
    devconnm = ApiClient(deviceconnect.URL_MGMT)
    for _ in redo.retrier(attempts=MAX_TRIES, sleeptime=1):
        r = devconnm.with_auth(utoken).call(
            "GET", deviceconnect.URL_MGMT_DEVICE, path_params={"id": devid}
        )
        if r.status_code == 200 and r.json()["status"] == "connected":
            return
    pytest.fail(f"timeout: waiting for device {devid} to connect")


def download_file(utoken, devid, path):
    devconnm = ApiClient(deviceconnect.URL_MGMT)
    return devconnm.with_auth(utoken).call(
        "GET",
        deviceconnect.URL_MGMT_FDOWNLOAD + "?path=" + urllib.parse.quote(path),
        path_params={"id": devid},
    )


def upload_file(utoken, devid, path, file, mode="600", uid="0", gid="0"):
    devconnm = ApiClient(deviceconnect.URL_MGMT)
    files = (
        ("path", (None, path)),
        ("mode", (None, mode)),
        ("uid", (None, uid)),
        ("gid", (None, gid)),
        ("file", (os.path.basename(path), file, "application/octet-stream")),
    )
    return devconnm.with_auth(utoken).call(
        "PUT",
        deviceconnect.URL_MGMT_FUPLOAD,
        path_params={"id": devid},
        files=files,
    )


@pytest.fixture(scope="module")
def legacy_device(mongo):
    """Clean the database, accept the device running a legacy mender-connect
    and wait for it to connect. Yields (user token, device ID, container)."""
    mongo.cleanup()
    CliDeviceauth().migrate()
    CliUseradm().migrate()
    user = create_user("user-filetransfer@acme.com", "correcthorse")

    r = ApiClient(useradm.URL_MGMT).call(
        "POST", useradm.URL_LOGIN, auth=(user.name, user.pwd)
    )
    assert r.status_code == 200
    utoken = r.text

    container = find_client_container(LEGACY_CLIENT_SERVICE)
    exit_code, output = container.exec_run(["mender-connect", "--version"])
    assert exit_code == 0
    # Make sure the client really runs the old version of mender-connect,
    # otherwise the test does not cover the legacy protocol.
    assert output.decode().split()[0] == LEGACY_CONNECT_VERSION

    devid = accept_device_by_hostname(utoken, container.attrs["Config"]["Hostname"])
    wait_for_connect(utoken, devid)

    yield utoken, devid, container

    ApiClient(deviceauth.URL_MGMT).with_auth(utoken).call(
        "DELETE", deviceauth.URL_DEVICE, path_params={"id": devid}
    )


class TestFileTransferCompat:
    def test_download_ok(self, legacy_device):
        utoken, devid, _ = legacy_device
        path = "/etc/mender/mender.conf"
        r = download_file(utoken, devid, path)

        assert r.status_code == 200, r.text
        assert b"ServerURL" in r.content
        assert (
            r.headers.get("Content-Disposition")
            == 'attachment; filename="' + os.path.basename(path) + '"'
        )
        assert r.headers.get("Content-Type") == "application/octet-stream"
        assert r.headers.get("X-Men-File-Path") == path
        assert r.headers.get("X-Men-File-Size") == str(len(r.content))

    def test_download_error_relative_path(self, legacy_device):
        utoken, devid, _ = legacy_device
        r = download_file(utoken, devid, "relative/path")
        assert r.status_code == 400, r.text
        assert r.json().get("error") == "path: must be absolute."

    def test_download_error_no_such_file_or_directory(self, legacy_device):
        utoken, devid, _ = legacy_device
        r = download_file(utoken, devid, "/does/not/exist")
        assert r.status_code == 400, r.text
        assert "/does/not/exist: no such file or directory" in r.json().get("error")

    def test_upload_and_download_ok(self, legacy_device):
        utoken, devid, container = legacy_device
        # 40MB of random data: big enough to span many chunks and exercise
        # the acknowledgement sliding window of the legacy protocol.
        content = os.urandom(40 * 1024 * 1024)
        uid = random.randint(100, 200)
        gid = random.randint(100, 200)
        fname = random_filename()
        path = f"/tmp/{fname}"

        r = upload_file(
            utoken,
            devid,
            path,
            io.BytesIO(content),
            mode="600",
            uid=str(uid),
            gid=str(gid),
        )
        assert r.status_code == 201, r.text

        # verify the file landed on the device intact
        exit_code, output = container.exec_run(["md5sum", path])
        assert exit_code == 0
        assert output.decode().split()[0] == hashlib.md5(content).hexdigest()

        r = download_file(utoken, devid, path)
        assert r.status_code == 200, r.text
        assert r.headers.get("Content-Disposition") == f'attachment; filename="{fname}"'
        assert r.headers.get("Content-Type") == "application/octet-stream"
        assert r.headers.get("X-Men-File-Mode") == "600"
        assert r.headers.get("X-Men-File-Uid") == str(uid)
        assert r.headers.get("X-Men-File-Gid") == str(gid)
        assert r.headers.get("X-Men-File-Path") == path
        assert r.headers.get("X-Men-File-Size") == str(len(content))
        assert hashlib.md5(r.content).hexdigest() == hashlib.md5(content).hexdigest()

        container.exec_run(["rm", "-f", path])

    def test_upload_error_path_is_relative(self, legacy_device):
        utoken, devid, _ = legacy_device
        r = upload_file(utoken, devid, "relative/path/dummy.txt", io.BytesIO(b"dummy"))
        assert r.status_code == 400, r.text
        assert r.json().get("error") == "path: must be absolute."

    def test_upload_error_file_does_not_exist(self, legacy_device):
        utoken, devid, _ = legacy_device
        r = upload_file(
            utoken, devid, "/does/not/exist/dummy.txt", io.BytesIO(b"dummy")
        )
        assert r.status_code == 400, r.text
        assert "failed to create target file" in r.json().get("error")
