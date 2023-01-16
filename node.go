package nimona

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"

	"nimona.io/internal/xsync"
)

type (
	Node struct {
		config   NodeConfig
		sessions *SessionManager
		networks *xsync.Map[NetworkID, nodeNetwork]
	}
	nodeNetwork struct {
		networkInfo NetworkInfo
	}
	NodeConfig struct {
		Dialer        Dialer
		Listener      Listener
		Resolver      Resolver
		PublicKey     PublicKey
		PrivateKey    PrivateKey
		Handlers      map[string]RequestHandlerFunc
		DocumentStore *DocumentStore
	}
)

func NewNode(cfg NodeConfig) (*Node, error) {
	ses, err := NewSessionManager(
		cfg.Dialer,
		cfg.Listener,
		cfg.PublicKey,
		cfg.PrivateKey,
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
		networks: xsync.NewMap[NetworkID, nodeNetwork](),
	}

	return n, nil
}

func (n *Node) ListNetworks() ([]NetworkInfo, error) {
	docs, err := n.config.DocumentStore.GetDocumentsByType("core/network/info")
	if err != nil {
		return nil, fmt.Errorf("error getting documents: %w", err)
	}

	var networks []NetworkInfo
	for _, doc := range docs {
		netInfo := &NetworkInfo{}
		err := doc.UnmarshalInto(netInfo)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling document: %w", err)
		}
		networks = append(networks, *netInfo)
	}

	return networks, nil
}

func (n *Node) JoinNetwork(ctx context.Context, nID NetworkID) (*NetworkInfo, error) {
	peerAddrs, err := n.config.Resolver.Resolve(nID)
	if err != nil {
		return nil, fmt.Errorf("error resolving network: %w", err)
	}

	var errs error
	var netInfo *NetworkInfo
	for _, peerAddr := range peerAddrs {
		ses, err := n.sessions.Dial(ctx, peerAddr)
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		}
		netInfo, err = RequestNetworkInfo(ctx, ses)
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		}
		break
	}

	if errs != nil {
		return nil, fmt.Errorf("error joining network: %w", errs)
	}

	if netInfo == nil {
		return nil, fmt.Errorf("missing response when joining network: %w", errs)
	}

	docID := NewDocumentIDFromCBOR(netInfo.RawBytes)
	err = n.config.DocumentStore.PutDocument(&DocumentEntry{
		DocumentID:       docID,
		DocumentType:     "core/network/info",
		DocumentEncoding: "cbor",
		DocumentBytes:    netInfo.RawBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("error storing network info: %w", err)
	}

	if netInfo == nil {
		return nil, fmt.Errorf("error joining network: %w", errs)
	}

	return netInfo, nil
}

func (n *Node) Close() error {
	return n.sessions.Close()
}
