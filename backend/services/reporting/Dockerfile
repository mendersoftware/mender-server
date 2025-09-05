FROM --platform=$BUILDPLATFORM golang:1.25.0 AS builder

ARG BUILDFLAGS="-trimpath"
ARG LDFLAGS="-s -w"
ARG TARGETARCH
ARG TARGETOS

WORKDIR /build
RUN \
  --mount=type=bind,source=.,target=/build/src \
  --mount=type=cache,target=/go/pkg/mod/ \
  --mount=type=cache,target=/root/.cache/go-build \
  --mount=type=cache,target=/tmp,id=gotmp \
  make -C src/backend/services/reporting build \
  CGO_ENABLED=0 \
  GOOS="${TARGETOS}" \
  GOARCH="${TARGETARCH}" \
  bindir="/build" \
  LDFLAGS="${LDFLAGS}" \
  BUILDFLAGS="${BUILDFLAGS}"

FROM scratch
ARG TARGETARCH
ARG TARGETOS
ARG USER=65534:65534
USER $USER
COPY --chown=$USER  backend/services/reporting/config.yaml /etc/reporting/config.yaml
COPY --from=builder --chown=$USER /build/reporting /usr/bin/reporting
ENTRYPOINT ["/usr/bin/reporting", "--config", "/etc/reporting/config.yaml"]
