FROM --platform=$BUILDPLATFORM node:24.11.0-alpine3.22 AS base
WORKDIR /usr/src/app
RUN apk add --no-cache gzip
COPY package-lock.json package.json ./
RUN npm ci
COPY . ./
ARG GIT_COMMIT_SHA
ENV GIT_COMMIT_SHA=$GIT_COMMIT_SHA
ARG SENTRY_ORG
ENV SENTRY_ORG=$SENTRY_ORG
ARG SENTRY_URL
ENV SENTRY_URL=$SENTRY_URL
RUN --mount=type=secret,id=sentryAuthToken export SENTRY_AUTH_TOKEN=$(cat /run/secrets/sentryAuthToken) && npm run build
RUN gzip -r -k dist/*

FROM nginxinc/nginx-unprivileged:1.29.2-alpine3.22-slim AS unprivileged
EXPOSE 8090
WORKDIR /var/www/mender-gui/dist
ARG GIT_COMMIT_TAG
ENV GIT_COMMIT_TAG=$GIT_COMMIT_TAG
ARG GIT_COMMIT_SHA
ENV GIT_COMMIT_SHA=$GIT_COMMIT_SHA
COPY --from=base /usr/src/app/dist .
COPY ./entrypoint.sh /usr/src/entrypoint.sh
COPY httpd.conf /etc/nginx/nginx.conf
RUN sed -i 's|/var/run/nginx.pid|/tmp/nginx.pid|g' /etc/nginx/nginx.conf && sed -i 's|listen 80;|listen 8090;|g' /etc/nginx/nginx.conf
ENTRYPOINT ["/usr/src/entrypoint.sh"]
HEALTHCHECK --interval=8s --timeout=15s --start-period=120s --retries=128 CMD wget --quiet --tries=1 --spider --output-document=/dev/null 127.0.0.1:8090/ui/
CMD ["nginx"]
