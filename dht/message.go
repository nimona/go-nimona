package dht

// Message types
const (
	MessageTypePing string = "PING"
	MessageTypePut         = "PUT"
	MessageTypeGet         = "GET"
)

// Key prefixes
const (
	KeyPrefixPeer         string = "nimona/peer/"
	KeyPrefixProvider            = "nimona/provider/"
	KeyPrefixKeyValuePair        = "nimona/kv/"
)

type messageGet struct {
	OriginPeer *messagePeer `json:"p"`
	QueryID    string       `json:"q"`
	Key        string       `json:"k"`
}

type messagePeer struct {
	ID        string   `json:"id"`
	Addresses []string `json:"addresses"`
}

type messagePut struct {
	OriginPeer *messagePeer `json:"p"`
	QueryID    string       `json:"q"`
	Key        string       `json:"k"`
	Values     []string     `json:"v"`
}
