#!/bin/bash
# Copyright 2022 Northern.tech AS
#
#    All Rights Reserved

set -ex

function log() {
    echo "$(date) ${TEST_NAME:-"unknown"}/$$ $HOSTNAME $@"
}

function equal_artifacts() {
    local a1="$1"
    local a2="$2"
    local tmpdir=$(mktemp -d)
    local i
    local rc=0

    while [ 1 ]; do
        [[ $(ls -s "$a1" | sed -e 's/ [ ]*.*//') -eq $(ls -s "$a2" | sed -e 's/ [ ]*.*//') ]] || {
            rc=1
            break
        }
        mkdir -p "${tmpdir}/${a1}" || {
            rc=2
            break
        }
        mkdir -p "${tmpdir}/${a2}" || {
            rc=3
            break
        }
        for i in "${a1}" "${a2}"; do
            tar -C "${tmpdir}/${i}" -xvf "${i}" || {
                rc=4
                break 2
            }
            tar -C "${tmpdir}/${i}/data" -xvf "${tmpdir}/${i}/data/0000.tar.gz" || {
                rc=4
                break 2
            }
        done
        diff "${tmpdir}/${a1}/data/rootfs-1.delta" "${tmpdir}/${a2}/data/rootfs-1.delta" || rc=5
        break
    done

    diff <(mender-artifact read "${a1}" | grep -vF "modified: ") <(mender-artifact read "${a2}" | grep -vF "modified: ") || rc=$?
    [[ "${tmpdir}" != "/" && ${#tmpdir} -gt ${#TMPDIR} ]] && rm -Rf "$tmpdir"
    return $rc
}

function test_main() {
    local rc
    local timeout_msg

    head -c $((1 * 1024 * 1024)) /dev/urandom > /tmp/rootfs-1
    head -c $((2 * 1024 * 1024)) /dev/urandom > /tmp/rootfs-2
    mender-artifact write rootfs-image --device-type my-device-1 --artifact-name rootfs-test-1 --file /tmp/rootfs-1 --output-path tests/data/artifact-1.mender
    mender-artifact write rootfs-image --device-type my-device-1 --artifact-name rootfs-test-2 --file /tmp/rootfs-1 --output-path tests/data/artifact-2.mender
    mender-binary-delta-generator tests/data/artifact-1.mender tests/data/artifact-2.mender --output-path tests/data/delta.mender -D "-W 204800" -- -- "-W20248843"
    generate-delta-worker.test generate-delta \
        --first-artifact-url file://tests/data/artifact-1.mender \
        --second-artifact-url file://tests/data/artifact-2.mender \
        --tenant-id 1 \
        --output-artifact-path /tmp/delta.mender \
        --output-decoder-flags "-W 204800" \
        --output-encoder-flags "-W20248843" || true
    cp /tmp/delta.mender tests/data/delta-generated.mender
    equal_artifacts /tmp/delta.mender tests/data/delta.mender || {
        log FAILED
        false
        return 1
    }
    log 1/2 PASSED

    head -c $((512 * 1024 * 1024)) /dev/urandom >> /tmp/rootfs-1
    head -c $((512 * 1024 * 1024)) /dev/urandom >> /tmp/rootfs-2
    mender-artifact write rootfs-image --device-type my-device-1 --artifact-name rootfs-test-1 --file /tmp/rootfs-1 --output-path tests/data/artifact-1.mender
    mender-artifact write rootfs-image --device-type my-device-1 --artifact-name rootfs-test-2 --file /tmp/rootfs-1 --output-path tests/data/artifact-2.mender
    timeout_msg=$(generate-delta-worker.test generate-delta \
        --timeout-seconds 1 \
        --first-artifact-url file://tests/data/artifact-1.mender \
        --second-artifact-url file://tests/data/artifact-2.mender \
        --tenant-id 1 \
        --output-artifact-path /tmp/delta.mender \
        --output-decoder-flags "-W 204800" \
        --output-encoder-flags "-W20248843" 2>&1 | tail -1 || true)
    [[ "$timeout_msg" == *"ERROR timeout reached" ]] || {
        log FAILED
        false
        return 1
    }
    log 2/2 PASSED
}
