FROM golang:1.11-alpine AS builder

RUN apk -U add curl git build-base
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

WORKDIR /go/src/nimona.io/go

COPY Gopkg.lock .
COPY Gopkg.toml .
RUN dep ensure --vendor-only

ADD . .

RUN mage build


FROM alpine:3.8

COPY --from=builder /go/src/nimona.io/go/bin/* /

ENTRYPOINT ["./nimona"]
CMD ["daemon"]
