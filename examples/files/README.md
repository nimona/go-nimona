
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
NIMONA_BIND_ADDRESS=0.0.0.0:18000 \
NIMONA_DEBUG_METRICS_PORT=6060 \
NIMONA_PEER_PRIVATE_KEY=ed25519.prv.nzNG38rXKFGTPqfNxBNGUyte2hpGgJP77br9GmUQiQ3e9HpqUMFuSavRfz5K5MWhwZskHr48uDD9X8Y2hw3Yg1q \
go run ./main.go serve <filename>
```

```sh
LOG_LEVEL=INFO \
NIMONA_DEBUG_METRICS_PORT=6061 \
NIMONA_BIND_ADDRESS=0.0.0.0:18001 \
NIMONA_PEER_PRIVATE_KEY=ed25519.prv.574sFU7bQUpCsYUnsw9RF4fveUBEYAfbu1DbmVnZ4ieuFkRyPWudHYeHeesrYRQJf2qFh2V5b98AmMseT7VGGEcm \
go run main.go get <hash of the shared file>
```
