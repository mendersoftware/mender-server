#!/bin/bash

# Post-install script to create initial  user for Mender Review Apps
# This script runs after Helm chart installation completes

set -e  # Exit on error
set -u  # Exit on undefined variable

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored messages
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

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

ADMIN_USERNAME="${REVIEW_APP_ADMIN_USERNAME:-admin-${RANDOM_SUFFIX}@mender.local}"
ADMIN_PASSWORD="${REVIEW_APP_ADMIN_PASSWORD:-${RANDOM_PASSWORD}}"

log_info "Generated credentials for this review app deployment"

log_info "Waiting for useradm pod to be ready..."

# Wait for the useradm pod to be created (up to 2 minutes)
USERADM_POD=""
for i in {1..12}; do
    USERADM_POD=$(kubectl get pods -n "${NAMESPACE}" -l app.kubernetes.io/component=useradm -o custom-columns=POD:metadata.name --no-headers 2>/dev/null | head -n1 || true)
    if [ -n "$USERADM_POD" ]; then
        break
    fi
    if [ $i -eq 12 ]; then
        log_error "Timeout waiting for useradm pod to be created"
        exit 1
    fi
    sleep 10
done

# Wait for the pod to become ready (up to 5 minutes)
if ! kubectl wait --for=condition=ready pod/"${USERADM_POD}" -n "${NAMESPACE}" --timeout=300s; then
    log_error "Timeout waiting for useradm pod ${USERADM_POD} to be ready"
    exit 1
fi

log_info "useradm pod is ready: ${USERADM_POD}"

log_info "Creating initial user..."

# Create initial user
log_info "Admin username: ${ADMIN_USERNAME}"

USER_ID=$(kubectl exec -n "${NAMESPACE}" "${USERADM_POD}" -- \
    useradm create-user \
    --username "${ADMIN_USERNAME}" \
    --password "${ADMIN_PASSWORD}" 2>&1 | tail -n1)

if [ -z "$USER_ID" ]; then
    log_error "Failed to create user"
    exit 1
fi

log_info "User created successfully with ID: ${USER_ID}"

# Wait a moment for workflows to execute
sleep 2

# Display success message with login credentials
echo ""
echo "========================================="
echo -e "${GREEN}Initial User Created Successfully!${NC}"
echo "========================================="
echo "User ID:         ${USER_ID}"
echo "Admin Username:  ${ADMIN_USERNAME}"
echo "Admin Password:  ${ADMIN_PASSWORD}"
echo "========================================="
echo ""
echo "IMPORTANT: Save these credentials - they are randomly generated for this deployment!"
echo ""
