FROM --platform=$BUILDPLATFORM golang:1.23.0 AS builder

ARG BUILDFLAGS="-trimpath"
ARG LDFLAGS="-s -w"
ARG TARGETARCH
ARG TARGETOS

WORKDIR /build
RUN \
  --mount=type=bind,source=.,target=/build/src \
  --mount=type=cache,target=/go/pkg/mod/ \
  --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=0 \
  GOOS="${TARGETOS}" \
  GOARCH="${TARGETARCH}" \
  go build -C src/backend -ldflags="${LDFLAGS}" ${BUILDFLAGS} \
  -o /build/useradm ./services/useradm

FROM scratch
ARG TARGETARCH
ARG TARGETOS
ARG USER=65534
USER $USER
COPY --chown=$USER  backend/services/useradm/config.yaml /etc/useradm/config.yaml
COPY --chown=$USER  backend/services/useradm/config/plans.yaml /etc/useradm/plans.yaml
COPY --from=builder --chown=$USER build/useradm /usr/bin/useradm
ENTRYPOINT ["/usr/bin/useradm", "--config", "/etc/useradm/config.yaml"]
