# GhostWire

Lightweight end-to-end-encrypted ephemeral chat for small ARM64 hosts such as Raspberry Pi 4.

## Features

- Go backend with SQLite WAL, REST registration/invites, WebSocket relay, and TTL janitor.
- SvelteKit + Tailwind dark UI with Web Crypto RSA-OAEP identity generation and encrypted payload relay.
- PWA service worker and Capacitor configuration for mobile wrapping.
- Multi-stage Docker image targeting Linux/ARM64.

## Development

```sh
cd frontend && npm install && npm run build
cd ../backend && go run ./cmd/ghostwire
```

The app listens on `:8080` by default and stores SQLite data in `ghostwire.db` unless `GHOSTWIRE_DB` is set.
