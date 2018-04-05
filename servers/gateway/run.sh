#!/usr/bin/env bash
docker pull fredhw/gateway
docker pull fredhw/summary
docker pull fredhw/messaging
docker pull fredhw/qeeg-api

docker rm -f gateway
docker rm -f summary1
docker rm -f summary2
docker rm -f devredis
docker rm -f mymongo
docker rm -f messaging1
docker rm -f messaging2
docker rm -f messaging3
docker rm -f qeeg1
docker rm -f qeeg2

export TLSCERT=/etc/letsencrypt/live/api.synapse-solutions.net/fullchain.pem
export TLSKEY=/etc/letsencrypt/live/api.synapse-solutions.net/privkey.pem
export DBADDR="mymongo:27017"

docker run -d \
--name devredis \
--network appnet \
redis

docker run -d \
--name mymongo \
--network appnet \
-v ~/data:/data/db \
mongo

docker run -d \
--name summary1 \
--network appnet \
fredhw/summary

docker run -d \
--name summary2 \
--network appnet \
fredhw/summary

docker run -d \
--name messaging1 \
--network appnet \
fredhw/messaging

docker run -d \
--name messaging2 \
--network appnet \
fredhw/messaging

docker run -d \
--name messaging3 \
--network appnet \
fredhw/messaging

docker run -d \
--name qeeg1 \
--network appnet \
fredhw/qeeg-api

docker run -d \
--name qeeg2 \
--network appnet \
fredhw/qeeg-api

docker run -d \
-p 443:443 \
--name gateway \
--network appnet \
-v /etc/letsencrypt:/etc/letsencrypt:ro \
-e TLSCERT=$TLSCERT \
-e TLSKEY=$TLSKEY \
-e SESSIONKEY="testing" \
-e REDISADDR="devredis:6379" \
-e DBADDR=$DBADDR \
-e MESSAGESSVC_ADDRS=messaging1,messaging2,messaging3 \
-e SUMMARYSVC_ADDRS=summary1,summary2 \
-e QEEGSVC_ADDRS=qeeg1,qeeg2 \
fredhw/gateway

