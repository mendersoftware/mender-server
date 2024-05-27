FROM scratch
ARG TARGETARCH
ARG TARGETOS
ARG USER=65534
ARG BIN_FILE=./dist/${TARGETOS}/${TARGETARCH}/useradm
USER $USER
COPY --chown=$USER backend/services/useradm/config.yaml /etc/useradm/config.yaml
COPY --chown=$USER backend/services/useradm/config/plans.yaml /etc/useradm/plans.yaml
COPY --chown=$USER ${BIN_FILE} /usr/bin/useradm
ENTRYPOINT ["/usr/bin/useradm", "--config", "/etc/useradm/config.yaml"]
