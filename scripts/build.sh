#!/usr/bin/env bash

# Exit if any command fails
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[1;36m'
NC='\033[0m' # No Color

rm -rf ./dist
./scripts/test.sh

if [[ -z "${VERSION}" ]]; then
    VERSION="0.0"
fi

BUILD_FLAGS="-X main.AppVersion=$VERSION -s -w"

echo -e "\n${CYAN}Building v${VERSION}...${NC}"

# The regular SQLite driver (which is faster) requires CGO

CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags="$BUILD_FLAGS" -buildvcs=false -trimpath -o "dist/data-tools-mac-arm64-${VERSION}" .
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags="$BUILD_FLAGS" -buildvcs=false -trimpath -o "dist/data-tools-mac-amd64-${VERSION}" .

# Use the alternative driver for these
GOOS=linux GOARCH=arm64 go build -ldflags="$BUILD_FLAGS" -buildvcs=false -trimpath -o "dist/data-tools-linux-arm64-${VERSION}" -tags alternative_driver .
GOOS=linux GOARCH=amd64 go build -ldflags="$BUILD_FLAGS" -buildvcs=false -trimpath -o "dist/data-tools-linux-amd64-${VERSION}" -tags alternative_driver .
GOOS=windows GOARCH=amd64 go build -ldflags="$BUILD_FLAGS" -buildvcs=false -trimpath -o "dist/data-tools-${VERSION}.exe" -tags alternative_driver .

docker build -t antfie/data-tools --build-arg BUILD_FLAGS="$BUILD_FLAGS" .

echo -e "\n${CYAN}Build Success${NC}"