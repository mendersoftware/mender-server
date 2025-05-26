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

import openapi_client as oas


def pytest_addoption(parser):
    parser.addoption(
        "--host", action="store", default="inventory", help="host running API"
    )
    parser.addoption(
        "--mongo-url",
        action="store",
        default="mongodb://mongo",
        help="Mongo URL (connection string)",
    )


def pytest_configure(config):
    lvl = logging.INFO
    host = config.getoption("host")
    oas.Configuration.set_default(oas.Configuration(host=f"http://{host}"))

    # Setup default tokens for Api Clients

    if config.getoption("verbose"):
        lvl = logging.DEBUG
    logging.basicConfig(level=lvl)
