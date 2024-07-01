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
import logging


def pytest_addoption(parser):
    parser.addoption(
        "--api",
        action="store",
        default="0.1.0",
        help="API version used in HTTP requests",
    )
    parser.addoption(
        "--host", action="store", default="inventory", help="host running API"
    )
    parser.addoption(
        "--mongo-url",
        action="store",
        default="mongodb://mongo",
        help="Mongo URL (connection string)",
    )
    parser.addoption(
        "--devices", action="store", default="1001", help="# of devices to test with"
    )
    parser.addoption(
        "--management-spec",
        action="store",
        default="management_api.yml.",
        help="management API spec",
    )
    parser.addoption(
        "--management-v2-spec",
        action="store",
        default="management_api_v2.yml.",
        help="management API v2 spec",
    )
    parser.addoption("--internal-spec", default="../docs/internal_api.yml")
    parser.addoption(
        "--inventory-items",
        action="store",
        default="inventory_items",
        help="file with inventory items",
    )


def pytest_configure(config):
    api_version = config.getoption("api")
    host = config.getoption("host")
    test_device_count = int(config.getoption("devices"))
    lvl = logging.INFO
    if config.getoption("verbose"):
        lvl = logging.DEBUG
    logging.basicConfig(level=lvl)
    # configure bravado related loggers to be less verbose
    logging.getLogger("swagger_spec_validator").setLevel(logging.INFO)
    logging.getLogger("bravado_core").setLevel(logging.INFO)
