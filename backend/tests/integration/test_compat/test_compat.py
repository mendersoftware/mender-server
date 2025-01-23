# Copyright 2025 Northern.tech AS
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
import time

from testutils.api.client import ApiClient
from testutils.infra.cli import CliUseradm, CliDeviceauth
import testutils.api.deviceauth as deviceauth
import testutils.api.useradm as useradm
import testutils.api.deployments as deployments
import testutils.api.inventory as inventory

from testutils.common import (
    clean_mongo,
    create_user,
    mongo,
)


@pytest.fixture(scope="function")
def user(clean_mongo):
    deviceauth_cli = CliDeviceauth()
    useradm_cli = CliUseradm()

    deviceauth_cli.migrate()
    useradm_cli.migrate()
    yield create_user("user-foo@acme.com", "correcthorse")


class TestClientCompat:
    def test_compat(self, user):
        expected_client_versions = {
            "3.5.3",
            "4.0.6",
            "3.0.2",
            "3.4.0",
            "3.1.1",
            "2.3.2",
            "2.4.2",
            "3.5.0",
            "2.6.1",
            "3.2.1",
            "2.0.1",
            "2.5.4",
            "2.2.1",
            "2.1.3",
            "3.3.2",
            "5.0.0",
        }
        max_tries = 512
        devauthm = ApiClient(deviceauth.URL_MGMT)
        uadm = ApiClient(useradm.URL_MGMT)
        deploymentsm = ApiClient(deployments.URL_MGMT)

        r = uadm.call("POST", useradm.URL_LOGIN, auth=(user.name, user.pwd))
        assert r.status_code == 200
        utoken = r.text

        r = devauthm.with_auth(utoken).call("GET", deviceauth.URL_DEVICES_COUNT)
        assert r.status_code == 200

        devices_per_page = 64
        qs_params = {"page": 1, "per_page": devices_per_page, "status": "pending"}
        pending = []
        j = max_tries
        while j > 0:
            r = devauthm.with_auth(utoken).call(
                "GET", deviceauth.URL_MGMT_DEVICES, qs_params=qs_params
            )
            assert r.status_code == 200
            time.sleep(1)
            j = j - 1
            pending = r.json()
            if len(pending) >= len(expected_client_versions):
                break
        assert len(pending) == len(expected_client_versions)

        for i in range(len(pending)):
            device = pending[i]
            r = devauthm.with_auth(utoken).call(
                "PUT",
                deviceauth.URL_AUTHSET_STATUS,
                deviceauth.req_status("accepted"),
                path_params={"did": device["id"], "aid": device["auth_sets"][0]["id"]},
            )
            assert r.status_code == 204

        qs_params = {"page": 1, "per_page": devices_per_page, "status": "accepted"}
        accepted = []
        j = max_tries
        while j > 0:
            r = devauthm.with_auth(utoken).call(
                "GET", deviceauth.URL_MGMT_DEVICES, qs_params=qs_params
            )
            assert r.status_code == 200
            time.sleep(1)
            j = j - 1
            accepted = r.json()
            if len(accepted) >= len(expected_client_versions):
                break
        assert len(accepted) == len(expected_client_versions)

        invm = ApiClient(inventory.URL_MGMT)
        seen_versions = set()
        device_id_to_mender_version = {}
        devices_ids = []
        for i in range(len(accepted)):
            device = accepted[i]
            j = max_tries
            while j > 0:
                r = invm.with_auth(utoken).call(
                    "GET", inventory.URL_DEVICE, path_params={"id": device["id"]}
                )
                assert r.status_code == 200

                inv_dev = r.json()
                mender_client_version_attributes = [
                    a
                    for a in inv_dev["attributes"]
                    if a["name"] == "mender_client_version"
                ]
                if len(mender_client_version_attributes) > 0:
                    break
                j = j - 1
                time.sleep(1)
            mender_client_version = mender_client_version_attributes[0]["value"]
            seen_versions.add(mender_client_version)
            device_id_to_mender_version[device["id"]] = mender_client_version
            devices_ids.append(device["id"])

        artifact_file = "/tests/data/date.mender"
        missing_versions = expected_client_versions.difference(seen_versions)
        assert (
            not missing_versions
        ), f"did not observe the following expected versions: {missing_versions}; versions seen: {seen_versions}"

        rsp = deploymentsm.with_auth(utoken).call(
            "POST",
            deployments.URL_DEPLOYMENTS_ARTIFACTS,
            files={
                (
                    "artifact",
                    (
                        artifact_file,
                        open(artifact_file, "rb"),
                        "application/octet-stream",
                    ),
                ),
            },
        )
        assert rsp.status_code == 201

        rsp = deploymentsm.with_auth(utoken).call(
            "POST",
            deployments.URL_DEPLOYMENTS,
            body={
                "artifact_name": "date",
                "devices": devices_ids,
                "name": "test-compat-deployment",
            },
        )
        assert rsp.status_code == 201

        deployment_id = rsp.headers.get("Location").split("/")[-1]
        j = max_tries
        while j > 0:
            rsp = deploymentsm.with_auth(utoken).call(
                "GET", deployments.URL_DEPLOYMENTS_ID.format(id=deployment_id)
            )
            assert rsp.status_code == 200

            dpl = rsp.json()
            if dpl["status"] == "finished":
                rsp = deploymentsm.with_auth(utoken).call(
                    "GET",
                    deployments.URL_DEPLOYMENTS_STATISTICS.format(id=deployment_id),
                )
                assert rsp.status_code == 200
                assert rsp.json()["failure"] == 0
                assert rsp.json()["success"] == dpl["device_count"]
                break
            elif j <= 0:
                pytest.fail("timeout: Waiting for devices to update")
            else:
                time.sleep(1)
            j = j - 1

        for i in range(len(devices_ids)):
            r = devauthm.with_auth(utoken).call(
                "DELETE", deviceauth.URL_DEVICE, path_params={"id": devices_ids[i]}
            )
            assert r.status_code == 204
