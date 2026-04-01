#!/usr/bin/env bash
# Source this script to export Review App deploy override variables.
# Requires: yq
# Usage: source compose/k8s/parse-review-config.sh
#
# Exports:
#   REVIEW_GUI_IMAGE           - full image reference or empty string
#   HELM_GUI_OVERRIDE_FLAG     - "-f /tmp/review-gui-override.yaml" or empty string
#   REVIEW_ENTERPRISE          - "true" or "false"
#   REVIEW_ENTERPRISE_IMAGE_TAG - build-<sha> tag for ENT-only services or empty string
#   HELM_ENT_OVERRIDE_FLAG     - "-f /tmp/review-ent-services-override.yaml" or empty string

set -euo pipefail

REVIEW_CONFIG_FILE="${CI_PROJECT_DIR:-.}/.gitlab/review_apps.yaml"

export REVIEW_GUI_IMAGE=""
export HELM_GUI_OVERRIDE_FLAG=""
export REVIEW_ENTERPRISE="false"
export REVIEW_ENTERPRISE_IMAGE_TAG=""
export HELM_ENT_OVERRIDE_FLAG=""

if [[ ! -f "${REVIEW_CONFIG_FILE}" ]]; then
    echo "No review_apps.yaml found - using pipeline defaults"
    return 0
fi

echo "Reading review config from ${REVIEW_CONFIG_FILE}..."

REVIEW_GUI_IMAGE="$(yq -r '.review_app.gui_image // ""' "${REVIEW_CONFIG_FILE}")"
[[ "${REVIEW_GUI_IMAGE}" == "null" ]] && REVIEW_GUI_IMAGE=""

if [[ -n "${REVIEW_GUI_IMAGE}" ]]; then
    echo "GUI image override: ${REVIEW_GUI_IMAGE}"

    GUI_IMAGE_TAG="${REVIEW_GUI_IMAGE##*:}"
    GUI_IMAGE_NO_TAG="${REVIEW_GUI_IMAGE%:*}"

    if [[ "${GUI_IMAGE_TAG}" == "${REVIEW_GUI_IMAGE}" ]]; then
        echo "ERROR: gui_image must include a tag (<image>:<tag>), got: ${REVIEW_GUI_IMAGE}"
        return 1
    fi

    GUI_IMAGE_REGISTRY="${GUI_IMAGE_NO_TAG%%/*}"
    GUI_IMAGE_REPO="${GUI_IMAGE_NO_TAG#*/}"
    GUI_IMAGE_REPO="${GUI_IMAGE_REPO%/*}"

    echo "  registry:   ${GUI_IMAGE_REGISTRY}"
    echo "  repository: ${GUI_IMAGE_REPO}"
    echo "  tag:        ${GUI_IMAGE_TAG}"

    cat > /tmp/review-gui-override.yaml <<EOF
gui:
  image:
    registry: "${GUI_IMAGE_REGISTRY}"
    repository: "${GUI_IMAGE_REPO}"
    tag: "${GUI_IMAGE_TAG}"
  imagePullSecrets:
    - name: gitlab-registry
EOF

    echo "GUI override values:"; cat /tmp/review-gui-override.yaml
    export HELM_GUI_OVERRIDE_FLAG="-f /tmp/review-gui-override.yaml"
else
    echo "No GUI image override - using default pipeline image"
fi

REVIEW_ENTERPRISE="$(yq -r '.review_app.enterprise // "false"' "${REVIEW_CONFIG_FILE}")"
[[ "${REVIEW_ENTERPRISE}" == "null" ]] && REVIEW_ENTERPRISE="false"
export REVIEW_ENTERPRISE

REVIEW_ENTERPRISE_IMAGE_TAG="$(yq -r '.review_app.enterprise_image_tag // ""' "${REVIEW_CONFIG_FILE}")"
[[ "${REVIEW_ENTERPRISE_IMAGE_TAG}" == "null" ]] && REVIEW_ENTERPRISE_IMAGE_TAG=""
export REVIEW_ENTERPRISE_IMAGE_TAG

if [[ "${REVIEW_ENTERPRISE}" == "true" && -z "${REVIEW_ENTERPRISE_IMAGE_TAG}" ]]; then
    echo "ERROR: enterprise_image_tag must be set when enterprise: true"
    return 1
fi

if [[ "${REVIEW_ENTERPRISE}" == "true" ]]; then
    echo "Enterprise mode enabled - ENT image tag: ${REVIEW_ENTERPRISE_IMAGE_TAG}"

    cat > /tmp/review-ent-services-override.yaml <<EOF
tenantadm:
  image:
    registry: "registry.gitlab.com"
    repository: "northern.tech/mender/mender-server-enterprise"
    tag: "${REVIEW_ENTERPRISE_IMAGE_TAG}"
auditlogs:
  image:
    registry: "registry.gitlab.com"
    repository: "northern.tech/mender/mender-server-enterprise"
    tag: "${REVIEW_ENTERPRISE_IMAGE_TAG}"
generate_delta_worker:
  image:
    registry: "registry.gitlab.com"
    repository: "northern.tech/mender/mender-server-enterprise"
    tag: "${REVIEW_ENTERPRISE_IMAGE_TAG}"
devicemonitor:
  image:
    registry: "registry.gitlab.com"
    repository: "northern.tech/mender/mender-server-enterprise"
    tag: "${REVIEW_ENTERPRISE_IMAGE_TAG}"
api_gateway:
  image:
    registry: "registry.gitlab.com"
    repository: "northern.tech/mender/mender-server-enterprise"
    tag: "${REVIEW_ENTERPRISE_IMAGE_TAG}"
device_gateway:
  image:
    registry: "registry.gitlab.com"
    repository: "northern.tech/mender/mender-server-enterprise"
    tag: "${REVIEW_ENTERPRISE_IMAGE_TAG}"
EOF

    echo "ENT services override values:"; cat /tmp/review-ent-services-override.yaml
    export HELM_ENT_OVERRIDE_FLAG="-f /tmp/review-ent-services-override.yaml"
fi
