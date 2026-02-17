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
import logging
import pytest
import time
import uuid
import redo

from testutils.api.client import ApiClient
from testutils.infra.cli import CliUseradm, CliDeviceauth
from testutils.infra.container_manager.kubernetes_manager import isK8S
import testutils.api.deviceauth as deviceauth
import testutils.api.useradm as useradm
import testutils.api.inventory as inventory
import testutils.api.inventory_v2 as inventory_v2

from testutils.common import (
    create_user,
    create_org,
    make_accepted_device,
    make_accepted_devices,
    make_pending_device,
)

WAITING_TIME_K8S = 5.0

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

class TestGetDevicesBase:
    def do_test_get_devices_ok(self, user, tenant_token=""):
        useradmm = ApiClient(useradm.URL_MGMT)
        devauthd = ApiClient(deviceauth.URL_DEVICES)
        devauthm = ApiClient(deviceauth.URL_MGMT)
        invm = ApiClient(inventory.URL_MGMT)

        # log in user
        r = useradmm.call("POST", useradm.URL_LOGIN, auth=(user.name, user.pwd))
        assert r.status_code == 200
        utoken = r.text

        # count existing devices
        r = invm.with_auth(utoken).call(
            "GET", inventory.URL_DEVICES, qs_params={"per_page": 1}
        )
        assert r.status_code == 200
        count = int(r.headers["X-Total-Count"])

        # prepare accepted devices
        make_accepted_devices(devauthd, devauthm, utoken, tenant_token, 40)

        # wait for devices to be provisioned
        for _ in redo.retrier(attempts=3, sleeptime=1):
            r = invm.with_auth(utoken).call(
                "GET", inventory.URL_DEVICES, qs_params={"per_page": 1}
            )
            if r.status_code == 200:
                break

        assert r.status_code == 200
        new_count = int(r.headers["X-Total-Count"])
        assert new_count == count + 40

    def do_test_filter_devices_ok(self, user, tenant_token=""):
        useradmm = ApiClient(useradm.URL_MGMT)
        devauthd = ApiClient(deviceauth.URL_DEVICES)
        devauthm = ApiClient(deviceauth.URL_MGMT)
        invm = ApiClient(inventory.URL_MGMT)
        invd = ApiClient(inventory.URL_DEV)

        # log in user
        r = useradmm.call("POST", useradm.URL_LOGIN, auth=(user.name, user.pwd))
        assert r.status_code == 200
        utoken = r.text

        r = invm.with_auth(utoken).call(
            "GET", inventory.URL_DEVICES, qs_params={"per_page": 1}
        )
        assert r.status_code == 200
        count = int(r.headers["X-Total-Count"])

        # prepare accepted devices
        devs = make_accepted_devices(devauthd, devauthm, utoken, tenant_token, 40)

        # wait for devices to be provisioned
        for _ in redo.retrier(attempts=3, sleeptime=1):
            r = invm.with_auth(utoken).call(
                "GET", inventory.URL_DEVICES, qs_params={"per_page": 1}
            )
            if r.status_code == 200:
                break

        assert r.status_code == 200
        new_count = int(r.headers["X-Total-Count"])
        assert new_count == count + 40

        # upload inventory attributes
        for i, d in enumerate(devs):
            payload = [{"name": "mac", "value": "de:ad:be:ef:06:" + str(i)}]
            r = invd.with_auth(d.token).call(
                "PATCH", inventory.URL_DEVICE_ATTRIBUTES, payload
            )
            assert r.status_code == 200

        # get device with exact mac value
        qs_params = {}
        qs_params["per_page"] = 100
        qs_params["mac"] = "de:ad:be:ef:06:7"
        r = invm.with_auth(utoken).call(
            "GET", inventory.URL_DEVICES, qs_params=qs_params
        )
        assert r.status_code == 200
        api_devs = r.json()
        assert len(api_devs) == 1

class TestGetDevices(TestGetDevicesBase):
    def test_get_devices_ok(self, user):
        self.do_test_get_devices_ok(user)

    def test_filter_devices_ok(self, user):
        self.do_test_filter_devices_ok(user)

class TestDevicePatchAttributes:
    def test_ok(self, user):
        useradmm = ApiClient(useradm.URL_MGMT)
        devauthd = ApiClient(deviceauth.URL_DEVICES)
        devauthm = ApiClient(deviceauth.URL_MGMT)
        invm = ApiClient(inventory.URL_MGMT)
        invd = ApiClient(inventory.URL_DEV)

        # log in user
        r = useradmm.call("POST", useradm.URL_LOGIN, auth=(user.name, user.pwd))
        assert r.status_code == 200
        utoken = r.text

        # prepare accepted devices
        devs = make_accepted_devices(devauthd, devauthm, utoken, "", 3)

        # wait for devices to be provisioned
        time.sleep(3)

        for i, d in enumerate(devs):
            payload = [
                {"name": "mac", "value": "mac-new-" + str(d.id)},
                {
                    # empty value for existing
                    "name": "sn",
                    "value": "",
                },
                {
                    # empty value for new
                    "name": "new-empty",
                    "value": "",
                },
            ]
            r = invd.with_auth(d.token).call(
                "PATCH", inventory.URL_DEVICE_ATTRIBUTES, payload
            )
            assert r.status_code == 200

        for d in devs:
            r = invm.with_auth(utoken).call(
                "GET", inventory.URL_DEVICE, path_params={"id": d.id}
            )
            assert r.status_code == 200

            api_dev = r.json()
            # Expected inventory count per scope:
            # {"inventory": 3, "identity": 1+2, "system": 3}
            # +2 comes from the id_data see MEN-3637
            assert len(api_dev["attributes"]) == 9
            # new scopes: identity and system holding authset status and
            #             time-stamp values respectively

            for a in api_dev["attributes"]:
                if a["name"] == "mac" and a["scope"] == "inventory":
                    assert a["value"] == "mac-new-" + str(api_dev["id"])
                elif a["name"] == "sn" and a["scope"] == "inventory":
                    assert a["value"] == ""
                elif a["name"] == "new-empty" and a["scope"] == "inventory":
                    assert a["value"] == ""
                elif a["name"] == "status" and a["scope"] == "identity":
                    assert a["value"] in ["accepted", "pending"]
                elif a["scope"] != "inventory":
                    # Check that the value is present
                    assert a["value"] != ""
                else:
                    assert False, "unexpected attribute " + a["name"]

    def test_fail_no_attr_value(self, user):
        useradmm = ApiClient(useradm.URL_MGMT)
        devauthd = ApiClient(deviceauth.URL_DEVICES)
        devauthm = ApiClient(deviceauth.URL_MGMT)
        invd = ApiClient(inventory.URL_DEV)

        # log in user
        r = useradmm.call("POST", useradm.URL_LOGIN, auth=(user.name, user.pwd))
        assert r.status_code == 200
        utoken = r.text

        # prepare accepted devices
        devs = make_accepted_devices(devauthd, devauthm, utoken, "", 1)

        # wait for devices to be provisioned
        time.sleep(3)

        for i, d in enumerate(devs):
            payload = [{"name": "mac"}]
            r = invd.with_auth(d.token).call(
                "PATCH", inventory.URL_DEVICE_ATTRIBUTES, payload
            )
            assert r.status_code == 400

def dict_to_inventoryattrs(d, scope="inventory"):
    attr_list = []
    for key, value in d.items():
        attr = {"name": key, "value": value}
        if scope is not None:
            attr["scope"] = scope
        attr_list.append(attr)

    return attr_list

def add_devices_to_tenant(tenant, dev_inventories):
    try:
        tenant.devices
    except AttributeError:
        tenant.devices = []

    useradmm = ApiClient(useradm.URL_MGMT)
    devauthd = ApiClient(deviceauth.URL_DEVICES)
    devauthm = ApiClient(deviceauth.URL_MGMT)
    invd = ApiClient(inventory.URL_DEV)

    user = tenant.users[0]
    utoken = useradmm.call("POST", useradm.URL_LOGIN, auth=(user.name, user.pwd)).text
    assert utoken != ""
    tenant.api_token = utoken

    for inv in dev_inventories:
        device = make_accepted_device(
            devauthd, devauthm, utoken, tenant_token=tenant.tenant_token
        )
        tenant.devices.append(device)

        attrs = dict_to_inventoryattrs(inv)
        rsp = invd.with_auth(device.token).call(
            "PATCH", inventory.URL_DEVICE_ATTRIBUTES, body=attrs
        )
        assert rsp.status_code == 200
        device.inventory = inv

    return tenant

class DeviceFilteringTests:
    @property
    def logger(self):
        try:
            return self._logger
        except AttributeError:
            self._logger = logging.getLogger(self.__class__.__name__)
        return self._logger

    def do_test_search_v2(self, user_token, devices, additional_test_cases = []):
        assert user_token
        assert len(devices) > 0

        test_cases = [
            {
                "name": "Test $eq single match",
                "request": {
                    "filters": [
                        {
                            "type": "$eq",
                            "attribute": "idx",
                            "value": 1,
                            "scope": "inventory",
                        }
                    ],
                },
                "status_code": 200,
                "response": [
                    {
                        "id": str(devices[1].id),
                        "attributes": dict_to_inventoryattrs(
                            devices[1].inventory, scope="inventory"
                        ),
                    }
                ],
            },
            {
                "name": "Test $eq no-match",
                "request": {
                    "filters": [
                        {
                            "type": "$eq",
                            "attribute": "id_data",
                            "value": "illegal_data",
                            "scope": "inventory",
                        }
                    ],
                },
                "status_code": 200,
                "response": [],
            },
            {
                "name": "Test $nin, sort by descending idx",
                "request": {
                    "filters": [
                        {
                            "type": "$nin",
                            "attribute": "artifact",
                            "value": ["v3"],
                            "scope": "inventory",
                        },
                    ],
                    "sort": [
                        {"attribute": "idx", "scope": "inventory", "order": "desc"},
                    ],
                },
                "status_code": 200,
                "response": [
                    {
                        "id": dev.id,
                        "attributes": dict_to_inventoryattrs(
                            dev.inventory, scope="inventory"
                        ),
                    }
                    # The following is just the python expression of the
                    # above operation.
                    for dev in sorted(
                        filter(
                            lambda dev: "v3" not in dev.inventory["artifact"],
                            devices,
                        ),
                        key=lambda dev: dev.inventory["idx"],
                        reverse=True,
                    )
                ],
            },
            {
                "name": "Error - missing type parameter",
                "request": {
                    "filters": [
                        {
                            "attribute": "artifact",
                            "value": "v1",
                            "scope": "inventory",
                        },
                    ],
                },
                "status_code": 400,
            },
            {
                "name": "Error - invalid filter scope",
                "request": {
                    "filters": [
                     {
                            "type": "$eq",
                            "attribute": "idx",
                            "value": 1,
                            "scope": "user_defined",
                        }
                    ],
                },
                "status_code": 400,
            },
            {
                "name": "Error - invalid sort scope",
                "request": {
                    "filters": [
                        {
                            "type": "$eq",
                            "attribute": "idx",
                            "value": 1,
                            "scope": "inventory",
                        }
                    ],
                    "sort":[
                        {"attribute": "idx", "scope": "user_defined", "order": "desc"},
                    ]
                },
                "status_code": 400,
            },
            {
                "name": "Error - invalid attribute scope",
                "request": {
                    "filters": [
                        {
                            "type": "$eq",
                            "attribute": "idx",
                            "value": 1,
                            "scope": "inventory",
                        }
                    ],
                    "attributes":[
                        { "attribute": "idx", "scope": "user_defined" },
                    ]
                },
                "status_code": 400,
            },
            {
                "name": "Error - invalid filter scope with path",
                "request": {
                    "filters": [
                        {
                            "type": "$eq",
                            "attribute": "idx",
                            "value": 1,
                            "scope": "../../../../Windows/system.ini",
                        },
                    ],
                },
                "status_code": 400,
            },
            {
                "name": "Error - valid mongo query unsupported operation",
                "request": {
                    "filters": [
                        {
                            "type": "$type",
                            "attribute": "artifact",
                            "value": ["int", "string", "array"],
                            "scope": "inventory",
                        },
                    ],
                },
                "status_code": 400,
            },
        ] + additional_test_cases

        invm_v2 = ApiClient(inventory_v2.URL_MGMT)
        for test_case in test_cases:
            self.do_search_test_case(invm_v2, inventory_v2.URL_SEARCH, user_token, test_case)

    def do_search_test_case(self, client, url, token, test_case):
        self.logger.info("Running test case: %s" % test_case["name"])
        rsp = client.with_auth(token).call(
            "POST", url, test_case["request"]
        )
        assert rsp.status_code == test_case["status_code"], (
            "Unexpected status code (%d) from /filters/search response: %s"
            % (rsp.status_code, rsp.text)
        )

        if rsp.status_code == 200 and "response" in test_case:
            body = rsp.json()
            if body is None:
                body = []
            self.logger.info(test_case["response"])
            self.logger.info(body)
            assert len(test_case["response"]) == len(body), (
                "Unexpected number of results: %s != %s"
                % (
                    [dev["id"] for dev in test_case["response"]],
                    [dev["id"] for dev in body],
                )
            )

            if len(test_case["response"]) > 0:
                if "sort" not in test_case["request"]:
                    body = sorted(body, key=lambda dev: dev["id"])
                    test_case["response"] = sorted(
                        test_case["response"], key=lambda dev: dev["id"]
                    )

                for i, dev in enumerate(test_case["response"]):
                    assert (
                        dev["id"] == body[i]["id"]
                    ), "Unexpected device in response"
                    assert_device_attributes(dev, body[i])

class TestDeviceFiltering(DeviceFilteringTests):
    @pytest.fixture(autouse=True)
    def setup_user_and_devices(self, clean_mongo):
        uuidv4 = str(uuid.uuid4())
        username, password = (
            "some.user+" + uuidv4 + "@example.com",
            "secretsecret",
        )
        user = create_user(username, password)

        useradmm = ApiClient(useradm.URL_MGMT)
        self.user_token = useradmm.call("POST", useradm.URL_LOGIN, auth=(user.name, user.pwd)).text
        assert self.user_token != ""

        self.devices = self.add_devices(self.user_token, [
            {"artifact": ["v1"], "idx": 0},
            {"artifact": ["v1"], "idx": 1},
            {"artifact": ["v1"], "idx": 2},
        ])

    def test_search_v2(self):
        self.do_test_search_v2(self.user_token, self.devices)

    def test_search_v2_internal(self):
        """
        This endpoint technically exists in open-source, but
        the URL expects a "tenant_id" parameter in the
        URL ("tenants" is a concept that doesn't exist in open-source)
        and when one isn't specified, the implementation is
        identical to search_v2 above afaict.

        As such, I see no need to maintain tests for this endpoint.
        """

    def add_devices(self, user_token, dev_inventories):
        devauthd = ApiClient(deviceauth.URL_DEVICES)
        devauthm = ApiClient(deviceauth.URL_MGMT)
        invd = ApiClient(inventory.URL_DEV)

        devices = []
        for inv in dev_inventories:
            device = make_accepted_device(
                devauthd, devauthm, user_token
            )

            attrs = dict_to_inventoryattrs(inv)
            rsp = invd.with_auth(device.token).call(
                "PATCH", inventory.URL_DEVICE_ATTRIBUTES, body=attrs
            )
            assert rsp.status_code == 200
            device.inventory = inv

            devices.append(device)

        return devices

def assert_device_attributes(dev, api_dev):
    for attr in dev["attributes"]:
        assert attr in api_dev["attributes"], (
            "Missing inventory attribute: %s; device attributes: %s"
            % (attr, api_dev["attributes"])
        )
