FROM scratch
ARG TARGETARCH
ARG TARGETOS
ARG USER=65534
ARG BIN_FILE=./dist/${TARGETOS}/${TARGETARCH}/devicemonitor
USER $USER
COPY --chown=$USER backend/services/devicemonitor/config.yaml /etc/devicemonitor/config.yaml
COPY --chown=$USER ${BIN_FILE} /usr/bin/devicemonitor
ENTRYPOINT ["/usr/bin/devicemonitor", "--config", "/etc/devicemonitor/config.yaml"]
