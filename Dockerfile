FROM golang:1.16-buster AS builder

RUN apt-get update && apt-get install -y ca-certificates openssl

ARG version=dev

WORKDIR /src/nimona.io

ENV CGO_ENABLED=1
ENV VERSION=$version

COPY . .

RUN make build
RUN make build-examples

###

FROM debian:buster-slim

COPY --from=builder /src/nimona.io/bin/bootstrap /bootstrap
COPY --from=builder /src/nimona.io/bin/sonar /sonar
COPY --from=builder /src/nimona.io/bin/examples /examples
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["/bootstrap"]
