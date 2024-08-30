FROM scratch
ARG TARGETARCH
ARG TARGETOS
ARG USER=65534
ARG BIN_FILE=./dist/deviceauth_${TARGETOS}-${TARGETARCH}
USER $USER
COPY --chown=$USER backend/services/deviceauth/config.yaml /etc/deviceauth/config.yaml
COPY --chown=$USER ${BIN_FILE} /usr/bin/deviceauth
ENTRYPOINT ["/usr/bin/deviceauth", "--config", "/etc/deviceauth/config.yaml"]
