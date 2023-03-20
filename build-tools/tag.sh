#!/usr/bin/env bash
set -e
read -p "Enter version number: " BUILD_TAG;

echo $BUILD_TAG

docker build --platform linux/arm64 -t "pun-sho:$BUILD_TAG" -f "Dockerfile" "."
docker build --platform linux/amd64 -t "pun-sho:$BUILD_TAG" -f "Dockerfile" "."
docker tag "pun-sho:$BUILD_TAG" "ghcr.io/doutorfinancas/pun-sho:$BUILD_TAG"

docker push "ghcr.io/doutorfinancas/pun-sho:$BUILD_TAG"

docker tag "pun-sho:$BUILD_TAG" "ghcr.io/doutorfinancas/pun-sho:latest"
docker push "ghcr.io/doutorfinancas/pun-sho:latest"
