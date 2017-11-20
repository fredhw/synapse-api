#!/usr/bin/env bash

set -e

./build.sh

ssh 344api 'bash -s' < run.sh