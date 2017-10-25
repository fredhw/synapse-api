#!/usr/bin/env bash

set -e

./build.sh

docker push fredhw/gateway

ssh root@165.227.62.58 'bash -s' < run.sh