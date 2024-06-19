FROM scratch as empty
FROM scratch
ARG TARGETARCH
ARG TARGETOS
ARG USER=65534
ARG BIN_FILE=./dist/${TARGETOS}/${TARGETARCH}/deployments
USER $USER
ENV TMPDIR=/tmp
# Create tempdir for generating configuration artifacts
COPY --from=empty --chown=$USER / $TMPDIR
COPY --chown=$USER backend/services/deployments/config.yaml /etc/deployments/config.yaml
COPY --chown=$USER ${BIN_FILE} /usr/bin/deployments
ENTRYPOINT ["/usr/bin/deployments", "--config", "/etc/deployments/config.yaml"]
