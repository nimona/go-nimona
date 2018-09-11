package dht

import (
	"nimona.io/go/blocks"
	"nimona.io/go/crypto"
	"nimona.io/go/peers"
)

func init() {
	blocks.RegisterContentType(&Provider{})
	blocks.RegisterContentType(&PeerInfoRequest{})
	blocks.RegisterContentType(&PeerInfoResponse{})
	blocks.RegisterContentType(&ProviderRequest{})
	blocks.RegisterContentType(&ProviderResponse{})
}

// PeerInfoRequest payload
type PeerInfoRequest struct {
	RequestID string            `json:"requestID,omitempty"`
	PeerID    string            `json:"peerID"`
	Signature *crypto.Signature `json:"-"`
}

func (p *PeerInfoRequest) GetType() string {
	return "dht.provider"
}

func (p *PeerInfoRequest) GetSignature() *crypto.Signature {
	return p.Signature
}

func (p *PeerInfoRequest) SetSignature(s *crypto.Signature) {
	p.Signature = s
}

func (p *PeerInfoRequest) GetAnnotations() map[string]interface{} {
	// no annotations
	return map[string]interface{}{}
}

func (p *PeerInfoRequest) SetAnnotations(a map[string]interface{}) {
	// no annotations
}

// ProviderRequest payload
type ProviderRequest struct {
	RequestID string            `json:"requestID,omitempty"`
	Key       string            `json:"key"`
	Signature *crypto.Signature `json:"-"`
}

func (p *ProviderRequest) GetType() string {
	return "dht.peer-info-request"
}

func (p *ProviderRequest) GetSignature() *crypto.Signature {
	return p.Signature
}

func (p *ProviderRequest) SetSignature(s *crypto.Signature) {
	p.Signature = s
}

func (p *ProviderRequest) GetAnnotations() map[string]interface{} {
	// no annotations
	return map[string]interface{}{}
}

func (p *ProviderRequest) SetAnnotations(a map[string]interface{}) {
	// no annotations
}

// Provider payload
type Provider struct {
	BlockIDs  []string          `json:"blockIDs"`
	Signature *crypto.Signature `json:"-"`
}

func (p *Provider) GetType() string {
	return "dht.peer-info-response"
}

func (p *Provider) GetSignature() *crypto.Signature {
	return p.Signature
}

func (p *Provider) SetSignature(s *crypto.Signature) {
	p.Signature = s
}

func (p *Provider) GetAnnotations() map[string]interface{} {
	// no annotations
	return map[string]interface{}{}
}

func (p *Provider) SetAnnotations(a map[string]interface{}) {
	// no annotations
}

type PeerInfoResponse struct {
	RequestID    string            `json:"requestID,omitempty"`
	PeerInfo     *peers.PeerInfo   `json:"peerInfo,omitempty"`
	ClosestPeers []*peers.PeerInfo `json:"closestPeers,omitempty"`
	Signature    *crypto.Signature `json:"-"`
}

func (p *PeerInfoResponse) GetType() string {
	return "dht.provider-request"
}

func (p *PeerInfoResponse) GetSignature() *crypto.Signature {
	return p.Signature
}

func (p *PeerInfoResponse) SetSignature(s *crypto.Signature) {
	p.Signature = s
}

func (p *PeerInfoResponse) GetAnnotations() map[string]interface{} {
	// no annotations
	return map[string]interface{}{}
}

func (p *PeerInfoResponse) SetAnnotations(a map[string]interface{}) {
	// no annotations
}

type ProviderResponse struct {
	RequestID    string            `json:"requestID,omitempty"`
	Providers    []*Provider       `json:"providers,omitempty"`
	ClosestPeers []*peers.PeerInfo `json:"closestPeers,omitempty"`
	Signature    *crypto.Signature `json:"-"`
}

func (p *ProviderResponse) GetType() string {
	return "dht.provider-response"
}

func (p *ProviderResponse) GetSignature() *crypto.Signature {
	return p.Signature
}

func (p *ProviderResponse) SetSignature(s *crypto.Signature) {
	p.Signature = s
}

func (p *ProviderResponse) GetAnnotations() map[string]interface{} {
	// no annotations
	return map[string]interface{}{}
}

func (p *ProviderResponse) SetAnnotations(a map[string]interface{}) {
	// no annotations
}
