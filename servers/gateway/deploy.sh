#!/usr/bin/env bash

set -e

./build.sh

ssh -i ./eeg-aws/dragonball.pem ec2-user@ec2-18-216-188-207.us-east-2.compute.amazonaws.com 'bash -s' < run.sh