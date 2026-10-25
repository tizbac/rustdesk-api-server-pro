FROM golang:alpine AS golang
WORKDIR /backend
COPY ./backend .
RUN go build


FROM node:20-alpine AS node
WORKDIR /frontend
COPY ./soybean-admin .
RUN rm -rf node_modules
RUN npm install -g pnpm@9 # newer pnpm ignores the shamefully-hoist of .npmrc and fails on unapproved build scripts
RUN pnpm i && pnpm build


# builds the rustdesk web client, see rustdesk/build_web.sh
FROM node:20-bookworm AS rustdesk-web
ARG FLUTTER_VERSION=3.24.5
RUN apt-get update && apt-get install -y --no-install-recommends \
        curl \
        git \
        protobuf-compiler \
        python3 \
        python-is-python3 \
        unzip \
        xz-utils \
        zip \
    && rm -rf /var/lib/apt/lists/*
RUN curl -fsSL "https://storage.googleapis.com/flutter_infra_release/releases/stable/linux/flutter_linux_${FLUTTER_VERSION}-stable.tar.xz" -o /tmp/flutter.tar.xz \
    && tar -xJf /tmp/flutter.tar.xz -C /opt \
    && rm /tmp/flutter.tar.xz
ENV PATH="/opt/flutter/bin:$PATH"
RUN git config --global --add safe.directory /opt/flutter
COPY ./rustdesk /rustdesk
WORKDIR /rustdesk
# flutter build web copies every file of flutter/web, so drop the yarn state again;
# rustdeskWeb.json is served by the api-server from server.yaml instead
RUN bash ./build_web.sh && rm -rf flutter/build/web/js/node_modules flutter/build/web/rustdeskWeb.json


FROM alpine:3.20.3
ENV ADMIN_USER=
ENV ADMIN_PASS=
WORKDIR /app
COPY ./docker/start.sh .
COPY --from=golang /backend/rustdesk-api-server-pro .
COPY --from=golang /backend/server.yaml ./data/server.yaml
COPY --from=node /frontend/dist ./dist
COPY --from=rustdesk-web /rustdesk/flutter/build/web ./web
RUN apk add tzdata
EXPOSE 8080
CMD [ "sh", "/app/start.sh"]
