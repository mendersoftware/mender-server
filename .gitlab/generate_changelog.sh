#!/bin/bash

set -e

RELEASE_VERSION=$1
CHANGELOG_SUFFIX=$2
GITHUB_REPO_URL=$3
CI_COMMIT_REF_NAME=$4

Generating changelog file CHANGELOG${CHANGELOG_SUFFIX:-}.md for release ${RELEASE_VERSION}
mv CHANGELOG${CHANGELOG_SUFFIX}.md.${CI_COMMIT_REF_NAME} CHANGELOG${CHANGELOG_SUFFIX}.md
if [ "${CHANGELOG_SUFFIX}" == "-saas" ]; then
    git cliff --unreleased --prepend CHANGELOG${CHANGELOG_SUFFIX}.md --github-repo ${GITHUB_REPO_URL} --use-branch-tags --tag ${RELEASE_VERSION}
else
    git cliff --unreleased --prepend CHANGELOG${CHANGELOG_SUFFIX}.md --github-repo ${GITHUB_REPO_URL} --use-branch-tags --tag ${RELEASE_VERSION} --ignore-tags saas
fi
git add CHANGELOG${CHANGELOG_SUFFIX}.md
