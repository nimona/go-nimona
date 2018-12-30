FROM golang:1.11.4 AS builder

WORKDIR /src/nimona.io

ADD . .

RUN ls -lah .

ENV CGO_ENABLED=0

RUN go run nimona.io/tools/nmake build
RUN cp -r ./bin /bin

###

FROM alpine:3.8

COPY --from=builder /bin/* /

RUN ls -lah /

ENTRYPOINT ["/nimona"]
CMD ["daemon"]
