# Copyright 2024 Northern.tech AS
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

from common import clean_db, mongo, internal_api
import internal_v1 as ia


class TestInternalApi:
    def test_create_tenant_ok(self, internal_api, clean_db):
        try:
            internal_api.create_tenant("foobar")
        except ia.ApiException as e:
            assert e.status == 201
