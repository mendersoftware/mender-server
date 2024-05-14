FROM scratch
ARG USER=65534
ARG BIN_FILE=./dist/${TARGETOS}/${TARGETARCH}/tenantadm
USER $USER
COPY --chown=$USER backend/services/tenantadm/config.yaml /etc/tenantadm/config.yaml
COPY --chown=$USER ${BIN_FILE} /usr/bin/tenantadm
ENTRYPOINT ["/usr/bin/tenantadm", "--config", "/etc/tenantadm/config.yaml"]
