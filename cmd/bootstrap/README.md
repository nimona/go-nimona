# Bootstrap

Bootstrap daemon.

## Env vars

* `NIMONA_PEER_PRIVATE_KEY` - Private key for peer.
* `NIMONA_PEER_BIND_ADDRESS` - Address (in the `host:port` format) to bind the
  peer to.
* `NIMONA_PEER_ANNOUNCE_ADDRESS` - Address (in the `host:port` format) to
  announce in addition to the bound ones. Mostly used to announce hosts.
* `NIMONA_PEER_BOOTSTRAPS` - Other bootstrap peers to use
  (in the `publicKey@tcps:ip:port` shorthand format).

## Example

```sh
LOG_LEVEL=info \
NIMONA_PEER_BIND_ADDRESS=0.0.0.0:17002 \
NIMONA_METRICS_BIND_ADDRESS=0.0.0.0:17001 \
NIMONA_PEER_PRIVATE_KEY=ed25519.prv.3Bock8sqT8UWEZargCJDWo9rsG4AcEMSubdm6fHLRxR4d5S41UVmUYLQSc9qjHkKPiaobE8JEaY6Bo4YJqnEG8Y9 \
go run ./cmd/bootstrap/main.go
```
