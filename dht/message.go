package dht

// Message types
const (
	MessageTypePing string = "messaging:dht:action:ping"
	MessageTypePut         = "messaging:dht:action:put"
	MessageTypeGet         = "messaging:dht:action:get"
)

type messageGet struct {
	OriginPeerID string            `json:"p"`
	QueryID      string            `json:"q"`
	Key          string            `json:"k"`
	Labels       map[string]string `json:"l"`
}

type messagePut struct {
	OriginPeerID string            `json:"p"`
	QueryID      string            `json:"q"`
	Key          string            `json:"k"`
	Value        string            `json:"v"`
	Labels       map[string]string `json:"l"`
}
