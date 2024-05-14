FROM scratch
ARG TARGETARCH
ARG TARGETOS
ARG USER=65534
ARG BIN_FILE=./dist/${TARGETOS}/${TARGETARCH}/deviceconfig
USER $USER
COPY --chown=$USER backend/services/deviceconfig/config.yaml /etc/deviceconfig/config.yaml
COPY --chown=$USER ${BIN_FILE} /usr/bin/deviceconfig
ENTRYPOINT ["/usr/bin/deviceconfig", "--config", "/etc/deviceconfig/config.yaml"]
