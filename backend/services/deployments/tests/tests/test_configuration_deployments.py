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

import bravado
import pytest
import requests
import os
import subprocess
import json

from uuid import uuid4

from urllib.parse import urlparse, parse_qs, urlencode, quote
from client import DeploymentsClient, ArtifactsClient

from bson.objectid import ObjectId
from common import api_client_int, mongo, clean_db, Device, Lock, MONGO_LOCK_FILE


from client import SimpleDeviceClient, InventoryClient


def inventory_add_dev(dev, tenant_id):
    inv = InventoryClient()
    inv.report_attributes(
        dev.fake_token_mt(tenant_id),
        [{"name": "device_type", "value": dev.device_type}],
    )


class TestInternalApiPostConfigurationDeployment:
    def test_ok(self, api_client_int, clean_db, mongo):
        with Lock(MONGO_LOCK_FILE) as l:
            tenant_id = str(ObjectId())
            _, r = api_client_int.create_tenant(tenant_id)
            assert r.status_code == 201

            dev = Device()
            inventory_add_dev(dev, tenant_id)
            deployment_id = "foo"
            deployment_id = str(uuid4())
            device_id = dev.devid
            url = api_client_int.make_api_url(
                "/tenants/{}/configuration/deployments/{}/devices/{}".format(
                    tenant_id, deployment_id, device_id
                )
            )
            configuration_deployment = {"name": "foo", "configuration": '{"foo":"bar"}'}
            rsp = requests.post(url, json=configuration_deployment, verify=False)
            assert rsp.status_code == 201
            loc = rsp.headers.get("Location", None)
            assert loc
            api_deployment_id = os.path.basename(loc)
            assert api_deployment_id == deployment_id

            # verify the deployment has been stored correctly in mongodb
            deployment = mongo[
                "deployment_service-{}".format(tenant_id)
            ].deployments.find_one({"_id": deployment_id})
            assert deployment is not None
            assert deployment["type"] == "configuration"
            assert deployment["configuration"]
            l.unlock()

    def test_fail_missing_name(self, api_client_int, clean_db):
        with Lock(MONGO_LOCK_FILE) as l:
            tenant_id = str(ObjectId())
            _, r = api_client_int.create_tenant(tenant_id)
            assert r.status_code == 201

            deployment_id = "foo"
            deployment_id = str(uuid4())
            device_id = "bar"
            url = api_client_int.make_api_url(
                "/tenants/{}/configuration/deployments/{}/devices/{}".format(
                    tenant_id, deployment_id, device_id
                )
            )
            configuration_deployment = {"configuration": '{"foo":"bar"}'}
            rsp = requests.post(url, json=configuration_deployment, verify=False)
            assert rsp.status_code == 400
            l.unlock()

    def test_fail_missing_configuration(self, api_client_int, clean_db):
        with Lock(MONGO_LOCK_FILE) as l:
            tenant_id = str(ObjectId())
            _, r = api_client_int.create_tenant(tenant_id)
            assert r.status_code == 201

            deployment_id = "foo"
            deployment_id = str(uuid4())
            device_id = "bar"
            url = api_client_int.make_api_url(
                "/tenants/{}/configuration/deployments/{}/devices/{}".format(
                    tenant_id, deployment_id, device_id
                )
            )
            configuration_deployment = {"name": "foo"}
            rsp = requests.post(url, json=configuration_deployment, verify=False)
            assert rsp.status_code == 400
            l.unlock()

    def test_fail_wrong_deployment_id(self, api_client_int, clean_db):
        with Lock(MONGO_LOCK_FILE) as l:
            tenant_id = str(ObjectId())
            _, r = api_client_int.create_tenant(tenant_id)
            assert r.status_code == 201

            dev = Device()
            inventory_add_dev(dev, tenant_id)
            deployment_id = "foo"
            deployment_id = "baz"
            device_id = dev.devid
            url = api_client_int.make_api_url(
                "/tenants/{}/configuration/deployments/{}/devices/{}".format(
                    tenant_id, deployment_id, device_id
                )
            )
            configuration_deployment = {"name": "foo", "configuration": '{"foo":"bar"}'}
            rsp = requests.post(url, json=configuration_deployment, verify=False)
            assert rsp.status_code == 400
            l.unlock()

    def test_fail_duplicate_deployment(self, api_client_int, clean_db):
        with Lock(MONGO_LOCK_FILE) as l:
            tenant_id = str(ObjectId())
            _, r = api_client_int.create_tenant(tenant_id)
            assert r.status_code == 201

            dev = Device()
            inventory_add_dev(dev, tenant_id)
            deployment_id = "foo"
            deployment_id = str(uuid4())
            device_id = dev.devid
            url = api_client_int.make_api_url(
                "/tenants/{}/configuration/deployments/{}/devices/{}".format(
                    tenant_id, deployment_id, device_id
                )
            )
            configuration_deployment = {"name": "foo", "configuration": '{"foo":"bar"}'}
            rsp = requests.post(url, json=configuration_deployment, verify=False)
            assert rsp.status_code == 201
            loc = rsp.headers.get("Location", None)
            assert loc
            api_deployment_id = os.path.basename(loc)
            assert api_deployment_id == deployment_id

            rsp = requests.post(url, json=configuration_deployment, verify=False)
            assert rsp.status_code == 409
            l.unlock()


class TestDevicesApiGetConfigurationDeploymentLink:
    """
    Tests /download/configuration/... download links.
    """

    @pytest.mark.parametrize(
        "test_set",
        [
            {"dev_type": "rpi4", "name": "foo", "config": '{"foo":"bar"}'},
            {"dev_type": "bb", "name": "bar", "config": '{"foo":"bar","baz":"qux"}'},
        ],
    )
    def test_ok(self, api_client_int, clean_db, mongo, test_set):
        """
        Happy path - correct link obtained from the service, leading to a successful download
        of a correct artifact.
        """

        # set up deployment
        with Lock(MONGO_LOCK_FILE) as l:
            tenant_id = str(ObjectId())
            _, r = api_client_int.create_tenant(tenant_id)
            assert r.status_code == 201

            deployment_id = str(uuid4())

            dev = Device()
            dev.device_type = test_set["dev_type"]
            inventory_add_dev(dev, tenant_id)

            configuration_deployment = {
                "name": test_set["name"],
                "configuration": test_set["config"],
            }

            make_deployment(
                api_client_int,
                tenant_id,
                deployment_id,
                dev.devid,
                configuration_deployment,
            )

            # obtain + verify deployment instructions
            dc = SimpleDeviceClient()
            nextdep = dc.get_next_deployment(
                dev.fake_token_mt(tenant_id),
                artifact_name="dontcare",
                device_type=test_set["dev_type"],
            )

            assert nextdep.artifact["artifact_name"] == test_set["name"]
            assert nextdep.artifact["source"]["uri"] is not None
            assert nextdep.artifact["source"]["expire"] is not None
            assert nextdep.artifact["device_types_compatible"] == [test_set["dev_type"]]

            # get/verify download contents
            r = requests.get(nextdep.artifact["source"]["uri"], verify=False)
            assert r.status_code == 200

            with open("/testing/out.mender", "wb+") as f:
                f.write(r.content)

            self.verify_artifact(
                "/testing/out.mender",
                test_set["name"],
                test_set["dev_type"],
                test_set["config"],
            )
            l.unlock()

    def test_failures(self, api_client_int, clean_db, mongo):
        """
        Simulate invalid or malicious download requests.
        """
        # for reference - get a real, working link to an actual deployment
        with Lock(MONGO_LOCK_FILE) as l:
            tenant_id = str(ObjectId())
            _, r = api_client_int.create_tenant(tenant_id)
            assert r.status_code == 201

            deployment_id = str(uuid4())
            dev = Device()
            inventory_add_dev(dev, tenant_id)
            configuration_deployment = {"name": "foo", "configuration": '{"foo":"bar"}'}

            make_deployment(
                api_client_int,
                tenant_id,
                deployment_id,
                dev.devid,
                configuration_deployment,
            )

            dc = SimpleDeviceClient()
            nextdep = dc.get_next_deployment(
                dev.fake_token_mt(tenant_id),
                artifact_name="dontcare",
                device_type="hammer",
            )
            uri = nextdep.artifact["source"]["uri"]
            qs = parse_qs(urlparse(uri).query)

            # now break the url in various ways

            # wrong deployment (signature error)
            uri_bad_depl = uri.replace(deployment_id, str(uuid4()))
            r = requests.get(uri_bad_depl, verify=False)
            assert r.status_code == 403

            # wrong tenant in url (signature error)
            other_tenant_id = str(ObjectId())
            _, r = api_client_int.create_tenant(other_tenant_id)
            assert r.status_code == 201

            uri_bad_tenant = uri.replace(tenant_id, other_tenant_id)
            r = requests.get(uri_bad_tenant, verify=False)
            assert r.status_code == 403

            # wrong dev type (signature error)
            other_dev = Device()
            other_dev.device_type = "foo"
            uri_bad_devtype = uri.replace(dev.device_type, other_dev.device_type)
            r = requests.get(uri_bad_devtype, verify=False)
            assert r.status_code == 403

            # wrong dev id (signature error)
            uri_bad_devid = uri.replace(dev.devid, other_dev.devid)
            r = requests.get(uri_bad_devid, verify=False)
            assert r.status_code == 403

            # wrong x-men-signature
            uri_bad_sig = uri.replace(
                qs["x-men-signature"][0], "mftJRzBafnvMXhmMBH3THQertiEk0dZKP075bjBKccc"
            )
            r = requests.get(uri_bad_sig, verify=False)
            assert r.status_code == 403

            # no x-men-signature
            uri_no_sig = uri.replace("&x-men-signature=", "")
            r = requests.get(uri_no_sig, verify=False)
            assert r.status_code == 400

            # no x-men-expire
            uri_no_exp = uri.replace("&x-men-expire=", "")
            r = requests.get(uri_no_exp, verify=False)
            assert r.status_code == 400
            l.unlock()

    def verify_artifact(self, fname, name, dtype, config):
        stdout = subprocess.check_output(["mender-artifact", "read", fname])
        stdout = stdout.decode("utf-8")
        # sanitize varying whitespaces
        stdout = " ".join(stdout.split())
        print(stdout)

        # type, name, dev type
        assert "Type: mender-configure" in stdout
        assert "Name: {}".format(name) in stdout
        assert "Version: 3" in stdout
        assert "Compatible devices: '[{}]'".format(dtype) in stdout

        # configuration contents
        metapos = stdout.index("Metadata")
        start, stop = stdout.index("{", metapos), stdout.index("}", metapos)
        assert json.loads(config) == json.loads(stdout[start : stop + 1])

        # provides
        assert (
            "Provides: data-partition.mender-configure.version: {}".format(name)
            in stdout
        )
        assert 'Clears Provides: ["data-partition.mender-configure.*"]' in stdout
        assert "Depends: Nothing" in stdout


class TestDeviceApiGetConfigurationDeploymentNext:
    """
    Verify expected failures when asking for a configuration upgrade,
    i.e. that devices that are not eligible won't get it.
    (happy path is tested in GetConfigurationDeploymentLink::test_ok)
    """

    def test_fail_no_upgrade(self, api_client_int, clean_db, mongo):
        # start with a valid deployment
        with Lock(MONGO_LOCK_FILE) as l:
            tenant_id = str(ObjectId())
            _, r = api_client_int.create_tenant(tenant_id)
            assert r.status_code == 201

            deployment_id = str(uuid4())
            dev = Device()
            inventory_add_dev(dev, tenant_id)
            configuration_deployment = {"name": "foo", "configuration": '{"foo":"bar"}'}

            make_deployment(
                api_client_int,
                tenant_id,
                deployment_id,
                dev.devid,
                configuration_deployment,
            )

            # try get upgrade
            # valid device id + type, but different tenant
            other_tenant_id = str(ObjectId())
            _, r = api_client_int.create_tenant(other_tenant_id)
            assert r.status_code == 201

            dc = SimpleDeviceClient()
            nodep = dc.get_next_deployment(
                dev.fake_token_mt(other_tenant_id),
                artifact_name="dontcare",
                device_type=dev.device_type,
            )
            assert nodep is None

            # correct tenant, incorrect device id (but correct type)
            otherdev = Device()
            nodep = dc.get_next_deployment(
                otherdev.fake_token_mt(tenant_id),
                artifact_name="dontcare",
                device_type=dev.device_type,
            )
            assert nodep is None

            # correct tenant, correct device id, incorrect type
            nodep = dc.get_next_deployment(
                otherdev.fake_token_mt(tenant_id),
                artifact_name="dontcare",
                device_type=otherdev.device_type,
            )
            assert nodep is None
            l.unlock()


def make_deployment(api_client_int, tenant_id, dep_id, dev_id, deployment):
    url = api_client_int.make_api_url(
        "/tenants/{}/configuration/deployments/{}/devices/{}".format(
            tenant_id, dep_id, dev_id
        )
    )

    rsp = requests.post(url, json=deployment, verify=False)

    assert rsp.status_code == 201
    loc = rsp.headers.get("Location", None)
    assert loc
    api_deployment_id = os.path.basename(loc)
    assert api_deployment_id == dep_id
    return api_deployment_id
