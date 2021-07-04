# Sonar (Testing tool)

Sonar is a testing tool used as part of our first-pass end to end tests.
It allows creating peers that will attempt to "ping" other specified
peers and wait for them to ping back.

Once sonar has managed to ping all its peers, it then waits to receive pings
from all other defined peers, and exits.

## Env vars

* `NIMONA_PEER_PRIVATE_KEY` - Private key for peer.
* `NIMONA_PEER_BIND_ADDRESS` - Address (in the `ip:port` format) to bind sonar to.
* `NIMONA_PEER_BOOTSTRAPS` - Bootstrap peers to use (in the `publicKey@tcps:ip:port`
  shorthand format).
* `NIMONA_SONAR_PING_PEERS` - Public keys of the peers to lookup and ping.

## Example

Create three peers, one that will be a bootstrap peer.

```txt
Peer 1 (bootstrap):
  * port: 17000
  * private: zrv1ZUTssgEcsJqcuwWcs9rcDNTMg1uptVpW6YkbUckxh5LErpZENo757V39dEidGsZuPBSbsgf3hZSrLFVPPRPdUQq
  * public: z6MkwRpBRDowAYyZ5rHTMfue7HeR22GhKgaSvNDSco4tJN5X

Peer 2:
  * port: 17001
  * private: zrv22oiqYfnUM7oKvk1nU2fBCWoaAdftuTxWjmdGMAik4XWck279nD9RknRCBNUFh3Dw6zLoHfMCNBhHKHADmiBfpew
  * public: z6Mkmctpx1kWoGoFLKv6Qgx2ayvmZ6XWVfWv5WwF3ZE9DZLP

Peer 3:
  * port: 17002
  * private: zrv413HQfEHGdWGVtq2zwdxa9YwcWxpSsZqVwmppSt12RBV1tYhSt7AZ8zGPhkgj14CkutXLBeosESBV4NyWc5HaoSo
  * public: z6MkptMyu7ikDQdrkHyvW3aMwhxzULbxLAeYxeZ7KE6xe8tP
```

```sh
NIMONA_LOG_LEVEL=error \
NIMONA_UPNP_DISABLE=true \
NIMONA_PEER_BIND_ADDRESS=0.0.0.0:17000 \
NIMONA_PEER_PRIVATE_KEY=zrv1ZUTssgEcsJqcuwWcs9rcDNTMg1uptVpW6YkbUckxh5LErpZENo757V39dEidGsZuPBSbsgf3hZSrLFVPPRPdUQq \
go run ./cmd/bootstrap/main.go
```

```sh
NIMONA_LOG_LEVEL=error \
NIMONA_UPNP_DISABLE=true \
NIMONA_PEER_BIND_ADDRESS=0.0.0.0:17001 \
NIMONA_PEER_PRIVATE_KEY=zrv22oiqYfnUM7oKvk1nU2fBCWoaAdftuTxWjmdGMAik4XWck279nD9RknRCBNUFh3Dw6zLoHfMCNBhHKHADmiBfpew \
NIMONA_SONAR_PING_PEERS=z6MkptMyu7ikDQdrkHyvW3aMwhxzULbxLAeYxeZ7KE6xe8tP \
NIMONA_PEER_BOOTSTRAPS=z6MkwRpBRDowAYyZ5rHTMfue7HeR22GhKgaSvNDSco4tJN5X@tcps:0.0.0.0:17000 \
go run ./cmd/sonar/main.go
```

```sh
NIMONA_LOG_LEVEL=error \
NIMONA_UPNP_DISABLE=true \
NIMONA_PEER_BIND_ADDRESS=0.0.0.0:17002 \
NIMONA_PEER_PRIVATE_KEY=zrv413HQfEHGdWGVtq2zwdxa9YwcWxpSsZqVwmppSt12RBV1tYhSt7AZ8zGPhkgj14CkutXLBeosESBV4NyWc5HaoSo \
NIMONA_SONAR_PING_PEERS=z6Mkmctpx1kWoGoFLKv6Qgx2ayvmZ6XWVfWv5WwF3ZE9DZLP \
NIMONA_PEER_BOOTSTRAPS=z6MkwRpBRDowAYyZ5rHTMfue7HeR22GhKgaSvNDSco4tJN5X@tcps:0.0.0.0:17000 \
go run ./cmd/sonar/main.go
```
