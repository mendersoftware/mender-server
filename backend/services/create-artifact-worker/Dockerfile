FROM --platform=$BUILDPLATFORM alpine:3.19.1 as dependencies
ARG TARGETARCH
ARG TARGETOS
ARG MENDER_ARTIFACT_VERSION=3.11.2
RUN apk add dpkg zstd && \
    wget "https://downloads.mender.io/repos/debian/pool/main/m/mender-artifact/mender-artifact_${MENDER_ARTIFACT_VERSION}-1%2bubuntu%2bnoble_${TARGETARCH}.deb" -O mender-artifact.deb && \
    dpkg -x mender-artifact.deb . && \
    mkdir /var/cache/create-artifact-worker

FROM alpine:3.19.1
ARG TARGETARCH
ARG TARGETOS
ARG USER=65534
ARG BIN_FILE=./dist/${TARGETOS}/${TARGETARCH}/create-artifact-worker
ARG BIN_WORKFLOWS=./dist/${TARGETOS}/${TARGETARCH}/workflows
RUN apk add libc6-compat xz
USER $USER

# Setup work directory
ENV CREATE_ARTIFACT_WORKDIR="/var/cache/create-artifact-worker"
COPY --from=dependencies --chown=$USER \
    /var/cache/create-artifact-worker \
    /var/cache/create-artifact-worker

# Install mender-artifact
ADD --chmod=0755 https://raw.githubusercontent.com/mendersoftware/mender/master/support/modules-artifact-gen/single-file-artifact-gen \
    /usr/bin/single-file-artifact-gen
COPY --from=dependencies --chown=$USER \
    /usr/bin/mender-artifact \
    /usr/bin/mender-artifact

COPY --chown=$USER backend/services/create-artifact-worker/config.yaml /etc/workflows/config.yaml
COPY --chown=$USER backend/services/create-artifact-worker/workflows /etc/workflows/definitions
COPY --chown=$USER ${BIN_WORKFLOWS} /usr/bin/workflows
COPY --chown=$USER ${BIN_FILE} /usr/bin/create-artifact
ENTRYPOINT ["/usr/bin/workflows", "--config", "/etc/workflows/config.yaml"]
CMD ["worker"]
