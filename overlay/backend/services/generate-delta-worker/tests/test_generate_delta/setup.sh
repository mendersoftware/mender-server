#!/bin/bash
# Copyright 2022 Northern.tech AS
#
#    All Rights Reserved

set -ex

export TEST_ROOT=$(dirname "$0")
[[ -f ${TEST_ROOT}/functions.sh ]] && . ${TEST_ROOT}/functions.sh

mkdir -p tests/data
