#!/bin/bash

# Post-install script to create initial user (and tenant for enterprise) for Mender Review Apps
# This script runs after Helm chart installation completes
#
# Usage:
#   ./create-initial-user.sh               # OSS mode
#   ./create-initial-user.sh --enterprise   # Enterprise mode (creates tenant + user)

set -e  # Exit on error
set -u  # Exit on undefined variable

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored messages
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1" >&2
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1" >&2
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

ENTERPRISE=false
if [ "${1:-}" = "--enterprise" ]; then
    ENTERPRISE=true
fi

# Required environment variables
REQUIRED_VARS=(
    "NAMESPACE"
    "RELEASE_NAME"
)

# Check required environment variables
log_info "Checking required environment variables..."
for var in "${REQUIRED_VARS[@]}"; do
    if [ -z "${!var:-}" ]; then
        log_error "Required environment variable $var is not set"
        exit 1
    fi
done
log_info "All required environment variables are set"

# Generate random credentials for this review app
# Can be overridden via environment variables if needed
RANDOM_SUFFIX=$(openssl rand -hex 4)
RANDOM_PASSWORD=$(openssl rand -base64 16 | tr -d "=+/" | cut -c1-16)

ADMIN_USERNAME="${REVIEW_APPS_ADMIN_USERNAME:-admin-${RANDOM_SUFFIX}@mender.local}"
ADMIN_PASSWORD="${REVIEW_APPS_ADMIN_PASSWORD:-${RANDOM_PASSWORD}}"

log_info "Generated credentials for this review app deployment"

# Wait for a pod matching the given label to be created and ready
wait_for_pod() {
    local component="$1"
    local pod=""

    log_info "Waiting for ${component} pod to be ready..."

    for i in {1..12}; do
        pod=$(kubectl get pods -n "${NAMESPACE}" -l "app.kubernetes.io/component=${component}" -o custom-columns=POD:metadata.name --no-headers 2>/dev/null | head -n1 || true)
        if [ -n "$pod" ]; then
            break
        fi
        if [ $i -eq 12 ]; then
            log_error "Timeout waiting for ${component} pod to be created"
            exit 1
        fi
        sleep 10
    done

    # Wait for the pod to become ready (up to 5 minutes)
    if ! kubectl wait --for=condition=ready pod/"${pod}" -n "${NAMESPACE}" --timeout=300s; then
        log_error "Timeout waiting for ${component} pod ${pod} to be ready"
        exit 1
    fi

    log_info "${component} pod is ready: ${pod}"
    echo "${pod}"
}

if [ "$ENTERPRISE" = true ]; then
    TENANT_NAME="${REVIEW_APP_TENANT_NAME:-review-${RANDOM_SUFFIX}}"

    POD=$(wait_for_pod "tenantadm")
    USERADM_POD=$(wait_for_pod "useradm")

    log_info "Creating initial tenant and admin user..."
    log_info "Tenant name:    ${TENANT_NAME}"
    log_info "Admin username: ${ADMIN_USERNAME}"

    TENANT_ID=$(kubectl exec -n "${NAMESPACE}" "${POD}" -- \
        tenantadm create-org \
        --name "${TENANT_NAME}" \
        --username "${ADMIN_USERNAME}" \
        --password "${ADMIN_PASSWORD}" \
        --addon "configure" \
        --addon "monitor" \
        --addon "troubleshoot" \
        --plan "enterprise" 2>&1 | tail -n1)

    if [ -z "$TENANT_ID" ]; then
        log_error "Failed to create tenant"
        exit 1
    fi

    log_info "Tenant created successfully with ID: ${TENANT_ID}"
else
    POD=$(wait_for_pod "useradm")

    log_info "Creating initial user..."
    log_info "Admin username: ${ADMIN_USERNAME}"

    USER_ID=$(kubectl exec -n "${NAMESPACE}" "${POD}" -- \
        useradm create-user \
        --username "${ADMIN_USERNAME}" \
        --password "${ADMIN_PASSWORD}" 2>&1 | tail -n1)

    if [ -z "$USER_ID" ]; then
        log_error "Failed to create user"
        exit 1
    fi

    log_info "User created successfully with ID: ${USER_ID}"
fi

# Wait a moment for workflows to execute
sleep 2

# Display success message with login credentials
echo ""
echo "========================================="
if [ "$ENTERPRISE" = true ]; then
    echo -e "${GREEN}Initial Tenant and User Created Successfully!${NC}"
    echo "========================================="
    echo "Tenant Name:     ${TENANT_NAME}"
    echo "Tenant ID:       ${TENANT_ID}"
else
    echo -e "${GREEN}Initial User Created Successfully!${NC}"
    echo "========================================="
    echo "User ID:         ${USER_ID}"
fi
echo "Admin Username:  ${ADMIN_USERNAME}"
echo "Admin Password:  ${ADMIN_PASSWORD}"
echo "========================================="
echo ""
echo "IMPORTANT: Save these credentials - they are randomly generated for this deployment!"
echo ""
