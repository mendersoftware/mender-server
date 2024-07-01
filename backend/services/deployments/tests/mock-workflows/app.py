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
import binascii
import os
import time

from flask import Flask, jsonify

app = Flask(__name__)


def gen_random_object_id():
    timestamp = "{0:x}".format(int(time.time()))
    rest = binascii.b2a_hex(os.urandom(8)).decode("ascii")
    return timestamp + rest


@app.route("/api/v1/workflow/<name>", methods=["POST"])
def generate_artifact(name: str):
    response = {"id": gen_random_object_id(), "name": name}
    return jsonify(response), 201


def set_status(code: str):
    return "", code


@app.route("/status/<int:code>")
def _set_status(code: int):
    return "", code


@app.route("/status/<int:code>/<path:subpath>")
def _set_status_subpath(code: int, subpath: str):
    return "", code
