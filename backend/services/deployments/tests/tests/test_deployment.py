#!/usr/bin/python
# Copyright 2024 Northern.tech AS
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
import io

from uuid import uuid4

import bravado
import requests

from client import (
    SimpleArtifactsClient,
    ArtifactsClientError,
    DeploymentsClient,
    InventoryClient,
    SimpleDeviceClient,
)
from common import (
    artifact_rootfs_from_data,
    Device,
    mongo,
)


class TestDeployment:
    d = DeploymentsClient()

    @staticmethod
    def inventory_add_dev(dev):
        inv = InventoryClient()
        inv.report_attributes(
            dev.fake_token, [{"name": "device_type", "value": dev.device_type}]
        )

    def test_deployments_get(self):
        res = self.d.client.Management_API.List_Deployments(
            Authorization="foo"
        ).result()
        self.d.log.debug("result: %s", res)

        # try with bogus image ID
        try:
            res = self.d.client.Management_API.Show_Deployment(
                Authorization="foo", id="foo"
            ).result()
        except bravado.exception.HTTPError as e:
            assert e.response.status_code == 400
        else:
            raise AssertionError("expected to fail")

    def test_deployments_new_bogus(self):

        # NOTE: cannot make requests with arbitary data through swagger client,
        # so we'll use requests directly instead
        rsp = requests.post(self.d.make_api_url("/deployments"), data="foobar")
        assert 400 <= rsp.status_code < 500
        # some broken JSON now
        rsp = requests.post(
            self.d.make_api_url("/deployments"),
            data='{"foo": }',
            headers={"Content-Type": "application/json"},
        )
        assert 400 <= rsp.status_code < 500

        baddeps = [
            self.d.make_new_deployment(
                name="foobar", artifact_name="someartifact", devices=[]
            ),
            self.d.make_new_deployment(
                name="", artifact_name="someartifact", devices=["foo"]
            ),
            self.d.make_new_deployment(name="adad", artifact_name="", devices=["foo"]),
            self.d.make_new_deployment(name="", artifact_name="", devices=["foo"]),
        ]
        for newdep in baddeps:
            # try bogus image data
            try:
                res = self.d.client.Management_API.Create_Deployment(
                    Authorization="foo", deployment=newdep
                ).result()
            except bravado.exception.HTTPError as e:
                assert e.response.status_code == 400
            else:
                raise AssertionError("expected to fail")

    def test_deployments_new_valid(self):
        """Add a new valid deployment, verify its status, verify device deployment
        status, abort and verify eveything once again"""
        dev = Device()

        self.d.log.info("fake device with ID: %s", dev.devid)

        self.inventory_add_dev(dev)

        data = b"foo_bar"
        artifact_name = "hammer-update-" + str(uuid4())
        # come up with an artifact
        with artifact_rootfs_from_data(
            name=artifact_name, data=data, devicetype=dev.device_type
        ) as art:
            ac = SimpleArtifactsClient()
            artid = ac.add_artifact(
                description="some description", size=art.size, data=art
            )

            newdep = self.d.make_new_deployment(
                name="fake deployment", artifact_name=artifact_name, devices=[dev.devid]
            )
            depid = self.d.add_deployment(newdep)

            # artifact is used in deployment, so attempts to remove it should
            # fail
            try:
                ac.delete_artifact(artid)
            except ArtifactsClientError as ace:
                #  artifact is used in deployment
                assert ace.response.status_code == 409
            else:
                raise AssertionError("expected to fail")

            dc = SimpleDeviceClient()
            nextdep = dc.get_next_deployment(
                dev.fake_token,
                artifact_name="different {}".format(artifact_name),
                device_type=dev.device_type,
            )

            dep = self.d.client.Management_API.Show_Deployment(
                Authorization="foo", id=depid
            ).result()[0]
            assert dep.artifact_name == artifact_name
            assert dep.id == depid
            assert dep.status == "pending"

            # fetch device status
            depdevs = self.d.client.Management_API.List_All_Devices_in_Deployment(
                Authorization="foo", deployment_id=depid
            ).result()[0]
            assert len(depdevs) == 1
            depdev = depdevs[0]
            assert depdev.status == "pending"
            assert depdev.id == dev.devid

            # verify statistics
            self.d.verify_deployment_stats(depid, expected={"pending": 1})

            # abort deployment
            self.d.abort_deployment(depid)

            # that it's 'finished' now
            aborted_dep = self.d.client.Management_API.Show_Deployment(
                Authorization="foo", id=depid
            ).result()[0]
            self.d.log.debug("deployment dep: %s", aborted_dep)
            assert aborted_dep.status == "finished"

            # verify statistics once again
            self.d.verify_deployment_stats(depid, expected={"aborted": 1})

            # fetch device status
            depdevs = self.d.client.Management_API.List_All_Devices_in_Deployment(
                Authorization="foo", deployment_id=depid
            ).result()[0]
            self.d.log.debug("deployment devices: %s", depdevs)
            assert len(depdevs) == 1
            depdev = depdevs[0]
            assert depdev.status == "aborted"

            # deleting artifact should succeed
            ac.delete_artifact(artid)

    def test_deployments_conflict(self):
        """Add a new valid deployment, verify its status, verify a new identical deployment cannot be created"""
        dev = Device()
        self.d.log.info("fake device with ID: %s", dev.devid)
        self.inventory_add_dev(dev)

        data = b"foo_bar"
        artifact_name = "hammer-update-" + str(uuid4())

        # create an artifact
        with artifact_rootfs_from_data(
            name=artifact_name, data=data, devicetype=dev.device_type
        ) as art:
            ac = SimpleArtifactsClient()
            ac.add_artifact(description="some description", size=art.size, data=art)
            newdep = self.d.make_new_deployment(
                name="fake deployment", artifact_name=artifact_name, devices=[dev.devid]
            )

            # create the deployment
            self.d.add_deployment(newdep)

            # creating the same deployment again will fail
            try:
                self.d.add_deployment(newdep)
            except bravado.exception.HTTPError as err:
                assert err.response.status_code == 409
            else:
                raise AssertionError("expected to fail")

    def test_single_device_deployment_device_not_in_inventory(self):
        """Try to create single device deployment for a device wich is not
        in the inventory"""
        dev = Device()

        self.d.log.info("fake device with ID: %s", dev.devid)

        data = b"foo_bar"
        artifact_name = "hammer-update-" + str(uuid4())
        # come up with an artifact
        with artifact_rootfs_from_data(
            name=artifact_name, data=data, devicetype=dev.device_type
        ) as art:
            ac = SimpleArtifactsClient()
            artid = ac.add_artifact(
                description="some description", size=art.size, data=art
            )

            newdep = self.d.make_new_deployment(
                name="fake deployment", artifact_name=artifact_name, devices=[dev.devid]
            )
            depid = self.d.add_deployment(newdep)

    def test_deployments_new_no_artifact(self):
        """Try to add deployment without an artifact, verify that it failed with 422"""
        dev = Device()

        self.d.log.info("fake device with ID: %s", dev.devid)

        self.inventory_add_dev(dev)

        artifact_name = "no-artifact " + str(uuid4())
        # come up with an artifact
        newdep = self.d.make_new_deployment(
            name="fake deployment", artifact_name=artifact_name, devices=[dev.devid]
        )
        try:
            self.d.add_deployment(newdep)
        except bravado.exception.HTTPError as err:
            assert err.response.status_code == 422
        else:
            raise AssertionError("expected to fail")

    def test_deplyments_get_devices(self):
        """Create deployments, get devices with pagination"""
        devices = []
        devices_qty = 30
        device_ids = []
        default_per_page = 20
        device_type = "test-hammer-type"

        # create devices
        for i in range(0, devices_qty):
            device = Device(device_type)
            self.inventory_add_dev(device)
            devices.append(device)
            device_ids.append(device.devid)

        data = b"foo_bar"
        artifact_name = "pagination-test-" + str(uuid4())
        # come up with an artifact
        with artifact_rootfs_from_data(
            name=artifact_name, data=data, devicetype=device_type
        ) as art:
            ac = SimpleArtifactsClient()
            ac.add_artifact(description="some description", size=art.size, data=art)

            new_dep = self.d.make_new_deployment(
                name="pagination deployment",
                artifact_name=artifact_name,
                devices=device_ids,
            )
            dep_id = self.d.add_deployment(new_dep)

            for dev in devices:
                dc = SimpleDeviceClient()
                dc.get_next_deployment(
                    dev.fake_token,
                    artifact_name="different {}".format(artifact_name),
                    device_type=dev.device_type,
                )

            # check default 'page' and 'per_page' values
            res = self.d.client.Management_API.List_Devices_in_Deployment(
                Authorization="foo", deployment_id=dep_id
            ).result()[0]
            assert len(res) == default_per_page

            # check custom 'per_page'
            res = self.d.client.Management_API.List_Devices_in_Deployment(
                Authorization="foo", deployment_id=dep_id, per_page=devices_qty
            ).result()[0]
            assert len(res) == devices_qty

            # check 2nd page
            devices_qty_on_second_page = devices_qty - default_per_page
            res = self.d.client.Management_API.List_Devices_in_Deployment(
                Authorization="foo",
                deployment_id=dep_id,
                page=2,
                per_page=default_per_page,
            ).result()[0]
            assert len(res) == devices_qty_on_second_page

    def test_device_deployments_simple(self, mongo):
        """Check that device can get next deployment, simple cases:
        - bogus token
        - valid update
        - device type incompatible with artifact
        """
        dev = Device()

        self.d.log.info("fake device with ID: %s", dev.devid)

        self.inventory_add_dev(dev)

        data = b"foo_bar"
        artifact_name = "hammer-update-" + str(uuid4())
        # come up with an artifact
        with artifact_rootfs_from_data(
            name=artifact_name, data=data, devicetype=dev.device_type
        ) as art:
            ac = SimpleArtifactsClient()
            with ac.with_added_artifact(
                description="desc", size=art.size, data=art
            ) as artid:

                newdep = self.d.make_new_deployment(
                    name="foo", artifact_name=artifact_name, devices=[dev.devid]
                )

                with self.d.with_added_deployment(newdep) as depid:
                    dc = SimpleDeviceClient()
                    self.d.log.debug("device token %s", dev.fake_token)

                    # try with some bogus token
                    try:
                        dc.get_next_deployment(
                            "foo-bar-baz",
                            artifact_name=artifact_name,
                            device_type=dev.device_type,
                        )
                    except bravado.exception.HTTPError as err:
                        assert 400 <= err.response.status_code < 500
                    else:
                        raise AssertionError("expected to fail")

                    # pretend we have another artifact installed
                    nextdep = dc.get_next_deployment(
                        dev.fake_token,
                        artifact_name="different {}".format(artifact_name),
                        device_type=dev.device_type,
                    )
                    self.d.log.info("device next: %s", nextdep)
                    assert nextdep

                    assert (
                        dev.device_type in nextdep.artifact["device_types_compatible"]
                    )

                    try:
                        # pretend our device type is different than expected
                        nextdep = dc.get_next_deployment(
                            dev.fake_token,
                            artifact_name="different {}".format(artifact_name),
                            device_type="other {}".format(dev.device_type),
                        )
                    except bravado.exception.HTTPError as err:
                        assert err.response.status_code == 409
                    else:
                        raise AssertionError("expected to fail")

                    # verify that device status was properly set
                    self.d.verify_deployment_stats(depid, expected={"failure": 1})
        last_device_deployment_status = mongo[
            "deployment_service"
        ].devices_last_status.find_one({"_id": dev.devid})
        assert last_device_deployment_status["_id"] == dev.devid
        assert last_device_deployment_status["device_deployment_status"] == 256

    def test_device_deployments_already_installed(self, mongo):
        """Check case with already installed artifact"""
        dev = Device()

        self.d.log.info("fake device with ID: %s", dev.devid)

        self.inventory_add_dev(dev)

        data = b"foo_bar"
        artifact_name = "hammer-update-" + str(uuid4())
        # come up with an artifact
        with artifact_rootfs_from_data(
            name=artifact_name, data=data, devicetype=dev.device_type
        ) as art:
            ac = SimpleArtifactsClient()
            with ac.with_added_artifact(
                description="desc", size=art.size, data=art
            ) as artid:

                newdep = self.d.make_new_deployment(
                    name="foo", artifact_name=artifact_name, devices=[dev.devid]
                )

                with self.d.with_added_deployment(newdep) as depid:
                    dc = SimpleDeviceClient()
                    self.d.log.debug("device token %s", dev.fake_token)

                    # pretend we have the same artifact installed already
                    # NOTE: asking for a deployment while having it already
                    # installed is special in the sense that the status of
                    # deployment for this device will be marked as 'already-installed'
                    nextdep = dc.get_next_deployment(
                        dev.fake_token,
                        artifact_name=artifact_name,
                        device_type=dev.device_type,
                    )
                    self.d.log.info("device next: %s", nextdep)
                    assert nextdep == None
                    # verify that device status was properly recorded
                    self.d.verify_deployment_stats(
                        depid, expected={"already-installed": 1}
                    )
        last_device_deployment_status = mongo[
            "deployment_service"
        ].devices_last_status.find_one({"_id": dev.devid})
        assert last_device_deployment_status["_id"] == dev.devid
        assert last_device_deployment_status["device_deployment_status"] == 3072

    def test_device_deployments_full(self, mongo):
        """Check that device can get next deployment, full cycle"""
        dev = Device()

        self.d.log.info("fake device with ID: %s", dev.devid)

        self.inventory_add_dev(dev)

        data = b"foo_bar"
        artifact_name = "hammer-update-" + str(uuid4())
        # come up with an artifact
        with artifact_rootfs_from_data(
            name=artifact_name, data=data, devicetype=dev.device_type
        ) as art:
            ac = SimpleArtifactsClient()
            with ac.with_added_artifact(
                description="desc", size=art.size, data=art
            ) as artid:

                newdep = self.d.make_new_deployment(
                    name="foo", artifact_name=artifact_name, devices=[dev.devid]
                )

                with self.d.with_added_deployment(newdep) as depid:
                    dc = SimpleDeviceClient()
                    self.d.log.debug("device token %s", dev.fake_token)

                    # pretend we have another artifact installed
                    nextdep = dc.get_next_deployment(
                        dev.fake_token,
                        artifact_name="different {}".format(artifact_name),
                        device_type=dev.device_type,
                    )
                    self.d.log.info("device next: %s", nextdep)
                    assert nextdep

                    assert (
                        dev.device_type in nextdep.artifact["device_types_compatible"]
                    )

                    self.d.verify_deployment_stats(depid, expected={"pending": 1})

                    for st in [
                        "downloading",
                        "pause_before_installing",
                        "installing",
                        "pause_before_committing",
                        "pause_before_rebooting",
                        "rebooting",
                    ]:
                        dc.report_status(
                            token=dev.fake_token, devdepid=nextdep.id, status=st
                        )
                        self.d.verify_deployment_stats(depid, expected={st: 1})

                    # we have reported some statuses now, but not the final
                    # status, try to get the next deployment
                    againdep = dc.get_next_deployment(
                        dev.fake_token,
                        artifact_name="different {}".format(artifact_name),
                        device_type=dev.device_type,
                    )
                    assert againdep
                    assert againdep.id == nextdep.id

                    # deployment should be marked as inprogress
                    dep = self.d.client.Management_API.Show_Deployment(
                        Authorization="foo", id=depid
                    ).result()[0]
                    assert dep.status == "inprogress"

                    # report final status
                    dc.report_status(
                        token=dev.fake_token, devdepid=nextdep.id, status="success"
                    )
                    self.d.verify_deployment_stats(depid, expected={"success": 1})

                    dep = self.d.client.Management_API.Show_Deployment(
                        Authorization="foo", id=depid
                    ).result()[0]
                    assert dep.status == "finished"

                    # report failure as final status
                    dc.report_status(
                        token=dev.fake_token, devdepid=nextdep.id, status="failure"
                    )
                    self.d.verify_deployment_stats(depid, expected={"failure": 1})

                    # deployment is finished, should get no more updates
                    nodep = dc.get_next_deployment(
                        dev.fake_token,
                        artifact_name="other {}".format(artifact_name),
                        device_type=dev.device_type,
                    )
                    assert nodep == None

                    # TODO: check this path; it looks like something which shouldn't be possible;
                    # TODO: update the test after verifying or fixing deployments service behavior
                    # as a joke, report rebooting now
                    # dc.report_status(
                    #    token=dev.fake_token, devdepid=nextdep.id, status="rebooting"
                    # )
                    # self.d.verify_deployment_stats(depid, expected={"rebooting": 1})
                    # deployment is still finished
                    # dep = self.d.client.Management_API.Show_Deployment(
                    #    Authorization="foo", id=depid
                    # ).result()[0]
                    # assert dep.status == "finished"

                    # go on, let's pretend that the artifact is already installed
                    # nodep = dc.get_next_deployment(
                    #    dev.fake_token,
                    #    artifact_name=artifact_name,
                    #    device_type=dev.device_type,
                    # )
                    # assert nodep == None
                    # self.d.verify_deployment_stats(
                    #    depid, expected={"already-installed": 1}
                    # )
        last_device_deployment_status = mongo[
            "deployment_service"
        ].devices_last_status.find_one({"_id": dev.devid})
        assert last_device_deployment_status["_id"] == dev.devid
        assert last_device_deployment_status["device_deployment_status"] == 256

    def test_device_deployments_logs(self):
        """Check that device can get next deployment, full cycle"""
        dev = Device()

        self.d.log.info("fake device with ID: %s", dev.devid)

        self.inventory_add_dev(dev)

        data = b"foo_bar"
        artifact_name = "hammer-update-" + str(uuid4())
        # come up with an artifact
        with artifact_rootfs_from_data(
            name=artifact_name, data=data, devicetype=dev.device_type
        ) as art:
            ac = SimpleArtifactsClient()
            with ac.with_added_artifact(
                description="desc", size=art.size, data=art
            ) as artid:

                newdep = self.d.make_new_deployment(
                    name="foo", artifact_name=artifact_name, devices=[dev.devid]
                )

                with self.d.with_added_deployment(newdep) as depid:
                    dc = SimpleDeviceClient()
                    self.d.log.debug("device token %s", dev.fake_token)

                    # pretend we have another artifact installed
                    nextdep = dc.get_next_deployment(
                        dev.fake_token,
                        artifact_name="different {}".format(artifact_name),
                        device_type=dev.device_type,
                    )
                    self.d.log.info("device next: %s", nextdep)
                    assert nextdep

                    dc.upload_logs(
                        dev.fake_token, nextdep.id, logs=["foo bar baz", "lorem ipsum"]
                    )

                    rsp = self.d.client.Management_API.Get_Deployment_Log_for_Device(
                        Authorization="foo", deployment_id=depid, device_id=dev.devid
                    ).result()[1]
                    logs = rsp.text
                    self.d.log.info("device logs\n%s", logs)

                    assert "lorem ipsum" in logs
                    assert "foo bar baz" in logs
