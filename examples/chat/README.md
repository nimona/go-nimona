
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

Create some keys for your peers.

```sh
go run ./cmd/keygen/main.go
```

```sh
LOG_LEVEL=fatal \
NIMONA_BIND_ADDRESS=0.0.0.0:18000 \
NIMONA_PEER_PRIVATE_KEY=<key1> \
go run ./examples/chat/*.go
```

```sh
LOG_LEVEL=fatal \
NIMONA_BIND_ADDRESS=0.0.0.0:18001 \
NIMONA_PEER_PRIVATE_KEY=<key2> \
go run ../examples/chat/*.go
```

```sh
LOG_LEVEL=fatal \
NIMONA_BIND_ADDRESS=0.0.0.0:18002 \
NIMONA_PEER_PRIVATE_KEY=<key3> \
go run ./examples/chat/*.go
```
