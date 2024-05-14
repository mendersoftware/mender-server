FROM scratch
ARG TARGETARCH
ARG TARGETOS
ARG USER=65534
ARG BIN_FILE=./dist/${TARGETOS}/${TARGETARCH}/inventory
USER $USER
COPY --chown=$USER backend/services/inventory/config.yaml /etc/inventory/config.yaml
COPY --chown=$USER ${BIN_FILE} /usr/bin/inventory
ENTRYPOINT ["/usr/bin/inventory", "--config", "/etc/inventory/config.yaml"]
