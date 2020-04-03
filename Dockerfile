FROM golang:1.11-alpine as builder

WORKDIR /app
COPY ./Makefile ./go.* ./*.go ./
COPY ./cmd/ ./cmd/

RUN \
  apk add --no-cache --virtual .build-deps \
  make=4.2.1-r2 git=2.22.2-r0 build-base=0.5-r1 && \
  make build && \
  apk del .build-deps

FROM alpine:3.10

COPY --from=builder /app/build/rawkafka* /bin/

WORKDIR /app
COPY ./request.avsc ./

ENV RAWKAFKA_SCHEMA_LOCATION=/app/request.avsc

COPY ./docker-entrypoint.sh /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]
CMD [ "rawkafka" ]
