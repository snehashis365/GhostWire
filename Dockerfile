# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM node:22-alpine AS frontend
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json* ./
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi
COPY frontend ./
RUN npm run build

FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS backend
ARG TARGETOS=linux
ARG TARGETARCH
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend ./
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-$(go env GOARCH)} go build -trimpath -ldflags="-s -w" -o /out/ghostwire ./cmd/ghostwire

FROM alpine:3.20 AS server
RUN apk add --no-cache ca-certificates wget && adduser -D -H -u 10001 ghostwire
WORKDIR /app
COPY --from=backend /out/ghostwire /usr/local/bin/ghostwire
COPY --from=frontend /src/frontend/build /app/public
ENV GHOSTWIRE_ADDR=:8787 \
    GHOSTWIRE_DB=/data/ghostwire.db \
    GHOSTWIRE_STATIC=/app/public
RUN mkdir -p /data && chown -R ghostwire:ghostwire /data /app
USER ghostwire
EXPOSE 8787
VOLUME ["/data"]
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 CMD wget -qO- http://127.0.0.1:8787/api/health || exit 1
CMD ["ghostwire"]
