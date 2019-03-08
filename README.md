[![CircleCI Image]](https://circleci.com/gh/nimona/go-nimona)
[![License Image]](https://github.com/nimona/go-nimona/blob/master/LICENSE)
[![Issues Image]](https://waffle.io/nimona/go-nimona)

# Nimona

Nimona’s main goal is to provide a number of layers/components to help with the challenges presented when dealing with decentralized and peer to peer applications.

## Architecture

As the various components and protocols start taking shape their specifications will live under [`docs`](./docs).

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
