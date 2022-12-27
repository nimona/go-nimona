package nimona

type PeerInfo struct {
	_         string     `cborgen:"$type,const=core/node.info"`
	Addresses []PeerAddr `cborgen:"addresses"`
}
