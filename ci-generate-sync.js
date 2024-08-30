const fs = require("fs");

const services = [
  "auditlogs",
  "create-artifact-worker",
  "deployments",
  "deviceauth",
  "deviceconfig",
  "deviceconnect",
  "devicemonitor",
  "inventory",
  "mender-gateway",
  "tenantadm",
  "useradm",
  "workflows",
];

const pipelineTemplate = (jobs, shared) => `stages:
  - version-bump
${shared}
${jobs.join("")}
`;

const getChangesRule = ({
  RULES_CHANGES_COMPARE_TO_REF,
  service,
}) => `- changes:
        paths: ["backend/services/${service}/**/*", "backend/pkg/**/*"]
        compare_to: "${RULES_CHANGES_COMPARE_TO_REF}"`;

const sharedTemplate = ({ CI_COMMIT_REF_NAME, CI_COMMIT_SHA }) => `
.trigger:mender-helm-version-bump:
  stage: version-bump
  allow_failure: true
  trigger:
    project: Northern.tech/Mender/mender-helm
    strategy: depend
  resource_group: mender-helm
  variables:
    SYNC_IMAGE_TAG: $DOCKER_PUBLISH_COMMIT_TAG
    DOCKER_PUBLISH_COMMIT_TAG: ${CI_COMMIT_REF_NAME}_${CI_COMMIT_SHA}
    GITHUB_HELM_REPO: "github.com/mendersoftware/mender-helm.git"
    CHART_DIR: "mender"
`;

const template = ({
  CI_COMMIT_REF_NAME,
  CI_PIPELINE_ID,
  CONTAINER_NAME,
  RULES_CHANGES_COMPARE_TO_REF,
  service,
}) => `
trigger:mender-helm-version-bump:${service}:staging:
  extends:
    - .trigger:mender-helm-version-bump
  rules:
    ${getChangesRule({ RULES_CHANGES_COMPARE_TO_REF, service })}
    - if: $CI_COMMIT_BRANCH =~ /^(staging)$/
  variables:
    SYNC_CONTAINER_NAME: ${CONTAINER_NAME}
    CONTAINER: ${service}
    SYNC_ENVIRONMENT: staging
    HELM_PATCH_VERSION: ${CI_PIPELINE_ID}-staging # pre-release version for trigger staging only deploy

trigger:mender-helm-version-bump:${service}:prod:
  extends:
    - .trigger:mender-helm-version-bump
  rules:
    - if: "$CI_COMMIT_TAG =~ /^saas-[a-zA-Z0-9.]+$/"
      when: manual
  variables:
    SYNC_ENVIRONMENT: prod
    HELM_PATCH_VERSION: ${CI_PIPELINE_ID}
    DOCKER_PUBLISH_COMMIT_TAG: ${CI_COMMIT_REF_NAME}

`;

const generate = ({ branch, commit = "", ref }) => {
  const jobs = services.map((service) =>
    template({
      CI_COMMIT_REF_NAME: "pr_124",
      CI_PIPELINE_ID: "123",
      RULES_CHANGES_COMPARE_TO_REF: "main",
      DOCKER_PUBLISH_COMMIT_TAG: `${ref ? ref : "master"}-${commit}`,
      CI_COMMIT_SHA: "124532t4r",
      CONTAINER_NAME: service,
      service,
    })
  );
  const pipeline = pipelineTemplate(
    jobs,
    sharedTemplate({ CI_COMMIT_REF_NAME: "pr_124", CI_COMMIT_SHA: "124532t4r" })
  );
  fs.writeFileSync("generated-helm-bump-jobs.yml", pipeline);
};

generate({ branch: "main" });
