# Bootstrap

Bootstrap is basically a Hyperspace v2 provider

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
NIMONA_PEER_PRIVATE_KEY=<private_key> \
go run ./cmd/bootstrap/main.go
```
