FROM --platform=$BUILDPLATFORM node:22-alpine AS frontend
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm install
COPY frontend ./
RUN npm run build

FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS backend
ARG TARGETOS=linux
ARG TARGETARCH=arm64
WORKDIR /src
COPY backend/go.mod backend/go.sum* ./backend/
WORKDIR /src/backend
RUN go mod download
COPY backend ./
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-s -w" -o /out/ghostwire ./cmd/ghostwire

FROM alpine:3.20
RUN adduser -D -H ghostwire
WORKDIR /app
COPY --from=backend /out/ghostwire /usr/local/bin/ghostwire
COPY --from=frontend /src/frontend/build /app/public
ENV GHOSTWIRE_ADDR=:8080 GHOSTWIRE_DB=/data/ghostwire.db GHOSTWIRE_STATIC=/app/public
RUN mkdir -p /data && chown -R ghostwire:ghostwire /data /app
USER ghostwire
EXPOSE 8080
CMD ["ghostwire"]
