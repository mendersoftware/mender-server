# Helm values template for Mender Review Apps
# This file uses environment variable substitution via envsubst
# Variables are provided by GitLab CI/CD pipeline

# Global configuration
global:
  # Domain for this review app
  # Example: feature-auth.staging.hosted.mender.io
  url: "https://${CI_COMMIT_REF_SLUG}.${REVIEW_APPS_DOMAIN}"

  enterprise: false

  # S3 storage configuration - use existing secret created during namespace setup
  storage: "aws"
  s3:
    existingSecret: "mender-s3-artifacts"

# Image configuration - override all services to use GitLab registry
# All images are tagged with build-${CI_COMMIT_SHA}
default:
  image:
    registry: "${CI_REGISTRY}"
    repository: "northern.tech/mender/${CI_PROJECT_NAME}"
    tag: "build-${CI_COMMIT_SHA}"

  # Image pull secrets for GitLab registry
  imagePullSecrets:
    - name: gitlab-registry
  affinity:
    nodeAffinity:
      preferredDuringSchedulingIgnoredDuringExecution:
        - preference:
            matchExpressions:
              - key: kubernetes.io/arch
                operator: In
                values:
                  - amd64
          weight: 90
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
          - matchExpressions:
              - key: node.kubernetes.io/nodegroup
                operator: In
                values:
                  - review
  tolerations:
    - effect: NoSchedule
      key: review
      operator: Equal
      value: "true"

deployments:
  nodeSelector:
    kubernetes.io/arch: amd64

device_auth:
  nodeSelector:
    kubernetes.io/arch: amd64

deviceconfig:
  nodeSelector:
    kubernetes.io/arch: amd64

deviceconnect:
  nodeSelector:
    kubernetes.io/arch: amd64

inventory:
  nodeSelector:
    kubernetes.io/arch: amd64

iot_manager:
  nodeSelector:
    kubernetes.io/arch: amd64

useradm:
  nodeSelector:
    kubernetes.io/arch: amd64

workflows:
  nodeSelector:
    kubernetes.io/arch: amd64

workflows_worker:
  nodeSelector:
    kubernetes.io/arch: amd64

create_artifact_worker:
  nodeSelector:
    kubernetes.io/arch: amd64

gui:
  nodeSelector:
    kubernetes.io/arch: amd64

ingress:
  enabled: true
  ingressClassName: alb
  path: /
  annotations:
    alb.ingress.kubernetes.io/actions.ssl-redirect:
      '{"Type": "redirect", "RedirectConfig":
      { "Protocol": "HTTPS", "Port": "443", "StatusCode": "HTTP_301"}}'
    alb.ingress.kubernetes.io/backend-protocol: HTTP
    alb.ingress.kubernetes.io/certificate-arn: "${REVIEW_APPS_ACM_CERTIFICATE_ARN}"
    alb.ingress.kubernetes.io/group.name: mender
    alb.ingress.kubernetes.io/healthcheck-path: /ui/
    alb.ingress.kubernetes.io/listen-ports: '[{"HTTP": 80}, {"HTTPS":443}]'
    alb.ingress.kubernetes.io/load-balancer-attributes: routing.http2.enabled=true,idle_timeout.timeout_seconds=600
    alb.ingress.kubernetes.io/scheme: internet-facing
    alb.ingress.kubernetes.io/ssl-policy: ELBSecurityPolicy-TLS13-1-2-Res-2021-06
    alb.ingress.kubernetes.io/target-type: ip
    alb.ingress.kubernetes.io/ip-address-type: dualstack
  hosts:
    - "${CI_COMMIT_REF_SLUG}.${REVIEW_APPS_DOMAIN}"
  tls:
    - secretName: mender-review-ingress-tls
      hosts:
        - "${CI_COMMIT_REF_SLUG}.${REVIEW_APPS_DOMAIN}"

