package nimona

type DocumentMapper interface {
	DocumentMap() DocumentMap
	FromDocumentMap(DocumentMap)
}

type MessageWrapper struct {
	Type string `nimona:"$type"`
}
