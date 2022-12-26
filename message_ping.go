package nimona

type (
	Ping struct {
		_     string `cborgen:"$type,const=test/ping"`
		Nonce string `cborgen:"nonce"`
	}
	Pong struct {
		_     string `cborgen:"$type,const=test/pong"`
		Nonce string `cborgen:"nonce"`
	}
)
