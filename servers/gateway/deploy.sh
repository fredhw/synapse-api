#!/usr/bin/env bash

set -e

./build.sh

docker push fredhw/gateway

ssh 344api 'bash -s' < run.sh