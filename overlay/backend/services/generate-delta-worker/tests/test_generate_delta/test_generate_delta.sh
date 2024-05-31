#!/bin/bash
# Copyright 2022 Northern.tech AS
#
#    All Rights Reserved

set -ex

export TEST_ROOT=$(dirname "$0")
[[ -f ${TEST_ROOT}/functions.sh ]] && . ${TEST_ROOT}/functions.sh
[[ -f ${TEST_ROOT}/config.sh ]] && . ${TEST_ROOT}/config.sh

${TEST_ROOT}/setup.sh
test_main $@
