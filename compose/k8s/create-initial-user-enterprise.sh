#!/bin/bash

set -e
set -u

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

REQUIRED_VARS=("NAMESPACE" "RELEASE_NAME")
log_info "Checking required environment variables..."
for var in "${REQUIRED_VARS[@]}"; do
    if [ -z "${!var:-}" ]; then
        log_error "Required environment variable $var is not set"
        exit 1
    fi
done

RANDOM_SUFFIX=$(openssl rand -hex 4)
RANDOM_PASSWORD=$(openssl rand -base64 16 | tr -d "=+/" | cut -c1-16)

TENANT_NAME="${REVIEW_APP_TENANT_NAME:-review-${RANDOM_SUFFIX}}"
ADMIN_USERNAME="${REVIEW_APP_ADMIN_USERNAME:-admin-${RANDOM_SUFFIX}@mender.local}"
ADMIN_PASSWORD="${REVIEW_APP_ADMIN_PASSWORD:-${RANDOM_PASSWORD}}"

log_info "Waiting for tenantadm pod to be ready..."

TENANTADM_POD=""
for i in {1..12}; do
    TENANTADM_POD=$(kubectl get pods -n "${NAMESPACE}" -l app.kubernetes.io/component=tenantadm -o custom-columns=POD:metadata.name --no-headers 2>/dev/null | head -n1 || true)
    if [ -n "$TENANTADM_POD" ]; then
        break
    fi
    if [ $i -eq 12 ]; then
        log_error "Timeout waiting for tenantadm pod to be created"
        exit 1
    fi
    sleep 10
done

if ! kubectl wait --for=condition=ready pod/"${TENANTADM_POD}" -n "${NAMESPACE}" --timeout=300s; then
    log_error "Timeout waiting for tenantadm pod ${TENANTADM_POD} to be ready"
    exit 1
fi

log_info "tenantadm pod is ready: ${TENANTADM_POD}"
log_info "Creating initial tenant and admin user..."
log_info "Tenant name:    ${TENANT_NAME}"
log_info "Admin username: ${ADMIN_USERNAME}"

TENANTADM_OUTPUT=$(kubectl exec -n "${NAMESPACE}" "${TENANTADM_POD}" -- \
    tenantadm create-org \
    --name "${TENANT_NAME}" \
    --username "${ADMIN_USERNAME}" \
    --password "${ADMIN_PASSWORD}" \
    --addon "configure" \
    --addon "monitor" \
    --addon "troubleshoot" \
    --plan "enterprise" 2>&1) || {
    log_error "tenantadm create-org failed: ${TENANTADM_OUTPUT}"
    exit 1
}
TENANT_ID=$(echo "${TENANTADM_OUTPUT}" | tail -n1)

if [ -z "$TENANT_ID" ]; then
    log_error "Failed to create tenant"
    exit 1
fi

log_info "Tenant created successfully with ID: ${TENANT_ID}"

sleep 2

echo ""
echo "========================================="
echo -e "${GREEN}Initial Tenant and User Created Successfully!${NC}"
echo "========================================="
echo "Tenant Name:     ${TENANT_NAME}"
echo "Tenant ID:       ${TENANT_ID}"
echo "Admin Username:  ${ADMIN_USERNAME}"
echo "Admin Password:  ${ADMIN_PASSWORD}"
echo "========================================="
echo ""
echo "IMPORTANT: Save these credentials - they are randomly generated for this deployment!"
echo ""
