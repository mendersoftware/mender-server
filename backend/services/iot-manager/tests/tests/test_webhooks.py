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

import json
import re
import time

import pytest

import internal_api.exceptions as intrnl_exceptions

from client import ManagementAPIClient, InternalAPIClient
from management_api import models as mgmt_models
from internal_api import models as intrnl_models
from internal_api.model_utils import ModelNormal, ModelComposed
from utils import compare_expectations

TEST_TENANT_ID = "123456789012345678901234"

ED25519_PUBKEY = """-----BEGIN PUBLIC KEY-----
MCowBQYDK2VwAyEA18UTNP3MwIpKGZDv25V8ExIlqYjKs7C9cKI2KQDDFqc=
-----END PUBLIC KEY-----"""

REGEX_UUID = re.compile(
    "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
)
REGEX_RFC3339 = re.compile(
    r"[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.?[0-9]{,9}(Z|[+-][0-9]{2}:00)"
)


class TestWebhooks:
    @pytest.fixture
    def setup_test_case(self, clean_mongo, clean_mmock):
        mgmt = ManagementAPIClient(TEST_TENANT_ID)
        mgmt.register_integration(
            mgmt_models.Integration(
                "webhook",
                mgmt_models.Credentials(
                    type="http",
                    http=mgmt_models.HTTPHttp(
                        url="http://mmock:8080", secret="deadbeef"
                    ),
                ),
            )
        )
        yield clean_mmock

    def test_webhook_provision_device(self, setup_test_case):
        mgmt = ManagementAPIClient(TEST_TENANT_ID)
        device_id = "9c91e5fa-9dcb-451b-b4d8-d30e6b1e9266"
        authset_id = "70338f23-b4e6-49ef-a126-097ce6a44140"
        mmock = setup_test_case
        intrnl = InternalAPIClient()
        assert isinstance(intrnl_models.NewDevice("123"), (ModelNormal, ModelComposed))
        dev = intrnl_models.NewDevice(
            device_id,
            status="accepted",
            auth_sets=[
                intrnl_models.AuthSet(
                    id=authset_id,
                    device_id=device_id,
                    pubkey=ED25519_PUBKEY,
                    identity_data={"sn": "12345"},
                )
            ],
        )

        intrnl.provision_device(
            TEST_TENANT_ID, dev,
        )

        # let the async processing complete
        time.sleep(1)

        expected_event = {
            "id": REGEX_UUID,
            "type": "device-provisioned",
            "data": {
                "id": device_id,
                "status": "accepted",
                "auth_sets": [
                    {
                        "id": REGEX_UUID,
                        "device_id": device_id,
                        "identity_data": {"sn": "12345"},
                        "pubkey": ED25519_PUBKEY,
                    }
                ],
            },
            "time": REGEX_RFC3339,
        }

        assert len(mmock.unmatched) == 0
        matched = mmock.matched
        assert len(matched) == 1
        req = matched[0]["request"]
        event = json.loads(req["body"])
        compare_expectations(
            expected_event, event,
        )
        events = mgmt.list_events()
        actual_event = events[0]
        # Convert model to dict
        eventd = actual_event.to_dict()
        # Convert datetime attribute to string
        eventd["time"] = eventd["time"].isoformat()
        # Verify expectations
        compare_expectations(expected_event, eventd)

    def test_webhook_decommission_device(self, setup_test_case):
        mgmt = ManagementAPIClient(TEST_TENANT_ID)
        device_id = "9c91e5fa-9dcb-451b-b4d8-d30e6b1e9266"
        authset_id = "70338f23-b4e6-49ef-a126-097ce6a44140"
        mmock = setup_test_case
        intrnl = InternalAPIClient()
        try:
            intrnl.decommission_device(TEST_TENANT_ID, device_id)
        except intrnl_exceptions.NotFoundException:
            pass

        # let the async processing complete
        time.sleep(1)

        expected_event = {
            "id": REGEX_UUID,
            "type": "device-decommissioned",
            "data": {"id": device_id},
            "time": REGEX_RFC3339,
        }

        assert len(mmock.unmatched) == 0
        matched = mmock.matched
        assert len(matched) == 1
        req = matched[0]["request"]
        event = json.loads(req["body"])
        compare_expectations(
            expected_event, event,
        )
        events = mgmt.list_events()
        actual_event = events[0]
        # Convert model to dict
        eventd = actual_event.to_dict()
        # Convert datetime attribute to string
        eventd["time"] = eventd["time"].isoformat()
        # Verify expectations
        compare_expectations(expected_event, eventd)

    def test_webhook_status_change(self, setup_test_case):
        mgmt = ManagementAPIClient(TEST_TENANT_ID)
        device_id = "9c91e5fa-9dcb-451b-b4d8-d30e6b1e9266"
        authset_id = "70338f23-b4e6-49ef-a126-097ce6a44140"
        mmock = setup_test_case
        intrnl = InternalAPIClient()
        try:
            intrnl.update_device_statuses(
                TEST_TENANT_ID, "rejected", [intrnl_models.InlineObject(device_id)]
            )
        except intrnl_exceptions.NotFoundException:
            pass

        # let the async processing complete
        time.sleep(1)

        expected_event = {
            "id": REGEX_UUID,
            "type": "device-status-changed",
            "data": {"id": device_id, "status": "rejected"},
            "time": REGEX_RFC3339,
        }

        assert len(mmock.unmatched) == 0
        matched = mmock.matched
        assert len(matched) == 1
        req = matched[0]["request"]
        event = json.loads(req["body"])
        compare_expectations(
            expected_event, event,
        )
        events = mgmt.list_events()
        actual_event = events[0]
        # Convert model to dict
        eventd = actual_event.to_dict()
        # Convert datetime attribute to string
        eventd["time"] = eventd["time"].isoformat()
        # Verify expectations
        compare_expectations(expected_event, eventd)
