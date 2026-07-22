# GhostWire

Lightweight end-to-end-encrypted ephemeral chat for small ARM64 hosts such as Raspberry Pi 4.

## Features

- Go backend with SQLite WAL, REST registration/invites, WebSocket relay, health checks, and TTL janitor.
- SvelteKit + Tailwind dark UI with Web Crypto RSA-OAEP identity generation and encrypted payload relay.
- PWA service worker and Capacitor configuration for mobile wrapping.
- Multi-stage Docker image targeting Linux/ARM64 first, with Linux/AMD64 and Windows server release artifacts.

## Containerized server deployment

Build and run the complete server container, including the compiled Go backend and static web assets:

```sh
docker compose up --build -d
```

The service listens on `http://localhost:8787`, stores SQLite data in the `ghostwire-data` volume, and applies a `128m` container memory limit.

Check container health:

```sh
curl -fsS http://localhost:8787/api/health
```

Stop the deployment:

```sh
docker compose down
```

Remove persisted SQLite data as well:

```sh
docker compose down -v
```

For Raspberry Pi 4 or other ARM64 hosts, build explicitly with Buildx:

```sh
docker buildx build --platform linux/arm64 -t ghostwire:arm64 .
```

For x86_64 hosts, build the same server target with:

```sh
docker buildx build --platform linux/amd64 -t ghostwire:amd64 .
```

Or build both supported Linux container architectures together:

```sh
docker buildx build --platform linux/arm64,linux/amd64 -t ghostwire:multiarch .
```

Podman can build and run the same containerfile locally:

```sh
podman build --target server -t ghostwire:local .
podman run --rm -p 8787:8787 -v ghostwire-data:/data ghostwire:local
```

Or use the helper scripts:

```sh
./scripts/deploy-container.sh
./scripts/build-arm64.sh
```

## Releases

Release artifacts are built by `.github/workflows/release.yml` whenever a `v*.*.*` tag is pushed or the workflow is run manually. The release contains:

- `ghostwire-client-web.tar.gz` — static SvelteKit client bundle.
- `ghostwire-server-linux-arm64.tar.gz` — primary Raspberry Pi / ARM64 server binary.
- `ghostwire-server-linux-amd64.tar.gz` — x86_64 Linux server binary.
- `ghostwire-server-windows-amd64.zip` — Windows x64 server executable.
- GHCR container images for `linux/arm64` and `linux/amd64`.

Create local release artifacts without GitHub Actions:

```sh
./scripts/release-local.sh
```

## Local development

```sh
cd frontend && npm install && npm run build
cd ../backend && GHOSTWIRE_STATIC=../frontend/build go run ./cmd/ghostwire
```

The app listens on `:8787` by default and stores SQLite data in `ghostwire.db` unless `GHOSTWIRE_DB` is set.
