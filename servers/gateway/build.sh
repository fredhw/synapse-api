#!/usr/bin/env bash
set -e
echo "building linux executable"
GOOS=linux go build
docker build -t fredhw/gateway .
docker push fredhw/gateway
go clean

cd ../messaging
docker build -t fredhw/messaging .
docker push fredhw/messaging

cd -

cd ../summary
GOOS=linux go build
docker build -t fredhw/summary .
docker push fredhw/summary
go clean

cd -

cd ../qeeg-api
docker build -t fredhw/qeeg-api .
docker push fredhw/qeeg-api

cd -