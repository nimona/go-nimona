package peer

const ConnectionInfoType = "nimona.io/peer.ConnectionInfo"

type ConnectionInfo struct {
	Owner         ID                `nimona:"@metadata.owner:s,type=nimona.io/peer.ConnectionInfo"`
	Timestamp     string            `nimona:"@metadata.timestamp:s"`
	Version       int64             `nimona:"version:i"`
	Addresses     []string          `nimona:"addresses:as"`
	Relays        []*ConnectionInfo `nimona:"relays:am"`
	ObjectFormats []string          `nimona:"objectFormats:as"`
}
