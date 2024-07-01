#!/bin/sh
# Copyright 2020 Northern.tech AS
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

generate_mock() {
    if [ -z "${GOFILE}" ] || [ -z "${GOLINE}" ]; then
        echo "ERROR: script not run in go generate context"
        return 1
    fi

    local REPO_ROOT=$(git rev-parse --show-toplevel)
    local PACKAGE_PATH=${PWD##$REPO_ROOT}
    # Line following should contain the interface definition, i.e.
    # type $INTERFACE interface {...}
    local INTERFACE=$(awk "NR==$(expr ${GOLINE} + 1)"'{if($0 ~ /type.*interface/){print $2}}' ${GOFILE})
    if [ -z "${INTERFACE}" ]; then
        echo "ERROR: misplaced go:generate comment: place comment on line above declaration"
        return 1
    fi

    mkdir -p ./mocks

    # Initialize mock file with copyright header
    awk '$1 !~ /^[/][/].*/ {print ""; exit} ; {print $0}' $GOFILE > "mocks/${INTERFACE}.go"

    docker run --rm -v ${REPO_ROOT}:/wd \
        -w /wd/${PACKAGE_PATH} \
        -u $(id -u):$(id -g) \
        vektra/mockery --name "${INTERFACE}" \
        --output ./mocks --print >> "mocks/${INTERFACE}.go"
}
generate_mock
