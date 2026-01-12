#!/bin/bash

set -e

RELEASE_VERSION=$1
CHANGELOG_SUFFIX=$2
GITHUB_REPO_URL=$3
PR_BODY_FILE=$4  # Optional: if provided, also generate PR body to this file

# Detect repository type
REPO_NAME=$(basename -s .git "$(git config --get remote.origin.url)")

echo "INFO - Generating changelog file CHANGELOG${CHANGELOG_SUFFIX}.md for release ${RELEASE_VERSION} in ${REPO_NAME}"

# Common git cliff flags
CLIFF_FLAGS="--prepend CHANGELOG${CHANGELOG_SUFFIX}.md --github-repo ${GITHUB_REPO_URL} --use-branch-tags --tag ${RELEASE_VERSION}"

# Determine range and ignore-tags based on release type
if [[ ! "$RELEASE_VERSION" =~ (rc|saas) ]]; then
    # Stable release: find last stable tag for full changelog
    echo "INFO - Stable release detected, finding last stable tag"

    # Extract major.minor from current version (e.g., 4.1 from v4.1.1)
    CURRENT_MAJOR_MINOR=$(echo "$RELEASE_VERSION" | sed -E 's/^v?([0-9]+\.[0-9]+)\..*/\1/')

    # Try to find the latest stable tag in the same major.minor series
    LAST_STABLE=$(git tag --list "v${CURRENT_MAJOR_MINOR}.*" --sort=-version:refname | grep -v -E '(rc|saas)' | head -n 1)

    if [ -z "$LAST_STABLE" ]; then
        # No previous stable in this series, find latest stable overall
        echo "INFO - First release in ${CURRENT_MAJOR_MINOR} series"
        LAST_STABLE=$(git tag --list 'v*' --sort=-version:refname | grep -v -E '(rc|saas)' | head -n 1)
    fi

    if [ -n "$LAST_STABLE" ]; then
        echo "INFO - Generating changelog from $LAST_STABLE to ${RELEASE_VERSION}"
        RANGE="${LAST_STABLE}..HEAD"
    else
        echo "WARN - No previous stable release found"
        RANGE="--unreleased"
    fi
else
    # Pre-release: only changes since last tag
    echo "INFO - Pre-release detected: ${RELEASE_VERSION}"
    RANGE="--unreleased"
fi

# Configure ignore-tags based on changelog type
if [ "${CHANGELOG_SUFFIX}" == "-saas" ]; then
    # Saas changelog: ignore rc tags
    IGNORE_PATTERN=".*-rc.*"
elif [ "${CHANGELOG_SUFFIX}" == "-rc" ]; then
    # RC changelog: ignore saas tags
    IGNORE_PATTERN=".*-saas.*"
elif [ "${CHANGELOG_SUFFIX}" == "-enterprise" ]; then
    # Enterprise changelog: ignore saas and rc tags (stable only)
    IGNORE_PATTERN=".*-(rc|saas).*"
else
    # Open source CHANGELOG.md: ignore both saas and rc tags
    IGNORE_PATTERN=".*-(rc|saas).*"
fi

CLIFF_FLAGS="${CLIFF_FLAGS} --ignore-tags ${IGNORE_PATTERN}"

# Debug output
echo "DEBUG - RANGE: ${RANGE}"
echo "DEBUG - CLIFF_FLAGS: ${CLIFF_FLAGS}"
echo "DEBUG - Full command: git cliff ${RANGE} ${CLIFF_FLAGS}"

# Generate changelog
git cliff ${RANGE} ${CLIFF_FLAGS}
git add CHANGELOG${CHANGELOG_SUFFIX}.md

# Optionally generate PR body using the same logic
if [ -n "$PR_BODY_FILE" ]; then
    echo "INFO - Generating PR body to ${PR_BODY_FILE}"
    # Use same range and ignore pattern, but output to file instead of prepending
    PR_CLIFF_FLAGS="--github-repo ${GITHUB_REPO_URL} --use-branch-tags --tag ${RELEASE_VERSION} --ignore-tags ${IGNORE_PATTERN} -o ${PR_BODY_FILE}"
    echo "DEBUG - PR body command: git cliff ${RANGE} ${PR_CLIFF_FLAGS}"
    git cliff ${RANGE} ${PR_CLIFF_FLAGS}

    # Safeguard: GitHub PR body has a limit (~65KB). Truncate if too long while preserving
    # the '---' markers at the beginning and end of the file.
    FILE_SIZE=$(wc -c < "${PR_BODY_FILE}")
    MAX_SIZE=60000  # Stay well under GitHub's limit to account for additional PR metadata

    if [ "$FILE_SIZE" -gt "$MAX_SIZE" ]; then
        echo "WARN - PR body is too large (${FILE_SIZE} bytes), truncating to ${MAX_SIZE} bytes"

        # Extract header (first line with ---)
        HEAD_MARKER=$(head -1 "${PR_BODY_FILE}")

        # Extract footer (last line with ---)
        TAIL_MARKER=$(tail -1 "${PR_BODY_FILE}")

        # Create temporary file with truncated content
        TMP_FILE="${PR_BODY_FILE}.tmp"

        # Calculate how many lines we can keep (rough estimate: avg 100 chars per line)
        MAX_LINES=$((MAX_SIZE / 100))

        # Build truncated file: header + content + truncation notice + footer
        echo "${HEAD_MARKER}" > "${TMP_FILE}"
        head -n $((MAX_LINES - 10)) "${PR_BODY_FILE}" | tail -n +2 >> "${TMP_FILE}"
        echo "" >> "${TMP_FILE}"
        echo "..." >> "${TMP_FILE}"
        echo "" >> "${TMP_FILE}"
        echo "_Note: Changelog truncated due to size limits. View full changelog in the repository files._" >> "${TMP_FILE}"
        echo "" >> "${TMP_FILE}"
        echo "${TAIL_MARKER}" >> "${TMP_FILE}"

        mv "${TMP_FILE}" "${PR_BODY_FILE}"
        echo "INFO - PR body truncated to $(wc -c < "${PR_BODY_FILE}") bytes"
    fi

    echo "INFO - PR body generated successfully"
fi
