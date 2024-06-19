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
import re
import time

from datetime import datetime
from typing import Union

import management_api
import utils


class TestManagementSearch:
    @pytest.fixture(scope="class", autouse=True)
    def setup_test_context(self, opensearch):
        # clean up any indices from previous tests
        indices = opensearch.cat.indices(format="json")
        for idx in indices:
            if not idx["index"].startswith("."):
                opensearch.delete_by_query(
                    index=[idx["index"]], body={"query": {"match_all": {}}}
                )

        for tenant_id, dep in self._test_set:
            utils.index_deployment(tenant_id, dep)

        time.sleep(5)

    class _TestCase:
        def __init__(
            self,
            search_terms: management_api.models.DeploymentSearchTerms,
            http_code: int,
            result: Union[list[management_api.models.Deployment], str],
            authorization: str = None,
        ):
            self.search_terms = search_terms
            self.http_code = http_code
            self.result = result
            self.authorization = authorization

    _test_set = [
        (
            "123456789012345678901234",
            management_api.models.Deployment(
                id="e2b8df68-a7d8-4380-9a91-69ce1d6684e0",
                tenant_id="123456789012345678901234",
                deployment_all_devices=False,
                deployment_artifact_name="mender-demo-artifact-3.4.0",
                deployment_created=datetime(2023, 1, 6, 4, 42, 40, 482000),
                deployment_type="software",
                deployment_autogenerate_deta=False,
                deployment_force_installation=True,
                deployment_group=None,
                deployment_id="ba4e785b-ccfc-4266-99f6-028f7cf557b0",
                deployment_max_devices=1,
                deployment_name="All devices",
                deployment_phased=False,
                deployment_retries=0,
                device_attempts=0,
                device_created=datetime(2023, 1, 6, 4, 42, 40, 482000),
                device_elapsed_seconds=8,
                device_finished=datetime(2023, 1, 6, 4, 42, 49, 177000),
                device_id="0a4eb6a1-861a-4d8c-a3cf-7517f0e7c7c3",
                device_is_log_available=False,
                device_retries=0,
                device_status="success",
                image_artifact_name="mender-demo-artifact-3.4.0",
                image_artifact_info_format="mender",
                image_artifact_info_version=3,
                image_signed=False,
                image_size=1982976,
                image_id="596c0e6f-02a6-46df-95a3-4d09b3607e9c",
                image_depends={
                    "device_type": [
                        "beaglebone",
                        "beaglebone-yocto",
                        "beaglebone-yocto-grub",
                        "generic-armv6",
                        "generic-x86_64",
                        "qemux86-64",
                        "raspberrypi0w",
                        "raspberrypi0-wifi",
                        "raspberrypi3",
                        "raspberrypi4",
                        "raspberrypi",
                    ]
                },
                image_provides={
                    "artifact_name": "mender-demo-artifact-3.4.0",
                    "data-partition.mender-demo-artifact.version": "3.4.0",
                },
                image_device_types=[
                    "beaglebone",
                    "beaglebone-yocto",
                    "beaglebone-yocto-grub",
                    "generic-armv6",
                    "generic-x86_64",
                    "qemux86-64",
                    "raspberrypi0w",
                    "raspberrypi0-wifi",
                    "raspberrypi3",
                    "raspberrypi4",
                    "raspberrypi",
                ],
            ),
        ),
        (
            "123456789012345678901234",
            management_api.models.Deployment(
                id="f32a5329-b344-44b4-9a2f-ff2b44968a35",
                tenant_id="123456789012345678901234",
                deployment_all_devices=False,
                deployment_artifact_name="mender-demo-artifact-3.4.0",
                deployment_created=datetime(2023, 1, 6, 4, 39, 59, 765000),
                deployment_type="software",
                deployment_autogenerate_deta=False,
                deployment_force_installation=False,
                deployment_group=None,
                deployment_id="1b543b1c-0a47-4024-bdc5-ce9722775580",
                deployment_max_devices=1,
                deployment_name="All devices",
                deployment_phased=False,
                deployment_retries=0,
                device_attempts=0,
                device_created=datetime(2023, 1, 6, 4, 39, 59, 765000),
                device_elapsed_seconds=13,
                device_finished=datetime(2023, 1, 6, 4, 40, 13, 269000),
                device_id="0a4eb6a1-861a-4d8c-a3cf-7517f0e7c7c3",
                device_is_log_available=False,
                device_retries=0,
                device_status="already-installed",
                image_artifact_name="",
                image_signed=False,
            ),
        ),
    ]

    @pytest.mark.parametrize(
        argnames="test_case",
        argvalues=[
            _TestCase(
                authorization=utils.generate_jwt(tenant_id="123456789012345678901234"),
                search_terms=management_api.models.DeploymentSearchTerms(
                    filters=[
                        management_api.models.DeploymentFilterTerm(
                            attribute="device_status",
                            value="already-installed",
                            type="$eq",
                        )
                    ],
                    sort=[
                        management_api.models.DeploymentSortTerm(
                            attribute="id", order="asc",
                        )
                    ],
                ),
                http_code=200,
                result=[_test_set[1][1]],  # f32a5329-b344-44b4-9a2f-ff2b44968a35
            ),
            _TestCase(
                authorization=utils.generate_jwt(tenant_id="123456789012345678901234"),
                search_terms=management_api.models.DeploymentSearchTerms(
                    filters=[
                        management_api.models.DeploymentFilterTerm(
                            attribute="device_status",
                            value="already-installed",
                            type="$ne",
                        )
                    ],
                    sort=[
                        management_api.models.DeploymentSortTerm(
                            attribute="id", order="asc",
                        )
                    ],
                ),
                http_code=200,
                result=[_test_set[0][1]],  # e2b8df68-a7d8-4380-9a91-69ce1d6684e0
            ),
            _TestCase(
                authorization=utils.generate_jwt(tenant_id="123456789012345678901234"),
                search_terms=management_api.models.DeploymentSearchTerms(
                    filters=[
                        management_api.models.DeploymentFilterTerm(
                            attribute="image_size",
                            value=1000000,
                            type="$gte",
                        )
                    ],
                    sort=[
                        management_api.models.DeploymentSortTerm(
                            attribute="id", order="asc",
                        )
                    ],
                ),
                http_code=200,
                result=[_test_set[0][1]],  # e2b8df68-a7d8-4380-9a91-69ce1d6684e0
            ),
            _TestCase(
                authorization=utils.generate_jwt(tenant_id="123456789012345678901234"),
                search_terms=management_api.models.DeploymentSearchTerms(
                    filters=[
                        management_api.models.DeploymentFilterTerm(
                            attribute="device_status",
                            value=[
                                "aborted",
                            ],
                            type="$nin",
                        )
                    ],
                    sort=[
                        management_api.models.DeploymentSortTerm(
                            attribute="device_created", order="desc",
                        )
                    ],
                ),
                http_code=200,
                result=[_test_set[0][1], _test_set[1][1]],
            ),
            _TestCase(
                authorization=utils.generate_jwt(tenant_id="123456789012345678901234"),
                search_terms=management_api.models.DeploymentSearchTerms(
                    filters=[
                        management_api.models.DeploymentFilterTerm(
                            attribute="device_elapsed_seconds",
                            value=10000000,
                            type="$lte",
                        )
                    ],
                    sort=[
                        management_api.models.DeploymentSortTerm(
                            attribute="device_created", order="asc",
                        )
                    ],
                ),
                http_code=200,
                result=[_test_set[1][1], _test_set[0][1]],
            ),
            _TestCase(
                authorization=utils.generate_jwt(tenant_id="anIllegalTenantID"),
                search_terms=management_api.models.DeploymentSearchTerms(
                    filters=[
                        management_api.models.DeploymentFilterTerm(
                            attribute="foo", value="bar", type="$eq",
                        )
                    ],
                ),
                http_code=200,
                result=[],
            ),
            _TestCase(
                search_terms=management_api.models.DeploymentSearchTerms(
                    filters=[
                        management_api.models.DeploymentFilterTerm(
                            attribute="foo", value="bar", type="$eq",
                        )
                    ],
                ),
                http_code=401,
                result=[],
            ),
        ],
        ids=[
            "ok, $eq",
            "ok, $ne",
            "ok, $gte",
            "ok, $nin + sort",
            "ok, $lte + sort",
            "error, missing index for tenant",
            "error, unauthorized access",
        ],
    )
    def test_search(self, test_case, setup_test_context):
        conf = None
        if test_case.authorization is not None:
            conf = management_api.Configuration.get_default_copy()
            conf.access_token = test_case.authorization
        api_client = management_api.ApiClient(configuration=conf)
        client = management_api.ManagementAPIClient(api_client=api_client)
        try:
            body, status, headers = client.search_deployments_with_http_info(
                deployment_search_terms=test_case.search_terms
            )
        except management_api.ApiException as r:
            body = r.body
            status = r.status
            headers = r.headers
        assert status == test_case.http_code
        if isinstance(test_case.result, str):
            assert isinstance(body, bytes)
            re.match(test_case.result, r.body.decode())
        elif len(test_case.result) > 0:
            assert isinstance(body, type(test_case.result))

            expected_ids = [dep.id for dep in test_case.result]
            actual_ids = [dep.id for dep in body]
            assert expected_ids == actual_ids
            for i, expected in enumerate(test_case.result):
                actual = body[i]
                # remove timezone
                actual.deployment_created = actual.deployment_created.replace(
                    tzinfo=None
                )
                actual.device_created = actual.device_created.replace(tzinfo=None)
                actual.device_finished = actual.device_finished.replace(tzinfo=None)
                # get the list of attributes
                expected_dict = expected.__dict__
                actual_dict = actual.__dict__
                # remove the extra attributes
                expected_dict.pop("local_vars_configuration", None)
                actual_dict.pop("local_vars_configuration", None)
                # compare and assert
                assert expected_dict == actual_dict
