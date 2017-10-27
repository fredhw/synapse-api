#!/usr/bin/env bash
set -e
echo "building linux executable"
GOOS=linux go build
docker build -t fredhw/gateway .
docker push fredhw/gateway
go clean