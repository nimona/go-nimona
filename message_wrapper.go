package nimona

type DocumentMapper interface {
	Document() *Document
	FromDocumentMap(*Document) error
}

type MessageWrapper struct {
	Type string `nimona:"$type"`
}
