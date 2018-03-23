package messenger

// Message for our wire protocol
type Message struct {
	Version   int
	Sender    string
	Recipient string
	Topics    []string
	Payload   []byte
	Checksum  []byte
}
