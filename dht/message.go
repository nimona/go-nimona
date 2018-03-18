package dht

// Message types
const (
	MessageTypePing string = "PING"
	MessageTypePut         = "PUT"
	MessageTypeGet         = "GET"
)

type messageGet struct {
	OriginPeer *messagePeer      `json:"p"`
	QueryID    string            `json:"q"`
	Key        string            `json:"k"`
	Labels     map[string]string `json:"l"`
}

type messagePeer struct {
	ID        string   `json:"id"`
	Addresses []string `json:"addresses"`
}

type messagePut struct {
	OriginPeer *messagePeer      `json:"p"`
	QueryID    string            `json:"q"`
	Key        string            `json:"k"`
	Value      string            `json:"v"`
	Labels     map[string]string `json:"l"`
}
