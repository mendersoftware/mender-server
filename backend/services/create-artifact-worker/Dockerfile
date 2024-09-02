FROM --platform=$BUILDPLATFORM golang:1.23.0 as builder
ARG TARGETARCH
ARG TARGETOS
ARG BUILDFLAGS="-trimpath"
ARG LDFLAGS="-s -w"
ARG MENDER_ARTIFACT_VERSION=3.11.2
WORKDIR /build
# Fetch and install mender-artifact
RUN wget "https://downloads.mender.io/repos/debian/pool/main/m/mender-artifact/mender-artifact_${MENDER_ARTIFACT_VERSION}-1%2bubuntu%2bnoble_${TARGETARCH}.deb" \
  -O mender-artifact.deb && \
  dpkg -x mender-artifact.deb / && \
  mkdir /var/cache/create-artifact-worker

# COPY . .
RUN \
  --mount=type=bind,source=.,target=/build/src \
  --mount=type=cache,target=/go/pkg/mod/ \
  --mount=type=cache,target=/root/.cache/go-build \
  GOOS="${TARGETOS}" \
  GOARCH="${TARGETARCH}" \
  CGO_ENABLED=0 \
  go build -C src/backend -ldflags="${LDFLAGS}" ${BUILDFLAGS} \
  -o /build/create-artifact-worker ./services/create-artifact-worker && \
  GOOS="${TARGETOS}" \
  GOARCH="${TARGETARCH}" \
  CGO_ENABLED=0 \
  go build -C src/backend -ldflags="${LDFLAGS}" ${BUILDFLAGS} \
  -o /build/workflows ./services/workflows

FROM alpine:3.19.1
ARG TARGETARCH
ARG TARGETOS
ARG USER=65534
RUN apk add libc6-compat xz
USER $USER

# Setup work directory
ENV CREATE_ARTIFACT_WORKDIR="/var/cache/create-artifact-worker"
COPY --from=builder --chown=$USER \
  /var/cache/create-artifact-worker \
  /var/cache/create-artifact-worker

# Install mender-artifact
ADD --chmod=0755 https://raw.githubusercontent.com/mendersoftware/mender/master/support/modules-artifact-gen/single-file-artifact-gen \
  /usr/bin/single-file-artifact-gen
COPY --from=builder --chown=$USER \
  /usr/bin/mender-artifact \
  /usr/bin/mender-artifact

COPY --chown=$USER  backend/services/create-artifact-worker/config.yaml /etc/workflows/config.yaml
COPY --chown=$USER  backend/services/create-artifact-worker/workflows /etc/workflows/definitions
COPY --from=builder --chown=$USER /build/workflows /usr/bin/workflows
COPY --from=builder --chown=$USER /build/create-artifact-worker /usr/bin/create-artifact
ENTRYPOINT ["/usr/bin/workflows", "--config", "/etc/workflows/config.yaml"]
CMD ["worker"]
