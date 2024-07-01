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
import os

from config import init


def pytest_addoption(parser):
    parser.addoption(
        "--host", action="store", default="localhost", help="host running API"
    )
    parser.addoption(
        "--inventory-host",
        action="store",
        default="mender-inventory:8080",
        help="host running API",
    )
    parser.addoption("--spec", default="../docs/management_api.yml")
    parser.addoption("--device-spec", default="../docs/devices_api.yml")
    parser.addoption("--internal-spec", default="../docs/internal_api.yml")
    parser.addoption("--mongo-url", default="mongodb://mongo", help="Mongo URL (Connection string)")
    parser.addoption(
        "--s3-bucket",
        default=os.environ.get("AWS_S3_BUCKET_NAME", "mender-artifact-storage"),
        help="The s3 bucket name",
    )
    parser.addoption(
        "--s3-key-id",
        default=os.environ.get("AWS_ACCESS_KEY_ID", "mender"),
        help="The key ID for s3",
    )
    parser.addoption(
        "--s3-secret-key",
        default=os.environ.get("AWS_SECRET_ACCESS_KEY", "correcthorsebatterystaple"),
        help="The access key secret for s3",
    )
    parser.addoption(
        "--s3-endpoint-url",
        default=os.environ.get("AWS_ENDPOINT_URL", "http://s3.mender.local:8080"),
        help="The endpoint URL for s3",
    )


def pytest_configure(config):
    lvl = logging.INFO
    if config.getoption("verbose"):
        lvl = logging.DEBUG
    logging.basicConfig(level=lvl)
    # configure bravado related loggers to be less verbose
    logging.getLogger("swagger_spec_validator").setLevel(logging.INFO)
    logging.getLogger("bravado_core").setLevel(logging.INFO)

    # capture global pytest cmdline config
    init(config)
