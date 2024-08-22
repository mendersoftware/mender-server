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
import io
import tempfile
import subprocess
import logging
import os
import abc
import random
import string
import json
import pytest
import time

from typing import List

import boto3

from hashlib import sha256
from contextlib import contextmanager
from base64 import urlsafe_b64encode
from client import CliClient, InternalApiClient, ArtifactsClient
from pymongo import MongoClient

DB_NAME = "deployment_service"
DB_MIGRATION_COLLECTION = "migration_info"
DB_VERSION = "1.2.1"

MONGO_LOCK_FILE = "/mongo-lock"


class Lock:
    def __init__(self, lock_file):
        letters = string.ascii_lowercase
        self.hash = "".join(random.choice(letters) for i in range(128))
        self.lock_file = lock_file

    def __enter__(self):
        while os.path.exists(self.lock_file):
            time.sleep(0.1)
        with open(self.lock_file, "w") as fh:
            fh.write(self.hash)
        return self

    def __exit__(self, type, value, traceback):
        self.unlock()

    def unlock(self):
        if not os.path.exists(self.lock_file):
            return
        with open(self.lock_file, "r") as fh:
            hash = fh.read()
            if hash != self.hash:
                return
        os.remove(self.lock_file)


class Artifact(metaclass=abc.ABCMeta):
    @abc.abstractproperty
    def size(self):
        pass

    @abc.abstractproperty
    def checksum(self):
        pass


class BytesArtifact(io.BytesIO, Artifact):
    def __init__(self, data):
        self._size = len(data)
        d = sha256()
        d.update(data)
        self._checksum = d.hexdigest()

        super().__init__(data)

    @property
    def size(self):
        return self._size

    @property
    def checksum(self):
        return self._checksum


class FileArtifact(io.RawIOBase, Artifact):
    def __init__(self, size, openedfile, data_file_name=""):
        self.file = openedfile
        self._size = size
        self.rdata_file_name = data_file_name

        d = sha256()
        with open(openedfile.name, "rb") as inf:
            fdata = inf.read()
            if fdata:
                d.update(fdata)

        self._checksum = d.hexdigest()

    def read(self, *args):
        return self.file.read(*args)

    @property
    def size(self):
        return self._size

    @property
    def checksum(self):
        return self._checksum

    @property
    def file_name(self):
        return self.file.name

    @property
    def data_file_name(self):
        return self.rdata_file_name

    @property
    def file_size(self):
        file_stats = os.stat(self.file.name)
        return file_stats.st_size


@contextmanager
def artifact_from_mender_file(path, data_file_name=""):
    with open(path, "rb") as infile:
        sz = str(os.stat(path).st_size)
        yield FileArtifact(sz, infile, data_file_name=data_file_name)


@contextmanager
def artifact_from_raw_data(data):
    if type(data) is str:
        data = data.encode()
    yield BytesArtifact(data)


@contextmanager
def artifact_rootfs_from_data(
    name: str = "foo", data: bytes = None, devicetype: str = "hammer", compression=""
):
    with tempfile.NamedTemporaryFile(prefix="menderout") as tmender:
        logging.info("writing mender artifact to temp file %s", tmender.name)

        with tempfile.NamedTemporaryFile(prefix="menderin") as tdata:
            logging.info("writing update data to temp file %s", tdata.name)
            tdata.write(data)
            tdata.flush()

            cmd = (
                f"mender-artifact write rootfs-image"
                + f' --device-type "{devicetype}"'
                + f' --file "{tdata.name}"'
                + f' --artifact-name "{name}"'
                + f' --output-path "{tmender.name}"'
                + f' --compression "{compression}"'
            )
            rc = subprocess.call(cmd, shell=True)
            if rc:
                logging.error("mender-artifact call '%s' failed with code %d", cmd, rc)
                raise RuntimeError(
                    "mender-artifact command '{}' failed with code {}".format(cmd, rc)
                )

            # bring up temp mender artifact
            with artifact_from_mender_file(
                tmender.name, data_file_name=tdata.name
            ) as fa:
                yield fa


@contextmanager
def artifact_update_module_from_data(
    name: str = "foo",
    data: bytes = None,
    devicetype: str = "hammer",
    update_type: str = "app",
):
    with tempfile.NamedTemporaryFile(prefix="menderout") as tmender:
        logging.info("writing mender artifact to temp file %s", tmender.name)

        with tempfile.NamedTemporaryFile(prefix="menderin") as tdata:
            logging.info("writing update data to temp file %s", tdata.name)
            tdata.write(data)
            tdata.flush()

            cmd = (
                f"mender-artifact write module-image"
                + f' -T "{update_type}"'
                + f' --device-type "{devicetype}"'
                + f' --file "{tdata.name}"'
                + f' --artifact-name "{name}"'
                + f' --output-path "{tmender.name}"'
            )
            rc = subprocess.call(cmd, shell=True)
            if rc:
                logging.error("mender-artifact call '%s' failed with code %d", cmd, rc)
                raise RuntimeError(
                    "mender-artifact command '{}' failed with code {}".format(cmd, rc)
                )

            # bring up temp mender artifact
            with artifact_from_mender_file(tmender.name) as fa:
                yield fa


@contextmanager
def artifact_bootstrap_from_data(
    name: str = "foo",
    devicetype: str = "hammer",
    provides: List = [],
    clears_provides: List = [],
):
    with tempfile.NamedTemporaryFile(prefix="menderout") as tmender:
        logging.info("writing mender artifact to temp file %s", tmender.name)

        provides_arg = "".join([" --provides {}".format(p) for p in provides])
        clears_provides_arg = "".join(
            [" --clears-provides {}".format(p) for p in clears_provides]
        )
        cmd = (
            f"mender-artifact write bootstrap-artifact"
            + f' --device-type "{devicetype}"'
            + f' --artifact-name "{name}"'
            + f' --output-path "{tmender.name}"'
            + f"{provides_arg}"
            + f"{clears_provides_arg}"
        )
        rc = subprocess.call(cmd, shell=True)
        if rc:
            logging.error("mender-artifact call '%s' failed with code %d", cmd, rc)
            raise RuntimeError(
                "mender-artifact command '{}' failed with code {}".format(cmd, rc)
            )

        # bring up temp mender artifact
        with artifact_from_mender_file(tmender.name) as fa:
            yield fa


@contextmanager
def artifacts_added_from_data(artifacts):
    data = b"foo_bar"
    out_artifacts = []
    ac = ArtifactsClient()

    for name, device_type in artifacts:
        # generate artifact
        with artifact_rootfs_from_data(
            name=name, data=data, devicetype=device_type
        ) as art:
            logging.info("uploading artifact")
            artid = ac.add_artifact("foo", art.size, art)
            out_artifacts.append(artid)

    yield out_artifacts

    for artid in out_artifacts:
        ac.delete_artifact(artid)


@contextmanager
def artifacts_update_module_added_from_data(artifacts):
    data = b"foo_bar"
    out_artifacts = []
    ac = ArtifactsClient()

    for name, device_type, update_type in artifacts:
        # generate artifact
        with artifact_update_module_from_data(
            name=name, data=data, devicetype=device_type, update_type=update_type
        ) as art:
            logging.info("uploading artifact")
            artid = ac.add_artifact("foo", art.size, art)
            out_artifacts.append(artid)

    yield out_artifacts

    for artid in out_artifacts:
        ac.delete_artifact(artid)


class Device:
    def __init__(self, device_type="hammer"):
        self.devid = "".join(
            [random.choice(string.ascii_letters + string.digits) for _ in range(10)]
        )
        self.device_type = device_type

    @property
    def fake_token(self):
        claims = json.dumps({"sub": self.devid, "iss": "Mender", "mender.device": True})
        hdr = '{"typ": "JWT"}'
        signature = "fake-signature"
        return ".".join(
            urlsafe_b64encode(p.encode()).decode().strip("=")
            for p in [hdr, claims, signature]
        )

    def fake_token_mt(self, tenant):
        claims = json.dumps(
            {
                "sub": self.devid,
                "iss": "Mender",
                "mender.tenant": tenant,
                "mender.device": True,
            }
        )
        hdr = '{"typ": "JWT"}'
        signature = "fake-signature"
        return ".".join(
            urlsafe_b64encode(p.encode()).decode().strip("=")
            for p in [hdr, claims, signature]
        )


@pytest.fixture(scope="session")
def cli(request):
    service = request.config.getoption("host").split(":")[0]
    return CliClient(service)


@pytest.fixture(scope="session")
def mongo(request):
    return MongoClient(request.config.getoption("mongo_url"))


@pytest.yield_fixture(scope="function")
def clean_db(mongo):
    mongo_cleanup(mongo)
    yield mongo
    mongo_cleanup(mongo)


@pytest.fixture(scope="session")
def s3_bucket(request):
    bucket_name = request.config.getoption("s3_bucket")
    key_id = request.config.getoption("s3_key_id")
    secret = request.config.getoption("s3_secret_key")
    endpoint = request.config.getoption("s3_endpoint_url")
    bucket = boto3.resource(
        "s3",
        aws_access_key_id=key_id,
        aws_secret_access_key=secret,
        endpoint_url=endpoint,
    ).Bucket(bucket_name)
    return bucket


@pytest.fixture(scope="function")
def clean_minio(s3_bucket):
    for obj in s3_bucket.objects.all():
        obj.delete()
    return s3_bucket


def mongo_cleanup(mongo):
    dbs = mongo.list_database_names()
    dbs = [
        d
        for d in dbs
        if d not in ["local", "admin", "config", "deployment_service", "workflows"]
    ]
    for d in dbs:
        mongo.drop_database(d)
    db = mongo["deployment_service"]
    collections = db.list_collection_names()
    for c in collections:
        if c == "migration_info":
            continue
        db[c].delete_many({})


@pytest.fixture(scope="session")
def api_client_int():
    return InternalApiClient()


def make_tenant_db(tenant_id):
    return "{}-{}".format(DB_NAME, tenant_id)
