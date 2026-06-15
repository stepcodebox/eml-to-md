#!/usr/bin/env bash
set -euo pipefail

# Build helper for eml-to-md.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
BIN_DIR="${ROOT_DIR}/bin"

BINARY_NAME="${BINARY_NAME:-eml2md}"
GOOS="${GOOS:-linux}"
GOARCH="${GOARCH:-amd64}"
CGO_ENABLED="${CGO_ENABLED:-0}"

cd "${ROOT_DIR}"
mkdir -p "${BIN_DIR}"

go mod tidy

CGO_ENABLED="${CGO_ENABLED}" GOOS="${GOOS}" GOARCH="${GOARCH}" \
	go build -trimpath -ldflags="-s -w" -o "${BIN_DIR}/${BINARY_NAME}" ./cmd/eml2md

echo "Built ${BIN_DIR}/${BINARY_NAME} (${GOOS}/${GOARCH}, CGO_ENABLED=${CGO_ENABLED})"
