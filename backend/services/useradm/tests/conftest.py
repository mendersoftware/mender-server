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

import internal_v1
import management_v1


def pytest_addoption(parser):
    parser.addoption(
        "--host", action="store", default="useradm", help="host running API"
    )
    parser.addoption(
        "--mongo-url", default="tenantadm", help="Mongo URL (connection string)"
    )


def pytest_configure(config):
    lvl = logging.INFO
    if config.getoption("verbose"):
        lvl = logging.DEBUG
    logging.basicConfig(level=lvl)

    host = config.getoption("host")

    # Configure generated API clients
    internal_v1.Configuration.set_default(
        internal_v1.Configuration(host="http://" + host)
    )
    management_v1.Configuration.set_default(
        management_v1.Configuration(host="http://" + host)
    )
