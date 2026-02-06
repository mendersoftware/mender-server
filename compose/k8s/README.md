# GitLab Review Apps for Mender Server on EKS

## Overview

This directory contains configuration files and scripts for deploying GitLab Review Apps to an existing Amazon EKS cluster. Review Apps automatically deploy feature branches to isolated Kubernetes namespaces, allowing developers to test changes in a staging-like environment before merging.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Implementation Components](#implementation-components)
- [AWS Infrastructure Setup](#aws-infrastructure-setup)
- [GitLab Configuration](#gitlab-configuration)
- [Usage](#usage)
- [Verification and Monitoring](#verification-and-monitoring)
- [Troubleshooting](#troubleshooting)
- [Architecture Decisions](#architecture-decisions)
- [Resource Estimates](#resource-estimates)

## Architecture

### Deployment Flow

```
GitLab Push → CI Pipeline → Build Images → Deploy to EKS
                                              ↓
                                    Namespace: mender-<branch-slug>
                                              ↓
                            ┌─────────────────┴─────────────────┐
                            ↓                                   ↓
                    Helm Install (mender chart)      AWS ALB Ingress
                            ↓                                   ↓
            ┌───────────────┼───────────────┐          TLS Certificate
            ↓               ↓               ↓                   ↓
     Backend Services   Frontend GUI   Dependencies    https://<branch>.staging.hosted.mender.io
                                            ↓
                                    MongoDB, NATS, Redis, SeaweedFS
```

### Key Components

1. **GitLab CI/CD Pipeline**: Orchestrates build and deployment
2. **Amazon EKS Cluster**: Target Kubernetes cluster
3. **Helm Chart**: External `mendersoftware/mender-helm` chart
4. **AWS ALB Ingress Controller**: Manages load balancers and ingress
5. **GitLab OIDC**: Authenticates to AWS without long-lived credentials
6. **AWS Certificate Manager**: Provides TLS certificates
7. **Route53**: DNS management for review app subdomains

### Namespace Strategy

Each review app is deployed to an isolated namespace:
- **Namespace Naming**: `mender-${CI_COMMIT_REF_SLUG}`
- **Naming**: `mender-$(echo -n ${CI_COMMIT_REF_SLUG} | md5sum | cut -c1-12)`
- **Example**: Branch `feature/new-auth` → Namespace `mender-feature-new-auth`
- **Isolation**: Each namespace contains all services and dependencies
- **Cleanup**: Namespace is deleted when branch is deleted or manually stopped

## Prerequisites

### 1. EKS Staging Cluster Preconfigured

Your EKS cluster must have:

- **Kubernetes Version**: 1.24 or later
- **AWS Load Balancer Controller**: Installed and configured
- **Sufficient Resources**: See [Resource Estimates](#resource-estimates)
- **VPC Configuration**: Public subnets for ALB, private subnets for pods
- **IAM OIDC Provider**: Associated with the cluster for service account authentication

### 2. AWS IAM Configuration

Managed by Terraform

### 3. EKS RBAC Configuration

The IAM role needs to be mapped to Kubernetes RBAC permissions. Managed by Terraform

### 4. AWS Certificate Manager (ACM)

Request or import a wildcard certificate for review app domains. Managed by
Terraform

### 5. DNS Configuration

You have to map `*.staging.hosted.mender.io` to the `mender` external ALB.

## Implementation Components

### Files in This Directory

1. **`README.md`** (this file): Comprehensive documentation
2. **`review-values.yaml.tpl`**: Helm values template for review apps
3. **`setup-review-namespace.sh`**: Script to prepare Kubernetes namespace

### GitLab CI/CD Configuration

The review apps functionality is implemented in `.gitlab-ci.yml` in the repository root:

- **Stage**: `review` (added after `build` stage)
- **Jobs**:
  - `review:deploy`: Deploys review app when branch is pushed
  - `review:stop`: Cleans up review app when branch is deleted or manually stopped

## GitLab Configuration

### Required CI/CD Variables

Add these variables to your GitLab project (Settings → CI/CD → Variables):

| Variable | Description |
|----------|-------------|
| `REVIEW_APPS_AWS_ASSUME_ROLE_ARN` | IAM role ARN for OIDC authentication |
| `REVIEW_APPS_EKS_CLUSTER_NAME` | Name of target EKS cluster |
| `REVIEW_APPS_AWS_REGION` | AWS region for EKS cluster |
| `REVIEW_APPS_DOMAIN` | Base domain for review apps |
| `REVIEW_APPS_ACM_CERTIFICATE_ARN` | ACM certificate for TLS |
| `REVIEW_APPS_REGISTRY_TOKEN` | Deploy token for Registry access (long-lived, doesn't expire with CI job) |
| `REVIEW_APPS_REGISTRY_USERNAME` | Username for the Deploy token for Registry access |

**Important**: We use GitLab deploy tokens instead of `CI_JOB_TOKEN` or `CI_REGISTRY_PASSWORD` because Karpenter may provision new nodes after the CI job completes. Deploy tokens are long-lived and allow those nodes to pull images successfully.

## Usage

### Automatic Deployment

Review apps are deployed automatically when you push to a non-protected branch: TODO: review this statement

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/my-new-feature
   ```

2. **Make changes and push**:
   ```bash
   git add .
   git commit -m "feat: Add new feature" -s
   git push origin feature/my-new-feature
   ```

3. **Monitor deployment**:
   - Go to GitLab → CI/CD → Pipelines
   - Wait for the `review:deploy` job to complete
   - Check the job output for the review app URL

4. **Access your review app**:
   - URL: `https://feature-my-new-feature.staging.hosted.mender.io`
   - The URL is also available in GitLab → Deployments → Environments

### Manual Stop

To manually stop a review app:

1. Go to GitLab → Deployments → Environments
2. Find your review app environment
3. Click the "Stop" button

### Automatic Cleanup

Review apps are automatically cleaned up when:
- The feature branch is deleted
- The merge request is merged (and branch is auto-deleted)
- After 7 days (TODO: review this)

## Verification and Monitoring

### Check Deployment Status

After pushing to a branch, verify the deployment:

```bash
# Set namespace variable
export NAMESPACE=mender-<your-branch-slug>

# Check namespace exists
kubectl get namespace $NAMESPACE

# Check Helm release
helm list -n $NAMESPACE

# Check all pods are running
kubectl get pods -n $NAMESPACE

# Check ingress configuration
kubectl get ingress -n $NAMESPACE

# Check services
kubectl get svc -n $NAMESPACE
```

## Troubleshooting

### Common Issues

#### 1. "Error: INSTALLATION FAILED: timed out waiting for the condition"

**Cause**: Pods are taking too long to start, usually due to image pull issues or resource constraints.

**Solution**:
```bash
# Check pod status
kubectl get pods -n $NAMESPACE

# Describe the problematic pod
kubectl describe pod <pod-name> -n $NAMESPACE

# Check events
kubectl get events -n $NAMESPACE --sort-by='.lastTimestamp'
```

Common causes:
- **ImagePullBackOff**: Check GitLab registry credentials
- **Insufficient resources**: Scale up cluster or reduce review app resources
- **PVC binding issues**: Check storage class and provisioner

#### 2. "502 Bad Gateway" or "503 Service Unavailable"

**Cause**: Backend pods are not ready or ingress routing is misconfigured.

**Solution**:
```bash
# Check pod readiness
kubectl get pods -n $NAMESPACE

# Check service endpoints
kubectl get endpoints -n $NAMESPACE

# Verify ingress backend configuration
kubectl describe ingress -n $NAMESPACE

# Test service directly (port-forward)
kubectl port-forward -n $NAMESPACE svc/gui 8080:80
```

#### 4. TLS Certificate Issues

**Cause**: ACM certificate is not properly configured or not in the correct region.

**Solution**:
- Verify certificate ARN in GitLab variables
- Ensure certificate is in the same region as the ALB
- Check certificate validation status in ACM console
- Verify ingress annotations include correct certificate ARN

#### 5. DNS Resolution Failures

**Cause**: DNS records not configured or propagation delay.

**Solution**:
```bash
# Check DNS resolution
dig <branch-slug>.staging.hosted.mender.io

# Verify wildcard DNS is set up
dig random-branch.staging.hosted.mender.io

# Check ALB DNS name
kubectl get ingress -n $NAMESPACE -o jsonpath='{.items[0].status.loadBalancer.ingress[0].hostname}'
```

## Additional Resources

- [GitLab Review Apps Documentation](https://docs.gitlab.com/ee/ci/review_apps/)
- [Mender Helm Chart](https://github.com/mendersoftware/mender-helm)
