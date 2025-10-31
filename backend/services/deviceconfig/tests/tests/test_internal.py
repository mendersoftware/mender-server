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

import internal_v1 as internal_api


class TestInternal:
    def test_internal_alive(self):
        client = internal_api.InternalAPIClient()
        r = client.device_config_internal_check_health_with_http_info(_preload_content=False)
        assert r.status == 204

    def test_internal_health(self):
        client = internal_api.InternalAPIClient()
        r = client.device_config_internal_check_liveliness_with_http_info(_preload_content=False)
        assert r.status == 204
