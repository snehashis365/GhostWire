#!/usr/bin/env sh
set -eu
(cd backend && GHOSTWIRE_STATIC=../frontend/build go run ./cmd/ghostwire)
