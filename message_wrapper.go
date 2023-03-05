package nimona

type Documenter interface {
	Document() *Document
	FromDocument(*Document) error
}

type MessageWrapper struct {
	Type string `nimona:"$type"`
}
