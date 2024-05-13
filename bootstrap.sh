#!/bin/sh

REPOSITORIES_PATH="${REPOSITORIES_PATH:-/tmp/repositories.mender}"

cd $(git rev-parse --show-toplevel)

mkdir -p backend/services

if test -n "$ENTERPRISE"; then
    cat > "$REPOSITORIES_PATH" << EOF
auditlogs
create-artifact-worker
deployments-enterprise
deviceauth-enterprise
deviceconfig
deviceconnect
devicemonitor
generate-delta-worker
inventory-enterprise
iot-manager
reporting
tenantadm
useradm-enterprise
workflows-enterprise
EOF
    echo "Replacing docker composition for enterprise"
    mv dev/docker-compose.enterprise.yml dev/docker-compose.yml
    mv backend/tests/docker-compose.enterprise.yml backend/tests/docker-compose.yml
else
    cat > "$REPOSITORIES_PATH" << EOF
create-artifact-worker
deployments
deviceauth
deviceconfig
deviceconnect
inventory
iot-manager
reporting
useradm
workflows
EOF
    echo "Removing enterprise docker composition"
    rm dev/docker-compose.enterprise.yml backend/tests/docker-compose.enterprise.yml
fi

# Backend repositories
while read repo; do
    service_path="backend/services/${repo%%-enterprise}"
    if ! test "${repo%%-enterprise}" = ${repo}; then
        echo "Replacing open source with enterprise sources: ${service_path}"
        rm -rf ${service_path}
    fi
    if ! test -d "${service_path}"; then
        git clone git@github.com:mendersoftware/${repo} "${service_path}"
    fi
    echo "Changing import paths for ${repo}"
    if ! test "${repo%%-enterprise}" = ${repo}; then
        echo "Stripping enterprise suffix from import paths"
        find "${service_path}" \
            -name '*.go' \
            -exec sed -i 's:"github.com/mendersoftware/'"${repo}"':"github.com/mendersoftware/mender-server/services/'"${repo%%-enterprise}"':' {} \;
    fi
    find "${service_path}" \
        -name '*.go' \
        -exec sed -i 's:"github.com/mendersoftware/\('"${repo}"'.*\)":"github.com/mendersoftware/mender-server/services/\1":' {} \;
done < "$REPOSITORIES_PATH"

git clone git@github.com:mendersoftware/go-lib-micro backend/pkg
rm -f backend/pkg/dummy.go

git clone git@github.com:mendersoftware/gui frontend

echo "Replacing import paths to go-lib-micro"
find backend \
    -name '*.go' \
    -exec sed -i 's:"github.com/mendersoftware/go-lib-micro/\(.*\)":"github.com/mendersoftware/mender-server/pkg/\1":' {} \;

echo "Removing git indexes"
find backend frontend -mindepth 1 -type d -name .git -prune -exec rm -rf {} \;

# Remove files that are not sources nor tests nor buildfiles
# To generate a list of all suffixes, you can run this here:
#
# find -type f -exec sh -c 'echo {} | sed -nE "s/^.*(\.[^/.]+)$/\1/p"' \; | sort -u
#
# The following contains a pruned list of file extensions as well as adding Dockerfile*
# and acceptance test directories to the skip list.
echo "Removing non-source files"
find backend frontend -mindepth 1 -type f \
    -and -not -name 'Makefile' \
    -and -not -path '*/tests/*' \
    -and -not -name '*.acceptance' \
    -and -not -name '*.acceptance-testing' \
    -and -not -name '*.cnf' \
    -and -not -name '*.conf' \
    -and -not -name '*.crt' \
    -and -not -name '*.dockerignore' \
    -and -not -name '*.editorconfig' \
    -and -not -name '*.eslintignore' \
    -and -not -name '*.eslintrc' \
    -and -not -name '*.gif' \
    -and -not -name '*.gitignore' \
    -and -not -name '*.gitkeeep' \
    -and -not -name '*.gitkeep' \
    -and -not -name '*.go' \
    -and -not -name '*.html' \
    -and -not -name '*.ico' \
    -and -not -name '*.idx' \
    -and -not -name '*.js' \
    -and -not -name '*.json' \
    -and -not -name '*.key' \
    -and -not -name '*.less' \
    -and -not -name '*.md' \
    -and -not -name '*.mustache' \
    -and -not -name '*.npmrc' \
    -and -not -name '*.openapi-generator-ignore' \
    -and -not -name '*.pack' \
    -and -not -name '*.pem' \
    -and -not -name '*.png' \
    -and -not -name '*.prettierrc' \
    -and -not -name '*.py' \
    -and -not -name '*.rev' \
    -and -not -name '*.sample' \
    -and -not -name '*.sh' \
    -and -not -name '*.snap' \
    -and -not -name '*.svg' \
    -and -not -name '*.ts' \
    -and -not -name '*.txt' \
    -and -not -name '*.woff' \
    -and -not -name '*.woff2' \
    -and -not -name '*.worker' \
    -and -not -name '*.yaml' \
    -and -not -name '*.yml' \
    -exec rm -f {} \;

# Remove empty directories (from last command)
echo "Removing empty directories"
find backend frontend \
    -mindepth 1 \
    -type d -empty \
    -prune -exec rm -vrf {} \;

cd backend
echo "Initializing go Modules for backend"
go mod init github.com/mendersoftware/mender-server
go mod tidy

cd "$(git rev-parse --show-toplevel)"

echo "Creating Makefiles and Dockerfiles"
while read repo; do
    repo=${repo%%-enterprise}
    cp Makefile.service backend/services/${repo}/Makefile
    if test "${repo%%-worker}" != "${repo}"; then
        # Adjust makefile to add dependency on workflows binary
        sed -i '/^docker: build$/a \\t\$(MAKE) -C ../workflows build' \
            backend/services/${repo}/Makefile
        sed -i '/^docker-acceptance: build-test$/a \\t\$(MAKE) -C ../workflows build-test' \
            backend/services/${repo}/Makefile
        cat << EOF > backend/services/${repo}/Dockerfile
FROM scratch
ARG TARGETARCH
ARG TARGETOS
ARG USER=65534
ARG BIN_FILE=./dist/\${TARGETOS}/\${TARGETARCH}/${repo}
ARG BIN_WORKFLOWS=./dist/\${TARGETOS}/\${TARGETARCH}/workflows
USER \$USER
COPY --chown=\$USER backend/services/${repo}/config.yaml /etc/${repo}/config.yaml
COPY --chown=\$USER \${BIN_WORKFLOWS} /usr/bin/workflows
COPY --chown=\$USER \${BIN_FILE} /usr/bin/${repo}
ENTRYPOINT ["/usr/bin/workflows", "--config", "/etc/${repo}/config.yaml"]
CMD ["worker"]
EOF
    else
        cat << EOF > backend/services/${repo}/Dockerfile
FROM scratch
ARG USER=65534
ARG BIN_FILE=./dist/\${TARGETOS}/\${TARGETARCH}/${repo}
USER \$USER
COPY --chown=\$USER backend/services/${repo}/config.yaml /etc/${repo}/config.yaml
COPY --chown=\$USER \${BIN_FILE} /usr/bin/${repo}
ENTRYPOINT ["/usr/bin/${repo}", "--config", "/etc/${repo}/config.yaml"]
EOF
    fi
    # FIXME: services depending on mender-artifact needs special care
    if test "${repo%%-enterprise}" = "deployments"; then
        sed -i 's/^BUILDFLAGS ?=/\0 -tags nopkcs11/' backend/services/${repo%%-enterprise}/Makefile
    fi

done < "$REPOSITORIES_PATH"

cp Makefile.backend ./backend/Makefile

# Test build and docker make targets
make -C backend docker
