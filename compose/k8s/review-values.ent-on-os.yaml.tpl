global:
  url: "https://${REVIEW_APPS_PROJECT_PREFIX}-${CI_COMMIT_REF_SLUG}.${REVIEW_APPS_DOMAIN}"

  enterprise: true

  storage: "aws"
  s3:
    existingSecret: "mender-s3-artifacts"

# Every backend service except gui comes from mender-server-enterprise's own
# published image. Must be default.image, not global.image: global propagates
# to the nats subchart, which reads global.image.registry for its own images.
default:
  image:
    registry: "registry.mender.io"
    repository: "mender-server-enterprise"
    tag: "main"

  imagePullSecrets:
    - name: gitlab-registry
    - name: mender-enterprise-registry

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

# Use Helm chart included services (not recommended for production, fine for review apps)
mongodb:
  enabled: true

nats:
  enabled: true

redis:
  enabled: true

# Enterprise-only services
generate_delta_worker:
  enabled: true
  nodeSelector:
    kubernetes.io/arch: amd64

auditlogs:
  nodeSelector:
    kubernetes.io/arch: amd64

tenantadm:
  nodeSelector:
    kubernetes.io/arch: amd64

devicemonitor:
  nodeSelector:
    kubernetes.io/arch: amd64

# Backend services (open source + enterprise)
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

# gui is the one exception - this pipeline's own build, not mender-server-enterprise's
gui:
  image:
    registry: "${CI_REGISTRY}"
    repository: "northern.tech/mender/${CI_PROJECT_NAME}"
    tag: "build-${CI_COMMIT_SHA}"
  imagePullSecrets:
    - name: gitlab-registry
  nodeSelector:
    kubernetes.io/arch: amd64

api_gateway:
  nodeSelector:
    kubernetes.io/arch: amd64

device_gateway:
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
    - "${REVIEW_APPS_PROJECT_PREFIX}-${CI_COMMIT_REF_SLUG}.${REVIEW_APPS_DOMAIN}"
  tls:
    - secretName: mender-review-ingress-tls
      hosts:
        - "${REVIEW_APPS_PROJECT_PREFIX}-${CI_COMMIT_REF_SLUG}.${REVIEW_APPS_DOMAIN}"
