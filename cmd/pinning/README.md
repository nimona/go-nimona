# Pinning Service

Allows pinning objects and streams.

## Env vars

* `NIMONA_PEER_PRIVATE_KEY` - Private key for peer.
* `NIMONA_PEER_BIND_ADDRESS` - Address (in the `host:port` format) to bind the
  peer to.
* `NIMONA_PEER_ANNOUNCE_ADDRESS` - Address (in the `host:port` format) to
  announce in addition to the bound ones. Mostly used to announce hosts.
* `NIMONA_PEER_BOOTSTRAPS` - Other bootstrap peers to use
  (in the `publicKey@tcps:ip:port` shorthand format).

## Commands

```sh
go run ./cmd/pinning/main.go serve
go run ./cmd/pinning/main.go list <pinning-service-peer-key>
go run ./cmd/pinning/main.go pin <pinning-service-peer-key> <object-hash>
```
