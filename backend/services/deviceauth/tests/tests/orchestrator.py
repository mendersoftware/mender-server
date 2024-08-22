#!/usr/bin/python3
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
import json
import os
import logging

from contextlib import contextmanager

import mockserver


def provision_device_handler(device_id=None, status=200):
    log = logging.getLogger("orchestartor.provision_device")

    def _provision_device(request):
        payload = json.loads(request.body.decode())
        if device_id is not None:
            assert payload.get("device_id", None) == device_id
        else:
            assert "device_id" in payload

        return (status, {}, "")

    return _provision_device


def decommission_device_handler(device_id=None, status=200):
    log = logging.getLogger("orchestartor.decommision_device")

    def _decommission_device(request):
        dreq = json.loads(request.body.decode())
        print("decommision request", dreq)
        # verify that devauth tries to decommision correct device
        assert dreq.get("device_id", None) == device_id
        # test is enforcing particular request ID
        assert dreq.get("request_id", None) == "delete_device"
        return (status, {}, "")

    return _decommission_device


def update_device_status_handler(device_id=None, status=200):
    log = logging.getLogger("orchestartor.update_device_status")

    def _update_device_status(request):
        dreq = json.loads(request.body.decode())
        print("update_device_status request", dreq)
        return (status, {}, "")

    return _update_device_status


def update_device_inventory_handler(device_id=None, status=200):
    log = logging.getLogger("orchestartor.update_device_inventory")

    def _update_device_inventory(request):
        dreq = json.loads(request.body.decode())
        print("update_device_inventory request", dreq)
        return (status, {}, "")

    return _update_device_inventory


def reindex_reporting(device_id=None, status=200):
    log = logging.getLogger("orchestartor.reindex_reporting")

    def _reindex_reporting(request):
        dreq = json.loads(request.body.decode())
        print("reindex_reporting request", dreq)
        return (status, {}, "")

    return _reindex_reporting


def get_fake_orchestrator_addr():
    return os.environ.get("FAKE_ORCHESTRATOR_ADDR", "0.0.0.0:9998")


ANY_DEVICE = None


@contextmanager
def run_fake_for_device_id(devid, status=None):
    if status is None:
        handlers = [
            (
                "POST",
                "/api/v1/workflow/provision_device",
                provision_device_handler(devid),
            ),
            (
                "POST",
                "/api/v1/workflow/decommission_device",
                decommission_device_handler(devid),
            ),
            (
                "POST",
                "/api/v1/workflow/update_device_status",
                update_device_status_handler(devid),
            ),
            (
                "POST",
                "/api/v1/workflow/update_device_inventory",
                update_device_inventory_handler(devid),
            ),
            (
                "POST",
                "/api/v1/workflow/reindex_reporting",
                reindex_reporting(devid),
            ),
        ]
    else:
        handlers = [
            (
                "POST",
                "/api/v1/workflow/provision_device",
                provision_device_handler(devid, status),
            ),
            (
                "POST",
                "/api/v1/workflow/decommission_device",
                decommission_device_handler(devid, status),
            ),
            (
                "POST",
                "/api/v1/workflow/update_device_status",
                update_device_status_handler(devid, status),
            ),
            (
                "POST",
                "/api/v1/workflow/update_device_inventory",
                update_device_inventory_handler(devid, status),
            ),
            (
                "POST",
                "/api/v1/workflow/reindex_reporting",
                reindex_reporting(devid, status),
            ),
        ]

    with mockserver.run_fake(get_fake_orchestrator_addr(), handlers=handlers) as server:
        yield server
