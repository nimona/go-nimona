package nimona

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
)

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

func (n *Node) JoinNetwork(ctx context.Context, nIDs NetworkIdentifier) (*NetworkInfo, error) {
	var peerAddrs []PeerAddr

	switch {
	case nIDs.NetworkAlias != nil:
		var err error
		peerAddrs, err = n.config.Resolver.Resolve(*nIDs.NetworkAlias)
		if err != nil {
			return nil, fmt.Errorf("error resolving network: %w", err)
		}
	case nIDs.NetworkInfo != nil:
		peerAddrs = nIDs.NetworkInfo.PeerAddresses
	case nIDs.NetworkIdentity != nil:
		return nil, fmt.Errorf("NetworkIdentity not yet implemented")
	default:
		return nil, fmt.Errorf("missing network identifier")
	}

	var errs error
	var netInfo *NetworkInfo
	for _, peerAddr := range peerAddrs {
		ses, dialErr := n.sessions.Dial(ctx, peerAddr)
		if dialErr != nil {
			errs = multierror.Append(errs, dialErr)
			continue
		}
		netInfo, dialErr = RequestNetworkInfo(ctx, ses)
		if dialErr != nil {
			errs = multierror.Append(errs, dialErr)
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

	netID := netInfo.NetworkIdentity()
	n.networks.Store(netID, nodeNetwork{
		networkInfo: *netInfo,
	})

	err := n.config.DocumentStore.PutDocument(netInfo.Document())
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
