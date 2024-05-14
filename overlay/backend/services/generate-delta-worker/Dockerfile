FROM scratch
ARG TARGETARCH
ARG TARGETOS
ARG USER=65534
ARG BIN_FILE=./dist/${TARGETOS}/${TARGETARCH}/generate-delta-worker
ARG BIN_WORKFLOWS=./dist/${TARGETOS}/${TARGETARCH}/workflows
USER $USER
COPY --chown=$USER backend/services/generate-delta-worker/config.yaml /etc/generate-delta-worker/config.yaml
COPY --chown=$USER ${BIN_WORKFLOWS} /usr/bin/workflows
COPY --chown=$USER ${BIN_FILE} /usr/bin/generate-delta-worker
ENTRYPOINT ["/usr/bin/workflows", "--config", "/etc/generate-delta-worker/config.yaml"]
CMD ["worker"]
