#!/usr/bin/env bash
set -e
docker build -t fredhw/messages .
go clean