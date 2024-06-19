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
import json
import random

import pytest
import time
import os
from os import urandom
from os.path import basename
from uuid import uuid4
from hashlib import sha256

import bravado
import requests

from client import ArtifactsClient, ArtifactsClientError
from common import (
    artifact_from_raw_data,
    artifact_rootfs_from_data,
    artifact_bootstrap_from_data,
    clean_db,
    clean_minio,
    s3_bucket,
    mongo,
    Lock,
    MONGO_LOCK_FILE,
)


class TestArtifact:
    ac = ArtifactsClient()

    def test_artifacts_all(self):
        res = self.ac.client.Management_API.List_Artifacts().result()
        self.ac.log.debug("result: %s", res)

    def test_artifacts_new_bogus_empty(self, clean_minio, clean_db):
        with Lock(MONGO_LOCK_FILE) as l:
            # try bogus image data
            try:
                res = self.ac.client.Management_API.Upload_Artifact(
                    Authorization="foo",
                    size=100,
                    artifact="".encode(),
                    description="bar",
                ).result()
            except bravado.exception.HTTPError as e:
                assert (
                    sum(1 for x in clean_minio.objects.all()) == 0
                )
                assert e.response.status_code == 400
            else:
                raise AssertionError("expected to fail")

    def test_artifacts_new_bogus_data(self, clean_minio, clean_db):
        with Lock(MONGO_LOCK_FILE) as l:
            with artifact_from_raw_data(b"foo_bar") as art:
                files = ArtifactsClient.make_upload_meta(
                    {
                        "description": "bar",
                        "size": str(art.size),
                        "artifact": ("firmware", art, "application/octet-stream", {},),
                    }
                )

                rsp = requests.post(self.ac.make_api_url("/artifacts"), files=files)
                l.unlock()
                assert (
                    sum(1 for x in clean_minio.objects.all()) == 0
                )
                assert rsp.status_code == 400

    def test_artifacts_valid(self, clean_minio, clean_db):
        with Lock(MONGO_LOCK_FILE) as l:
            artifact_name = str(uuid4())
            description = f"description for foo {artifact_name}"
            device_type = f"project-{str(uuid4())}"
            data = b"foo_bar"

            # generate artifact
            with artifact_rootfs_from_data(
                name=artifact_name, data=data, devicetype=device_type
            ) as art:
                self.ac.log.info("uploading artifact")
                artid = self.ac.add_artifact(description, art.size, art)

                # artifacts listing should not be empty now
                res = self.ac.client.Management_API.List_Artifacts().result()
                self.ac.log.debug("result: %s", res)
                assert len(res[0]) > 0

                res = self.ac.client.Management_API.Show_Artifact(
                    Authorization="foo", id=artid
                ).result()[0]
                self.ac.log.info("artifact: %s", res)

                # verify its data
                assert res.id == artid
                assert res.name == artifact_name
                assert res.description == description
                assert res.size == int(art.size)
                assert device_type in res.device_types_compatible
                assert len(res.updates) == 1
                update = res.updates[0]
                assert len(update.files) == 1
                uf = update.files[0]
                assert uf.size == len(data)
                assert uf.checksum
                # TODO: verify uf signature once it's supported
                # assert uf.signature

                # try to fetch the update
                res = self.ac.client.Management_API.Download_Artifact(
                    Authorization="foo", id=artid
                ).result()[0]
                self.ac.log.info("download result %s", res)
                assert res.uri
                # fetch it now (disable SSL verification)
                rsp = requests.get(res.uri, verify=False, stream=True)

                assert rsp.status_code == 200
                assert (
                    sum(1 for x in clean_minio.objects.all()) == 1
                )

                # receive artifact and compare its checksum
                dig = sha256()
                while True:
                    rspdata = rsp.raw.read()
                    if rspdata:
                        dig.update(rspdata)
                    else:
                        break

                self.ac.log.info(
                    "artifact checksum %s expecting %s", dig.hexdigest(), art.checksum,
                )
                assert dig.hexdigest() == art.checksum

                # delete it now
                self.ac.delete_artifact(artid)

                # should be unavailable now
                try:
                    res = self.ac.client.Management_API.Show_Artifact(
                        Authorization="foo", id=artid
                    ).result()
                except bravado.exception.HTTPError as e:
                    assert e.response.status_code == 404
                else:
                    raise AssertionError("expected to fail")
            l.unlock()

    def test_artifacts_bootstrap_valid(self, clean_minio, clean_db):
        with Lock(MONGO_LOCK_FILE) as l:
            artifact_name = str(uuid4())
            description = f"description for foo {artifact_name}"
            device_type = f"project-{str(uuid4())}"

            # generate artifact
            with artifact_bootstrap_from_data(
                name=artifact_name, devicetype=device_type
            ) as art:
                self.ac.log.info("uploading artifact")
                artid = self.ac.add_artifact(description, art.size, art)

                # artifacts listing should not be empty now
                res = self.ac.list_artifacts().json()
                self.ac.log.debug("result: %s", res)
                assert len(res[0]) > 0

                res = self.ac.show_artifact(artid).json()
                self.ac.log.debug("result: %s", res)

                # verify its data
                assert res["id"] == artid
                assert res["name"] == artifact_name
                assert res["description"] == description
                assert res["size"] == int(art.size)
                assert device_type in res["device_types_compatible"]
                assert len(res["updates"]) == 1
                update = res["updates"][0]
                assert update["type_info"]["type"] is None
                assert update["files"] is None

                # try to fetch the update
                res = self.ac.client.Management_API.Download_Artifact(
                    Authorization="foo", id=artid
                ).result()[0]
                self.ac.log.info("download result %s", res)
                assert res.uri
                # fetch it now (disable SSL verification)
                rsp = requests.get(res.uri, verify=False, stream=True)

                assert rsp.status_code == 200
                assert (
                    sum(1 for x in clean_minio.objects.all()) == 1
                )

                # receive artifact and compare its checksum
                dig = sha256()
                while True:
                    rspdata = rsp.raw.read()
                    if rspdata:
                        dig.update(rspdata)
                    else:
                        break

                self.ac.log.info(
                    "artifact checksum %s expecting %s", dig.hexdigest(), art.checksum,
                )
                assert dig.hexdigest() == art.checksum

                # delete it now
                self.ac.delete_artifact(artid)

                # should be unavailable now
                try:
                    res = self.ac.client.Management_API.Show_Artifact(
                        Authorization="foo", id=artid
                    ).result()
                except bravado.exception.HTTPError as e:
                    assert e.response.status_code == 404
                else:
                    raise AssertionError("expected to fail")
            l.unlock()

    @pytest.mark.usefixtures("clean_minio", "clean_db")
    def test_artifacts_valid_multipart(self):
        """
        Uploads an artifact > 10MiB to cover the multipart upload scenario.
        """
        with Lock(MONGO_LOCK_FILE) as l:
            artifact_name = str(uuid4())
            description = "description for foo " + artifact_name
            device_type = "project-" + str(uuid4())
            data = urandom(1024 * 1024 * 15)

            # generate artifact
            with artifact_rootfs_from_data(
                name=artifact_name, data=data, devicetype=device_type
            ) as art:
                self.ac.log.info("uploading artifact")
                artid = self.ac.add_artifact(description, art.size, art)

                # artifacts listing should not be empty now
                res = self.ac.client.Management_API.List_Artifacts().result()
                self.ac.log.debug("result: %s", res)
                assert len(res[0]) > 0

                res = self.ac.client.Management_API.Show_Artifact(
                    Authorization="foo", id=artid
                ).result()[0]
                self.ac.log.info("artifact: %s", res)

                # verify its data
                assert res.id == artid
                assert res.name == artifact_name
                assert res.description == description
                assert res.size == int(art.size)
                assert device_type in res.device_types_compatible
                assert len(res.updates) == 1
                update = res.updates[0]
                assert len(update.files) == 1
                uf = update.files[0]
                assert uf["size"] == len(data)
                assert uf["checksum"]
            l.unlock()

    def test_single_artifact(self):
        # try with bogus image ID
        with Lock(MONGO_LOCK_FILE) as l:
            try:
                res = self.ac.client.Management_API.Show_Artifact(
                    Authorization="foo", id="foo"
                ).result()
            except bravado.exception.HTTPError as e:
                assert e.response.status_code == 400
            else:
                raise AssertionError("expected to fail")

            # try with nonexistent image ID
            try:
                res = self.ac.client.Management_API.Show_Artifact(
                    Authorization="foo", id=uuid4()
                ).result()
            except bravado.exception.HTTPError as e:
                assert e.response.status_code == 404
            else:
                raise AssertionError("expected to fail")
            l.unlock()

    def test_artifacts_generate_valid(self, clean_minio, clean_db):
        with Lock(MONGO_LOCK_FILE) as l:
            artifact_name = str(uuid4())
            description = "description for foo " + artifact_name
            device_type = "project-" + str(uuid4())
            data = b"foo_bar"

            # generate artifact
            self.ac.log.info("uploading artifact")
            artid = self.ac.generate_artifact(
                name=artifact_name,
                description=description,
                device_types_compatible=device_type,
                type="single_file",
                args="",
                data=data,
            )

            # the file has been stored
            assert sum(1 for x in clean_minio.objects.all()) == 1
            l.unlock()

    @pytest.mark.usefixtures("clean_minio", "clean_db")
    def test_compressed_artifacts_valid(self):
        """Create and upload artifacts with different compressions"""
        with Lock(MONGO_LOCK_FILE) as l:
            compressions = ["gzip", "lzma"]
            for comp in compressions:
                artifact_name = str(uuid4())
                description = "description for foo " + artifact_name
                device_type = "project-" + str(uuid4())
                data = b"foo_bar"

                with artifact_rootfs_from_data(
                    name=artifact_name,
                    data=data,
                    devicetype=device_type,
                    compression=comp,
                ) as art:
                    self.ac.log.info(
                        "uploading artifact (compression: {})".format(comp)
                    )
                    self.ac.add_artifact(description, art.size, art)
            l.unlock()


class TestDirectUpload:
    def test_upload(self, clean_db):
        with Lock(MONGO_LOCK_FILE) as l:
            mgo = clean_db
            ac = ArtifactsClient()

            url = ac.make_upload_url()
            doc = mgo.deployment_service.uploads.find_one({"_id": url.id})
            assert doc is not None, "Upload intent not found in database"
            assert doc["status"] == 0

            with artifact_rootfs_from_data(data=b"", compression="none") as artie:
                requests.put(
                    url.uri,
                    artie.read(),
                    headers={"Content-Type": "application/octet-stream"},
                    verify=False,
                )
            rsp = ac.complete_upload(url.id)
            assert rsp.status_code == 202, "Unexpected HTTP status code"

            doc = mgo.deployment_service.uploads.find_one({"_id": url.id})
            assert doc["status"] > 0

            # Retry for half a minute
            for _ in range(60):
                try:
                    ac.show_artifact(artid=url.id)
                except ArtifactsClientError:
                    time.sleep(0.5)
                break
            else:
                raise TimeoutError("Timeout waiting for artifact to be processed")
            l.unlock()

    def test_upload_with_meta(self, clean_db):
        with Lock(MONGO_LOCK_FILE) as l:
            mgo = clean_db
            ac = ArtifactsClient()

            url = ac.make_upload_url()
            doc = mgo.deployment_service.uploads.find_one({"_id": url.id})
            assert doc is not None, "Upload intent not found in database"
            assert doc["status"] == 0

            with artifact_rootfs_from_data(data=b"", compression="none") as artie:
                requests.put(
                    url.uri,
                    artie.read(),
                    headers={"Content-Type": "application/octet-stream"},
                    verify=False,
                )
                artifact_size = int(artie.size)
                file_name = basename(artie.data_file_name)
            file_size = random.randint(1023, 65536)
            file_checksum = "cxvbfg4h34erdsafcxvbdny4w3t"

            rsp = ac.complete_upload(
                url.id,
                body=json.dumps(
                    {
                        "size": artifact_size,
                        "updates": [
                            {
                                "type_info": {"type": "directory"},
                                "files": [
                                    {
                                        "name": file_name,
                                        "checksum": file_checksum,
                                        "size": file_size,
                                    }
                                ],
                            }
                        ],
                    }
                ),
            )
            assert rsp.status_code == 202, "Unexpected HTTP status code"

            propagation_timeout_s = 16
            time.sleep(propagation_timeout_s)
            doc = mgo.deployment_service.releases.find_one({"_id": "foo"})
            assert (
                doc["artifacts"][0]["meta_artifact"]["updates"][0]["files"][0]["size"]
                == file_size
            )
            assert (
                doc["artifacts"][0]["meta_artifact"]["updates"][0]["files"][0][
                    "checksum"
                ]
                == file_checksum
            )
            assert (
                doc["artifacts"][0]["meta_artifact"]["updates"][0]["files"][0]["name"]
                == file_name
            )

            doc = mgo.deployment_service.uploads.find_one({"_id": url.id})
            assert doc["status"] > 0

            # Retry for half a minute
            for _ in range(60):
                try:
                    ac.show_artifact(artid=url.id)
                except ArtifactsClientError:
                    time.sleep(0.5)
                break
            else:
                raise TimeoutError("Timeout waiting for artifact to be processed")
            l.unlock()
