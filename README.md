<h1 align="center">
  <img src="https://user-images.githubusercontent.com/88447/67148586-4010a580-f2a1-11e9-9ece-91acf37b0c6f.png" alt="nimona" width="250px">
</h1>
<h4 align="center">a new internet stack; or something like it.</h4>

⚠️ **Note:** Nimona's architecture is getting an ovehaul based on the findings of the first version and more information should be popping up in the `v1/main` branch soon.

---

# Nimona

Nimona’s main goal is to provide a number of layers/components to help with
the challenges presented when dealing with decentralized and peer to peer
applications.

## Development

### Requirements

- go 1.18+ with modules enabled
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

## Library Architecture

![Library Architecture](./README-lib-architecture.drawio.svg)

### Network

Package `exchange` is responsible for a number of things around connections and
object exchange, as well as relaying objects to inaccessible peers.

```go
type Network interface {
    Subscribe(
        filters ...EnvelopeFilter,
    ) EnvelopeSubscription
    Send(
        ctx context.Context,
        object object.Object,
        recipient *peer.ConnectionInfo,
    ) error
    Listen(
        ctx context.Context,
        bindAddress string,
    ) (Listener, error)
}
```

### Resolver

Package `resolver` is responsible for looking up peers on the network that
fulfill specific requirements.

```go
type Resolver interface {
    Lookup(
        ctx context.Context,
        opts ...LookupOption,
    ) (<-chan *peer.ConnectionInfo, error)
}
```

The currently available `LookupOption` are the following, and can be used
on their own or in groups.

```go
func LookupByDigest(hash tilde.Digest) LookupOption { ... }
func LookupByDID(id did.DID) LookupOption { ... }
```

<!-- Links -->

[Go environment]: https://golang.org/doc/install

<!-- Badge images -->

[Actions Status]: https://github.com/nimona/go-nimona/workflows/CI/badge.svg?style=flat
[License Status]: https://img.shields.io/github/license/nimona/go-nimona.svg?style=flat
