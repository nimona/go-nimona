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
