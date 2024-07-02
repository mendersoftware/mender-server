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

import time
import requests
import os


def main():
    es_url = os.getenv("OPENSEARCH_URL")
    r_url = os.getenv("REPORTING_URL") + "/api/internal/v1/reporting/alive"
    for url in [es_url, r_url]:
        for i in range(300):
            try:
                r = requests.get(url)
                assert 200 <= r.status_code < 400
            except (requests.RequestException, AssertionError):
                time.sleep(1)
                pass
            else:
                break
        else:
            raise TimeoutError("timed out waiting for '%s'" % url)


if __name__ == "__main__":
    main()
