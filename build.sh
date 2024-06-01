#!/usr/bin/env bash
set -e

VERSION=35

docker login

docker build . -t marcohoyer/zcangate:$VERSION
docker push marcohoyer/zcangate:$VERSION
