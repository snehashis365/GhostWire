#!/usr/bin/env sh
set -eu
docker buildx build --platform linux/arm64 -t ghostwire:arm64 .
