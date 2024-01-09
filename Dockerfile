FROM alpine:latest

RUN apk add --update curl jq bash

ADD playlottery.sh /playlottery.sh

ENTRYPOINT /playlottery.sh
