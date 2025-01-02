# Copyright 2024 Northern.tech AS
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
import logging
import pytest
import random
import time
import base64
import json
import uuid
import requests
from os import path

from testutils.api.client import ApiClient
from testutils.infra.cli import CliUseradm, CliDeviceauth
from testutils.infra.mongo import MongoClient
from testutils.infra.container_manager.kubernetes_manager import isK8S
import testutils.api.deviceauth as deviceauth
import testutils.api.useradm as useradm
import testutils.api.tenantadm as tenantadm
import testutils.api.deployments as deployments
import testutils.api.inventory as inventory
import testutils.util.crypto as crypto

from testutils.common import (
    Device,
    Authset,
    clean_mongo,
    mongo_cleanup,
    mongo,
    create_user,
    create_org,
    create_random_authset,
    create_authset,
    get_device_by_id_data,
    change_authset_status,
    wait_until_healthy,
    useExistingTenant,
)

@pytest.fixture(scope="function")
def clean_mongo_client(mongo):
    """Fixture setting up a clean (i.e. empty database). Yields
    common.MongoClient connected to the DB.

    Useful for tests with multiple testcases:
    - protects the whole test func as usual
    - but also allows calling MongoClient.cleanup() between cases
    """
    mongo_cleanup(mongo)
    yield mongo
    mongo_cleanup(mongo)

@pytest.fixture(scope="function")
def clean_migrated_mongo(clean_mongo):
    deviceauth_cli = CliDeviceauth()
    useradm_cli = CliUseradm()

    deviceauth_cli.migrate()
    useradm_cli.migrate()

    yield clean_mongo


@pytest.fixture(scope="function")
def clean_migrated_mongo_mt(clean_mongo):
    deviceauth_cli = CliDeviceauth()
    useradm_cli = CliUseradm()
    for t in ["tenant1", "tenant2"]:
        deviceauth_cli.migrate(t)
        useradm_cli.migrate(t)

    yield clean_mongo


@pytest.fixture(scope="function")
def user(clean_migrated_mongo):
    yield create_user("user-foo@acme.com", "correcthorse")


@pytest.fixture(scope="function")
def devices(clean_migrated_mongo, user):
    useradm_mgmt_v1 = ApiClient(useradm.URL_MGMT)
    devauthm = ApiClient(deviceauth.URL_MGMT)
    devauthd = ApiClient(deviceauth.URL_DEVICES)

    r = useradm_mgmt_v1.call("POST", useradm.URL_LOGIN, auth=(user.name, user.pwd))
    assert r.status_code == 200
    utoken = r.text

    devices = []

    for _ in range(5):
        aset = create_random_authset(devauthd, devauthm, utoken)
        dev = Device(aset.did, aset.id_data, aset.pubkey)
        devices.append(dev)

    yield devices


@pytest.fixture(scope="function")
def tenants_user(clean_migrated_mongo_mt):
    uuidv4 = str(uuid.uuid4())
    tenant, username, password = (
        "test.mender.io-" + uuidv4,
        "some.user+" + uuidv4 + "@example.com",
        "secretsecret",
    )
    tenant = create_org(tenant, username, password)
    yield tenant


@pytest.fixture(scope="function")
def tenants_users_devices(clean_migrated_mongo_mt, tenants_users):
    useradm_mgmt_v1 = ApiClient(useradm.URL_MGMT)
    devauthm = ApiClient(deviceauth.URL_MGMT)
    devauthd = ApiClient(deviceauth.URL_DEVICES)

    for t in tenants_users:
        user = t.users[0]
        r = useradm_mgmt_v1.call("POST", useradm.URL_LOGIN, auth=(user.name, user.pwd))
        assert r.status_code == 200
        utoken = r.text

        for _ in range(5):
            aset = create_random_authset(devauthd, devauthm, utoken, t.tenant_token)
            dev = Device(aset.did, aset.id_data, aset.pubkey, t.tenant_token)
            dev.status = aset.status
            t.devices.append(dev)

    yield tenants_users


def make_devs_with_authsets(user, tenant_token=""):
    """ create a good number of devices, some with >1 authsets, with different statuses.
        returns DevWithAuthsets objects."""
    useradmm = ApiClient(useradm.URL_MGMT)

    # log in user
    r = useradmm.call("POST", useradm.URL_LOGIN, auth=(user.name, user.pwd))
    assert r.status_code == 200

    utoken = r.text

    devices = []

    def keygen_rsa():
        return crypto.get_keypair_rsa()

    def keygen_ec_256():
        return crypto.get_keypair_ec(crypto.EC_CURVE_256)

    def keygen_ed():
        return crypto.get_keypair_ed()

    # some vanilla 'pending' devices, single authset
    for _ in range(3):
        dev = make_pending_device(utoken, keygen_rsa, 1, tenant_token=tenant_token)
        devices.append(dev)

    for _ in range(2):
        dev = make_pending_device(utoken, keygen_ec_256, 1, tenant_token=tenant_token)
        devices.append(dev)

    dev = make_pending_device(utoken, keygen_ed, 1, tenant_token=tenant_token)
    devices.append(dev)

    # some pending devices with > 1 authsets
    for i in range(2):
        dev = make_pending_device(utoken, keygen_rsa, 3, tenant_token=tenant_token)
        devices.append(dev)

    for i in range(2):
        dev = make_pending_device(utoken, keygen_ec_256, 3, tenant_token=tenant_token)
        devices.append(dev)

    dev = make_pending_device(utoken, keygen_ed, 3, tenant_token=tenant_token)
    devices.append(dev)

    # some 'accepted' devices, single authset
    for _ in range(3):
        dev = make_accepted_device_with_multiple_authsets(
            utoken, keygen_rsa, 1, tenant_token=tenant_token
        )
        devices.append(dev)

    for _ in range(2):
        dev = make_accepted_device_with_multiple_authsets(
            utoken, keygen_ec_256, 1, tenant_token=tenant_token
        )
        devices.append(dev)

    dev = make_accepted_device_with_multiple_authsets(
        utoken, keygen_ed, 1, tenant_token=tenant_token
    )
    devices.append(dev)

    # some 'accepted' devices with >1 authsets
    for _ in range(2):
        dev = make_accepted_device_with_multiple_authsets(
            utoken, keygen_rsa, 3, tenant_token=tenant_token
        )
        devices.append(dev)

    for _ in range(2):
        dev = make_accepted_device_with_multiple_authsets(
            utoken, keygen_ec_256, 2, tenant_token=tenant_token
        )
        devices.append(dev)

    dev = make_accepted_device_with_multiple_authsets(
        utoken, keygen_ed, 2, tenant_token=tenant_token
    )
    devices.append(dev)

    # some rejected devices
    for _ in range(2):
        dev = make_rejected_device(utoken, keygen_rsa, 3, tenant_token=tenant_token)
        devices.append(dev)

    for _ in range(2):
        dev = make_rejected_device(utoken, keygen_ec_256, 2, tenant_token=tenant_token)
        devices.append(dev)

    dev = make_rejected_device(utoken, keygen_ed, 2, tenant_token=tenant_token)
    devices.append(dev)

    # preauth'd devices
    dev = make_preauthd_device(utoken, keygen_rsa)
    devices.append(dev)

    dev = make_preauthd_device(utoken, keygen_ec_256)
    devices.append(dev)

    dev = make_preauthd_device(utoken, keygen_ed)
    devices.append(dev)

    # preauth'd devices with extra 'pending' sets
    for i in range(2):
        dev = make_preauthd_device_with_pending(
            utoken, keygen_rsa, num_pending=2, tenant_token=tenant_token
        )
        devices.append(dev)

    dev = make_preauthd_device_with_pending(
        utoken, keygen_ec_256, num_pending=2, tenant_token=tenant_token
    )
    devices.append(dev)

    dev = make_preauthd_device_with_pending(
        utoken, keygen_ed, num_pending=2, tenant_token=tenant_token
    )
    devices.append(dev)

    devices.sort(key=lambda dev: dev.id)
    devices.sort(key=lambda dev: dev.status)
    return devices


@pytest.fixture(scope="function")
def devs_authsets(user):
    yield make_devs_with_authsets(user)


@pytest.fixture(scope="function")
def tenants_devs_authsets(tenants_users):
    for t in tenants_users:
        devs = make_devs_with_authsets(t.users[0], t.tenant_token)
        t.devices = devs

    yield tenants_users


def rand_id_data():
    mac = ":".join(["{:02x}".format(random.randint(0x00, 0xFF), "x") for i in range(6)])
    sn = "".join(["{}".format(random.randint(0x00, 0xFF)) for i in range(6)])

    return {"mac": mac, "sn": sn}


def make_pending_device(utoken, keygen, num_auth_sets=1, tenant_token=""):
    devauthm = ApiClient(deviceauth.URL_MGMT)
    devauthd = ApiClient(deviceauth.URL_DEVICES)

    id_data = rand_id_data()

    dev = None
    for i in range(num_auth_sets):
        priv, pub = keygen()
        new_set = create_authset(
            devauthd, devauthm, id_data, pub, priv, utoken, tenant_token=tenant_token
        )

        if dev is None:
            dev = Device(new_set.did, new_set.id_data, utoken, tenant_token)

        dev.authsets.append(new_set)

    dev.status = "pending"

    return dev


def make_accepted_device_with_multiple_authsets(
    utoken, keygen, num_auth_sets=1, num_accepted=1, tenant_token=""
):
    devauthm = ApiClient(deviceauth.URL_MGMT)

    dev = make_pending_device(utoken, keygen, num_auth_sets, tenant_token=tenant_token)

    for i in range(num_accepted):
        aset_id = dev.authsets[i].id
        change_authset_status(devauthm, dev.id, aset_id, "accepted", utoken)

        dev.authsets[i].status = "accepted"

    dev.status = "accepted"

    return dev


def make_rejected_device(utoken, keygen, num_auth_sets=1, tenant_token=""):
    devauthm = ApiClient(deviceauth.URL_MGMT)

    dev = make_pending_device(utoken, keygen, num_auth_sets, tenant_token=tenant_token)

    for i in range(num_auth_sets):
        aset_id = dev.authsets[i].id
        change_authset_status(devauthm, dev.id, aset_id, "rejected", utoken)

        dev.authsets[i].status = "rejected"

    dev.status = "rejected"

    return dev


def make_preauthd_device(utoken, keygen):
    devauthm = ApiClient(deviceauth.URL_MGMT)

    priv, pub = keygen()
    id_data = rand_id_data()

    body = deviceauth.preauth_req(id_data, pub)
    r = devauthm.with_auth(utoken).call("POST", deviceauth.URL_MGMT_DEVICES, body)
    assert r.status_code == 201

    api_dev = get_device_by_id_data(devauthm, id_data, utoken)
    assert len(api_dev["auth_sets"]) == 1
    aset = api_dev["auth_sets"][0]

    dev = Device(api_dev["id"], id_data, pub)
    dev.authsets.append(
        Authset(aset["id"], dev.id, id_data, pub, priv, "preauthorized")
    )

    dev.status = "preauthorized"

    return dev


def make_preauthd_device_with_pending(utoken, keygen, num_pending=1, tenant_token=""):
    devauthm = ApiClient(deviceauth.URL_MGMT)
    devauthd = ApiClient(deviceauth.URL_DEVICES)

    dev = make_preauthd_device(utoken, keygen)

    for i in range(num_pending):
        priv, pub = crypto.get_keypair_rsa()
        aset = create_authset(
            devauthd,
            devauthm,
            dev.id_data,
            pub,
            priv,
            utoken,
            tenant_token=tenant_token,
        )
        dev.authsets.append(
            Authset(aset.id, aset.did, dev.id_data, pub, priv, "pending")
        )

    return dev


def filter_and_page_devs(devs, page=None, per_page=None, status=None):
    if status is not None:
        devs = [d for d in devs if d.status == status]

    if page is None:
        page = 1

    if per_page is None:
        per_page = 20

    lo = (page - 1) * per_page
    hi = lo + per_page

    return devs[lo:hi]


def compare_aset(authset, api_authset):
    assert authset.id == api_authset["id"]
    assert authset.id_data == api_authset["identity_data"]
    assert crypto.compare_keys(authset.pubkey, api_authset["pubkey"])
    assert authset.status == api_authset["status"]


@pytest.fixture(scope="function")
def tenants_with_plans(clean_mongo):
    uuidv4 = str(uuid.uuid4())
    tenant, username, password = (
        "test.mender.io-" + uuidv4,
        "some.user+" + uuidv4 + "@example.com",
        "secretsecret",
    )
    tos = create_org(tenant, username, password, plan="os")
    tos.plan = "os"
    #
    uuidv4 = str(uuid.uuid4())
    tenant, username, password = (
        "test.mender.io-" + uuidv4,
        "some.user+" + uuidv4 + "@example.com",
        "secretsecret",
    )
    tpro = create_org(tenant, username, password, plan="professional")
    tpro.plan = "professional"
    #
    uuidv4 = str(uuid.uuid4())
    tenant, username, password = (
        "test.mender.io-" + uuidv4,
        "some.user+" + uuidv4 + "@example.com",
        "secretsecret",
    )
    tent = create_org(tenant, username, password, plan="enterprise")
    tent.plan = "enterprise"

    return [tos, tpro, tent]


class TestClientCompat:
    def test_compat(self, user):
        expected_client_versions = [
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
        ]
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
        seen_versions = {}
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
            seen_versions[mender_client_version] = True
            device_id_to_mender_version[device["id"]] = mender_client_version
            devices_ids.append(device["id"])

        artifact_file = "tests/data/date.mender"
        for i in range(len(expected_client_versions)):
            assert seen_versions[expected_client_versions[i]]

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

        m = MongoClient("mender-mongo:27017")
        m.cleanup()
