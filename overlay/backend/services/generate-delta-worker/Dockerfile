ARG WORKFLOWS_VERSION=master
ARG TARGETARCH
ARG TARGETOS
ARG USER=65534
ARG BIN_FILE=./dist/${TARGETOS}/${TARGETARCH}/generate-delta-worker
ARG BIN_WORKFLOWS=./dist/${TARGETOS}/${TARGETARCH}/workflows
#FIXME: We need to fetch mender-binary-delta-generator somehow
ARG BIN_GENERATE_DELTA=./dist/${TARGETOS}/${TARGETARCH}/mender-binary-delta-generator
FROM --platform=$BUILDPLATFORM alpine:3.19.1 as mender-artifact-get
ARG MENDER_ARTIFACT_VERSION=3.9.0
RUN apk --update --no-cache add dpkg zstd
RUN deb_filename=mender-artifact_${MENDER_ARTIFACT_VERSION}-1%2Bubuntu%2Bjammy_${TARGETARCH}.deb && \
    wget "https://downloads.mender.io/repos/debian/pool/main/m/mender-artifact/${deb_filename}" \
    --output-document=/mender-artifact.deb && dpkg-deb --extract /mender-artifact.deb /

FROM alpine:3.19.1

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

COPY ${BIN_GENERATE_DELTA} /usr/bin/mender-binary-delta-generator
COPY --from=mender-artifact-get /usr/bin/mender-artifact /usr/bin/mender-artifact
COPY --from=mender-artifact-get /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY backend/services/generate-delta-worker/workflows/generate_delta_artifact.json /etc/workflows/definitions/generate_delta_artifact.json
COPY backend/services/generate-delta-worker/config.yaml /etc/workflows/config.yaml
COPY ${BIN_FILE} /usr/bin/
COPY ${BIN_WORKFLOWS} /usr/bin/
RUN chmod +x /usr/bin/mender-artifact /usr/bin/mender-binary-delta-generator

ENTRYPOINT ["/usr/bin/workflows", "--config", "/etc/workflows/config.yaml"]
CMD ["worker"]
