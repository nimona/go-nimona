[![CircleCI Image]](https://circleci.com/gh/nimona/go-nimona)
[![License Image]](https://github.com/nimona/go-nimona/blob/master/LICENSE)
[![Issues Image]](https://waffle.io/nimona/go-nimona)

# Nimona

Nimona’s main goal is to provide a number of layers/components to help with the challenges presented when dealing with decentralized and peer to peer applications.

## Architecture

As the various components and protocols start taking shape their specifications will live under `docs`.

## Development

### Requirements

- Go 1.11.x with modules enabled
- Make

### Getting Started

```
git clone https://github.com/nimona/go-nimona.git go-nimona
cd go-nimona
make deps
```

### Process / Workflow

Nimona is developed using [Git Common-Flow](https://commonflow.org/), which is
essentially [GitHub Flow](http://scottchacon.com/2011/08/31/github-flow.html)
with the addition of versioned releases, and optional release branches.

In addition to the Common-Flow spec, contributors are also highly encouraged to
[sign commits](https://git-scm.com/book/en/v2/Git-Tools-Signing-Your-Work).

## Usage

### Installation in Provider

You can install the daemon in a supported provider.

```
nimona daemon install --platform do --token <> --ssh-fingerprint <> --hostname <>
```

#### Supported Flags

* **--platform** the provider to be used for the deployment
* **--hostname** the hostname that nimona will use, if defined the dns will also be updated
* **--token** the access token required to authenticate with the provider
* **--ssh-fingerprint** the ssh fingerprint for the key that will be added to the server (needs to exist in the provider)
* **--size** size of the server, default for DO *s-1vcpu-1gb*
* **--region** region that the server will be deployed, default *lon1*

#### Suppored Providers

* do - DigitalOcean

#### Commands

[CircleCI Image]: https://img.shields.io/circleci/project/github/nimona/go-nimona.svg?style=flat-square
[Coveralls Image]: https://img.shields.io/coveralls/github/nimona/go-nimona.svg?style=flat-square
[License Image]: https://img.shields.io/github/license/nimona/go-nimona.svg?style=flat-square
[Issues Image]: https://img.shields.io/waffle/label/nimona/go-nimona/in%20progress.svg?style=flat-square

[Go environment]: https://golang.org/doc/install

