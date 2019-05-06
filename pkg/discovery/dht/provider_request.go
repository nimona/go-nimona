package dht

//go:generate go run nimona.io/tools/objectify -schema nimona.io/dht/provider.request -type ProviderRequest -in provider_request.go -out provider_request_generated.go

// ProviderRequest payload
type ProviderRequest struct {
	RequestID string `json:"requestID,omitempty"`
	Key       string `json:"key"`
}
