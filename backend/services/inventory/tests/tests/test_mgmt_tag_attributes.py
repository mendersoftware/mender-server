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

from typing import List

import pytest

import openapi_client as oas
import openapi_client.exceptions as api_exceptions

from client import make_authenticated_client

LIMIT_TAGS = 20


@pytest.mark.usefixtures("clean_db")
class TestTagAttributes:
    @classmethod
    def check_tags(
        cls,
        client: oas.InventoryManagementV1Api,
        device_id: str,
        expected: List[oas.Tag],
    ):
        res = client.get_device_inventory(device_id)
        assert res.attributes is not None
        actual_tags = {
            (tag.name, tag.value.actual_instance)
            for tag in filter(lambda attr: attr.scope == "tags", res.attributes)
        }
        expected_tags = {(tag.name, tag.value) for tag in expected}
        assert actual_tags == expected_tags

    def test_set_tag_attributes_without_etag(self, inventory_attributes):
        internal_client = oas.InventoryInternalV1Api()
        management_client = oas.InventoryManagementV1Api(
            make_authenticated_client(is_device=False)
        )
        did = "some-device-id"
        internal_client.initialize_device(
            tenant_id="",
            device_new=oas.DeviceNew(id=did, attributes=inventory_attributes),
        )
        tags = [oas.Tag(name="n1", value="v_1", description="desc_1")]
        management_client.assign_tags(did, tags)

        TestTagAttributes.check_tags(management_client, did, tags)

    def test_update_tag_attributes_without_etag(self, inventory_attributes):
        internal_client = oas.InventoryInternalV1Api()
        management_client = oas.InventoryManagementV1Api(
            make_authenticated_client(is_device=False)
        )
        did = "some-device-id"
        internal_client.initialize_device(
            tenant_id="",
            device_new=oas.DeviceNew(id=did, attributes=inventory_attributes),
        )
        tags = [
            oas.Tag(name="n_1", value="v_1", description="desc_1"),
            oas.Tag(name="n_2", value="v_2", description="desc_2"),
        ]
        management_client.add_tags(did, tags)

        TestTagAttributes.check_tags(management_client, did, tags)

    def test_replace_tag_attributes_without_etag(self, inventory_attributes):
        internal_client = oas.InventoryInternalV1Api()
        management_client = oas.InventoryManagementV1Api(
            make_authenticated_client(is_device=False)
        )
        did = "some-device-id"
        internal_client.initialize_device(
            tenant_id="",
            device_new=oas.DeviceNew(id=did, attributes=inventory_attributes),
        )
        tags = [
            oas.Tag(name="n_3", value="v_3", description="desc_3"),
        ]
        management_client.assign_tags(did, tags)

        TestTagAttributes.check_tags(management_client, did, tags)

    def test_update_tag_attributes_with_etag(self, inventory_attributes):
        internal_client = oas.InventoryInternalV1Api()
        management_client = oas.InventoryManagementV1Api(
            make_authenticated_client(is_device=False)
        )
        did = "some-device-id"
        internal_client.initialize_device(
            tenant_id="",
            device_new=oas.DeviceNew(id=did, attributes=inventory_attributes),
        )
        tags = [
            oas.Tag(name="n_4", value="v_4", description="desc_4"),
        ]
        management_client.assign_tags(did, tags)

        rsp = management_client.get_device_inventory_with_http_info(id=did)
        assert rsp.headers is not None
        assert "Etag" in rsp.headers
        etag_one = rsp.headers.get("Etag")

        tags_new = [
            oas.Tag(name="n_5", value="v_5", description="desc_5"),
        ]
        management_client.add_tags(did, tags_new, if_match=etag_one)

        rsp = management_client.get_device_inventory_with_http_info(id=did)
        assert rsp.headers is not None
        assert "Etag" in rsp.headers
        etag_two = rsp.headers.get("Etag")
        assert etag_one != etag_two

        assert rsp.data.attributes is not None

        actual_tags = {
            tag.name
            for tag in filter(lambda attr: attr.scope == "tags", rsp.data.attributes)
        }
        expected_tags = {tag.name for tag in tags_new + tags}
        assert expected_tags == actual_tags

    def test_replace_tag_attributes_with_etag(self, inventory_attributes):
        internal_client = oas.InventoryInternalV1Api()
        management_client = oas.InventoryManagementV1Api(
            make_authenticated_client(is_device=False)
        )
        did = "some-device-id"
        internal_client.initialize_device(
            tenant_id="",
            device_new=oas.DeviceNew(id=did, attributes=inventory_attributes),
        )
        tags = [
            oas.Tag(name="n_4", value="v_4", description="desc_4"),
        ]
        management_client.assign_tags(did, tags)

        rsp = management_client.get_device_inventory_with_http_info(id=did)
        assert rsp.headers is not None
        assert "Etag" in rsp.headers
        etag_one = rsp.headers.get("Etag")

        tags_new = [
            oas.Tag(name="n_5", value="v_5", description="desc_5"),
        ]
        management_client.assign_tags(did, tags_new, if_match=etag_one)

        rsp = management_client.get_device_inventory_with_http_info(id=did)
        assert rsp.headers is not None
        assert "Etag" in rsp.headers
        etag_two = rsp.headers.get("Etag")
        assert etag_one != etag_two

        assert rsp.data.attributes is not None

        actual_tags = {
            tag.name
            for tag in filter(lambda attr: attr.scope == "tags", rsp.data.attributes)
        }
        expected_tags = {tag.name for tag in tags_new}
        assert expected_tags == actual_tags

    def test_update_tag_attributes_with_wrong_etag(self, inventory_attributes):
        internal_client = oas.InventoryInternalV1Api()
        management_client = oas.InventoryManagementV1Api(
            make_authenticated_client(is_device=False)
        )
        did = "some-device-id"
        internal_client.initialize_device(
            tenant_id="",
            device_new=oas.DeviceNew(id=did, attributes=inventory_attributes),
        )
        tags = [
            oas.Tag(name="n_5", value="v_5", description="desc_5"),
        ]
        try:
            management_client.add_tags(did, tags, if_match="bad/tag")
        except api_exceptions.ApiException as e:
            assert (
                e.status == 412
            ), "Expected http status code 412 (Precondition Failed)"

    def test_replace_tag_attributes_with_wrong_etag(self, inventory_attributes):
        internal_client = oas.InventoryInternalV1Api()
        management_client = oas.InventoryManagementV1Api(
            make_authenticated_client(is_device=False)
        )
        did = "some-device-id"
        internal_client.initialize_device(
            tenant_id="",
            device_new=oas.DeviceNew(id=did, attributes=inventory_attributes),
        )
        tags = [
            oas.Tag(name="n_5", value="v_5", description="desc_5"),
        ]
        try:
            management_client.assign_tags(did, tags, if_match="bad/tag")
        except api_exceptions.ApiException as e:
            assert (
                e.status == 412
            ), "Expected http status code 412 (Precondition Failed)"

    def test_set_tags_fails_because_of_limits(self, inventory_attributes):
        internal_client = oas.InventoryInternalV1Api()
        management_client = oas.InventoryManagementV1Api(
            make_authenticated_client(is_device=False)
        )
        did = "some-device-id"
        internal_client.initialize_device(
            tenant_id="",
            device_new=oas.DeviceNew(id=did, attributes=inventory_attributes),
        )
        tags_body = [
            oas.Tag(name=f"n_{i}", value=f"v_{i}") for i in range(LIMIT_TAGS + 1)
        ]
        with pytest.raises(api_exceptions.BadRequestException):
            management_client.assign_tags(did, tags_body)
        with pytest.raises(api_exceptions.BadRequestException):
            management_client.add_tags(did, tags_body)

        TestTagAttributes.check_tags(management_client, did, [])
