FROM golang:1.12.5 AS builder

WORKDIR /src/nimona.io

ADD . .

RUN ls -lah .

ENV CGO_ENABLED=0

RUN make build
RUN cp -r ./bin/nimona /bin/nimona

###

FROM alpine:3.9

COPY --from=builder /bin/* /

RUN ls -lah /

ENTRYPOINT ["/nimona"]
CMD ["daemon"]
