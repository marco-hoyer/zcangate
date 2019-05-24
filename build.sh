#!/usr/bin/env bash

VERSION=6

docker login

docker build . -t marcohoyer/zcangate:$VERSION
docker push marcohoyer/zcangate:$VERSION
