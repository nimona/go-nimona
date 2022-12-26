package nimona

type NodeInfo struct {
	_         string     `cborgen:"$type,const=core/node.info"`
	Addresses []NodeAddr `cborgen:"addresses"`
}
