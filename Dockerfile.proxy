FROM golang:1.12.7 AS builder

WORKDIR /src/nimona.io

ADD . .

RUN make build-proxy

###

FROM alpine:3.10

RUN apk --no-cache add ca-certificates && update-ca-certificates

COPY --from=builder /src/nimona.io/bin/proxy /proxy

ENTRYPOINT ["/proxy"]
