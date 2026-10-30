#!/bin/bash
#
# Usage: GITHUB_APP_TOKEN="$(.gitlab/github-app-token.sh)"
#
# Prints a short-lived GitHub App installation token on stdout, and nothing else.
# The token is scoped to the repositories and permissions below. It expires after 1 hour.
#
# Required variables:
#   GH_APP_ID               GitHub App ID
#   GH_APP_INSTALLATION_ID  installation ID of the App on the owner of GITHUB_REPO_URL
#   GH_APP_PRIVATE_KEY_B64  App private key (PEM), base64 on a single line
#   GITHUB_REPO_URL         "owner/repo"
#
# Optional variables:
#   GH_APP_REPOSITORIES     space-separated repo names (same owner) to scope the token to,
#                           default: the repo in GITHUB_REPO_URL
#   GH_APP_PERMISSIONS      JSON object of permissions, a subset of the App's own,
#                           default: contents, pull_requests and issues write (release-please)
#
set -eo pipefail

for var in GH_APP_ID GH_APP_INSTALLATION_ID GH_APP_PRIVATE_KEY_B64 GITHUB_REPO_URL; do
    [ -n "${!var}" ] || { echo "ERROR - ${var} is not set" >&2; exit 1; }
done

repositories="${GH_APP_REPOSITORIES:-${GITHUB_REPO_URL#*/}}"
default_permissions='{"contents":"write","pull_requests":"write","issues":"write"}'
permissions="${GH_APP_PERMISSIONS:-${default_permissions}}"

body=$(jq -nc --arg repos "${repositories}" --argjson perms "${permissions}" \
    '{repositories: ($repos | split(" ") | map(select(length > 0))), permissions: $perms}') \
    || { echo "ERROR - GH_APP_PERMISSIONS is not a valid JSON object" >&2; exit 1; }

b64url() { openssl base64 -A | tr '+/' '-_' | tr -d '='; }

# iat is backdated to tolerate clock skew; exp stays under GitHub's 10 minute maximum.
now=$(date +%s)
header=$(printf '{"alg":"RS256","typ":"JWT"}' | b64url)
payload=$(printf '{"iat":%d,"exp":%d,"iss":"%s"}' $((now - 60)) $((now + 540)) "${GH_APP_ID}" | b64url)

# The key never touches disk: process substitution feeds the decoded PEM to openssl.
signature=$(printf '%s.%s' "${header}" "${payload}" \
    | openssl dgst -sha256 -sign <(printf '%s' "${GH_APP_PRIVATE_KEY_B64}" | base64 -d) \
    | b64url) || { echo "ERROR - could not sign the JWT, check GH_APP_PRIVATE_KEY_B64" >&2; exit 1; }

response=$(curl -fsS -X POST \
    -H "Authorization: Bearer ${header}.${payload}.${signature}" \
    -H "Accept: application/vnd.github+json" \
    "https://api.github.com/app/installations/${GH_APP_INSTALLATION_ID}/access_tokens" \
    -d "${body}") \
    || { echo "ERROR - installation token request failed" >&2; exit 1; }

token=$(echo "${response}" | jq -er '.token') \
    || { echo "ERROR - no token in the installation token response" >&2; exit 1; }
echo "${token}"
