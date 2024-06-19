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

import pytest
import uuid
import time

import msgpack

import management_api

from common import Device, management_api_with_params, management_api_connect


@pytest.mark.usefixtures("timeout")
class _TestConnect:
    def test_connect(self, clean_mongo, tenant_id=None):
        """
        Tests end-to-end connection between user and device and checks
        for basic expected responses.
        """

        dev = Device(tenant_id=tenant_id)
        user_id = str(uuid.uuid4())
        api_mgmt = management_api_with_params(user_id=user_id, tenant_id=tenant_id)
        try:
            api_mgmt.connect(
                "00000000-0000-0000-0000-000000000000",
                connection="Upgrade",
                upgrade="websocket",
                sec_websocket_key="dGhlIHNhbXBsZSBub25jZQ==",
                sec_websocket_version=13,
            )
        except management_api.ApiException as e:
            assert e.status == 404
            assert "device not found" in str(e.body)
        else:
            raise Exception("Expected status code 404")

        try:
            api_mgmt.connect(
                dev.id,
                connection="Upgrade",
                upgrade="websocket",
                sec_websocket_key="dGhlIHNhbXBsZSBub25jZQ==",
                sec_websocket_version=13,
            )
        except management_api.ApiException as e:
            assert e.status == 404
            assert "device not connected" in str(e.body)
        else:
            raise Exception("Expected status code 404")

        with dev.connect() as dev_conn:
            for i in range(15):
                obj = api_mgmt.get_device(dev.id)
                if obj.status == 'connected':
                    break
                time.sleep(1)
            else:
                raise Exception("Device not connecting to deviceconnect", obj)

            try:
                api_mgmt.connect(dev.id)
            except management_api.ApiException as e:
                assert e.status == 400
                assert "the client is not using the websocket protocol" in e.body
            else:
                raise Exception("Expected status code 400")

            try:
                api_mgmt.connect(
                    dev.id,
                    connection="Upgrade",
                    upgrade="websocket",
                    sec_websocket_key="dGhlIHNhbXBsZSBub25jZQ==",
                    sec_websocket_version=13,
                )
            except management_api.ApiException as e:
                assert e.status == 101
            else:
                raise Exception("Expected status code 101")

            with management_api_connect(
                dev.id, user_id=user_id, tenant_id=tenant_id
            ) as user_conn:
                user_conn.send(
                    msgpack.dumps(
                        {
                            "hdr": {
                                "proto": 1,
                                "typ": "start",
                                "props": {"status": "ok"},
                            },
                        }
                    )
                )
                msg = dev_conn.recv()
                rsp = msgpack.loads(msg)
                assert "hdr" in rsp, "Message does not contain header"
                assert (
                    "sid" in rsp["hdr"]
                ), "Forwarded message should contain session ID"
                assert rsp == {
                    "hdr": {
                        "proto": 1,
                        "typ": "start",
                        "props": {
                            "status": "ok",
                            "user_id": user_id,
                        },
                        "sid": rsp["hdr"]["sid"],
                    },
                }
                dev_conn.send(
                    msgpack.dumps(
                        {
                            "hdr": {
                                "proto": 1,
                                "typ": "shell",
                                "props": {"status": "ok"},
                                "sid": rsp["hdr"]["sid"],
                            },
                            "body": b"sh-5.0$ ",
                        }
                    )
                )
                msg = user_conn.recv()
                assert msgpack.loads(msg) == {
                    "hdr": {
                        "proto": 1,
                        "typ": "shell",
                        "props": {"status": "ok"},
                        "sid": rsp["hdr"]["sid"],
                    },
                    "body": b"sh-5.0$ ",
                }

        for i in range(15):
            obj = api_mgmt.get_device(dev.id)
            if obj.status != 'connected':
                break
            time.sleep(1)

        try:
            api_mgmt.connect(
                dev.id,
                connection="Upgrade",
                upgrade="websocket",
                sec_websocket_key="dGhlIHNhbXBsZSBub25jZQ==",
                sec_websocket_version=13,
            )
        except management_api.ApiException as e:
            assert e.status == 404
        else:
            raise Exception("Expected status code 404")


class TestConnect(_TestConnect):
    def test_connect(self, clean_mongo):
        super().test_connect(clean_mongo)


class TestConnectEnterprise(_TestConnect):
    def test_connect(self, clean_mongo, tenant):
        super().test_connect(clean_mongo, tenant)
