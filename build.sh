#!/usr/bin/env bash

VERSION=2

docker login

env GOOS=linux GOARCH=arm GOARM=6 go build
docker build . -t marcohoyer/zcangate:$VERSION
docker push marcohoyer/zcangate:$VERSION
