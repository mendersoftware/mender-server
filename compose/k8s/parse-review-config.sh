#!/usr/bin/env bash
# Source this script to export Review App deploy override variables.
# Requires: yq
# Usage: source compose/k8s/parse-review-config.sh
#
# Exports:
#   REVIEW_GUI_IMAGE      - full image reference or empty string
#   HELM_GUI_OVERRIDE_FLAG - "-f /tmp/review-gui-override.yaml" or empty string

set -euo pipefail

REVIEW_CONFIG_FILE="${CI_PROJECT_DIR:-.}/.gitlab/review_apps.yaml"

export REVIEW_GUI_IMAGE=""
export HELM_GUI_OVERRIDE_FLAG=""

if [[ ! -f "${REVIEW_CONFIG_FILE}" ]]; then
    echo "No review_apps.yaml found - using pipeline defaults"
    return 0
fi

echo "Reading review config from ${REVIEW_CONFIG_FILE}..."

REVIEW_GUI_IMAGE="$(yq -r '.review_app.gui_image // ""' "${REVIEW_CONFIG_FILE}")"
[[ "${REVIEW_GUI_IMAGE}" == "null" ]] && REVIEW_GUI_IMAGE=""

if [[ -z "${REVIEW_GUI_IMAGE}" ]]; then
    echo "No GUI image override - using default pipeline image"
    return 0
fi

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
