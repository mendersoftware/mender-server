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

import os

import mender_client


def pytest_addoption(parser):
    parser.addoption(
        "--host",
        action="store",
        default=os.environ["TEST_HOST"] if "TEST_HOST" in os.environ else "localhost",
        help="Address for host hosting iot-manager API (env: TEST_HOST)",
    )


def pytest_configure(config):
    host = config.getoption("host")
    mender_client.Configuration.set_default(
        mender_client.Configuration(host="http://" + host)
    )
