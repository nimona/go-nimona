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

Create three peers, one that will be used as a bootstrap peer.

```txt
Peer 1:
  * port: 17000
  * private: ed25519.prv.Jf3xha8ZqEnFv9T9UDcN41nFFfZpc9MY4tzUnpgGHx8ZwKQ6uXX6PGY1nHLQAKhPiFtV4YEqMsCd5vjkdRyC5nJ
  * public: ed25519.J9AfT7J2SbXen83NuyVEQ7UkpCaLJbnw41nLrR82HnSW

Peer 2:
  * port: 17001
  * private: ed25519.prv.2bAdgQxfcJsGRMccgMXkGSPQt396g77KKq8y6fEeQbxpnPqS5Ujh1DXTNU539wW5ispS1McLyKjrJDsgxYKneyCZ
  * public: ed25519.3ykKbHUoHE8Sa9P6ckzrsXzGw3HC9iV4vnTcNrrcBBmP

Peer 3:
  * port: 17002
  * private: ed25519.prv.4i1anFeotM4TKnjLsJFLgwERtq4rD5yaR6AQ5HuChgNBfzrApXpQYA8WT83bMSc8CLj76LbJfdSKn3HiKmSpn25U
  * public: ed25519.9CA3BuLzPrxHAHET8zicCtTku5zAaPsA6WRFp4PRARx2
```

```sh
LOG_LEVEL=error \
NIMONA_BIND_ADDRESS=0.0.0.0:17000 \
NIMONA_PEER_PRIVATE_KEY=ed25519.prv.Jf3xha8ZqEnFv9T9UDcN41nFFfZpc9MY4tzUnpgGHx8ZwKQ6uXX6PGY1nHLQAKhPiFtV4YEqMsCd5vjkdRyC5nJ \
NIMONA_SONAR_PING_PEERS=ed25519.3ykKbHUoHE8Sa9P6ckzrsXzGw3HC9iV4vnTcNrrcBBmP,ed25519.9CA3BuLzPrxHAHET8zicCtTku5zAaPsA6WRFp4PRARx2 \
go run ./cmd/sonar/main.go
```

```sh
LOG_LEVEL=error \
NIMONA_BIND_ADDRESS=0.0.0.0:17001 \
NIMONA_PEER_PRIVATE_KEY=ed25519.prv.2bAdgQxfcJsGRMccgMXkGSPQt396g77KKq8y6fEeQbxpnPqS5Ujh1DXTNU539wW5ispS1McLyKjrJDsgxYKneyCZ \
NIMONA_SONAR_PING_PEERS=ed25519.J9AfT7J2SbXen83NuyVEQ7UkpCaLJbnw41nLrR82HnSW,ed25519.9CA3BuLzPrxHAHET8zicCtTku5zAaPsA6WRFp4PRARx2 \
NIMONA_PEER_BOOTSTRAPS=ed25519.J9AfT7J2SbXen83NuyVEQ7UkpCaLJbnw41nLrR82HnSW@tcps:0.0.0.0:17000 \
go run ./cmd/sonar/main.go
```

```sh
LOG_LEVEL=error \
NIMONA_BIND_ADDRESS=0.0.0.0:17002 \
NIMONA_PEER_PRIVATE_KEY=ed25519.prv.4i1anFeotM4TKnjLsJFLgwERtq4rD5yaR6AQ5HuChgNBfzrApXpQYA8WT83bMSc8CLj76LbJfdSKn3HiKmSpn25U \
NIMONA_SONAR_PING_PEERS=ed25519.J9AfT7J2SbXen83NuyVEQ7UkpCaLJbnw41nLrR82HnSW,ed25519.3ykKbHUoHE8Sa9P6ckzrsXzGw3HC9iV4vnTcNrrcBBmP \
NIMONA_PEER_BOOTSTRAPS=ed25519.J9AfT7J2SbXen83NuyVEQ7UkpCaLJbnw41nLrR82HnSW@tcps:0.0.0.0:17000 \
go run ./cmd/sonar/main.go
```
