# Bootstrap

Bootstrap daemon.

## Env vars

* `PEER_PRIVATE_KEY` - Private key for peer.
* `BIND_ADDRESS` - Address (in the `host:port` format) to bind the peer to.
* `ANNOUNCE_ADDRESS` - Address (in the `host:port` format) to announce in
  addition to the bound ones. Mostly used to announce hosts.
* `BOOTSTRAP_PEERS` - Other bootstrap peers to use
  (in the `publicKey@tcps:ip:port` shorthand format).
