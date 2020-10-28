FROM golang:1.15-buster AS builder

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

ENTRYPOINT ["/bootstrap"]
