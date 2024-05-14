FROM scratch
ARG USER=65534
ARG BIN_FILE=./dist/${TARGETOS}/${TARGETARCH}/deviceauth
USER $USER
COPY --chown=$USER backend/services/deviceauth/config.yaml /etc/deviceauth/config.yaml
COPY --chown=$USER ${BIN_FILE} /usr/bin/deviceauth
ENTRYPOINT ["/usr/bin/deviceauth", "--config", "/etc/deviceauth/config.yaml"]
