FROM node:24-bookworm-slim AS builder

WORKDIR /src/admin
COPY admin/package.json admin/pnpm-lock.yaml admin/pnpm-workspace.yaml ./
RUN corepack enable && pnpm install --frozen-lockfile

COPY admin/ ./
ARG SSG_PUBLIC_ORIGIN=https://img.example.com
ENV SSG_PUBLIC_ORIGIN=${SSG_PUBLIC_ORIGIN}
ENV VITE_API_BASE_URL=
RUN pnpm run build:ssg

FROM nginx:1.27-alpine
COPY --from=builder /src/admin/dist/ /usr/share/nginx/html/
COPY deploy/docker/nginx.conf /etc/nginx/conf.d/default.conf
