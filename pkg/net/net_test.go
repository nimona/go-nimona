package net

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"nimona.io/internal/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/net/peer"
)

func TestNetDiscoverer(t *testing.T) {
	disc1 := discovery.NewDiscoverer()
	disc2 := discovery.NewDiscoverer()

	_, _, l1 := newPeer(t, "", disc1)
	_, _, l2 := newPeer(t, "", disc2)

	ctx := context.New()

	disc1.Add(l2.GetPeerInfo())
	disc2.Add(l1.GetPeerInfo())


	q1 := &peer.PeerInfoRequest{
		Keys: []string{
			 l2.GetPeerKey().Fingerprint(),
		},
	}
	ps2, err := disc1.Discover(ctx, q1)
	p2 := ps2[0]
	assert.NoError(t, err)
	// assert.Equal(t, n2.key.PublicKey, p2.SignerKey)
	assert.Equal(t,
		l2.GetPeerKey().Fingerprint(),
		p2.Fingerprint(),
	)

	q2 := &peer.PeerInfoRequest{
		Keys: []string{
			 l1.GetPeerKey().Fingerprint(),
		},
	}
	ps1, err := disc2.Discover(ctx, q2)
	p1 := ps1[0]
	assert.NoError(t, err)
	// assert.Equal(t, n1.key.PublicKey, p1.SignerKey)
	assert.Equal(t,
		l1.GetPeerKey().Fingerprint(),
		p1.Fingerprint(),
	)
}





func TestNetConnectionSuccess(t *testing.T) {

	disc1 := discovery.NewDiscoverer()
	disc2 := discovery.NewDiscoverer()
	address1 := fmt.Sprintf("0.0.0.0:%d", 0)

	ctx := context.New()

	BindLocal = true
	_, n1, l1 := newPeer(t, "", disc1)
	_, n2, l2 := newPeer(t, "", disc2)

	// we need to start listening before we add the peerInfo
	// otherwise the addresses are not populated
	sconn, err := n1.Listen(ctx, address1)
	assert.NoError(t, err)

	disc1.Add(l2.GetPeerInfo())
	disc2.Add(l1.GetPeerInfo())

	peer1Addr := l1.GetPeerInfo().Address()

	done := make(chan bool)

	go func() {
		cconn, err := n2.Dial(ctx, peer1Addr)
		assert.NoError(t, err)
		err = Write((peer.PeerInfoRequest{}).ToObject(), cconn)
		assert.NoError(t, err)
		done <- true

	}()

	sc := <-sconn

	err = Write((peer.PeerInfoRequest{}).ToObject(), sc)
	assert.NoError(t, err)

	<-done
}

func TestNetConnectionFailureMiddleware(t *testing.T) {

	disc1 := discovery.NewDiscoverer()
	disc2 := discovery.NewDiscoverer()
	address1 := fmt.Sprintf("0.0.0.0:%d", 0)

	ctx := context.New()

	BindLocal = true
	_, n1, l1 := newPeer(t, "", disc1)
	_, n2, l2 := newPeer(t, "", disc2)

	// we need to start listening before we add the peerInfo
	// otherwise the addresses are not populated
	fm := fakeMid{}

	sconn, err := n1.Listen(ctx, address1)
	n1.AddMiddleware(fm.Handle())
	assert.NoError(t, err)

	disc1.Add(l2.GetPeerInfo())
	disc2.Add(l1.GetPeerInfo())

	peer1Addr := l1.GetPeerInfo().Address()

	done := make(chan bool)

	go func() {
		cconn, err := n2.Dial(ctx, peer1Addr)
		assert.Error(t, err)
		assert.Nil(t, cconn)
		done <- true
	}()

	<-done
	assert.Len(t, sconn, 0)
}

func newPeer(t *testing.T, relayAddress string, discover discovery.Discoverer) (
	*crypto.PrivateKey, *network, *LocalInfo) {
	pk, err := crypto.GenerateKey()
	assert.NoError(t, err)

	relayAddresses := []string{}
	if relayAddress != "" {
		relayAddresses = append(relayAddresses, relayAddress)
	}

	localInfo, err := NewLocalInfo("", pk)
	n, err := New(discover, localInfo)
	assert.NoError(t, err)

	tcpTr := NewTCPTransport(localInfo, relayAddresses)
	n.AddTransport("tcps", tcpTr)

	return pk, n.(*network), localInfo
}

type fakeMid struct {
}

func (fm *fakeMid) Handle() MiddlewareHandler {
	return func(ctx context.Context, conn *Connection) (*Connection, error) {
		return conn, errors.New("what?")
	}
}
