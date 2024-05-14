FROM scratch
ARG USER=65534
ARG BIN_FILE=./dist/${TARGETOS}/${TARGETARCH}/deployments
USER $USER
COPY --chown=$USER backend/services/deployments/config.yaml /etc/deployments/config.yaml
COPY --chown=$USER ${BIN_FILE} /usr/bin/deployments
ENTRYPOINT ["/usr/bin/deployments", "--config", "/etc/deployments/config.yaml"]
