FROM scratch
ARG TARGETARCH
ARG TARGETOS
ARG USER=65534
ARG BIN_FILE=./dist/${TARGETOS}/${TARGETARCH}/auditlogs
USER $USER
COPY --chown=$USER backend/services/auditlogs/config.yaml /etc/auditlogs/config.yaml
COPY --chown=$USER ${BIN_FILE} /usr/bin/auditlogs
ENTRYPOINT ["/usr/bin/auditlogs", "--config", "/etc/auditlogs/config.yaml"]
