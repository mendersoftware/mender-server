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
  make -C src/backend/services/inventory build \
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
ARG BIN_FILE=./dist/${TARGETOS}/${TARGETARCH}/inventory
USER $USER
COPY --chown=$USER  backend/services/inventory/config.yaml /etc/inventory/config.yaml
COPY --from=builder --chown=$USER /build/inventory /usr/bin/inventory
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/usr/bin/inventory", "--config", "/etc/inventory/config.yaml"]
