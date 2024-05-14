FROM scratch
ARG TARGETARCH
ARG TARGETOS
ARG USER=65534
ARG BIN_FILE=./dist/${TARGETOS}/${TARGETARCH}/create-artifact-worker
ARG BIN_WORKFLOWS=./dist/${TARGETOS}/${TARGETARCH}/workflows
USER $USER
COPY --chown=$USER backend/services/create-artifact-worker/config.yaml /etc/create-artifact-worker/config.yaml
COPY --chown=$USER ${BIN_WORKFLOWS} /usr/bin/workflows
COPY --chown=$USER ${BIN_FILE} /usr/bin/create-artifact-worker
ENTRYPOINT ["/usr/bin/workflows", "--config", "/etc/create-artifact-worker/config.yaml"]
CMD ["worker"]
