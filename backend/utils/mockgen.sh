#!/bin/sh
# Copyright 2023 Northern.tech AS
#
#    All Rights Reserved

generate_mock() {
	if [ -z "${GOFILE}" ] || [ -z "${GOLINE}" ]; then
		echo "ERROR: script not run in go generate context"
		return 1
	fi

	local REPO_ROOT=$(git rev-parse --show-toplevel)
	local PACKAGE_PATH=$(go list -f '{{.Dir}}')
	# Line following should contain the interface definition, i.e.
	# type $INTERFACE interface {...}
	local INTERFACE=$(awk "NR==$(expr ${GOLINE} + 1)"'{if($0 ~ /type.*interface/){print $2}}' ${GOFILE})
	if [ -z "${INTERFACE}" ]; then
		echo "ERROR: misplaced go:generate comment: place comment on line above declaration"
		return 1
	fi

	mkdir -p ./mocks

	# Initialize mock file with copyright header
	awk '$1 !~ /^[/][/].*/ {print ""; exit} ; {print $0}' $GOFILE >"mocks/${INTERFACE}.go"

	docker run --rm -v "${REPO_ROOT}:${REPO_ROOT}" \
		-w ${PACKAGE_PATH} \
		-u $(id -u):$(id -g) \
		vektra/mockery:v2.45 --name "${INTERFACE}" \
		--output ./mocks --print >>"mocks/${INTERFACE}.go"
}
generate_mock
