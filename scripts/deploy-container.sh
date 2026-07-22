#!/usr/bin/env sh
set -eu
IMAGE=${IMAGE:-ghostwire:local}
docker compose build ghostwire
docker compose up -d ghostwire
echo "GhostWire is starting at http://localhost:8787 using image ${IMAGE}"
