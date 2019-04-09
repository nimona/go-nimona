package handshake

import (
	"context"

	"go.uber.org/zap"
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

		if conn.IsOutgoing {
			return hs.handleOutgoing(ctx, conn)
		}

		// TODO error failed handle connection
		return nil, nil

	}
}

func (hs *Handshake) handleIncoming(ctx context.Context,
	conn *net.Connection) (*net.Connection, error) {

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

	// store who is on the other side
	// TODO Exchange relies on this nees to be somewhere else?
	conn.RemotePeerKey = synAck.PeerInfo.SignerKey
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

	return conn, nil

}

func (hs *Handshake) handleOutgoing(ctx context.Context, conn *net.Connection) (
	*net.Connection, error) {
	synObj, err := net.Read(conn)
	if err != nil {
		log.DefaultLogger.Warn("waiting for syn failed", zap.Error(err))
		// TODO close conn?
		return nil, err
	}

	syn := &Syn{}
	if err := syn.FromObject(synObj); err != nil {
		log.DefaultLogger.Warn("could not convert obj to syn")
		// TODO close conn?
		return nil, err
	}

	// store the remote peer
	conn.RemotePeerKey = syn.PeerInfo.SignerKey
	hs.discoverer.Add(syn.PeerInfo)

	synAck := &SynAck{
		Nonce:    syn.Nonce,
		PeerInfo: hs.local.GetPeerInfo(),
	}

	sao := synAck.ToObject()
	if err := crypto.Sign(sao, hs.local.GetPeerKey()); err != nil {
		log.DefaultLogger.Warn("could not sign for syn ack object", zap.Error(err))
		// TODO close conn?
		return nil, nil
	}
	if err := net.Write(sao, conn); err != nil {
		log.DefaultLogger.Warn("sending for syn-ack failed", zap.Error(err))
		// TODO close conn?
		return nil, nil
	}

	ackObj, err := net.Read(conn)
	if err != nil {
		log.DefaultLogger.Warn("waiting for ack failed", zap.Error(err))
		// TODO close conn?
		return nil, nil
	}

	ack := &Ack{}
	if err := ack.FromObject(ackObj); err != nil {
		// TODO close conn?
		log.DefaultLogger.Warn("could not convert obj to syn ack")
		return nil, nil
	}

	if ack.Nonce != syn.Nonce {
		log.DefaultLogger.Warn("validating syn to ack nonce failed")
		// TODO close conn?
		return nil, nil
	}

	return conn, nil
}
