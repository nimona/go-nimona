package nimona

import (
	"fmt"

	"nimona.io/internal/xsync"
)

type (
	Node struct {
		config   *NodeConfig
		sessions *SessionManager
		networks *xsync.Map[NetworkIdentity, nodeNetwork]
	}
	nodeNetwork struct {
		networkInfo NetworkInfo
	}
	NodeConfig struct {
		Dialer        Dialer
		Listener      Listener
		Resolver      Resolver
		PeerConfig    *PeerConfig
		Handlers      map[string]RequestHandlerFunc
		DocumentStore *DocumentStore
	}
)

func NewNode(cfg *NodeConfig) (*Node, error) {
	if cfg == nil {
		return nil, fmt.Errorf("missing config")
	}

	ses, err := NewSessionManager(
		cfg.Dialer,
		cfg.Listener,
		cfg.PeerConfig.GetPublicKey(),
		cfg.PeerConfig.GetPrivateKey(),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating session manager: %w", err)
	}

	for docType, handler := range cfg.Handlers {
		ses.RegisterHandler(docType, handler)
	}

	n := &Node{
		config:   cfg,
		sessions: ses,
		networks: xsync.NewMap[NetworkIdentity, nodeNetwork](),
	}

	return n, nil
}
