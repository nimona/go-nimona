package net

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	igd "github.com/emersion/go-upnp-igd"
	ucodec "github.com/ugorji/go/codec"
	"go.uber.org/zap"

	"nimona.io/go/crypto"
	"nimona.io/go/encoding"
	"nimona.io/go/log"
	"nimona.io/go/peers"
)

// Network interface
type Network interface {
	Dial(ctx context.Context, address string) (*Connection, error)
	Listen(ctx context.Context, addrress string) (chan *Connection, error)
	GetPeerInfo() *peers.PeerInfo
	Resolver() Resolver
}

// NewNetwork creates a new p2p network using an address book
func NewNetwork(key *crypto.Key, hostname string) (Network, error) {
	if key == nil {
		return nil, errors.New("missing key")
	}

	if _, ok := key.Materialize().(*ecdsa.PrivateKey); !ok {
		return nil, errors.New("network currently requires an ecdsa private key")
	}

	return &network{
		key:      key,
		resolver: NewResolver(),
		hostname: hostname,
	}, nil
}

// network allows dialing and listening for p2p connections
type network struct {
	key           *crypto.Key
	resolver      Resolver
	hostname      string
	addressesLock sync.RWMutex
	addresses     []string
}

// Dial to a peer and return a net.Conn or error
func (n *network) Dial(ctx context.Context, address string) (*Connection, error) {
	logger := log.Logger(ctx)

	addressType := strings.Split(address, ":")[0]
	switch addressType {
	case "peer":
		peerID := strings.Replace(address, "peer:", "", 1)
		if peerID == n.key.GetPublicKey().HashBase58() {
			return nil, errors.New("cannot dial our own peer")
		}
		logger.Debug("dialing peer", zap.String("peer", address))
		peerInfo, err := n.Resolver().Resolve(peerID)
		if err != nil {
			return nil, err
		}

		if len(peerInfo.Addresses) == 0 {
			return nil, ErrNoAddresses
		}

		for _, addr := range peerInfo.Addresses {
			conn, err := n.Dial(ctx, addr)
			if err == nil {
				return conn, nil
			}
		}

		return nil, ErrAllAddressesFailed

	case "tcps":
		config := tls.Config{
			InsecureSkipVerify: true,
		}
		addr := strings.Replace(address, "tcps:", "", 1)
		dialer := net.Dialer{Timeout: time.Second}
		logger.Debug("dialing", zap.String("address", addr))
		tcpConn, err := tls.DialWithDialer(&dialer, "tcp", addr, &config)
		if err != nil {
			return nil, err
		}

		if tcpConn == nil {
			return nil, ErrAllAddressesFailed
		}

		conn := &Connection{
			Conn:     tcpConn,
			RemoteID: "", // we don't really know who the other side is
		}

		nonce := RandStringBytesMaskImprSrc(8)
		syn := &HandshakeSyn{
			Nonce:    nonce,
			PeerInfo: n.GetPeerInfo(),
		}
		so := syn.ToObject()
		if err := crypto.Sign(so, n.key); err != nil {
			return nil, err
		}

		if err := Write(so, conn); err != nil {
			return nil, err
		}

		synAckObj, err := Read(conn)
		if err != nil {
			return nil, err
		}

		synAck := &HandshakeSynAck{}
		if err := synAck.FromObject(synAckObj); err != nil {
			return nil, err
		}

		if synAck.Nonce != nonce {
			return nil, errors.New("invalid handhshake.syn-ack")
		}

		// store who is on the other side - peer id
		conn.RemoteID = synAck.Signer.HashBase58()
		n.Resolver().Add(synAck.PeerInfo)

		ack := &HandshakeAck{
			Nonce: nonce,
		}
		ao := ack.ToObject()
		if err := crypto.Sign(ao, n.key); err != nil {
			return nil, err
		}

		if err := Write(ao, conn); err != nil {
			return nil, err
		}

		return conn, nil
	default:
		logger.Info("not sure how to dial", zap.String("address", address), zap.String("type", addressType))
	}

	return nil, ErrNoAddresses
}

// Listen on an address
// TODO do we need to return a listener?
func (n *network) Listen(ctx context.Context, address string) (chan *Connection, error) {
	logger := log.Logger(ctx).Named("network")
	cert, err := crypto.GenerateCertificate(n.key)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	config := tls.Config{
		Certificates: []tls.Certificate{*cert},
	}
	config.NextProtos = []string{"nimona/1"} // TODO(geoah) is this of any actual use?
	config.Time = func() time.Time { return now }
	config.Rand = rand.Reader
	tcpListener, err := tls.Listen("tcp", address, &config)
	if err != nil {
		return nil, err
	}

	port := tcpListener.Addr().(*net.TCPAddr).Port
	logger.Info("Listening and service nimona", zap.Int("port", port))
	addresses := GetAddresses(tcpListener)
	devices := make(chan igd.Device, 10)

	if n.hostname != "" {
		addresses = append(addresses, fmtAddress(n.hostname, port))
	}

	upnp := true
	upnpFlag := os.Getenv("UPNP")
	if upnpFlag != "" {
		upnp, _ = strconv.ParseBool(upnpFlag)
	}
	if upnp {
		logger.Info("Trying to find external IP and open port")
		go func() {
			if err := igd.Discover(devices, 2*time.Second); err != nil {
				logger.Error("could not discover devices", zap.Error(err))
			}
		}()
		for device := range devices {
			externalAddress, err := device.GetExternalIPAddress()
			if err != nil {
				logger.Error("could not get external ip", zap.Error(err))
				continue
			}
			desc := "nimona"
			ttl := time.Hour * 24 * 365
			if _, err := device.AddPortMapping(igd.TCP, port, port, desc, ttl); err != nil {
				logger.Error("could not add port mapping", zap.Error(err))
			} else {
				addresses = append(addresses, fmtAddress(externalAddress.String(), port))
			}
		}
	}

	logger.Info("Started listening", zap.Strings("addresses", addresses))
	n.addAddress(addresses...)

	cconn := make(chan *Connection, 10)
	go func() {

		for {
			tcpConn, err := tcpListener.Accept()
			if err != nil {
				log.DefaultLogger.Warn("could not accept connection", zap.Error(err))
				// TODO close conn?
				return
			}

			conn := &Connection{
				Conn:     tcpConn,
				RemoteID: "unknown: handshaking",
			}

			synObj, err := Read(conn)
			if err != nil {
				log.DefaultLogger.Warn("waiting for syn failed", zap.Error(err))
				// TODO close conn?
				continue
			}

			syn := &HandshakeSyn{}
			if err := syn.FromObject(synObj); err != nil {
				log.DefaultLogger.Warn("could not convert obj to syn")
				// TODO close conn?
				continue
			}

			// store the remote peer
			n.Resolver().Add(syn.PeerInfo)

			synAck := &HandshakeSynAck{
				Nonce:    syn.Nonce,
				PeerInfo: n.GetPeerInfo(),
			}
			sao := synAck.ToObject()
			if err := crypto.Sign(sao, n.key); err != nil {
				log.DefaultLogger.Warn("could not sign for syn ack block", zap.Error(err))
				// TODO close conn?
				continue
			}
			if err := Write(sao, conn); err != nil {
				log.DefaultLogger.Warn("sending for syn-ack failed", zap.Error(err))
				// TODO close conn?
				continue
			}

			ackObj, err := Read(conn)
			if err != nil {
				log.DefaultLogger.Warn("waiting for ack failed", zap.Error(err))
				// TODO close conn?
				continue
			}

			ack := &HandshakeSynAck{}
			if err := ack.FromObject(ackObj); err != nil {
				// TODO close conn?
				log.DefaultLogger.Warn("could not convert obj to syn ack")
				continue
			}

			if ack.Nonce != syn.Nonce {
				log.DefaultLogger.Warn("validating syn to ack nonce failed")
				// TODO close conn?
				continue
			}

			conn.RemoteID = ack.Signer.HashBase58()
			cconn <- conn
		}
	}()

	return cconn, nil
}

func Write(o *encoding.Object, conn *Connection) error {
	conn.Conn.SetWriteDeadline(time.Now().Add(time.Second))
	if o == nil {
		log.DefaultLogger.Error("block for fw cannot be nil")
		return errors.New("missing block")
	}

	b, err := encoding.Marshal(o)
	if err != nil {
		return err
	}

	if os.Getenv("DEBUG_BLOCKS") == "true" {
		b, _ := json.MarshalIndent(o.ToMap(), "", "  ")
		log.DefaultLogger.Info(string(b), zap.String("remoteID", conn.RemoteID), zap.String("direction", "outgoing"))
	}

	if _, err := conn.Conn.Write(b); err != nil {
		return err
	}

	SendBlockEvent(
		"outgoing",
		o.GetType(),
		len(b),
	)

	return nil
}

func Read(conn *Connection) (*encoding.Object, error) {
	logger := log.DefaultLogger

	pDecoder := ucodec.NewDecoder(conn.Conn, encoding.RawCborHandler())
	r := &ucodec.Raw{}
	if err := pDecoder.Decode(r); err != nil {
		return nil, err
	}

	defer func() {
		if r := recover(); r != nil {
			logger.Error("Recovered while processing", zap.Any("r", r))
		}
	}()

	o, err := encoding.NewObjectFromBytes([]byte(*r))
	if err != nil {
		return nil, err
	}

	if o.GetSignature() != nil {
		if err := crypto.Verify(o); err != nil {
			return nil, err
		}
	} else {
		fmt.Println("--------------------------------------------------------")
		fmt.Println("----- BLOCK NOT SIGNED ---------------------------------")
		fmt.Println("--------------------------------------------------------")
		fmt.Println("-----", o.GetType())
		fmt.Println("-----", o)
		fmt.Println("--------------------------------------------------------")
	}

	SendBlockEvent(
		"incoming",
		o.GetType(),
		pDecoder.NumBytesRead(),
	)
	if os.Getenv("DEBUG_BLOCKS") == "true" {
		b, _ := json.MarshalIndent(o.ToMap(), "", "  ")
		logger.Info(string(b), zap.String("remoteID", conn.RemoteID), zap.String("direction", "incoming"))
	}
	return o, nil
}

func (n *network) Resolver() Resolver {
	return n.resolver
}

func (n *network) addAddress(addrs ...string) {
	n.addressesLock.Lock()
	if n.addresses == nil {
		n.addresses = []string{}
	}
	n.addresses = append(n.addresses, addrs...)
	n.addressesLock.Unlock()
}

// GetPeerInfo returns the local peer info
func (n *network) GetPeerInfo() *peers.PeerInfo {
	addrs := []string{}
	n.addressesLock.RLock()
	addrs = append(addrs, n.addresses...)
	n.addressesLock.RUnlock()
	// TODO cache peer info and reuse
	p := &peers.PeerInfo{
		Addresses: addrs,
		SignerKey: n.key.GetPublicKey(),
	}
	o := p.ToObject()
	if err := crypto.Sign(o, n.key); err != nil {
		panic(err)
	}
	p.FromObject(o)
	return p
}
