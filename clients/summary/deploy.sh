#!/usr/bin/env bash

set -e

./build.sh

docker push fredhw/summary

ssh root@165.227.23.185 'bash -s' < run.sh
