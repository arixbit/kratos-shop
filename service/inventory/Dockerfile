ARG GOPROXY=https://proxy.golang.org,direct
ARG VERSION=unknown

FROM golang:1.27rc2 AS builder

ARG GOPROXY
ARG VERSION

COPY . /src
WORKDIR /src

RUN GOPROXY=$GOPROXY make build VERSION=$VERSION

FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates \
		curl \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

COPY --from=builder /src/bin /app

WORKDIR /app

EXPOSE 50055
VOLUME /data/conf

CMD ["./inventory", "-conf", "/data/conf"]
