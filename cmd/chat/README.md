
# Chat

Chat is a proof of concept app that allows peers to join public conversations
and send messages.

## Env vars

* `NIMONA_PEER_PRIVATE_KEY` - Private key for peer. (optional)
* `NIMONA_PEER_BIND_ADDRESS` - Address (in the `ip:port` format) to bind sonar 
  to. (optional)
* `NIMONA_PEER_BOOTSTRAPS` - Bootstrap peers to use (in the
  `publicKey@tcps:ip:port`  shorthand format). (optional)
* `NIMONA_CHAT_CHANNEL_HASH` - Channel to join (optional)

## Example

```sh
LOG_LEVEL=fatal \
NIMONA_BIND_ADDRESS=0.0.0.0:18000 \
NIMONA_PEER_PRIVATE_KEY=ed25519.prv.2iFcWsLBbgtbLNX78kYfuA8ZCzaYENmsYvZVMqcLBtPrXAPbZC73T4Wo3ZMeZf93KqvNsYae9wSbsqC6P5VDod8H \
go run ./cmd/chat/*.go
```

```sh
LOG_LEVEL=fatal \
NIMONA_BIND_ADDRESS=0.0.0.0:18001 \
NIMONA_PEER_PRIVATE_KEY=ed25519.prv.3ZJpzEB9QWzprYvbL8FdNDosv7a6gg6otrc8nHLdoyeJnxbngDcvxQtMX3Y8fkG8Dsgo58GtDzxua8YnHYBeJBub \
go run ../cmd/chat/*.go
```

```sh
LOG_LEVEL=fatal \
NIMONA_BIND_ADDRESS=0.0.0.0:18002 \
NIMONA_PEER_PRIVATE_KEY=ed25519.prv.32KvrZFsw39TrabSNPU9oFapT7ygRHWGSL1DqiD36CZf3odwZP5TLkLNdCeN7zk6oRuMwPqRP2wDGuH1N4ukb2Vs \
go run ./cmd/chat/*.go
```
