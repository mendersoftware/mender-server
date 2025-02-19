#!/usr/bin/python
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
import logging
import devices_api
import internal_api
import management_api


def pytest_addoption(parser):
    parser.addoption(
        "--host", action="store", default="deviceauth", help="host running API"
    )
    parser.addoption(
        "--mongo-url",
        default="mongodb://mongo",
        help="The MongoDB URL (connection string)",
    )
    parser.addoption("--spec", default="../docs/internal_api.yml")
    parser.addoption("--management-spec", default="../docs/management_api.yml")


def pytest_configure(config):
    lvl = logging.INFO
    if config.getoption("verbose"):
        lvl = logging.DEBUG
    logging.basicConfig(level=lvl)
    host = config.getoption("host")
    devices_api.Configuration.set_default(
        devices_api.Configuration(
            host="http://" + host + "/api/devices/v1/authentication"
        )
    )
    internal_api.Configuration.set_default(
        internal_api.Configuration(host="http://" + host + "/api/internal/v1/devauth")
    )
    management_api.Configuration.set_default(
        management_api.Configuration(
            host="http://" + host + "/api/management/v2/devauth"
        )
    )
