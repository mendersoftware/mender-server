FROM --platform=$BUILDPLATFORM golang:1.23.0 AS builder
ARG LDFLAGS="-s -w"
ARG BUILDFLAGS="-trimpath"
WORKDIR /build
RUN \
  --mount=type=bind,source=.,target=/build/src \
  --mount=type=cache,target=/go/pkg/mod/ \
  --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=0 go build -C src/backend -ldflags="${LDFLAGS}" ${BUILDFLAGS} \
  -o /build/deviceconfig ./services/deviceconfig

FROM scratch
ARG TARGETARCH
ARG TARGETOS
ARG USER=65534
USER $USER
COPY --chown=$USER  backend/services/deviceconfig/config.yaml /etc/deviceconfig/config.yaml
COPY --from=builder --chown=$USER /build/deviceconfig /usr/bin/deviceconfig
ENTRYPOINT ["/usr/bin/deviceconfig", "--config", "/etc/deviceconfig/config.yaml"]
