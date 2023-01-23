package nimona

type DocumentType string

const (
	DocumentTypeDocumentID  DocumentType = "nimona://doc:"
	DocumentTypePeerAddress DocumentType = "nimona://peer:addr:"
	DocumentTypePeerKey     DocumentType = "nimona://peer:key:"

	DocumentTypeIdentity      DocumentType = "nimona://id:"       // ...<keystreamID>
	DocumentTypeIdentityAlias DocumentType = "nimona://id:alias:" // ...<handle>@<hostname>

	DocumentTypeNetwork      DocumentType = "nimona://net:"       // ...<keystreamID>
	DocumentTypeNetworkAlias DocumentType = "nimona://net:alias:" // ...<hostname>
)

func (t DocumentType) String() string {
	return string(t)
}
