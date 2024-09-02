FROM golang:1.23.0 AS builder

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
  -o /build/reporting ./services/reporting

FROM scratch
ARG TARGETARCH
ARG TARGETOS
ARG USER=65534
USER $USER
COPY --chown=$USER  backend/services/reporting/config.yaml /etc/reporting/config.yaml
COPY --from=builder --chown=$USER /build/reporting /usr/bin/reporting
ENTRYPOINT ["/usr/bin/reporting", "--config", "/etc/reporting/config.yaml"]
