FROM scratch
ARG TARGETARCH
ARG TARGETOS
ARG USER=65534
ARG BIN_FILE=./dist/${TARGETOS}/${TARGETARCH}/workflows
USER $USER
COPY --chown=$USER backend/services/workflows/config.yaml /etc/workflows/config.yaml
COPY --chown=$USER ${BIN_FILE} /usr/bin/workflows
ENTRYPOINT ["/usr/bin/workflows", "--config", "/etc/workflows/config.yaml"]
