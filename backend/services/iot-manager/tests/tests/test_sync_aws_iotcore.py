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
import re
import uuid

from os import path

import docker

from client import ManagementAPIClient
from management_api import models
from utils import compare_expectations


class TestSyncAWSIoTCore:
    tenant_devices = {
        "TestSyncAWSIoTCore01": [
            {
                "id": "44a66a7e-4b09-4665-b343-0c5d3ad9a2ed",
                "deviceauth": "accepted",
                "hub": "enabled",
            },
            {
                "id": "cea63fbb-0b0b-4a92-bd47-8d310a220a16",
                "deviceauth": "rejected",
                "hub": "disabled",
            },
            {
                "id": "8ba0bc48-a669-462e-873f-f7de2f0e04c3",
                "deviceauth": "no auth",
                "hub": "disabled",
            },
            {
                "id": "895a1ee2-1be0-4e78-889a-7d753f53dbd5",
                "deviceauth": "accepted",
                "hub": "enabled",
            },
            {
                "id": "a74ea639-8fae-4b17-ac3e-f9759dbdbfd9",
                "deviceauth": "rejected",
                "hub": "disabled",
            },
            # Inconsistent devices
            {
                # Inconsistent status
                "id": "1424a387-3431-425b-9f44-1c8eba21812d",
                "deviceauth": "accepted",
                "hub": "disabled",
            },
            {
                # Inconsistent status
                "id": "966095ec-6bdd-4a76-8498-3c0dffdb9ee2",
                "deviceauth": "rejected",
                "hub": "enabled",
            },
            # 3x Devices stored locally but not mirrored everywhere
            {"id": "9b2083e4-83b6-41d3-b089-4d2ec137620b"},
            {"id": "93406e21-8e3f-4435-9786-a294a70298ee", "deviceauth": "accepted"},
            {"id": "7abb6133-ad97-44ba-a159-674242ee565e", "hub": "disabled"},
        ],
        "TestSyncAWSIoTCore02": [
            # All devices are in sync (2x batches)
            {
                "id": "1e657abe-ad58-4d20-af7a-3a3449a405e7",
                "deviceauth": "no auth",
                "hub": "disabled",
            },
            {
                "id": "43396b5b-ff5c-45f2-ab91-150957f037c9",
                "deviceauth": "no auth",
                "hub": "disabled",
            },
            {
                "id": "5a629188-f221-49b1-aece-4fc336597834",
                "deviceauth": "accepted",
                "hub": "enabled",
            },
            {
                "id": "653e15bc-3797-4eb0-b71e-46383096f512",
                "deviceauth": "accepted",
                "hub": "enabled",
            },
            {
                "id": "917ae5d9-b7ae-4ae4-970e-7b2801ba84c3",
                "deviceauth": "accepted",
                "hub": "enabled",
            },
            {
                "id": "9d3360c0-147e-4018-828f-4331a0e2a8ef",
                "deviceauth": "rejected",
                "hub": "disabled",
            },
            {
                "id": "a43eefc2-2370-468c-b614-710e16a8379f",
                "deviceauth": "accepted",
                "hub": "enabled",
            },
            {
                "id": "a7050cae-be5c-4ea7-87c9-d4b62253f21b",
                "deviceauth": "accepted",
                "hub": "enabled",
            },
            {
                "id": "aba6808c-ca58-40d5-97b7-fba13b68a42d",
                "deviceauth": "rejected",
                "hub": "disabled",
            },
            {
                "id": "b216b66d-c09c-4cb5-9f50-68ea45cdd6f4",
                "deviceauth": "rejected",
                "hub": "disabled",
            },
            {
                "id": "b5be0df7-ddb6-449f-b456-9ecf5dc54de5",
                "deviceauth": "rejected",
                "hub": "disabled",
            },
            {
                "id": "b7221134-6058-4f5e-9d51-ccc7d58d648b",
                "deviceauth": "rejected",
                "hub": "disabled",
            },
            {
                "id": "b9c488af-4c79-4b1f-b1bb-5626ace246de",
                "deviceauth": "accepted",
                "hub": "enabled",
            },
            {
                "id": "c650152a-e7d3-44b0-9ca7-10ec78caaff3",
                "deviceauth": "accepted",
                "hub": "enabled",
            },
            {
                "id": "c65d71e9-cfaf-4d5a-9155-b3b043901b8c",
                "deviceauth": "accepted",
                "hub": "enabled",
            },
            {
                "id": "cdfa43e7-eb2e-4fb0-81e0-93d7d8521cbf",
                "deviceauth": "accepted",
                "hub": "enabled",
            },
            {
                "id": "cf1baca8-49f0-4dc5-91bd-c742292ea497",
                "deviceauth": "accepted",
                "hub": "enabled",
            },
            {
                "id": "e1e686ae-925e-4d5b-a12a-ade8ff11b855",
                "deviceauth": "accepted",
                "hub": "enabled",
            },
            {
                "id": "eb8a74a6-dc6b-4930-95d2-5d1db5da1be0",
                "deviceauth": "accepted",
                "hub": "enabled",
            },
            {
                "id": "fd06ea0c-0ff6-4751-b7a9-44c5ce40e2fe",
                "deviceauth": "rejected",
                "hub": "disabled",
            },
        ],
    }

    expected_requests = [
        {
            "request": {
                "host": "mender-device-auth",
                "method": "GET",
                "path": "/api/internal/v1/devauth/tenants/TestSyncAWSIoTCore01/devices",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/deviceauth_get_devices_TenantSync01.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/7abb6133-ad97-44ba-a159-674242ee565e",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_7abb6133-ad97-44ba-a159-674242ee565e.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/7abb6133-ad97-44ba-a159-674242ee565e/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_inactive_7abb6133-ad97-44ba-a159-674242ee565e.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "DELETE",
                "path": "/things/7abb6133-ad97-44ba-a159-674242ee565e/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_delete_things_principals.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "PUT",
                "path": "/certificates/inactivePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_update_certificate_inactive.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "DELETE",
                "path": "/certificates/inactivePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_delete_certificate.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/7abb6133-ad97-44ba-a159-674242ee565e/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_inactive_7abb6133-ad97-44ba-a159-674242ee565e.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/7abb6133-ad97-44ba-a159-674242ee565e/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_inactive_7abb6133-ad97-44ba-a159-674242ee565e.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/7abb6133-ad97-44ba-a159-674242ee565e/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_inactive_7abb6133-ad97-44ba-a159-674242ee565e.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/7abb6133-ad97-44ba-a159-674242ee565e/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_inactive_7abb6133-ad97-44ba-a159-674242ee565e.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/7abb6133-ad97-44ba-a159-674242ee565e/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_inactive_7abb6133-ad97-44ba-a159-674242ee565e.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/9b2083e4-83b6-41d3-b089-4d2ec137620b",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_not_found_9b2083e4-83b6-41d3-b089-4d2ec137620b.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/1424a387-3431-425b-9f44-1c8eba21812d",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_1424a387-3431-425b-9f44-1c8eba21812d.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/1424a387-3431-425b-9f44-1c8eba21812d/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_inactive_1424a387-3431-425b-9f44-1c8eba21812d.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/inactivePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_inactive.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/1424a387-3431-425b-9f44-1c8eba21812d",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_1424a387-3431-425b-9f44-1c8eba21812d.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/1424a387-3431-425b-9f44-1c8eba21812d/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_inactive_1424a387-3431-425b-9f44-1c8eba21812d.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/inactivePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_inactive.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/inactivePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_inactive.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "PUT",
                "path": "/certificates/inactivePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_update_certificate_active.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/44a66a7e-4b09-4665-b343-0c5d3ad9a2ed",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_44a66a7e-4b09-4665-b343-0c5d3ad9a2ed.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/44a66a7e-4b09-4665-b343-0c5d3ad9a2ed/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_active_44a66a7e-4b09-4665-b343-0c5d3ad9a2ed.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/activePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_active.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/cea63fbb-0b0b-4a92-bd47-8d310a220a16",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_cea63fbb-0b0b-4a92-bd47-8d310a220a16.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/cea63fbb-0b0b-4a92-bd47-8d310a220a16/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_inactive_cea63fbb-0b0b-4a92-bd47-8d310a220a16.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/inactivePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_inactive.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/895a1ee2-1be0-4e78-889a-7d753f53dbd5",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_895a1ee2-1be0-4e78-889a-7d753f53dbd5.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/895a1ee2-1be0-4e78-889a-7d753f53dbd5/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_active_895a1ee2-1be0-4e78-889a-7d753f53dbd5.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/activePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_active.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/8ba0bc48-a669-462e-873f-f7de2f0e04c3",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_8ba0bc48-a669-462e-873f-f7de2f0e04c3.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/8ba0bc48-a669-462e-873f-f7de2f0e04c3/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_inactive_8ba0bc48-a669-462e-873f-f7de2f0e04c3.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/inactivePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_inactive.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/93406e21-8e3f-4435-9786-a294a70298ee",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_not_found_93406e21-8e3f-4435-9786-a294a70298ee.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/966095ec-6bdd-4a76-8498-3c0dffdb9ee2",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_966095ec-6bdd-4a76-8498-3c0dffdb9ee2.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/966095ec-6bdd-4a76-8498-3c0dffdb9ee2/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_active_966095ec-6bdd-4a76-8498-3c0dffdb9ee2.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/activePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_active.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/966095ec-6bdd-4a76-8498-3c0dffdb9ee2",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_966095ec-6bdd-4a76-8498-3c0dffdb9ee2.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/966095ec-6bdd-4a76-8498-3c0dffdb9ee2/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_active_966095ec-6bdd-4a76-8498-3c0dffdb9ee2.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/activePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_active.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/activePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_active.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "PUT",
                "path": "/certificates/activePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_update_certificate_inactive.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/a74ea639-8fae-4b17-ac3e-f9759dbdbfd9",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_a74ea639-8fae-4b17-ac3e-f9759dbdbfd9.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/a74ea639-8fae-4b17-ac3e-f9759dbdbfd9/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_inactive_a74ea639-8fae-4b17-ac3e-f9759dbdbfd9.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/inactivePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_inactive.yml",
            },
        },
        {
            "request": {
                "host": "mender-device-auth",
                "method": "GET",
                "path": "/api/internal/v1/devauth/tenants/TestSyncAWSIoTCore02/devices",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/deviceauth_get_devices_TenantSync02_batch_1.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/1e657abe-ad58-4d20-af7a-3a3449a405e7",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_1e657abe-ad58-4d20-af7a-3a3449a405e7.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/1e657abe-ad58-4d20-af7a-3a3449a405e7/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_inactive_1e657abe-ad58-4d20-af7a-3a3449a405e7.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/inactivePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_inactive.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/43396b5b-ff5c-45f2-ab91-150957f037c9",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_43396b5b-ff5c-45f2-ab91-150957f037c9.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/43396b5b-ff5c-45f2-ab91-150957f037c9/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_inactive_43396b5b-ff5c-45f2-ab91-150957f037c9.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/inactivePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_inactive.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/5a629188-f221-49b1-aece-4fc336597834",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_5a629188-f221-49b1-aece-4fc336597834.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/5a629188-f221-49b1-aece-4fc336597834/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_active_5a629188-f221-49b1-aece-4fc336597834.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/activePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_active.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/653e15bc-3797-4eb0-b71e-46383096f512",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_653e15bc-3797-4eb0-b71e-46383096f512.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/653e15bc-3797-4eb0-b71e-46383096f512/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_active_653e15bc-3797-4eb0-b71e-46383096f512.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/activePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_active.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/917ae5d9-b7ae-4ae4-970e-7b2801ba84c3",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_917ae5d9-b7ae-4ae4-970e-7b2801ba84c3.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/917ae5d9-b7ae-4ae4-970e-7b2801ba84c3/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_active_917ae5d9-b7ae-4ae4-970e-7b2801ba84c3.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/activePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_active.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/9d3360c0-147e-4018-828f-4331a0e2a8ef",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_9d3360c0-147e-4018-828f-4331a0e2a8ef.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/9d3360c0-147e-4018-828f-4331a0e2a8ef/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_inactive_9d3360c0-147e-4018-828f-4331a0e2a8ef.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/inactivePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_inactive.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/a43eefc2-2370-468c-b614-710e16a8379f",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_a43eefc2-2370-468c-b614-710e16a8379f.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/a43eefc2-2370-468c-b614-710e16a8379f/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_active_a43eefc2-2370-468c-b614-710e16a8379f.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/activePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_active.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/a7050cae-be5c-4ea7-87c9-d4b62253f21b",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_a7050cae-be5c-4ea7-87c9-d4b62253f21b.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/a7050cae-be5c-4ea7-87c9-d4b62253f21b/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_active_a7050cae-be5c-4ea7-87c9-d4b62253f21b.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/activePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_active.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/aba6808c-ca58-40d5-97b7-fba13b68a42d",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_aba6808c-ca58-40d5-97b7-fba13b68a42d.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/aba6808c-ca58-40d5-97b7-fba13b68a42d/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_inactive_aba6808c-ca58-40d5-97b7-fba13b68a42d.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/inactivePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_inactive.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/b216b66d-c09c-4cb5-9f50-68ea45cdd6f4",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_b216b66d-c09c-4cb5-9f50-68ea45cdd6f4.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/b216b66d-c09c-4cb5-9f50-68ea45cdd6f4/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_inactive_b216b66d-c09c-4cb5-9f50-68ea45cdd6f4.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/inactivePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_inactive.yml",
            },
        },
        {
            "request": {
                "host": "mender-device-auth",
                "method": "GET",
                "path": "/api/internal/v1/devauth/tenants/TestSyncAWSIoTCore02/devices",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/deviceauth_get_devices_TenantSync02_batch_2.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/b5be0df7-ddb6-449f-b456-9ecf5dc54de5",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_b5be0df7-ddb6-449f-b456-9ecf5dc54de5.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/b5be0df7-ddb6-449f-b456-9ecf5dc54de5/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_inactive_b5be0df7-ddb6-449f-b456-9ecf5dc54de5.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/inactivePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_inactive.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/b7221134-6058-4f5e-9d51-ccc7d58d648b",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_b7221134-6058-4f5e-9d51-ccc7d58d648b.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/b7221134-6058-4f5e-9d51-ccc7d58d648b/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_inactive_b7221134-6058-4f5e-9d51-ccc7d58d648b.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/inactivePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_inactive.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/b9c488af-4c79-4b1f-b1bb-5626ace246de",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_b9c488af-4c79-4b1f-b1bb-5626ace246de.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/b9c488af-4c79-4b1f-b1bb-5626ace246de/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_active_b9c488af-4c79-4b1f-b1bb-5626ace246de.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/activePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_active.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/c650152a-e7d3-44b0-9ca7-10ec78caaff3",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_c650152a-e7d3-44b0-9ca7-10ec78caaff3.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/c650152a-e7d3-44b0-9ca7-10ec78caaff3/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_active_c650152a-e7d3-44b0-9ca7-10ec78caaff3.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/activePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_active.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/c65d71e9-cfaf-4d5a-9155-b3b043901b8c",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_c65d71e9-cfaf-4d5a-9155-b3b043901b8c.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/c65d71e9-cfaf-4d5a-9155-b3b043901b8c/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_active_c65d71e9-cfaf-4d5a-9155-b3b043901b8c.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/activePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_active.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/cdfa43e7-eb2e-4fb0-81e0-93d7d8521cbf",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_cdfa43e7-eb2e-4fb0-81e0-93d7d8521cbf.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/cdfa43e7-eb2e-4fb0-81e0-93d7d8521cbf/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_active_cdfa43e7-eb2e-4fb0-81e0-93d7d8521cbf.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/activePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_active.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/cf1baca8-49f0-4dc5-91bd-c742292ea497",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_cf1baca8-49f0-4dc5-91bd-c742292ea497.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/cf1baca8-49f0-4dc5-91bd-c742292ea497/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_active_cf1baca8-49f0-4dc5-91bd-c742292ea497.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/activePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_active.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/e1e686ae-925e-4d5b-a12a-ade8ff11b855",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_e1e686ae-925e-4d5b-a12a-ade8ff11b855.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/e1e686ae-925e-4d5b-a12a-ade8ff11b855/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_active_e1e686ae-925e-4d5b-a12a-ade8ff11b855.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/activePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_active.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/eb8a74a6-dc6b-4930-95d2-5d1db5da1be0",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_eb8a74a6-dc6b-4930-95d2-5d1db5da1be0.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/eb8a74a6-dc6b-4930-95d2-5d1db5da1be0/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_active_eb8a74a6-dc6b-4930-95d2-5d1db5da1be0.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/activePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_active.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/fd06ea0c-0ff6-4751-b7a9-44c5ce40e2fe",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_thing_fd06ea0c-0ff6-4751-b7a9-44c5ce40e2fe.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/things/fd06ea0c-0ff6-4751-b7a9-44c5ce40e2fe/principals",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_list_things_principals_inactive_fd06ea0c-0ff6-4751-b7a9-44c5ce40e2fe.yml",
            },
        },
        {
            "request": {
                "host": "iot.region.amazonaws.com",
                "method": "GET",
                "path": "/certificates/inactivePrincipalId",
            },
            "result": {
                "match": True,
                "uri": "test_sync_aws_iotcore/iotcore_describe_certificate_inactive.yml",
            },
        },
    ]

    def test_sync(self, clean_mongo, clean_mmock, cli_iot_manager):
        mgo = clean_mongo
        mmock = clean_mmock
        dc = docker.from_env()
        for tenant_id, devices in self.tenant_devices.items():
            client = ManagementAPIClient(tenant_id)
            _, code, hdr = client.register_integration(
                models.Integration(
                    provider="iot-core",
                    credentials=models.Credentials(
                        type="aws",
                        aws=models.AWSCredentialsAws(
                            access_key_id="access_key_id",
                            secret_access_key="secret_access_key",
                            device_policy_name="device-policy",
                            region="region",
                        ),
                    ),
                ),
                _return_http_data_only=False,
            )
            assert code == 201
            assert "Location" in hdr
            location_basename = path.basename(hdr.get("Location"))
            assert location_basename != ""
            integration_id = uuid.UUID(location_basename)

            dev_docs = [
                {
                    "_id": dev["id"],
                    "tenant_id": tenant_id,
                    "integration_ids": [integration_id],
                }
                for dev in devices
            ]
            mgo.iot_manager.devices.insert_many(dev_docs)

        code, output = cli_iot_manager.sync_devices(batch_size=10)
        assert code == 0, output.decode("ascii")

        # You can use the following snippet to generate the list of expected API calls:
        #
        # expected = []
        # for i, match in enumerate(matched_requests):
        #     expected.append(
        #         {
        #             "request": {
        #                 "host": match["request"]["host"],
        #                 "method": match["request"]["method"],
        #                 "path": match["request"]["path"],
        #             },
        #             "result": {"match": True, "uri": match["result"]["uri"],},
        #         }
        #     )
        # from pprint import pprint
        # pprint(expected)

        assert (
            mmock.unmatched == []
        ), "%d requests did match expected request criteria" % len(mmock.unmatched)

        matched_requests = mmock.matched
        assert len(self.expected_requests) == len(matched_requests)
        for i, match in enumerate(matched_requests):
            compare_expectations(self.expected_requests[i], match)
