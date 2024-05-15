ARG WORKFLOWS_VERSION=master
ARG TARGETARCH
ARG TARGETOS
ARG USER=65534
ARG BIN_FILE=./dist/${TARGETOS}/${TARGETARCH}/generate-delta-worker
ARG BIN_WORKFLOWS=./dist/${TARGETOS}/${TARGETARCH}/workflows
FROM --platform=$BUILDPLATFORM alpine:3.19.1 as mender-artifact-get
ARG MENDER_ARTIFACT_VERSION=3.9.0
RUN apk --update --no-cache add dpkg zstd
RUN deb_filename=mender-artifact_${MENDER_ARTIFACT_VERSION}-1%2Bubuntu%2Bjammy_${TARGETARCH}.deb && \
    wget "https://downloads.mender.io/repos/debian/pool/main/m/mender-artifact/${deb_filename}" \
    --output-document=/mender-artifact.deb && dpkg-deb --extract /mender-artifact.deb /

FROM alpine:3.19.1
ARG MENDER_ARTIFACT_VERSION=3.11.2
ARG MENDER_BINARY_DELTA_VERSION=1.4.1
ARG AWS_ACCESS_KEY_ID
ARG AWS_SECRET_ACCESS_KEY
ARG AWS_SESSION_TOKEN
ENV AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID
ENV AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY
ENV AWS_SESSION_TOKEN=$AWS_SESSION_TOKEN

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
    aws-cli \
    bash

RUN sed -i 's/ash/bash/g' /etc/passwd

COPY mender-binary-delta-generator-${TARGETARCH} /usr/bin/mender-binary-delta-generator
COPY --from=mender-artifact-get /usr/bin/mender-artifact /usr/bin/mender-artifact
COPY --from=mender-artifact-get /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY ./workflows/generate_delta_artifact.json /etc/workflows/definitions/generate_delta_artifact.json
COPY ./config.yaml /etc/workflows/config.yaml
COPY ${BIN_FILE} /usr/bin/
COPY ${BIN_WORKFLOWS} /usr/bin/
RUN chmod +x /usr/bin/mender-artifact /usr/bin/mender-binary-delta-generator

ENTRYPOINT ["/usr/bin/workflows", "--config", "/etc/workflows/config.yaml"]
CMD ["worker"]
