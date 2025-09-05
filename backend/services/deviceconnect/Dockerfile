FROM --platform=$BUILDPLATFORM golang:1.25.0 AS builder
ARG TARGETOS
ARG TARGETARCH
ARG LDFLAGS="-s -w"
ARG BUILDFLAGS="-trimpath"
WORKDIR /build
RUN \
  --mount=type=bind,source=.,target=/build/src \
  --mount=type=cache,target=/go/pkg/mod/ \
  --mount=type=cache,target=/root/.cache/go-build \
  --mount=type=cache,target=/tmp,id=gotmp \
  make -C src/backend/services/deviceconnect build \
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
COPY --chown=$USER  backend/services/deviceconnect/config.yaml /etc/deviceconnect/config.yaml
COPY --from=builder --chown=$USER /build/deviceconnect /usr/bin/deviceconnect
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/usr/bin/deviceconnect", "--config", "/etc/deviceconnect/config.yaml"]
