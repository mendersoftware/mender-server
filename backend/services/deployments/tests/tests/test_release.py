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
import pytest
from uuid import uuid4

from client import DeploymentsClient, ArtifactsClient
from client import management_v1_client
from common import (
    artifacts_added_from_data,
    artifact_bootstrap_from_data,
    artifacts_update_module_added_from_data,
    clean_db,
    clean_minio,
    s3_bucket,
    mongo,
    Lock,
    MONGO_LOCK_FILE,
)


class TestRelease:
    d = DeploymentsClient()

    @pytest.mark.usefixtures("clean_db")
    def test_releases_no_artifacts(self):
        rsp = management_v1_client(jwt="foo").list_releases()
        assert len(rsp) == 0

    @pytest.mark.usefixtures("clean_minio", "clean_db")
    def test_get_all_releases(self):
        with Lock(MONGO_LOCK_FILE) as l:
            with artifacts_added_from_data(
                [
                    ("foo", "device-type-1"),
                    ("foo", "device-type-2"),
                    ("bar", "device-type-2"),
                ]
            ):
                releases = management_v1_client(jwt="foo").list_releases()
                assert len(releases) == 2
                release1 = releases[0]
                release2 = releases[1]

                assert release1.name == "bar"
                assert len(release1.artifacts) == 1
                r1a = release1.artifacts[0]
                assert r1a.name == "bar"
                assert r1a.device_types_compatible == ["device-type-2"]

                assert release2.name == "foo"
                assert len(release2.artifacts) == 2

                r2a1 = release2.artifacts[0]
                r2a2 = release2.artifacts[1]
                l.unlock()
                assert r2a1.name == "foo"
                assert r2a1.device_types_compatible == ["device-type-1"]
                assert r2a2.name == "foo"
                assert r2a2.device_types_compatible == ["device-type-2"]

    @pytest.mark.usefixtures("clean_minio", "clean_db")
    def test_get_release_with_bootstrap_artifact(self):
        with Lock(MONGO_LOCK_FILE) as l:
            artifact_name = str(uuid4())
            description = f"description for foo {artifact_name}"
            device_type = f"project-{str(uuid4())}"
            provides = ["foo:bar", "something:cool"]
            clears_provides = ["nothing.really.useful.*"]

            # generate artifact
            with artifact_bootstrap_from_data(
                name=artifact_name,
                devicetype=device_type,
                provides=provides,
                clears_provides=clears_provides,
            ) as art:
                ac = ArtifactsClient()
                ac.add_artifact(description, art.size, art)
                releases = management_v1_client(jwt="foo").list_releases()
                assert len(releases) == 1
                release1 = releases[0]

                assert release1.name == artifact_name
                assert len(release1.artifacts) == 1
                r1a = release1.artifacts[0]
                assert r1a.name == artifact_name
                assert device_type in r1a.device_types_compatible
                provides_dict = dict(p.split(":") for p in provides)
                for p in provides_dict:
                    assert p in r1a.artifact_provides
                for c in clears_provides:
                    assert c in r1a.clears_artifact_provides
                assert len(r1a.updates) == 1
                r1au = r1a.updates[0]
                l.unlock()
                assert r1au.files is None
                assert r1au.type_info.type is None

    @pytest.mark.usefixtures("clean_minio", "clean_db")
    def test_get_releases_by_name(self):
        with Lock(MONGO_LOCK_FILE) as l:
            with artifacts_added_from_data(
                [
                    ("foo", "device-type-1"),
                    ("foo", "device-type-2"),
                    ("bar", "device-type-2"),
                ]
            ):
                releases = management_v1_client(jwt="foo").list_releases(name="bar")
                assert len(releases) == 1
                release = releases[0]
                assert release.name == "bar"
                assert len(release.artifacts) == 1
                artifact = release.artifacts[0]
                l.unlock()
                assert artifact.name == "bar"
                assert artifact.device_types_compatible == ["device-type-2"]

    @pytest.mark.usefixtures("clean_minio", "clean_db")
    def test_get_releases_by_update_type(self):
        with Lock(MONGO_LOCK_FILE) as l:
            with artifacts_update_module_added_from_data(
                [
                    ("foo", "device-type-1", "app"),
                    ("foo", "device-type-2", "single-file"),
                    ("bar", "device-type-2", "directory"),
                ]
            ):
                releases = management_v1_client(jwt="foo").list_releases(
                    update_type="app"
                )
                assert len(releases) == 1
                release = releases[0]
                assert release.name == "foo"
                assert len(release.artifacts) > 0
                artifact = release.artifacts[0]
                assert artifact.name == "foo"
                assert artifact.device_types_compatible == ["device-type-1"]

                releases = management_v1_client(jwt="foo").list_releases(
                    update_type="single-file"
                )
                assert len(releases) == 1
                release = releases[0]
                assert release.name == "foo"
                assert len(release.artifacts) > 0
                artifact = release.artifacts[1]
                assert artifact.name == "foo"
                assert artifact.device_types_compatible == ["device-type-2"]

                releases = management_v1_client(jwt="foo").list_releases(
                    update_type="directory"
                )
                assert len(releases) == 1
                release = releases[0]
                assert release.name == "bar"
                assert len(release.artifacts) > 0
                artifact = release.artifacts[0]
                l.unlock()
                assert artifact.name == "bar"
                assert artifact.device_types_compatible == ["device-type-2"]

    @pytest.mark.usefixtures("clean_minio", "clean_db")
    def test_get_releases_with_pagination_by_update_type(self):
        with Lock(MONGO_LOCK_FILE) as l:
            with artifacts_update_module_added_from_data(
                [
                    ("foo", "device-type-1", "app"),
                    ("foo", "device-type-2", "single-file"),
                    ("bar", "device-type-2", "directory"),
                ]
            ):
                releases = management_v1_client(
                    jwt="foo"
                ).list_releases_with_pagination(update_type="app")
                assert len(releases) == 1
                release = releases[0]
                assert release.name == "foo"
                assert len(release.artifacts) > 0
                artifact = release.artifacts[0]
                assert artifact.name == "foo"
                assert artifact.device_types_compatible == ["device-type-1"]

                releases = management_v1_client(
                    jwt="foo"
                ).list_releases_with_pagination(update_type="single-file")
                assert len(releases) == 1
                release = releases[0]
                assert release.name == "foo"
                assert len(release.artifacts) > 0
                artifact = release.artifacts[1]
                assert artifact.name == "foo"
                assert artifact.device_types_compatible == ["device-type-2"]

                releases = management_v1_client(
                    jwt="foo"
                ).list_releases_with_pagination(update_type="directory")
                assert len(releases) == 1
                release = releases[0]
                assert release.name == "bar"
                assert len(release.artifacts) > 0
                artifact = release.artifacts[0]
                l.unlock()
                assert artifact.name == "bar"
                assert artifact.device_types_compatible == ["device-type-2"]

    @pytest.mark.usefixtures("clean_minio", "clean_db")
    def test_get_releases_by_name_no_result(self):
        with Lock(MONGO_LOCK_FILE) as l:
            with artifacts_added_from_data(
                [
                    ("foo", "device-type-1"),
                    ("foo", "device-type-2"),
                    ("bar", "device-type-2"),
                ]
            ):
                releases = management_v1_client(jwt="foo").list_releases(name="baz")
                l.unlock()
                assert len(releases) == 0

    @pytest.mark.usefixtures("clean_minio", "clean_db")
    def test_get_releases_paginated_by_name_no_result(self):
        with Lock(MONGO_LOCK_FILE) as l:
            with artifacts_added_from_data(
                [
                    ("foo", "device-type-1"),
                    ("foo", "device-type-2"),
                    ("bar", "device-type-2"),
                ]
            ):
                releases = management_v1_client(
                    jwt="foo"
                ).list_releases_with_pagination(name="baz")
                l.unlock()
                assert len(releases) == 0
