FROM scratch
ARG USER=65534
ARG BIN_FILE=./dist/${TARGETOS}/${TARGETARCH}/reporting
USER $USER
COPY --chown=$USER backend/services/reporting/config.yaml /etc/reporting/config.yaml
COPY --chown=$USER ${BIN_FILE} /usr/bin/reporting
ENTRYPOINT ["/usr/bin/reporting", "--config", "/etc/reporting/config.yaml"]
