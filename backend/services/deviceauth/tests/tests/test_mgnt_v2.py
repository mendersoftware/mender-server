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
import bravado

from common import (
    Device,
    DevAuthorizer,
    device_auth_req,
    make_devices,
    devices,
    clean_migrated_db,
    clean_db,
    mongo,
    cli,
    management_api,
    internal_api,
    device_api,
    get_keypair,
)

import orchestrator


class TestDeleteDevice:
    def test_delete_device(self, management_api, internal_api, devices):
        # try delete an existing device, verify decommissioning workflow was started
        # setup single device and poke devauth
        dev, _ = devices[0]
        ourdev = management_api.get_single_device()
        assert ourdev

        try:
            with orchestrator.run_fake_for_device_id(ourdev.id):
                rsp = management_api.decommission_device(
                    ourdev.id,
                    {
                        "X-MEN-RequestID": "delete_device",
                        "Authorization": "Bearer foobar",
                    },
                )
            print("decommission request finished with status:", rsp.status_code)
        except bravado.exception.HTTPError as e:
            assert e.response.status_code == 204

        try:
            with orchestrator.run_fake_for_device_id(ourdev.id):
                rsp = internal_api.delete_device(
                    ourdev.id,
                    headers={
                        "X-MEN-RequestID": "delete_device",
                        "Authorization": "Bearer foobar",
                    },
                )
            print("delete request finished with status:", rsp.status_code)
        except bravado.exception.HTTPError as e:
            assert e.response.status_code == 204

        found = None
        status_code = None
        try:
            found = management_api.get_device(id=ourdev.id)
        except bravado.exception.HTTPError as e:
            status_code = e.response.status_code

        assert status_code == 404
        assert not found

    def test_delete_device_ochestrator_failure(self, management_api, devices):
        # try delete an existing device, verify it is failing when orchestrator
        # is failing
        dev, _ = devices[0]
        ourdev = management_api.get_single_device()
        assert ourdev

        with orchestrator.run_fake_for_device_id(ourdev.id, 500) as server:
            rsp = management_api.decommission_device(
                ourdev.id,
                {
                    "X-MEN-RequestID": "delete_device",
                    "Authorization": "Bearer foobar",
                },
            )
        print("decommission request finished with status:", rsp.status_code)
        assert rsp.status_code == 500

    def test_delete_device_nonexistent(self, management_api):
        # try delete a nonexistent device
        rsp = management_api.decommission_device("some-devid-foo")
        assert rsp.status_code == 404

    def test_device_accept_reject_cycle(self, devices, device_api, management_api):
        d, da = devices[0]
        url = device_api.auth_requests_url

        dev = management_api.get_single_device()

        assert dev
        devid = dev.id

        print("found device with ID:", dev.id)
        aid = dev.auth_sets[0].id

        with orchestrator.run_fake_for_device_id(devid) as server:
            _, rsp = management_api.accept_device(devid, aid)
            assert rsp.status_code == 204

            # device is accepted, we should get a token now
            rsp = device_auth_req(url, da, d)
            assert rsp.status_code == 200

            da.parse_rsp_payload(d, rsp.text)

            assert len(d.token) > 0

            # reject it now
            _, rsp = management_api.reject_device(devid, aid)
            print("RSP:", rsp)
            assert rsp.status_code == 204

            # device is rejected, should get unauthorized
            rsp = device_auth_req(url, da, d)
            assert rsp.status_code == 401

    def test_device_accept_orchestrator_failure(
        self, devices, device_api, management_api
    ):
        d, da = devices[0]
        url = device_api.auth_requests_url

        dev = management_api.get_single_device()

        assert dev
        devid = dev.id

        print("found device with ID:", dev.id)
        aid = dev.auth_sets[0].id

        status = None
        try:
            with orchestrator.run_fake_for_device_id(devid, 500) as server:
                management_api.accept_device(devid, aid)
        except bravado.exception.HTTPError as e:
            status = e.response.status_code
        assert status == 500
