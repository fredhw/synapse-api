#!/usr/bin/env bash
docker pull fredhw/gateway

docker rm -f gateway
docker rm -f devredis
docker rm -f mymongo

export TLSCERT=/etc/letsencrypt/live/api.fredhw.me/fullchain.pem
export TLSKEY=/etc/letsencrypt/live/api.fredhw.me/privkey.pem

docker run -d \
-p 443:443 \
--name gateway \
--network appnet \
-v /etc/letsencrypt:/etc/letsencrypt:ro \
-e TLSCERT=$TLSCERT \
-e TLSKEY=$TLSKEY \
-e SESSIONKEY="testing" \
-e REDISADDR="devredis:6379" \
-e DBADDR="mymongo:27017" \
fredhw/gateway

docker run --name mymongo --network appnet -d -v ~/data:/data/db mongo
docker run --name devredis --network appnet -d redis