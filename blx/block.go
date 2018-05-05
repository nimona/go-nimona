package blx

const (
	MessageTypeTransfer string = "blx:action:transfer"
	MessageTypeRequest         = "blx:action:request"
)

type Block struct {
	Key  string            `json:"k"`
	Data []byte            `json:"d"`
	Meta map[string][]byte `json:"m"`
}

type BlockRequest struct {
	Key string `json:"k"`
}
