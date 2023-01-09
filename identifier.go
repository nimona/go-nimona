package nimona

type ResourceType string

const (
	ResourceTypePeerAddress   ResourceType = "nimona://peer:addr:"
	ResourceTypeDocumentID    ResourceType = "nimona://doc:"
	ResourceTypePeerKey       ResourceType = "nimona://peer:key:"
	ResourceTypeNetworkHandle ResourceType = "nimona://network:handle:"
	ResourceTypeUserHandle    ResourceType = "nimona://user:handle:"
)

func (t ResourceType) String() string {
	return string(t)
}
