#!/bin/bash

docker_compose_cmd="docker compose"

MENDER_SERVER_PATH=$(git rev-parse --show-toplevel)
export MENDER_SERVER_PATH

DOWNLOAD_REQUIREMENTS="true"

COMPOSE_FILES_BASE="-f $MENDER_SERVER_PATH/backend/tests/docker/docker-compose.backend-tests.yml"
COMPOSE_FILES_OPEN=" \
        -f $MENDER_SERVER_PATH/backend/tests/docker-compose.yml \
        -f $MENDER_SERVER_PATH/backend/tests/docker/docker-compose.backend-tests.yml \
        "
COMPOSE_FILES_ENTERPRISE="${COMPOSE_FILES_OPEN} \
        -f $MENDER_SERVER_PATH/backend/tests/docker/docker-compose.backend-tests-enterprise.yml \
        "
COMPOSE_FILES_COMPAT="${COMPOSE_FILES_OPEN} \
        -f $MENDER_SERVER_PATH/backend/tests/docker/docker-compose.backend-tests-compat.yml \
        "
COMPOSE_FILE_STORAGE_AZURE="-f $MENDER_SERVER_PATH/backend/tests/docker/docker-compose.storage.azblob.yml"

COMPOSE_FILE_AZURE_SETUP="-f $MENDER_SERVER_PATH/backend/tests/docker/docker-compose.azblob.setup.yml"

COMPOSE_FILES=""

COMPOSE_UP_EXTRA_ARGS="--remove-orphans"

compose_cmd() {
    $docker_compose_cmd -p backend-tests $COMPOSE_FILES $@
}

PYTEST_FILTER_OPEN="not Enterprise and not Multitenant"
PYTEST_FILTER_ENTERPRISE="Enterprise"
PYTEST_FILTER=""

PYTEST_REPORT_OPEN="--self-contained-html \
        --junit-xml=/tests/logs/results_integration_open.xml \
        --html=/tests/logs/report_integration_open.html"
PYTEST_REPORT_ENTERPRISE="--self-contained-html \
        --junit-xml=/tests/logs/tests/results_integration_enterprise.xml \
        --html=/tests/logs/report_integration_enterprise.html"
PYTEST_REPORT=""

PYTEST_ADDOPTS=""

DOCS="$MENDER_SERVER_PATH/backend/tests/docs"

usage() {
    echo "runner script for backend-specific integration tests"
    echo ""
    echo "./backend-tests"
    echo -e "\t-h --help"
    echo -e "\t-s --suite <SUITE>\trun specific test suite"
    echo -e "\t                  \t<SUITE> can be 'open' (default), 'enterprise', 'all'"
    echo -e "\t--no-download     \tdo not download the external dependencies"
    echo -e "\t-c --skip-cleanup \tleave containers running after tests"
    echo -e "\t other args will be passed to the testing container's py.test command"
    echo ""
    echo -e "examples:"
    echo -e "run default ST setup:"
    echo -e "\t./run"
    echo -e "run tests Enterprise tests"
    echo -e "\t./run -s enterprise"
    echo -e "run specific test TestGetDevices in both setups"
    echo -e "\t./run -s all -k TestGetDevices"
}

TEST_SUITES=("open")
if [ -d "${MENDER_SERVER_PATH}/backend/services/auditlogs" ]; then
    TEST_SUITES=("enterprise")
fi

parse_args() {
    while [ $# -gt 0 ]; do
        case "$1" in
            -h | --help)
                usage
                exit
                ;;
            -s | --suite)
                shift 1
                case "$1" in
                    open) ;;
                    enterprise)
                        TEST_SUITES=("enterprise")
                        ;;
                    compat)
                        TEST_SUITES=("compat")
                        ;;
                    all)
                        TEST_SUITES=("open" "enterprise")
                        ;;
                    *)
                        usage
                        exit
                        ;;
                esac
                ;;
            -f | -f=*)
                usage
                exit 1
                ;;
            --no-download)
                DOWNLOAD_REQUIREMENTS=""
                ;;
            -c | --skip-cleanup)
                SKIP_CLEANUP=1
                ;;
            --)
                shift 1
                USER_PYTEST_ADDOPTS="$USER_PYTEST_ADDOPTS $@"
                return 0
                ;;
            *)
                USER_PYTEST_ADDOPTS="$USER_PYTEST_ADDOPTS $1"
                ;;
        esac
        shift 1
    done
}

get_runner_requirements() {
    for svc_path in $(ls -d $MENDER_SERVER_PATH/backend/services/*); do
        svc=$(basename $svc_path)
        if [ -d "${svc_path}/docs" ]; then
            mkdir -p "${DOCS}/${svc}"
            find "${svc_path}/docs" -type f -name "*_api*.yml" -exec cp {} "${DOCS}/${svc}/" \;
        fi
    done
}

run_tests() {
    local RUN_ARGS="--use-aliases"
    # Need to start the backend first
    compose_cmd up $COMPOSE_UP_EXTRA_ARGS -d
    if [ -n "$AZURE_IOTHUB_CONNECTIONSTRING" ]; then
        RUN_ARGS="${RUN_ARGS} -e AZURE_IOTHUB_CONNECTIONSTRING=${AZURE_IOTHUB_CONNECTIONSTRING}"
    fi
    # ... then the runner
    compose_cmd run $RUN_ARGS integration-tester
    return $?
}

prepare_pytest_args() {
    filter="none"
    for val in $USER_PYTEST_ADDOPTS; do
        if [ "$val" == "-k" ]; then
            filter="next"
        elif [ "$filter" == "next" ]; then
            PYTEST_FILTER="$PYTEST_FILTER and $val"
            filter="done"
        else
            PYTEST_ADDOPTS="$PYTEST_ADDOPTS $val"
        fi
    done

    echo "-- using PYTEST_FILTER=$PYTEST_FILTER"
    PYTEST_ADDOPTS="$PYTEST_ADDOPTS -k '$PYTEST_FILTER' $PYTEST_REPORT"

    export PYTEST_ADDOPTS
}

cleanup() {
    if [ -z $SKIP_CLEANUP ]; then
        compose_cmd down -v --remove-orphans
    else
        # Remove stopped container created by $docker_compose_cmd run
        compose_cmd rm -f
    fi
    if [ -n "$AZURE_STORAGE_CONTAINER_PREFIX" ]; then
        # Remove the blob storage container if we created one
        echo "Removing blob storage container '$AZURE_STORAGE_CONTAINER_NAME'"
        $docker_compose_cmd $COMPOSE_FILE_AZURE_SETUP run --rm \
            azblob-container-teardown $AZURE_STORAGE_CONTAINER_NAME
        $docker_compose_cmd $COMPOSE_FILE_AZURE_SETUP down
    fi
}

parse_args "$@"

if [[ -n "$DOWNLOAD_REQUIREMENTS" ]]; then
    get_runner_requirements
fi

script_failed=0

for suite in "${TEST_SUITES[@]}"; do
    case "$suite" in
        open)
            COMPOSE_FILES="$COMPOSE_FILES_OPEN"
            PYTEST_FILTER="$PYTEST_FILTER_OPEN"
            PYTEST_REPORT="$PYTEST_REPORT_OPEN"
            ;;
        compat)
            COMPOSE_FILES="$COMPOSE_FILES_COMPAT"
            PYTEST_FILTER="$PYTEST_FILTER_OPEN"
            PYTEST_REPORT="$PYTEST_REPORT_OPEN"
            ;;
        enterprise)
            COMPOSE_FILES="$COMPOSE_FILES_ENTERPRISE"
            PYTEST_FILTER="$PYTEST_FILTER_ENTERPRISE"
            PYTEST_REPORT="$PYTEST_REPORT_ENTERPRISE"
            ;;
    esac

    # Are we running in Azure mode?
    if [ -n "$AZURE_STORAGE_CONTAINER_PREFIX" ]; then
        # Create a container (bucket) will be removed in cleanup
        rand_suffix=$(cat /dev/urandom | tr -dc 'a-z' | head -c 8)
        date_suffix=$(date '+%Y%m%d-%H%M%S')
        container="${AZURE_STORAGE_CONTAINER_PREFIX}-${date_suffix}-${rand_suffix}"
        export AZURE_STORAGE_CONTAINER_NAME="${container}"
        echo "Creating container '$AZURE_STORAGE_CONTAINER_NAME'"
        $docker_compose_cmd $COMPOSE_FILE_AZURE_SETUP run --rm \
            azblob-container-setup $AZURE_STORAGE_CONTAINER_NAME
        ec=$?
        if [ $ec -ne 0 ]; then
            echo -n 'Failed to setup blob storage for testing'
            exit $ec
        fi
    fi
    if [ -n "$K8S" ]; then
        COMPOSE_FILES="${COMPOSE_FILES_BASE}"
    elif [ -n "$AZURE_STORAGE_CONTAINER_NAME" ]; then
        COMPOSE_FILES="${COMPOSE_FILES} ${COMPOSE_FILE_STORAGE_AZURE}"
        COMPOSE_UP_EXTRA_ARGS="${COMPOSE_UP_EXTRA_ARGS} --scale minio=0"
        PYTEST_ADDOPTS="$PYTEST_ADDOPTS -m storage_test"
    fi

    prepare_pytest_args
    run_tests
    run_tests_retcode=$?
    if [ $script_failed -eq 0 ]; then
        script_failed=$run_tests_retcode
    fi

    if [ $run_tests_retcode -ne 0 ]; then
        tmppath=$(mktemp -p "${MENDER_SERVER_PATH}/backend/tests/logs" logs.XXXXXX)
        echo "-- tests failed, dumping logs to $tmppath"
        compose_cmd logs > "$tmppath" 2>&1
    fi

    cleanup

done

exit $script_failed
