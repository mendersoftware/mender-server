#!/bin/sh
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

# tests are supposed to be located in the same directory as this file

DIR=$(readlink -f $(dirname $0))

export TEST_HOST=${TESTING_HOST:="mender-iot-manager:8080"}

export PYTHONDONTWRITEBYTECODE=1

pip3 install --quiet --force-reinstall -r /testing/requirements.txt

py.test -vv -s --tb=short --verbose \
	--host="$TEST_HOST" \
	--junitxml=$DIR/results.xml \
	$DIR/tests "$@"
