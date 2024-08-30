FROM scratch
ARG TARGETARCH
ARG TARGETOS
ARG USER=65534
ARG BIN_FILE=./dist/deviceconnect_${TARGETOS}-${TARGETARCH}
USER $USER
COPY --chown=$USER backend/services/deviceconnect/config.yaml /etc/deviceconnect/config.yaml
COPY --chown=$USER ${BIN_FILE} /usr/bin/deviceconnect
ENTRYPOINT ["/usr/bin/deviceconnect", "--config", "/etc/deviceconnect/config.yaml"]
