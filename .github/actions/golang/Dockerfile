FROM golang:1.12.6

LABEL "name"="Nimona lint, test, and build"
LABEL "maintainer"="George Antoniadis <george@noodles.gr>"
LABEL "version"="0.1.0"

LABEL "com.github.actions.icon"="code"
LABEL "com.github.actions.color"="green-dark"
LABEL "com.github.actions.name"="Nimona lint, test, and build"
LABEL "com.github.actions.description"="This is an Action to lint and test nimona."

ENV GOMODULE111=off

RUN go get -u github.com/goreleaser/goreleaser
RUN go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
RUN go get -u github.com/shurcooL/vfsgen

ENTRYPOINT ["make"]