FROM golang:1.11.4 AS builder
WORKDIR /src/nimona.io
ADD . .
RUN go run nimona.io/tools/nmake build

FROM alpine:3.8
COPY --from=builder /src/nimona.io/bin/* /
ENTRYPOINT ["./nimona"]
CMD ["daemon"]
