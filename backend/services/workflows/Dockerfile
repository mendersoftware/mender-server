FROM scratch
ARG TARGETARCH
ARG TARGETOS
ARG USER=65534
ARG BIN_FILE=./dist/workflows_${TARGETOS}-${TARGETARCH}
USER $USER
COPY --chown=$USER backend/services/workflows/worker/workflows /etc/workflows/definitions
COPY --chown=$USER backend/services/workflows/config.yaml /etc/workflows/config.yaml
COPY --chown=$USER ${BIN_FILE} /usr/bin/workflows
ENTRYPOINT ["/usr/bin/workflows", "--config", "/etc/workflows/config.yaml"]
