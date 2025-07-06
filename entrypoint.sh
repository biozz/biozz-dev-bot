#!/bin/sh
set -e

ARCH=$(uname -m)
case "$ARCH" in
    x86_64 | amd64)
        exec ./app "$@"
        ;;
    aarch64 | arm64)
        exec ./app-arm64 "$@"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac
