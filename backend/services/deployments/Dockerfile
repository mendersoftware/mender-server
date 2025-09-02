FROM --platform=$BUILDPLATFORM golang:1.25.0 AS builder

ARG BUILDFLAGS="-tags nopkcs11 -trimpath"
ARG LDFLAGS="-s -w"
ARG TARGETARCH
ARG TARGETOS

WORKDIR /build
RUN \
  --mount=type=bind,source=.,target=/build/src \
  --mount=type=cache,target=/go/pkg/mod/ \
  --mount=type=cache,target=/root/.cache/go-build \
  --mount=type=cache,target=/tmp,id=gotmp \
  make -C src/backend/services/deployments build \
  CGO_ENABLED=0 \
  GOOS="${TARGETOS}" \
  GOARCH="${TARGETARCH}" \
  bindir="/build" \
  LDFLAGS="${LDFLAGS}" \
  BUILDFLAGS="${BUILDFLAGS}"

FROM scratch AS empty
FROM scratch
ARG TARGETARCH
ARG TARGETOS
ARG USER=65534:65534
USER $USER
ENV TMPDIR=/tmp
# Create tempdir for generating configuration artifacts
COPY --from=empty --chown=$USER / $TMPDIR
COPY --chown=$USER  backend/services/deployments/config.yaml /etc/deployments/config.yaml
COPY --from=builder --chown=$USER /build/deployments /usr/bin/deployments
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/usr/bin/deployments", "--config", "/etc/deployments/config.yaml"]
