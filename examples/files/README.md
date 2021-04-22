
# Files

Files is a proof of concept cli tool that allows users to exchange files.

## Env vars

* `NIMONA_PEER_PRIVATE_KEY` - Private key for peer. (optional)
* `NIMONA_PEER_BIND_ADDRESS` - Address (in the `ip:port` format) to bind sonar
  to. (optional)
* `NIMONA_PEER_BOOTSTRAPS` - Bootstrap peers to use (in the
  `publicKey@tcps:ip:port`  shorthand format). (optional)
* `NIMONA_DEBUG_METRICS_PORT` - Enable application runtime statistics.

```sh
LOG_LEVEL=INFO \
NIMONA_BIND_PEER_ADDRESS=0.0.0.0:18000 \
NIMONA_DEBUG_METRICS_PORT=6060 \
NIMONA_PEER_PRIVATE_KEY= <private key> \
go run *.go serve <filename>
```

```sh
LOG_LEVEL=INFO \
NIMONA_DEBUG_METRICS_PORT=6061 \
NIMONA_BIND_PEER_ADDRESS=0.0.0.0:18001 \
NIMONA_PEER_PRIVATE_KEY= <private key> \
go run *.go get <hash of the shared file>
```
