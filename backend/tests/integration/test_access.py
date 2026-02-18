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
import logging
import json
import pytest
import uuid
import os
import time
import redo

import mender_client

from testutils.common import Tenant, User, update_tenant, create_user, create_org
from testutils.infra.cli import CliTenantadm
from testutils.api.client import ApiClient
import testutils.api.deviceconnect as deviceconnect
import testutils.api.deviceconfig as deviceconfig
import testutils.api.auditlogs as auditlogs
import testutils.api.useradm as useradm
import testutils.api.tenantadm as tenantadm
import testutils.api.tenantadm_v2 as tenantadm_v2
import testutils.integration.stripe as stripeutils

logger = logging.getLogger("testAccess")

def device_connect_insert_device(mongo, device_id, tenant_id, status="connected"):
    devices_collection = mongo.client["deviceconnect"]["devices"]
    devices_collection.delete_one(
        {"_id": device_id, "tenant_id": tenant_id,}
    )
    devices_collection.insert_one(
        {
            "_id": device_id,
            "tenant_id": tenant_id,
            "created_ts": "2022-10-26T16:28:18.796Z",
            "status": "disconnected",
            "updated_ts": "2022-10-26T16:28:51.031Z",
        }
    )

def device_config_insert_device(mongo, device_id, tenant_id, status="connected"):
    devices_collection = mongo.client["deviceconfig"]["devices"]
    devices_collection.delete_one(
        {"_id": device_id, "tenant_id": tenant_id,}
    )
    devices_collection.insert_one(
        {
            "_id": device_id,
            "tenant_id": tenant_id,
            "reported_ts": "2022-10-26T16:28:18.796Z",
            "updated_ts": "2022-10-26T16:28:51.031Z",
            "reported": [{"key": "timezone", "value": "UTC"}],
        }
    )

class _TestAccessBase:
    """Access checking functions.

    Probe a selected EP from every addon feature to see if it's enabled or not.
    Other endpoints are spelled out in detail in acceptance and unit tests for
    useradm/deviceauth access layers (assume they're restricted correctly as well).
    """

    # troubleshoot
    def check_access_remote_term(self, auth, devid, forbid=False):

        api_client = mender_client.ApiClient()
        api_client.configuration.access_token = auth
        devconn = mender_client.DeviceConnectManagementAPIApi(api_client=api_client)

        logger.info(
            f"using {auth} to call {deviceconnect.URL_MGMT_DEVICE} with devid={devid}"
        )
        rsp = devconn.device_connect_management_get_device_without_preload_content(devid)

        if forbid:
            assert rsp.status == 403, f"unexpected status code {rsp.status}, body follows: {rsp.data}"
        else:
            assert rsp.status == 200, f"unexpected status code {rsp.status}, body follows: {rsp.data}"

    def check_access_file_transfer(self, auth, devid, forbid=False):
        api_client = mender_client.ApiClient()
        api_client.configuration.access_token = auth
        devconn = mender_client.DeviceConnectManagementAPIApi(api_client=api_client)
        rsp = devconn.device_connect_management_download_without_preload_content(devid, path="/etc/mender/mender.conf")
        if forbid:
            assert rsp.status == 403, f"unexpected status code {rsp.status}, body follows: {rsp.data}"
        else:
            assert rsp.status in [404, 409], f"unexpected status code {rsp.status}, body follows: {rsp.data}"

        rsp = devconn.device_connect_management_upload_without_preload_content(devid, path="/etc/mender/mender.conf")
        if forbid:
            assert rsp.status == 403, f"unexpected status code {rsp.status}, body follows: {rsp.data}"
        else:
            assert rsp.status != 403, f"unexpected status code {rsp.status}, body follows: {rsp.data}"

    def check_access_auditlogs(self, auth, forbid=False):
        # FIXME: Cannot use generated client due to auditlogs spec being closed source.
        alogs = ApiClient(auditlogs.URL_MGMT)

        res = alogs.with_auth(auth).call("GET", auditlogs.URL_LOGS)

        if forbid:
            assert res.status_code == 403
        else:
            assert res.status_code == 200

    def check_access_sessionlogs(self, auth, forbid=False):
        api_client = mender_client.ApiClient()
        api_client.configuration.access_token = auth
        devconn = mender_client.DeviceConnectManagementAPIApi(api_client=api_client)

        rsp = devconn.device_connect_management_playback_without_preload_content(session_id="foo")

        if forbid:
            assert rsp.status == 403, f"unexpected status code {rsp.status}, body follows: {rsp.data}"
        else:
            assert rsp.status != 403, f"unexpected status code {rsp.status}, body follows: {rsp.data}"

    # configure
    def check_access_deviceconfig(self, auth, devid, forbid=False):
        api_client = mender_client.ApiClient()
        api_client.configuration.access_token = auth
        devconf = mender_client.DeviceConfigureManagementAPIApi(api_client=api_client)
        rsp = devconf.device_config_management_get_device_configuration_without_preload_content(device_id=devid)

        if forbid:
            assert rsp.status == 403, f"unexpected status code {rsp.status}, body follows: {rsp.data}"
        else:
            assert rsp.status == 200, f"unexpected status code {rsp.status}, body follows: {rsp.data}"

    # rbac (no addon)
    def check_access_rbac(self, auth, forbid=False):
        uadm = ApiClient(useradm.URL_MGMT)
        res = uadm.with_auth(auth).call("GET", useradm.URL_ROLES)

        if forbid:
            assert res.status_code == 403
        else:
            assert res.status_code == 200

class TestAccess(_TestAccessBase):
    """Onprem OS.

    Quite a few addon features are available here (despite being
    hidden behind paid addons in hosted).
    """

    def test_ok(self, mongo):
        devid = str(uuid.uuid4())
        email = "mender_tests@" + str(uuid.uuid4()) + ".com"
        password = str(uuid.uuid4())
        user = create_user(email, password)
        r = ApiClient(useradm.URL_MGMT).call(
            "POST", useradm.URL_LOGIN, auth=(user.name, user.pwd)
        )
        assert r.status_code == 200
        auth = r.text

        device_connect_insert_device(mongo, device_id=devid, tenant_id="")
        device_config_insert_device(mongo, device_id=devid, tenant_id="")

        self.check_access_remote_term(auth, devid)
        self.check_access_file_transfer(auth, devid)
        self.check_access_sessionlogs(auth)
        self.check_access_deviceconfig(auth, devid)
