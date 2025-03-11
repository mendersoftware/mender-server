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

from typing import Optional
import pytest
import logging
import uuid
import time

from testutils.api import (
    deviceauth,
    deviceconfig,
    iot_manager as iot,
    useradm,
)
from testutils.api.client import ApiClient
from testutils.infra.cli import CliUseradm, CliDeviceauth
from testutils.common import (
    Device,
    User,
    MockedHttp,
    create_org,
    create_user,
    clean_mongo,
    mongo,
    make_device_with_inventory,
    decommission_device,
    submit_inventory,
)

EVENT_PROPAGATION_TIMEOUT_S = 8


@pytest.fixture(scope="function")
def clean_migrated_mongo(clean_mongo):
    deviceauth_cli = CliDeviceauth()
    useradm_cli = CliUseradm()

    deviceauth_cli.migrate()
    useradm_cli.migrate()

    yield clean_mongo


@pytest.fixture(scope="function")
def user(clean_migrated_mongo):
    uuidv4 = str(uuid.uuid4())
    u = create_user(f"user+{uuidv4}@example.com", "correcthorse")
    rsp = ApiClient(useradm.URL_MGMT).call(
        "POST", useradm.URL_LOGIN, auth=(u.name, u.pwd)
    )
    assert rsp.status_code == 200
    u.token = rsp.text
    yield u


class TestInventoryWebhooks:
    """Test the inventory webhooks."""

    @classmethod
    def setup_class(cls):
        cls.attributes = [
            {"name": "os", "value": "Super"},
            {"name": "location", "value": "Here"},
        ]
        cls.new_attributes = [
            {"name": "os", "value": "Super"},
            {"name": "location", "value": "Not Here"},
        ]
        cls.mocked_http = MockedHttp()

        cls.api_devauth_devices = ApiClient(base_url=deviceauth.URL_DEVICES)
        cls.api_devauth_mgmt = ApiClient(base_url=deviceauth.URL_MGMT)
        cls.api_iot = ApiClient(base_url=iot.URL_MGMT)

        cls.devices = list()
        cls.logger = logging.getLogger(cls.__class__.__name__)

    @classmethod
    def teardown_class(cls):
        cls.logger.info(
            f"Inventory webhook test teardown - removing devices: {cls.devices}"
        )

    def _prepare_device(self, user: User) -> Device:
        """Create an accepted device."""
        tenant_token = getattr(getattr(user, "tenant", {}), "tenant_token", "")
        dev = make_device_with_inventory(
            self.attributes, user.token, tenant_token=tenant_token,
        )
        self.devices.append(dev)
        return dev

    def _create_webhooks(
        self, user: User, id, scopes=["deviceauth"], expected_responose_code=201
    ):
        target_webhook_path = "/webhook/deviceauth/" + id
        integration = {
            "provider": "webhook",
            "credentials": {
                "type": "http",
                "http": {"url": "http://mock-httpd:1080" + target_webhook_path,},
            },
            "description": "web hook test " + id,
            # "scopes": scopes,
        }
        response = (
            self.api_iot.with_auth(user.token)
            .with_header("Content-Type", "application/json")
            .call("POST", iot.URL_INTEGRATIONS, body=integration)
        )
        assert response.status_code == expected_responose_code

    def _get_events(self, user: User, expected_responose_code=200, integration_id=None):
        url = iot.URL_EVENTS
        if integration_id:
            url = url + "?integration_id=" + integration_id
        response = self.api_iot.with_auth(user.token).call("GET", url)
        assert response.status_code == expected_responose_code
        return response.json()

    def _get_integrations(self, user: User, expected_response_code=200):
        response = self.api_iot.with_auth(user.token).call("GET", iot.URL_INTEGRATIONS)
        assert response.status_code == expected_response_code
        return response.json()

    def _delete_integration(
        self, user: User, integration_id, expected_response_code=204
    ):
        response = self.api_iot.with_auth(user.token).call(
            "DELETE", iot.URL_INTEGRATIONS + "/" + integration_id
        )
        assert response.status_code == expected_response_code

    def test_deviceauth_webhook(
        self, user: User,
    ):
        """Let's create a webhook in only deviceauth scope."""
        dev = self._prepare_device(user)
        new_uuid = str(uuid.uuid4())
        self._create_webhooks(user, new_uuid, scopes=["deviceauth"])
        submit_inventory(self.new_attributes, dev.token)
        time.sleep(EVENT_PROPAGATION_TIMEOUT_S)
        submit_inventory(self.attributes, dev.token)
        time.sleep(EVENT_PROPAGATION_TIMEOUT_S)
        submit_inventory(self.new_attributes, dev.token)
        time.sleep(EVENT_PROPAGATION_TIMEOUT_S)
        target_webhook_path = "/webhook/deviceauth/" + new_uuid
        assert not self.mocked_http.request_seen(target_webhook_path)

    def test_webhooks_all_events(
        self, user: User,
    ):
        """Let's create a webhook in both supported scopes and see all the events."""
        new_uuid = str(uuid.uuid4())
        self._create_webhooks(user, new_uuid, scopes=["deviceauth"])
        dev = self._prepare_device(user)
        submit_inventory(self.new_attributes, dev.token)
        decommission_device(user.token, dev.id)
        time.sleep(EVENT_PROPAGATION_TIMEOUT_S)
        target_webhook_path = "/webhook/deviceauth/" + new_uuid
        all_requests = self.mocked_http.request_get_all(target_webhook_path)
        expected_data_fields_by_type = {
            "device-status-changed": ["id", "status"],
            "device-status-changed": ["id", "status"],
            "device-provisioned": ["id", "status", "auth_sets", "created_ts"],
            "device-decommissioned": ["id"],
        }
        i = 0
        for event_type in [
            "device-status-changed",
            "device-status-changed",
            "device-provisioned",
            "device-decommissioned",
        ]:
            assert all_requests[i]["type"] == event_type
            assert "time" in all_requests[i]
            for field in expected_data_fields_by_type[event_type]:
                assert field in all_requests[i]["data"]
            i = i + 1
        i = 0
        for device_status in ["pending", "accepted", "accepted"]:
            assert all_requests[i]["data"]["status"] == device_status
            assert all_requests[i]["data"]["id"] == dev.id
            i = i + 1
        expected_events_count = (
            4  # we have two for status of the device, and provision and decommission
        )
        assert len(all_requests) == expected_events_count

        events = self._get_events(user)
        assert len(events) == expected_events_count
        i = 4
        for event_type in [
            "device-status-changed",
            "device-status-changed",
            "device-provisioned",
            "device-decommissioned",
        ]:
            i = i - 1
            assert events[i]["type"] == event_type

        # lets get events by integration id
        integrations = self._get_integrations(user)
        assert len(integrations) > 0
        integration_id = integrations[0]["id"]
        events = self._get_events(user, integration_id=integration_id)
        assert len(events) == expected_events_count
        events = self._get_events(user, integration_id=str(uuid.uuid4()))
        assert len(events) == 0
