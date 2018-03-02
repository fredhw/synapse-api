#!/usr/bin/env bash

set -e

./build.sh

ssh root@159.65.104.35 'bash -s' < run.sh