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

FLAGS="-X main.AppVersion=$VERSION -s -w"

echo -e "\n${CYAN}Building v${VERSION}...${NC}"
GOOS=darwin GOARCH=arm64 go build -ldflags="$FLAGS" -buildvcs=false -trimpath -o "dist/data-tools-mac-arm64-${VERSION}" .
GOOS=darwin GOARCH=amd64 go build -ldflags="$FLAGS" -buildvcs=false -trimpath -o "dist/data-tools-mac-amd64-${VERSION}" .
GOOS=linux GOARCH=arm64 go build -ldflags="$FLAGS" -buildvcs=false -trimpath -o "dist/data-tools-linux-arm64-${VERSION}" .
GOOS=linux GOARCH=amd64 go build -ldflags="$FLAGS" -buildvcs=false -trimpath -o "dist/data-tools-linux-amd64-${VERSION}" .
GOOS=windows GOARCH=amd64 go build -ldflags="$FLAGS" -buildvcs=false -trimpath -o "dist/data-tools-${VERSION}.exe" .

echo -e "\n${CYAN}Build Success${NC}"