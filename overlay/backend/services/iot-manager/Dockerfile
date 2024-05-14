FROM scratch
ARG USER=65534
ARG BIN_FILE=./dist/${TARGETOS}/${TARGETARCH}/iot-manager
USER $USER
COPY --chown=$USER backend/services/iot-manager/config.yaml /etc/iot-manager/config.yaml
COPY --chown=$USER ${BIN_FILE} /usr/bin/iot-manager
ENTRYPOINT ["/usr/bin/iot-manager", "--config", "/etc/iot-manager/config.yaml"]
