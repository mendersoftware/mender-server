ARG MENDER_BINARY_DELTA_VERSION=master
FROM --platform=$BUILDPLATFORM alpine:3.19.1 as mender-artifact-get
ARG TARGETARCH
ARG TARGETOS
ARG WORKFLOWS_VERSION=master
#FIXME: We need to fetch mender-binary-delta-generator somehow
ARG MENDER_ARTIFACT_VERSION=3.11.2
RUN apk --update --no-cache add dpkg zstd
RUN wget "https://downloads.mender.io/repos/debian/pool/main/m/mender-artifact/mender-artifact_${MENDER_ARTIFACT_VERSION}-1%2bubuntu%2bjammy_${TARGETARCH}.deb" -O mender-artifact.deb
RUN dpkg -x mender-artifact.deb .

FROM registry.mender.io/mendersoftware/mender-binary-delta:${MENDER_BINARY_DELTA_VERSION} as mender-binary-delta
FROM alpine:3.19.1

ARG TARGETARCH
ARG TARGETOS
ARG USER=65534
ARG BIN_FILE=./dist/${TARGETOS}/${TARGETARCH}/generate-delta-worker
ARG BIN_WORKFLOWS=./dist/${TARGETOS}/${TARGETARCH}/workflows

RUN apk add --no-cache \
    xz \
    libc6-compat \
    binutils \
    file \
    rsync \
    parted \
    e2fsprogs \
    xfsprogs \
    pigz \
    dosfstools \
    wget \
    make \
    bash

RUN sed -i 's/ash/bash/g' /etc/passwd

COPY --from=mender-binary-delta \
    /usr/bin/mender-binary-delta-generator \
    /usr/bin/mender-binary-delta-generator
COPY --from=mender-artifact-get /usr/bin/mender-artifact /usr/bin/mender-artifact
COPY --from=mender-artifact-get /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY backend/services/generate-delta-worker/workflows/generate_delta_artifact.json /etc/workflows/definitions/generate_delta_artifact.json
COPY backend/services/generate-delta-worker/config.yaml /etc/workflows/config.yaml
COPY ${BIN_FILE} /usr/bin/
COPY ${BIN_WORKFLOWS} /usr/bin/
RUN chmod +x /usr/bin/mender-artifact /usr/bin/mender-binary-delta-generator

ENTRYPOINT ["/usr/bin/workflows", "--config", "/etc/workflows/config.yaml"]
CMD ["worker"]
