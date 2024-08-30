FROM alpine:3.19.1 as alpine

FROM scratch
ARG TARGETARCH
ARG TARGETOS
ARG USER=65534
ARG BIN_FILE=./dist/iot-manager_${TARGETOS}-${TARGETARCH}
USER $USER
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --chown=$USER backend/services/iot-manager/config.yaml /etc/iot-manager/config.yaml
COPY --chown=$USER ${BIN_FILE} /usr/bin/iot-manager
ENTRYPOINT ["/usr/bin/iot-manager", "--config", "/etc/iot-manager/config.yaml"]
