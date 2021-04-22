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
  * private: bagacnag4afafnpql33c4aezi6wqduwc45qd3evanoowrz66vjpapoz7zzmakbpnvq2jerpsoygzdincom76dovh7szmo5r5eug5atlk6tsf53hqyz4
  * public: bahwqdag4aeqllbusjc7e5qnsgq2e4z74g5kp7fsy53d2jin2bgwv5hel3wpbrty

Peer 2:
  * port: 17001
  * private: bagacnag4afagf2gpy7rkgni3ytui7kkyfek2mnc25qru4hh35owk4mdkorc632bhv2mglq52q4n3yi6x4nehagak3p2tqjow5ss2zfttatri6xsdvm
  * public: bahwqdag4aeqcpluymxb3vby3xqr5py2ioamavw7vhas5n3ffvslhgbhcr5pehky

Peer 3:
  * port: 17002
  * private: bagacnag4afamw4hk6fk2yqr7fyug4cpf6fs6pzhkyttmwt4m27emmc76sdjcg6o7xfkyfrhw5gs3sx6r5aorj66xnveeo7wpeu456c4vsp3ga4z6ie
  * public: bahwqdag4aeqn7okvqlcpn2nfxfp5d2a5ct55o3kii57m6jjz34fzle7wmbzt4qi
```

```sh
NIMONA_LOG_LEVEL=error \
NIMONA_UPNP_DISABLE=true \
NIMONA_PEER_BIND_ADDRESS=0.0.0.0:17000 \
NIMONA_PEER_PRIVATE_KEY=bagacnag4afafnpql33c4aezi6wqduwc45qd3evanoowrz66vjpapoz7zzmakbpnvq2jerpsoygzdincom76dovh7szmo5r5eug5atlk6tsf53hqyz4 \
go run ./cmd/bootstrap/main.go
```

```sh
NIMONA_LOG_LEVEL=error \
NIMONA_UPNP_DISABLE=true \
NIMONA_PEER_BIND_ADDRESS=0.0.0.0:17001 \
NIMONA_PEER_PRIVATE_KEY=bagacnag4afagf2gpy7rkgni3ytui7kkyfek2mnc25qru4hh35owk4mdkorc632bhv2mglq52q4n3yi6x4nehagak3p2tqjow5ss2zfttatri6xsdvm \
NIMONA_SONAR_PING_PEERS=bahwqdag4aeqn7okvqlcpn2nfxfp5d2a5ct55o3kii57m6jjz34fzle7wmbzt4qi \
NIMONA_PEER_BOOTSTRAPS=bahwqdag4aeqllbusjc7e5qnsgq2e4z74g5kp7fsy53d2jin2bgwv5hel3wpbrty@tcps:0.0.0.0:17000 \
go run ./cmd/sonar/main.go
```

```sh
NIMONA_LOG_LEVEL=error \
NIMONA_UPNP_DISABLE=true \
NIMONA_PEER_BIND_ADDRESS=0.0.0.0:17002 \
NIMONA_PEER_PRIVATE_KEY=bagacnag4afamw4hk6fk2yqr7fyug4cpf6fs6pzhkyttmwt4m27emmc76sdjcg6o7xfkyfrhw5gs3sx6r5aorj66xnveeo7wpeu456c4vsp3ga4z6ie \
NIMONA_SONAR_PING_PEERS=bahwqdag4aeqcpluymxb3vby3xqr5py2ioamavw7vhas5n3ffvslhgbhcr5pehky \
NIMONA_PEER_BOOTSTRAPS=bahwqdag4aeqllbusjc7e5qnsgq2e4z74g5kp7fsy53d2jin2bgwv5hel3wpbrty@tcps:0.0.0.0:17000 \
go run ./cmd/sonar/main.go
```
