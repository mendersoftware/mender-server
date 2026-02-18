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

import time
import logging
import pytest
import uuid
import redo

from testutils.api.client import ApiClient
import testutils.api.useradm as useradm
import testutils.api.deviceauth as deviceauth
import testutils.api.deployments as deployments
import testutils.api.inventory as inventory
from testutils.infra.cli import CliTenantadm, CliUseradm, CliDeviceauth
from testutils.common import (
    Device,
    create_org,
    create_random_authset,
    change_authset_status,
    create_user,
    useExistingTenant,
)

logging.basicConfig(format="%(asctime)s %(message)s")
logger = logging.getLogger("test_decomission")
logger.setLevel(logging.INFO)

@pytest.fixture(scope="function")
def clean_migrated_mongo(clean_mongo):
    deviceauth_cli = CliDeviceauth()
    useradm_cli = CliUseradm()

    deviceauth_cli.migrate()
    useradm_cli.migrate()

    yield clean_mongo

@pytest.fixture(scope="function")
def user(clean_migrated_mongo):
    yield create_user("user-foo@acme.com", "correcthorse")

@pytest.fixture(scope="function")
def devices(clean_migrated_mongo, user):
    useradmm = ApiClient(useradm.URL_MGMT)
    devauthm = ApiClient(deviceauth.URL_MGMT)
    devauthd = ApiClient(deviceauth.URL_DEVICES)

    r = useradmm.call("POST", useradm.URL_LOGIN, auth=(user.name, user.pwd))
    assert r.status_code == 200
    utoken = r.text

    devices = []

    for _ in range(2):
        aset = create_random_authset(devauthd, devauthm, utoken)
        dev = Device(aset.did, aset.id_data, aset.pubkey)
        dev.authsets.append(aset)
        devices.append(dev)

    yield devices

class TestDeviceDecomissioningBase:
    def do_test_ok(self, user, device, tenant_token=None):
        devauthd = ApiClient(deviceauth.URL_DEVICES)
        devauthm = ApiClient(deviceauth.URL_MGMT)
        useradmm = ApiClient(useradm.URL_MGMT)
        deploymentsd = ApiClient(deployments.URL_DEVICES)
        inventoryd = ApiClient(inventory.URL_DEV)
        inventorym = ApiClient(inventory.URL_MGMT)

        r = useradmm.call("POST", useradm.URL_LOGIN, auth=(user.name, user.pwd))
        assert r.status_code == 200
        utoken = r.text

        aset = device.authsets[0]
        change_authset_status(devauthm, aset.did, aset.id, "accepted", utoken)

        # request auth
        body, sighdr = deviceauth.auth_req(
            aset.id_data, aset.pubkey, aset.privkey, tenant_token
        )

        r = devauthd.call("POST", deviceauth.URL_AUTH_REQS, body, headers=sighdr)
        assert r.status_code == 200
        dtoken = r.text

        # wait for the device provisioning workflow to do its job
        for _ in redo.retrier(attempts=60, sleeptime=1):
            r = inventorym.with_auth(utoken).call(
                "GET", inventory.URL_DEVICE, path_params={"id": aset.did}
            )
            if r.status_code == 200:
                break
        else:
            assert False, "device not added to the inventory"

        # check if the device can access API by patching device inventory
        payload = [{"name": "mac", "value": "foo"}]
        r = inventoryd.with_auth(dtoken).call(
            "PATCH", inventory.URL_DEVICE_ATTRIBUTES, payload
        )
        assert r.status_code == 200

        # decommission
        r = devauthm.with_auth(utoken).call(
            "DELETE", deviceauth.URL_DEVICE.format(id=aset.did)
        )

        # check device is rejected
        r = deploymentsd.with_auth(dtoken).call(
            "GET",
            deployments.URL_NEXT,
            qs_params={"device_type": "foo", "artifact_name": "bar"},
        )
        assert r.status_code == 401

        # check device gone from inventory
        # this may take some time because it's done as an async job (workflow)
        timeout = 60 * 3
        for _ in redo.retrier(attempts=timeout, sleeptime=1):
            r = inventorym.with_auth(utoken).call(
                "GET", inventory.URL_DEVICE, path_params={"id": aset.did}
            )
            if r.status_code == 404:
                break
        else:
            assert False, "device not removed from the inventory"

        # check device gone from deviceauth
        timeout = 60
        for _ in redo.retrier(attempts=timeout, sleeptime=1):
            r = devauthm.with_auth(utoken).call(
                "GET", deviceauth.URL_DEVICE.format(id=aset.did)
            )
            if r.status_code == 404:
                break
        else:
            assert False, "device not removed from the deviceauth"

class TestDeviceDecomissioning(TestDeviceDecomissioningBase):
    def test_ok(self, user, devices):
        self.do_test_ok(user, devices[0])
