#!/usr/bin/env bash

docker pull fredhw/messages
docker rm -f 344client

export TLSCERT=/etc/letsencrypt/live/fredhw.me/fullchain.pem
export TLSKEY=/etc/letsencrypt/live/fredhw.me/privkey.pem

docker run -d \
--name 344client \
-p 80:80 -p 443:443 \
-v /etc/letsencrypt:/etc/letsencrypt:ro \
-e TLSCERT=$TLSCERT \
-e TLSKEY=$TLSKEY \
fredhw/messages