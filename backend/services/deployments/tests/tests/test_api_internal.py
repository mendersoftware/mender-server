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

import bravado
import pytest
import requests
import uuid

from uuid import uuid4

from bson.objectid import ObjectId
from common import (
    api_client_int,
    artifact_rootfs_from_data,
    mongo,
    clean_db,
    clean_minio,
    s3_bucket,
    Lock,
    MONGO_LOCK_FILE,
)
from client import SimpleArtifactsClient, ArtifactsClientError


class TestInternalApiTenantCreate:
    def test_create_ok(self, api_client_int, clean_db):
        with Lock(MONGO_LOCK_FILE) as l:
            _, r = api_client_int.create_tenant("foobar")
            l.unlock()
            assert r.status_code == 201

            assert "deployment_service-foobar" in clean_db.list_database_names()
            assert (
                "migration_info"
                in clean_db["deployment_service-foobar"].list_collection_names()
            )

    def test_create_twice(self, api_client_int, clean_db):
        with Lock(MONGO_LOCK_FILE) as l:
            _, r = api_client_int.create_tenant("foobar")
            assert r.status_code == 201

            # creating once more should not fail
            _, r = api_client_int.create_tenant("foobar")
            assert r.status_code == 201
            l.unlock()

    def test_create_empty(self, api_client_int):
        try:
            _, r = api_client_int.create_tenant("")
        except bravado.exception.HTTPError as e:
            assert e.response.status_code == 400

    @pytest.mark.usefixtures("clean_minio")
    def test_artifacts_valid(self, api_client_int, mongo):
        with Lock(MONGO_LOCK_FILE) as l:
            artifact_name = str(uuid4())
            description = "description for foo " + artifact_name
            device_type = "project-" + str(uuid4())
            data = b"foo_bar"

            tenant_id = str(ObjectId())
            _, r = api_client_int.create_tenant(tenant_id)
            assert r.status_code == 201

            # generate artifact
            with artifact_rootfs_from_data(
                name=artifact_name, data=data, devicetype=device_type
            ) as art:
                artifacts_client = SimpleArtifactsClient()

                artifacts_client.log.info("uploading artifact")
                artid = api_client_int.add_artifact(
                    tenant_id, description, art.size, art
                )
                assert artid is not None

                # verify the artifact has been stored correctly in mongodb
                artifact = mongo[
                    "deployment_service-{}".format(tenant_id)
                ].images.find_one({"_id": artid})
                assert artifact is not None
                #
                assert artifact["_id"] == artid
                assert artifact["meta_artifact"]["name"] == artifact_name
                assert artifact["meta"]["description"] == description
                assert artifact["size"] == int(art.size)
                assert (
                    device_type in artifact["meta_artifact"]["device_types_compatible"]
                )
                assert len(artifact["meta_artifact"]["updates"]) == 1
                update = artifact["meta_artifact"]["updates"][0]
                assert len(update["files"]) == 1
                uf = update["files"][0]
                assert uf["size"] == len(data)
                assert uf["checksum"]
            l.unlock()

    @pytest.mark.usefixtures("clean_minio")
    def test_artifacts_fails_invalid_artifact_id(self, api_client_int):
        artifact_name = str(uuid4())
        description = "description for foo " + artifact_name
        device_type = "project-" + str(uuid4())
        data = b"foo_bar"

        tenant_id = str(ObjectId())

        # generate artifact
        with artifact_rootfs_from_data(
            name=artifact_name, data=data, devicetype=device_type
        ) as art:
            artifacts_client = SimpleArtifactsClient()

            artifacts_client.log.info("uploading artifact")
            with pytest.raises(ArtifactsClientError):
                api_client_int.add_artifact(
                    tenant_id, description, -1, art, "wrong_uuid4"
                )


class TestInternalApiGetLastDeviceDeploymentStatus:
    DEVICE_DEPLOYMENT_FAILED = 256
    DEVICE_DEPLOYMENT_PENDING = 2304
    DEVICE_DEPLOYMENT_SUCCESS = 2560

    def test_get_statuses(self, api_client_int, mongo, clean_db):
        with Lock(MONGO_LOCK_FILE) as l:
            # insert something in the db
            device_ids = [str(uuid.uuid4()), str(uuid.uuid4())]
            deployment_id = "acaf62f0-6a6f-45e4-9c52-838ee593cb62"
            device_deployment_id = "b14a36d3-c1a9-408c-b128-bfb4808604f1"
            devices = [
                {
                    "_id": device_ids[0],
                    "deployment_id": deployment_id,
                    "device_deployment_id": device_deployment_id,
                    "device_deployment_status": self.DEVICE_DEPLOYMENT_SUCCESS,
                    "tenant_id": "",
                },
                {
                    "_id": device_ids[1],
                    "deployment_id": deployment_id,
                    "device_deployment_id": device_deployment_id,
                    "device_deployment_status": self.DEVICE_DEPLOYMENT_SUCCESS,
                    "tenant_id": "",
                },
            ]
            for i in range(len(devices)):
                mongo["deployment_service"].devices_last_status.insert_one(devices[i])

            for i in range(len(devices)):
                devices_ids = [device_ids[i]]
                r, c = api_client_int.get_last_device_deployment_status(devices_ids, "")
                assert c.status_code == 200
                r = r["device_deployment_last_statuses"]
                assert len(r) == len(devices_ids)
                assert r[0]["device_id"] == device_ids[i]
                assert r[0]["device_deployment_id"] == device_deployment_id
                assert r[0]["device_deployment_status"] == "success"

            mongo["deployment_service"].devices_last_status.delete_many({})

            for i in range(len(devices)):
                mongo["deployment_service"].devices_last_status.insert_one(devices[i])
            devices_ids = device_ids
            r, c = api_client_int.get_last_device_deployment_status(devices_ids, "")
            assert c.status_code == 200
            r = r["device_deployment_last_statuses"]
            assert len(r) == len(device_ids)

            mongo["deployment_service"].devices_last_status.delete_many({})
            devices_ids = device_ids
            r, c = api_client_int.get_last_device_deployment_status(devices_ids, "")
            assert c.status_code == 200
            r = r["device_deployment_last_statuses"]
            assert len(r) == 0
            l.unlock()
