package handshake

import (
	"nimona.io/internal/context"
	"nimona.io/internal/log"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/net"
)

// NewHandshake ...
func New(local *net.LocalInfo, discoverer discovery.Discoverer) net.Middleware {
	return &Handshake{
		local:      local,
		discoverer: discoverer,
	}
}

// Handshake ..
type Handshake struct {
	discoverer discovery.Discoverer
	local      *net.LocalInfo
}

// TODO needs to be able to handle both server and client interactions
func (hs *Handshake) Handle() net.MiddlewareHandler {
	return func(ctx context.Context, conn *net.Connection) (
		*net.Connection, error) {
		if conn.IsIncoming {
			return hs.handleIncoming(ctx, conn)
		}
		return hs.handleOutgoing(ctx, conn)
	}
}

func (hs *Handshake) handleIncoming(ctx context.Context,
	conn *net.Connection) (*net.Connection, error) {

	logger := log.FromContext(ctx)
	logger.Debug("handling inc connection, sending syn")

	nonce := net.RandStringBytesMaskImprSrc(8)
	syn := &Syn{
		Nonce:    nonce,
		PeerInfo: hs.local.GetPeerInfo(),
	}
	so := syn.ToObject()
	if err := crypto.Sign(so, hs.local.GetPeerKey()); err != nil {
		return nil, err
	}

	if err := net.Write(so, conn); err != nil {
		return nil, err
	}

	logger.Debug("sent syn, waiting syn-ack")

	synAckObj, err := net.Read(conn)
	if err != nil {
		return nil, err
	}

	synAck := &SynAck{}
	if err := synAck.FromObject(synAckObj); err != nil {
		return nil, err
	}

	if synAck.Nonce != nonce {
		return nil, net.ErrNonce
	}

	logger.Debug("got syn-ack, sending ack")

	// store who is on the other side
	// TODO Exchange relies on this nees to be somewhere else?
	conn.RemotePeerKey = synAck.PeerInfo.Signature.PublicKey
	hs.discoverer.Add(synAck.PeerInfo)

	ack := &Ack{
		Nonce: nonce,
	}

	ao := ack.ToObject()
	if err := crypto.Sign(ao, hs.local.GetPeerKey()); err != nil {
		return nil, err
	}

	if err := net.Write(ao, conn); err != nil {
		return nil, err
	}

	logger.Debug("sent acl, done")

	return conn, nil

}

func (hs *Handshake) handleOutgoing(ctx context.Context, conn *net.Connection) (
	*net.Connection, error) {
	logger := log.FromContext(ctx).Named("net/middleware/handshake")
	logger.Debug("handling out connection, waiting for syn")

	synObj, err := net.Read(conn)
	if err != nil {
		logger.Warn("waiting for syn failed", log.Error(err))
		// TODO close conn?
		return nil, err
	}

	syn := &Syn{}
	if err := syn.FromObject(synObj); err != nil {
		logger.Warn("could not convert obj to syn")
		// TODO close conn?
		return nil, err
	}

	logger.Debug("got syn, sending syn-ack")

	// store the remote peer
	conn.RemotePeerKey = syn.PeerInfo.Signature.PublicKey
	hs.discoverer.Add(syn.PeerInfo)

	synAck := &SynAck{
		Nonce:    syn.Nonce,
		PeerInfo: hs.local.GetPeerInfo(),
	}

	sao := synAck.ToObject()
	if err := crypto.Sign(sao, hs.local.GetPeerKey()); err != nil {
		logger.Warn(
			"could not sign for syn ack object", log.Error(err))
		// TODO close conn?
		return nil, nil
	}
	if err := net.Write(sao, conn); err != nil {
		logger.Warn("sending for syn-ack failed", log.Error(err))
		// TODO close conn?
		return nil, nil
	}

	logger.Debug("sent syn-ack, waiting ack")

	ackObj, err := net.Read(conn)
	if err != nil {
		logger.Warn("waiting for ack failed", log.Error(err))
		// TODO close conn?
		return nil, nil
	}

	ack := &Ack{}
	if err := ack.FromObject(ackObj); err != nil {
		// TODO close conn?
		logger.Warn("could not convert obj to syn ack")
		return nil, nil
	}

	if ack.Nonce != syn.Nonce {
		logger.Warn("validating syn to ack nonce failed")
		// TODO close conn?
		return nil, nil
	}

	logger.Debug("got ack, done")

	return conn, nil
}
