#!/usr/bin/env bash

# Setup script for Mender Review Apps Kubernetes namespace
# This script prepares the namespace with required secrets and labels

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

# Required environment variables from GitLab CI/CD
REQUIRED_VARS=(
    "NAMESPACE"
    "CI_PROJECT_PATH_SLUG"
    "CI_ENVIRONMENT_SLUG"
    "CI_REGISTRY"
    "REVIEW_APPS_REGISTRY_USERNAME"
    "REVIEW_APPS_REGISTRY_TOKEN"
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

# Verify kubectl is installed and configured
if ! command -v kubectl &> /dev/null; then
    log_error "kubectl command not found. Please install kubectl."
    exit 1
fi

# Test cluster connectivity
log_info "Testing Kubernetes cluster connectivity..."
if ! kubectl cluster-info &> /dev/null; then
    log_error "Cannot connect to Kubernetes cluster. Check your kubeconfig."
    exit 1
fi
log_info "Successfully connected to Kubernetes cluster"

# Create namespace if it doesn't exist
log_info "Creating namespace: ${NAMESPACE}"
if kubectl get namespace "${NAMESPACE}" &> /dev/null; then
    log_warn "Namespace ${NAMESPACE} already exists"
else
    kubectl create namespace "${NAMESPACE}"
    log_info "Created namespace ${NAMESPACE}"
fi

# Label the namespace for GitLab integration
log_info "Labeling namespace with GitLab metadata..."
kubectl label namespace "${NAMESPACE}" \
    app.gitlab.com/app="${CI_PROJECT_PATH_SLUG}" \
    app.gitlab.com/env="${CI_ENVIRONMENT_SLUG}" \
    mender.io/review-app="true" \
    --overwrite
log_info "Namespace labels applied"

# Create GitLab registry pull secret
log_info "Creating GitLab registry pull secret..."

# Delete existing secret if present
if kubectl get secret gitlab-registry -n "${NAMESPACE}" &> /dev/null; then
    log_warn "Secret gitlab-registry already exists in namespace ${NAMESPACE}, deleting..."
    kubectl delete secret gitlab-registry -n "${NAMESPACE}"
fi

# Create new secret
# Note: Using deploy token credentials (long-lived, created via Terraform)
# These persist indefinitely, allowing nodes provisioned after CI job completes to pull images
kubectl create secret docker-registry gitlab-registry \
    --docker-server="${CI_REGISTRY}" \
    --docker-username="${REVIEW_APPS_REGISTRY_USERNAME}" \
    --docker-password="${REVIEW_APPS_REGISTRY_TOKEN}" \
    --docker-email="gitlab-ci@${CI_PROJECT_PATH_SLUG}.local" \
    -n "${NAMESPACE}"

log_info "GitLab registry pull secret created"

# Verify secret was created
if kubectl get secret gitlab-registry -n "${NAMESPACE}" &> /dev/null; then
    log_info "Successfully verified GitLab registry secret"
else
    log_error "Failed to create GitLab registry secret"
    exit 1
fi

# Deploy SeaweedFS for S3-compatible artifact storage
log_info "Deploying SeaweedFS for artifact storage..."

# Generate random access keys for SeaweedFS (like Vagrantfile does)
ADMIN_KEY=$(pwgen -s 32 1)
ADMIN_SECRET=$(pwgen -s 64 1)
READ_KEY=$(pwgen -s 32 1)
READ_SECRET=$(pwgen -s 64 1)

# Create SeaweedFS S3 configuration
SEAWEEDFS_CONFIG=$(cat <<EOF
{
  "identities": [
    {
      "name": "anvAdmin",
      "credentials": [
        {
          "accessKey": "$ADMIN_KEY",
          "secretKey": "$ADMIN_SECRET"
        }
      ],
      "actions": ["Admin", "Read", "Write"]
    },
    {
      "name": "anvReadOnly",
      "credentials": [
        {
          "accessKey": "$READ_KEY",
          "secretKey": "$READ_SECRET"
        }
      ],
      "actions": ["Read"]
    }
  ]
}
EOF
)

# Create SeaweedFS configuration secret
log_info "Creating SeaweedFS S3 secret..."
kubectl create secret generic seaweedfs-mender-s3-secret \
    --from-literal=admin_access_key_id="$ADMIN_KEY" \
    --from-literal=admin_secret_access_key="$ADMIN_SECRET" \
    --from-literal=read_access_key_id="$READ_KEY" \
    --from-literal=read_secret_access_key="$READ_SECRET" \
    --from-literal=seaweedfs_s3_config="$SEAWEEDFS_CONFIG" \
    -n "${NAMESPACE}" \
    --dry-run=client -o yaml | kubectl apply -f -

# Create Mender S3 artifacts secret (used by Mender services)
log_info "Creating Mender S3 artifacts secret..."
kubectl create secret generic mender-s3-artifacts \
    --from-literal=AWS_AUTH_KEY="$ADMIN_KEY" \
    --from-literal=AWS_AUTH_SECRET="$ADMIN_SECRET" \
    --from-literal=AWS_BUCKET="mender-artifacts-storage-seaweedfs" \
    --from-literal=AWS_FORCE_PATH_STYLE="true" \
    --from-literal=AWS_URI="http://seaweedfs-s3.${NAMESPACE}.svc.cluster.local:8333" \
    --from-literal=AWS_REGION="us-east-1" \
    -n "${NAMESPACE}" \
    --dry-run=client -o yaml | kubectl apply -f -

# Deploy SeaweedFS using Helm
log_info "Installing SeaweedFS Helm chart..."
helm repo add seaweedfs https://seaweedfs.github.io/seaweedfs/helm
helm repo update

helm upgrade --install seaweedfs seaweedfs/seaweedfs \
    -n "${NAMESPACE}" \
    --set master.replicas=1 \
    --set volume.replicas=1 \
    --set filer.enabled=true \
    --set filer.replicas=1 \
    --set filer.s3.enabled=true \
    --set filer.s3.port=8333 \
    --set filer.s3.enableAuth=true \
    --set filer.s3.existingConfigSecret=seaweedfs-mender-s3-secret \
    --set filer.s3.createBuckets[0].name=mender-artifacts-storage-seaweedfs \
    --set s3.enabled=true \
    --set s3.enableAuth=true \
    --set s3.existingConfigSecret=seaweedfs-mender-s3-secret \
    --set global.enableSecurity=false \
    --set master.resources.requests.memory=256Mi \
    --set master.resources.requests.cpu=100m \
    --set master.resources.limits.memory=512Mi \
    --set master.resources.limits.cpu=250m \
    --set volume.resources.requests.memory=256Mi \
    --set volume.resources.requests.cpu=100m \
    --set volume.resources.limits.memory=512Mi \
    --set volume.resources.limits.cpu=250m \
    --set filer.resources.requests.memory=256Mi \
    --set filer.resources.requests.cpu=100m \
    --set filer.resources.limits.memory=512Mi \
    --set filer.resources.limits.cpu=250m \
    --wait --timeout 5m

log_info "SeaweedFS deployed successfully"

# Display namespace info
log_info "Namespace setup complete!"
echo ""
echo "========================================="
echo "Namespace: ${NAMESPACE}"
echo "Labels:"
kubectl get namespace "${NAMESPACE}" --show-labels
echo ""
echo "Secrets:"
kubectl get secrets -n "${NAMESPACE}"
echo ""
echo "SeaweedFS Pods:"
kubectl get pods -n "${NAMESPACE}" -l app.kubernetes.io/name=seaweedfs
echo "========================================="
