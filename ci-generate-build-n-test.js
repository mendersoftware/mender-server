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

const pipelineTemplate = (jobs) => `stages:
  - build
  - test
${jobs.join("")}
`;

const getChangesRule = ({
  RULES_CHANGES_COMPARE_TO_REF,
  service,
}) => `- changes:
        paths: ["backend/services/${service}/**/*", "backend/pkg/**/*"]
        compare_to: "${RULES_CHANGES_COMPARE_TO_REF}"`;

const buildTemplate = ({
  CI_COMMIT_REF_NAME,
  CI_DEPENDENCY_PROXY_DIRECT_GROUP_IMAGE_PREFIX,
  CI_PIPELINE_ID,
  CI_REGISTRY_IMAGE,
  DOCKER_PUBLISH_COMMIT_TAG,
  RULES_CHANGES_COMPARE_TO_REF,
  service,
}) => `
build:${service}:docker-multiplatform:
  image: "registry.gitlab.com/northern.tech/mender/mender-test-containers:docker-multiplatform-buildx-v1-master"
  stage: build
  rules:
    ${getChangesRule({ RULES_CHANGES_COMPARE_TO_REF, service })}
    - if: '$CI_COMMIT_TAG =~ /^saas-[a-zA-Z0-9.]+$/'
      when: never
    - when: on_success
  tags:
    - hetzner-amd-beefy
  services:
    - name: ${CI_DEPENDENCY_PROXY_DIRECT_GROUP_IMAGE_PREFIX}/docker:20.10.21-dind
      alias: docker
  needs: []
  variables:
    DOCKER_BUILDKIT: 1
    GITLAB_REGISTRY_TAG: '${CI_REGISTRY_IMAGE}:${CI_PIPELINE_ID}'
  before_script:
    - *dind-login
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
  script:
    - echo "building ${service} with tags \${GITLAB_REGISTRY_TAG} and ${CI_REGISTRY_IMAGE}:${CI_COMMIT_REF_NAME}"
    - docker context create builder
    - docker buildx create builder --use --driver-opt network=host --buildkitd-flags '--debug --allow-insecure-entitlement network.host'
    - docker buildx build
      --cache-to type=registry,ref=${CI_REGISTRY_IMAGE}:ci_cache,mode=max
      --cache-from type=registry,ref=${CI_REGISTRY_IMAGE}:ci_cache
      --tag \${GITLAB_REGISTRY_TAG}
      --tag ${CI_REGISTRY_IMAGE}:${CI_COMMIT_REF_NAME}
      --file backend/services/${service}/Dockerfile
      --build-arg GIT_COMMIT_TAG="${DOCKER_PUBLISH_COMMIT_TAG}"
      --platform $MULTIPLATFORM_PLATFORMS
      --provenance false
      --push
      \${EXTRA_DOCKER_ARGS}
      .
`;

const testTemplate = ({
  GOLANG_VERSION,
  RULES_CHANGES_COMPARE_TO_REF,
  service,
}) => `
test:${service}:unit:
  image: golang:${GOLANG_VERSION}
  stage: test
  rules:
    ${getChangesRule({ RULES_CHANGES_COMPARE_TO_REF, service })}
    - if: '$CI_COMMIT_TAG =~ /^saas-[a-zA-Z0-9.]+$/'
      when: never
    - when: on_success
  tags:
    - hetzner-amd-beefy
  needs:
    - job: build:${service}:docker-multiplatform
      artifacts: true
  script:
    - go test -trimpath -ldflags -s -w ./...
`;

const generate = ({ branch, commit = "", ref }) => {
  const jobs = services.reduce((accu, service) => {
    accu.push(
      buildTemplate({
        CI_COMMIT_REF_NAME: "pr_124",
        CI_DEPENDENCY_PROXY_DIRECT_GROUP_IMAGE_PREFIX: "menderRegistry",
        CI_PIPELINE_ID: "123",
        CI_REGISTRY_IMAGE: `mendersoftware/${service}`,
        DOCKER_PUBLISH_COMMIT_TAG: `${ref ? ref : "master"}-${commit}`,
        RULES_CHANGES_COMPARE_TO_REF: "main",
        branch,
        service,
      })
    );
    accu.push(
      testTemplate({
        GOLANG_VERSION: "1.23.0",
        RULES_CHANGES_COMPARE_TO_REF: "main",
        service,
      })
    );
    return accu;
  }, []);
  const pipeline = pipelineTemplate(jobs);
  fs.writeFileSync("generated-build-n-test.yml", pipeline);
};

generate({ branch: "main" });
