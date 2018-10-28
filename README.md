[![CircleCI Image]](https://circleci.com/gh/nimona/go-nimona)
[![Coveralls Image]](https://coveralls.io/github/nimona/go-nimona)
[![License Image]](https://github.com/nimona/go-nimona/blob/master/LICENSE)
[![Issues Image]](https://waffle.io/nimona/go-nimona)

# Nimona

Nimona’s main goal is to provide a number of layers/components to help with the challenges presented when dealing with decentralized and peer to peer applications.

## Architecture

For a technical overview, please refer to the [documentation introduction](https://nimona.io).

## Development

### Installation

Assuming you have a working [Go environment] with Go 1.10 or higher:

```
go get -d nimona.io/go
cd $GOPATH/src/nimona.io/go
dep ensure 
```

### Running

You can either `go install nimona.io/go/cmd/nimona` or run it from 
source every time with `go run nimona.io/go/cmd/nimona`.

#### Commands

[CircleCI Image]: https://img.shields.io/circleci/project/github/nimona/go-nimona.svg?style=flat-square
[Coveralls Image]: https://img.shields.io/coveralls/github/nimona/go-nimona.svg?style=flat-square
[License Image]: https://img.shields.io/github/license/nimona/go-nimona.svg?style=flat-square
[Issues Image]: https://img.shields.io/waffle/label/nimona/go-nimona/in%20progress.svg?style=flat-square

[Go environment]: https://golang.org/doc/install
