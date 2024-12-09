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
from multiprocessing import Process
from flask import Flask
import logging
log = logging.getLogger('werkzeug')
log.setLevel(logging.ERROR)

app = Flask(__name__)
server = None

@app.route("/api/internal/v1/tenantadm/tenants/verify", methods=["POST"])
def verify():
    return "", 200

class fake_tenantadm:
    def __enter__(self):
        self.server = Process(target=app.run, kwargs={'host': '0.0.0.0'})
        self.server.daemon=True
        self.server.start()
    def __exit__(self, type, value, traceback):
        self.server.terminate()
        self.server.join()
