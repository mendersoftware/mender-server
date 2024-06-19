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

from typing import Union

import internal_api
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

        for tenant_id, dev in self._test_set:
            utils.index_device(tenant_id, dev)

        time.sleep(5)

    class _TestCase:
        def __init__(
            self,
            search_terms: management_api.models.DeviceSearchTerms,
            http_code: int,
            result: Union[list[management_api.models.Device], str],
            authorization: str = None,
        ):
            self.search_terms = search_terms
            self.http_code = http_code
            self.result = result
            self.authorization = authorization

    _test_set = [
        (
            "123456789012345678901234",
            internal_api.models.Device(
                id="463e12dd-1adb-4f62-965e-b0a9ba2c93ff",
                attributes=[
                    management_api.models.DeviceAttribute(
                        name="string",
                        value="Lorem ipsum dolor sit amet",
                        scope="inventory",
                    ),
                    management_api.models.DeviceAttribute(
                        name="number", value=2 ** 47, scope="inventory"
                    ),
                ],
            ),
        ),
        (
            "123456789012345678901234",
            internal_api.models.Device(
                id="d8b04e01-690d-41ce-8c6d-ab079a04d488",
                attributes=[
                    management_api.models.DeviceAttribute(
                        name="string",
                        value="consectetur adipiscing elit",
                        scope="inventory",
                    ),
                    management_api.models.DeviceAttribute(
                        name="number", value=420.69, scope="inventory"
                    ),
                ],
            ),
        ),
        (
            "123456789012345678901234",
            internal_api.models.Device(
                id="ad707aab-916b-4ec9-a43f-0031b2bcf9ad",
                attributes=[
                    management_api.models.DeviceAttribute(
                        name="string", value="sed do eiusmod tempor", scope="inventory",
                    ),
                ],
            ),
        ),
        (
            "123456789012345678901234",
            internal_api.models.Device(
                id="85388603-5852-437f-89c4-7549502893d5",
                attributes=[
                    management_api.models.DeviceAttribute(
                        name="string", value="incididunt ut labore", scope="inventory",
                    ),
                    management_api.models.DeviceAttribute(
                        name="number", value=0.0, scope="inventory"
                    ),
                    management_api.models.DeviceAttribute(
                        name="geo-lat", value="59.94", scope="inventory"
                    ),
                    management_api.models.DeviceAttribute(
                        name="geo-lon", value="10.72", scope="inventory"
                    ),
                ],
            ),
        ),
        (
            "098765432109876543210987",
            internal_api.models.Device(
                id="98efdb94-26c2-42eb-828d-12a5d6eb698c",
                attributes=[
                    management_api.models.DeviceAttribute(
                        name="string", value="sample text", scope="inventory",
                    ),
                    management_api.models.DeviceAttribute(
                        name="number", value=1234, scope="inventory"
                    ),
                ],
            ),
        ),
    ]

    @pytest.mark.parametrize(
        argnames="test_case",
        argvalues=[
            _TestCase(
                authorization=utils.generate_jwt(tenant_id="123456789012345678901234"),
                search_terms=management_api.models.DeviceSearchTerms(
                    filters=[
                        management_api.models.DeviceFilterTerm(
                            attribute="string",
                            value="Lorem ipsum dolor sit amet",
                            type="$ne",
                            scope="inventory",
                        )
                    ],
                    sort=[
                        management_api.models.DeviceSortTerm(
                            attribute="string", scope="inventory", order="desc"
                        )
                    ],
                ),
                http_code=200,
                result=[
                    management_api.models.Device(
                        id="ad707aab-916b-4ec9-a43f-0031b2bcf9ad",
                        attributes=[
                            management_api.models.DeviceAttribute(
                                name="string",
                                value="sed do eiusmod tempor",
                                scope="inventory",
                            ),
                        ],
                    ),
                    management_api.models.Device(
                        id="85388603-5852-437f-89c4-7549502893d5",
                        attributes=[
                            management_api.models.DeviceAttribute(
                                name="string",
                                value="incididunt ut labore",
                                scope="inventory",
                            ),
                            management_api.models.DeviceAttribute(
                                name="number", value=0.0, scope="inventory"
                            ),
                            management_api.models.DeviceAttribute(
                                name="geo-lat", value="59.94", scope="inventory"
                            ),
                            management_api.models.DeviceAttribute(
                                name="geo-lon", value="10.72", scope="inventory"
                            ),
                        ],
                    ),
                    management_api.models.Device(
                        id="d8b04e01-690d-41ce-8c6d-ab079a04d488",
                        attributes=[
                            management_api.models.DeviceAttribute(
                                name="string",
                                value="consectetur adipiscing elit",
                                scope="inventory",
                            ),
                            management_api.models.DeviceAttribute(
                                name="number", value=420.69, scope="inventory"
                            ),
                        ],
                    ),
                ],
            ),
            _TestCase(
                authorization=utils.generate_jwt(tenant_id="123456789012345678901234"),
                search_terms=management_api.models.DeviceSearchTerms(
                    filters=[
                        management_api.models.DeviceFilterTerm(
                            attribute="number",
                            value=2 ** 47,
                            type="$gte",
                            scope="inventory",
                        )
                    ],
                ),
                http_code=200,
                result=[
                    management_api.models.Device(
                        id="463e12dd-1adb-4f62-965e-b0a9ba2c93ff",
                        attributes=[
                            management_api.models.DeviceAttribute(
                                name="number", value=2 ** 47, scope="inventory"
                            ),
                            management_api.models.DeviceAttribute(
                                name="string",
                                value="Lorem ipsum dolor sit amet",
                                scope="inventory",
                            ),
                        ],
                    ),
                ],
            ),
            _TestCase(
                authorization=utils.generate_jwt(tenant_id="123456789012345678901234"),
                search_terms=management_api.models.DeviceSearchTerms(
                    filters=[
                        management_api.models.DeviceFilterTerm(
                            attribute="string",
                            value=[
                                "Lorem ipsum dolor sit amet",
                                "consectetur adipiscing elit",
                            ],
                            type="$nin",
                            scope="inventory",
                        )
                    ],
                    sort=[
                        management_api.models.DeviceSortTerm(
                            attribute="number", scope="inventory", order="asc"
                        )
                    ],
                ),
                http_code=200,
                result=[
                    management_api.models.Device(
                        id="85388603-5852-437f-89c4-7549502893d5",
                        attributes=[
                            management_api.models.DeviceAttribute(
                                name="string",
                                value="incididunt ut labore",
                                scope="inventory",
                            ),
                            management_api.models.DeviceAttribute(
                                name="number", value=0.0, scope="inventory"
                            ),
                            management_api.models.DeviceAttribute(
                                name="geo-lat", value="59.94", scope="inventory"
                            ),
                            management_api.models.DeviceAttribute(
                                name="geo-lon", value="10.72", scope="inventory"
                            ),
                        ],
                    ),
                    management_api.models.Device(
                        id="ad707aab-916b-4ec9-a43f-0031b2bcf9ad",
                        attributes=[
                            management_api.models.DeviceAttribute(
                                name="string",
                                value="sed do eiusmod tempor",
                                scope="inventory",
                            ),
                        ],
                    ),
                ],
            ),
            _TestCase(
                authorization=utils.generate_jwt(tenant_id="123456789012345678901234"),
                search_terms=management_api.models.DeviceSearchTerms(
                    filters=[
                        management_api.models.DeviceFilterTerm(
                            attribute="number",
                            value=2 ** 47,
                            type="$lte",
                            scope="inventory",
                        )
                    ],
                    sort=[
                        management_api.models.DeviceSortTerm(
                            attribute="number", scope="inventory", order="asc"
                        )
                    ],
                ),
                http_code=200,
                result=[
                    management_api.models.Device(
                        id="85388603-5852-437f-89c4-7549502893d5",
                        attributes=[
                            management_api.models.DeviceAttribute(
                                name="string",
                                value="incididunt ut labore",
                                scope="inventory",
                            ),
                            management_api.models.DeviceAttribute(
                                name="number", value=0.0, scope="inventory"
                            ),
                            management_api.models.DeviceAttribute(
                                name="geo-lat", value="59.94", scope="inventory"
                            ),
                            management_api.models.DeviceAttribute(
                                name="geo-lon", value="10.72", scope="inventory"
                            ),
                        ],
                    ),
                    management_api.models.Device(
                        id="d8b04e01-690d-41ce-8c6d-ab079a04d488",
                        attributes=[
                            management_api.models.DeviceAttribute(
                                name="string",
                                value="consectetur adipiscing elit",
                                scope="inventory",
                            ),
                            management_api.models.DeviceAttribute(
                                name="number", value=420.69, scope="inventory"
                            ),
                        ],
                    ),
                    management_api.models.Device(
                        id="463e12dd-1adb-4f62-965e-b0a9ba2c93ff",
                        attributes=[
                            management_api.models.DeviceAttribute(
                                name="string",
                                value="Lorem ipsum dolor sit amet",
                                scope="inventory",
                            ),
                            management_api.models.DeviceAttribute(
                                name="number", value=2 ** 47, scope="inventory"
                            ),
                        ],
                    ),
                ],
            ),
            _TestCase(
                authorization=utils.generate_jwt(tenant_id="123456789012345678901234"),
                search_terms=management_api.models.DeviceSearchTerms(
                    filters=[
                        management_api.models.DeviceFilterTerm(
                            attribute="string",
                            value="consectetur adipiscing elit",
                            type="$ne",
                            scope="inventory",
                        )
                    ],
                    sort=[
                        management_api.models.DeviceSortTerm(
                            attribute="string", scope="inventory", order="asc"
                        )
                    ],
                ),
                http_code=200,
                result=[
                    management_api.models.Device(
                        id="463e12dd-1adb-4f62-965e-b0a9ba2c93ff",
                        attributes=[
                            management_api.models.DeviceAttribute(
                                name="string",
                                value="Lorem ipsum dolor sit amet",
                                scope="inventory",
                            ),
                            management_api.models.DeviceAttribute(
                                name="number", value=2 ** 47, scope="inventory"
                            ),
                        ],
                    ),
                    management_api.models.Device(
                        id="85388603-5852-437f-89c4-7549502893d5",
                        attributes=[
                            management_api.models.DeviceAttribute(
                                name="string",
                                value="incididunt ut labore",
                                scope="inventory",
                            ),
                            management_api.models.DeviceAttribute(
                                name="number", value=0.0, scope="inventory"
                            ),
                            management_api.models.DeviceAttribute(
                                name="geo-lat", value="59.94", scope="inventory"
                            ),
                            management_api.models.DeviceAttribute(
                                name="geo-lon", value="10.72", scope="inventory"
                            ),
                        ],
                    ),
                    management_api.models.Device(
                        id="ad707aab-916b-4ec9-a43f-0031b2bcf9ad",
                        attributes=[
                            management_api.models.DeviceAttribute(
                                name="string",
                                value="sed do eiusmod tempor",
                                scope="inventory",
                            ),
                        ],
                    ),
                ],
            ),
            _TestCase(
                authorization=utils.generate_jwt(tenant_id="123456789012345678901234"),
                search_terms=management_api.models.DeviceSearchTerms(
                    filters=[
                        management_api.models.DeviceFilterTerm(
                            attribute="number",
                            value=False,
                            type="$exists",
                            scope="inventory",
                        )
                    ],
                ),
                http_code=200,
                result=[
                    management_api.models.Device(
                        id="ad707aab-916b-4ec9-a43f-0031b2bcf9ad",
                        attributes=[
                            management_api.models.DeviceAttribute(
                                name="string",
                                value="sed do eiusmod tempor",
                                scope="inventory",
                            ),
                        ],
                    ),
                ],
            ),
            _TestCase(
                authorization=utils.generate_jwt(tenant_id="123456789012345678901234"),
                search_terms=management_api.models.DeviceSearchTerms(
                    filters=[
                        management_api.models.DeviceFilterTerm(
                            attribute="string",
                            value="(Lorem|consectetur).*",
                            type="$regex",
                            scope="inventory",
                        )
                    ],
                    sort=[
                        management_api.models.DeviceSortTerm(
                            attribute="string", scope="inventory", order="asc"
                        )
                    ],
                ),
                http_code=200,
                result=[
                    management_api.models.Device(
                        id="463e12dd-1adb-4f62-965e-b0a9ba2c93ff",
                        attributes=[
                            management_api.models.DeviceAttribute(
                                name="string",
                                value="Lorem ipsum dolor sit amet",
                                scope="inventory",
                            ),
                            management_api.models.DeviceAttribute(
                                name="number", value=2 ** 47, scope="inventory"
                            ),
                        ],
                    ),
                    management_api.models.Device(
                        id="d8b04e01-690d-41ce-8c6d-ab079a04d488",
                        attributes=[
                            management_api.models.DeviceAttribute(
                                name="string",
                                value="consectetur adipiscing elit",
                                scope="inventory",
                            ),
                            management_api.models.DeviceAttribute(
                                name="number", value=420.69, scope="inventory"
                            ),
                        ],
                    ),
                ],
            ),
            _TestCase(
                authorization=utils.generate_jwt(tenant_id="123456789012345678901234"),
                search_terms=management_api.models.DeviceSearchTerms(
                    filters=[
                        management_api.models.DeviceFilterTerm(
                            attribute="latest_deployment_status",
                            value="success",
                            type="$eq",
                            scope="system",
                        )
                    ],
                    sort=[
                        management_api.models.DeviceSortTerm(
                            attribute="string", scope="inventory", order="asc"
                        )
                    ],
                ),
                http_code=200,
                result=[
                    management_api.models.Device(
                        id="463e12dd-1adb-4f62-965e-b0a9ba2c93ff",
                        attributes=[
                            management_api.models.DeviceAttribute(
                                name="string",
                                value="Lorem ipsum dolor sit amet",
                                scope="inventory",
                            ),
                            management_api.models.DeviceAttribute(
                                name="number", value=2 ** 47, scope="inventory"
                            ),
                            management_api.models.DeviceAttribute(
                                name="latest_deployment_status",
                                value="success",
                                scope="system",
                            ),
                        ],
                    ),
                ],
            ),
            _TestCase(
                authorization=utils.generate_jwt(tenant_id="123456789012345678901234"),
                search_terms=management_api.models.DeviceSearchTerms(
                    geo_distance_filter=management_api.models.GeoDistanceFilter(
                        geo_distance=management_api.models.GeoDistance(
                            distance="100km",
                            location=management_api.models.GeoPoint(
                                lat=59.91, lon=10.55,
                            ),
                        ),
                    ),
                ),
                http_code=200,
                result=[
                    management_api.models.Device(
                        id="85388603-5852-437f-89c4-7549502893d5",
                        attributes=[
                            management_api.models.DeviceAttribute(
                                name="string",
                                value="incididunt ut labore",
                                scope="inventory",
                            ),
                            management_api.models.DeviceAttribute(
                                name="number", value=0.0, scope="inventory"
                            ),
                            management_api.models.DeviceAttribute(
                                name="geo-lat", value="59.94", scope="inventory"
                            ),
                            management_api.models.DeviceAttribute(
                                name="geo-lon", value="10.72", scope="inventory"
                            ),
                        ],
                    ),
                ],
            ),
            _TestCase(
                authorization=utils.generate_jwt(tenant_id="123456789012345678901234"),
                search_terms=management_api.models.DeviceSearchTerms(
                    geo_distance_filter=management_api.models.GeoDistanceFilter(
                        geo_distance=management_api.models.GeoDistance(
                            distance="10km",
                            location=management_api.models.GeoPoint(
                                lat=59.91, lon=10.55,
                            ),
                        ),
                    ),
                ),
                http_code=200,
                result=[],
            ),
            _TestCase(
                authorization=utils.generate_jwt(tenant_id="123456789012345678901234"),
                search_terms=management_api.models.DeviceSearchTerms(
                    geo_bounding_box_filter=management_api.models.GeoBoundingBoxFilter(
                        geo_bounding_box=management_api.models.GeoBoundingBox(
                            location=management_api.models.BoundingBox(
                                top_left=management_api.models.GeoPoint(
                                    lat=59.94, lon=10.70,
                                ),
                                bottom_right=management_api.models.GeoPoint(
                                    lat=59.93, lon=10.73,
                                ),
                            ),
                        ),
                    ),
                ),
                http_code=200,
                result=[
                    management_api.models.Device(
                        id="85388603-5852-437f-89c4-7549502893d5",
                        attributes=[
                            management_api.models.DeviceAttribute(
                                name="string",
                                value="incididunt ut labore",
                                scope="inventory",
                            ),
                            management_api.models.DeviceAttribute(
                                name="number", value=0.0, scope="inventory"
                            ),
                            management_api.models.DeviceAttribute(
                                name="geo-lat", value="59.94", scope="inventory"
                            ),
                            management_api.models.DeviceAttribute(
                                name="geo-lon", value="10.72", scope="inventory"
                            ),
                        ],
                    ),
                ],
            ),
            _TestCase(
                authorization=utils.generate_jwt(tenant_id="123456789012345678901234"),
                search_terms=management_api.models.DeviceSearchTerms(
                    geo_bounding_box_filter=management_api.models.GeoBoundingBoxFilter(
                        geo_bounding_box=management_api.models.GeoBoundingBox(
                            location=management_api.models.BoundingBox(
                                top_left=management_api.models.GeoPoint(
                                    lat=59.93, lon=10.73,
                                ),
                                bottom_right=management_api.models.GeoPoint(
                                    lat=59.92, lon=10.75,
                                ),
                            ),
                        ),
                    ),
                ),
                http_code=200,
                result=[],
            ),
            _TestCase(
                authorization=utils.generate_jwt(tenant_id="anIllegalTenantID"),
                search_terms=management_api.models.DeviceSearchTerms(
                    filters=[
                        management_api.models.DeviceFilterTerm(
                            attribute="foo", value="bar", type="$eq", scope="inventory",
                        )
                    ],
                ),
                http_code=200,
                result=[],
            ),
            _TestCase(
                search_terms=management_api.models.DeviceSearchTerms(
                    filters=[
                        management_api.models.DeviceFilterTerm(
                            attribute="foo", value="bar", type="$eq", scope="inventory",
                        )
                    ],
                ),
                http_code=401,
                result=[],
            ),
        ],
        ids=[
            "ok, $ne",
            "ok, $gte",
            "ok, $nin + sort",
            "ok, $lte + sort",
            "ok, $ne + sort",
            "ok, $exists",
            "ok, $regex + sort",
            "ok, latest_deployment_status",
            "ok, geo distance",
            "ok, geo distance, no results",
            "ok, geo bounding box",
            "ok, geo bounding box, no results",
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
            body, status, headers = client.search_with_http_info(
                device_search_terms=test_case.search_terms
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

            expected_ids = [dev.id for dev in test_case.result]
            actual_ids = [dev.id for dev in body]
            assert expected_ids == actual_ids
            for i, expected in enumerate(test_case.result):
                actual = body[i]
                for attr in expected.attributes:
                    assert attr in actual.attributes
