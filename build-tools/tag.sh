#!/usr/bin/env bash
set -e
read -p "Enter version number: " BUILD_TAG;

echo $BUILD_TAG

docker buildx build \
--push \
--platform linux/arm64,linux/amd64 \
--tag "ghcr.io/doutorfinancas/pun-sho:$BUILD_TAG" \
--tag "ghcr.io/doutorfinancas/pun-sho:latest" \
-f "Dockerfile" "."
