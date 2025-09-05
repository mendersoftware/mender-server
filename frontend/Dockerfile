FROM --platform=$BUILDPLATFORM node:24.1.0-alpine3.20 AS base
WORKDIR /usr/src/app
COPY package-lock.json package.json ./
RUN npm ci
COPY . ./
RUN npm run build

FROM nginxinc/nginx-unprivileged:1.27.5-alpine3.21-slim AS unprivileged
EXPOSE 8090
WORKDIR /var/www/mender-gui/dist
ARG GIT_COMMIT_TAG
ENV GIT_COMMIT_TAG=$GIT_COMMIT_TAG
COPY --from=base /usr/src/app/dist .
COPY ./entrypoint.sh /usr/src/entrypoint.sh
COPY httpd.conf /etc/nginx/nginx.conf
RUN sed -i 's|/var/run/nginx.pid|/tmp/nginx.pid|g' /etc/nginx/nginx.conf && sed -i 's|listen 80;|listen 8090;|g' /etc/nginx/nginx.conf
ENTRYPOINT ["/usr/src/entrypoint.sh"]
HEALTHCHECK --interval=8s --timeout=15s --start-period=120s --retries=128 CMD wget --quiet --tries=1 --spider --output-document=/dev/null 127.0.0.1:8090/ui/
CMD ["nginx"]
