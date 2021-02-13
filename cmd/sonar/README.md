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
  * private: bagacmacaslgzj7n6d4fwxs3by7nuqog45gwm5khjgtrm2eqpntui544swrlxc4cxrivfd2utkafyfkealenybtoxgsbqi5ow3wjfnj3iiyrmcuq
  * public: bahwqcabaofyfpcrkkhvjgualqkuiawi3qdg5onedar25nxmsk2twqrrcyfja

Peer 2:
  * port: 17001
  * private: bagacmacatzmnsq5kjjs2xmmhu6pnruu54uvmonpmjkpjd32sptnszpsejm6hyhbnffcrryogndoa2yhe2g4xcr7stib5w6yggu5yepqvz7mnzwi
  * public: bahwqcabapqoc2kkfddq4m2g4bvqojunzofd7fgqd3n5qmnj3qi7blt6y3tmq

Peer 3:
  * port: 17002
  * private: bagacmacaekai32qmxeol6thr5ml6vpym3mx74zlelifwdhw3smynevgv726mlqvzatc67d5apqkebqdxvggxrfx2ifyfkb53fftqnde5vlctx3i
  * public: bahwqcabayxblsbgf56h2a7auidahpkmnpclpuqlqkud3wklha2gj3kwfhpwq
```

```sh
NIMONA_LOG_LEVEL=error \
NIMONA_UPNP_DISABLE=true \
NIMONA_PEER_BIND_ADDRESS=0.0.0.0:17000 \
NIMONA_PEER_PRIVATE_KEY=bagacmacaslgzj7n6d4fwxs3by7nuqog45gwm5khjgtrm2eqpntui544swrlxc4cxrivfd2utkafyfkealenybtoxgsbqi5ow3wjfnj3iiyrmcuq \
go run ./cmd/bootstrap/main.go
```

```sh
NIMONA_LOG_LEVEL=error \
NIMONA_UPNP_DISABLE=true \
NIMONA_PEER_BIND_ADDRESS=0.0.0.0:17001 \
NIMONA_PEER_PRIVATE_KEY=bagacmacatzmnsq5kjjs2xmmhu6pnruu54uvmonpmjkpjd32sptnszpsejm6hyhbnffcrryogndoa2yhe2g4xcr7stib5w6yggu5yepqvz7mnzwi \
NIMONA_SONAR_PING_PEERS=bahwqcabayxblsbgf56h2a7auidahpkmnpclpuqlqkud3wklha2gj3kwfhpwq \
NIMONA_PEER_BOOTSTRAPS=bahwqcabaofyfpcrkkhvjgualqkuiawi3qdg5onedar25nxmsk2twqrrcyfja@tcps:0.0.0.0:17000 \
go run ./cmd/sonar/main.go
```

```sh
NIMONA_LOG_LEVEL=error \
NIMONA_UPNP_DISABLE=true \
NIMONA_PEER_BIND_ADDRESS=0.0.0.0:17002 \
NIMONA_PEER_PRIVATE_KEY=bagacmacaekai32qmxeol6thr5ml6vpym3mx74zlelifwdhw3smynevgv726mlqvzatc67d5apqkebqdxvggxrfx2ifyfkb53fftqnde5vlctx3i \
NIMONA_SONAR_PING_PEERS=bahwqcabapqoc2kkfddq4m2g4bvqojunzofd7fgqd3n5qmnj3qi7blt6y3tmq \
NIMONA_PEER_BOOTSTRAPS=bahwqcabaofyfpcrkkhvjgualqkuiawi3qdg5onedar25nxmsk2twqrrcyfja@tcps:0.0.0.0:17000 \
go run ./cmd/sonar/main.go
```
