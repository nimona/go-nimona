FROM golang:1.12.5 AS builder

WORKDIR /src/nimona.io

ADD . .

RUN make build

###

FROM alpine:3.9

RUN apk --no-cache add ca-certificates && update-ca-certificates

COPY --from=builder /src/nimona.io/bin/nimona /nimona

ENTRYPOINT ["/nimona"]
