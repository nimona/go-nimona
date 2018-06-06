package wire

// Envelope for wrapping the message
type Envelope struct {
	Version uint8  `json:"version"`
	Message []byte `json:"message"`
}
