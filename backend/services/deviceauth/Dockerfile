FROM golang:1.23.0 AS builder
ARG TARGETOS
ARG TARGETARCH
ARG LDFLAGS="-s -w"
ARG BUILDFLAGS="-trimpath"
WORKDIR /build
RUN \
  --mount=type=bind,source=.,target=/build/src \
  --mount=type=cache,target=/go/pkg/mod/ \
  --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=0 \
  GOOS="${TARGETOS}" \
  GOARCH="${TARGETARCH}" \
  go build -C src/backend -ldflags="${LDFLAGS}" ${BUILDFLAGS} \
  -o /build/deviceauth ./services/deviceauth

FROM scratch
ARG TARGETARCH
ARG TARGETOS
ARG USER=65534
ARG BIN_FILE=./dist/${TARGETOS}/${TARGETARCH}/deviceauth
USER $USER
COPY --chown=$USER  backend/services/deviceauth/config.yaml /etc/deviceauth/config.yaml
COPY --from=builder --chown=$USER /build/deviceauth /usr/bin/deviceauth
ENTRYPOINT ["/usr/bin/deviceauth", "--config", "/etc/deviceauth/config.yaml"]
