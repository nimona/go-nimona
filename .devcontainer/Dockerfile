FROM golang:latest

WORKDIR /master
RUN git clone --depth 1 https://github.com/golang/go.git
WORKDIR /master/go/src
RUN ./make.bash

ENV PATH "/master/go/bin:${PATH}"
RUN echo "PATH=${PATH}" >> /etc/bash.bashrc && echo "${PATH}"

RUN go install golang.org/x/tools/gopls@latest
