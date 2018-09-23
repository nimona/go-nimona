package net

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	igd "github.com/emersion/go-upnp-igd"
	ucodec "github.com/ugorji/go/codec"
	"go.uber.org/zap"

	"nimona.io/go/blocks"
	"nimona.io/go/codec"
	"nimona.io/go/log"
	"nimona.io/go/peers"
)

// Networker interface for mocking Network
type Networker interface {
	Dial(ctx context.Context, peerID string) (*Connection, error)
	Listen(ctx context.Context, addrress string) (chan *Connection, error)
}

// NewNetwork creates a new p2p network using an address book
func NewNetwork(addressBook *peers.AddressBook) (*Network, error) {
	return &Network{
		addressBook: addressBook,
	}, nil
}

// Network allows dialing and listening for p2p connections
type Network struct {
	addressBook *peers.AddressBook
}

// Dial to a peer and return a net.Conn or error
func (n *Network) Dial(ctx context.Context, peerID string) (*Connection, error) {
	logger := log.Logger(ctx)
	peerInfo, err := n.addressBook.GetPeerInfo(peerID)
	if err != nil {
		return nil, err
	}

	if len(peerInfo.Addresses) == 0 {
		return nil, ErrNoAddresses
	}

	var tcpConn net.Conn
	for _, addr := range peerInfo.Addresses {
		if !strings.HasPrefix(addr, "tcp:") {
			continue
		}
		addr = strings.Replace(addr, "tcp:", "", 1)
		dialer := net.Dialer{Timeout: time.Second}
		logger.Debug("dialing", zap.String("address", addr))
		newTcpConn, err := dialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			continue
		}
		tcpConn = newTcpConn
		break
	}

	if tcpConn == nil {
		return nil, ErrAllAddressesFailed
	}

	conn := &Connection{
		Conn:     tcpConn,
		RemoteID: peerID,
	}

	signer := n.addressBook.GetLocalPeerKey()
	nonce := RandStringBytesMaskImprSrc(8)
	syn := &HandshakeSyn{
		Nonce: nonce,
	}
	if err := Write(syn, conn, blocks.SignWith(signer)); err != nil {
		return nil, err
	}

	typedSynAck, err := Read(conn)
	if err != nil {
		return nil, err
	}

	if typedSynAck.(*HandshakeSynAck).Nonce != nonce {
		return nil, errors.New("invalid handhshake.syn-ack")
	}

	ack := &HandshakeAck{
		Nonce: nonce,
	}
	if err := Write(ack, conn, blocks.SignWith(signer)); err != nil {
		return nil, err
	}

	return conn, nil
}

// Listen on an address
// TODO do we need to return a listener?
func (n *Network) Listen(ctx context.Context, addr string) (chan *Connection, error) {
	logger := log.Logger(ctx).Named("network")
	tcpListener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	port := tcpListener.Addr().(*net.TCPAddr).Port
	logger.Info("Listening and service nimona", zap.Int("port", port))
	addresses := GetAddresses(tcpListener)
	devices := make(chan igd.Device, 10)

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
				addresses = append(addresses, fmt.Sprintf("tcp:%s:%d", externalAddress.String(), port))
			}
		}
	}

	logger.Info("Started listening", zap.Strings("addresses", addresses))
	n.addressBook.AddLocalPeerAddress(addresses...)

	cconn := make(chan *Connection, 10)
	go func() {
		signer := n.addressBook.GetLocalPeerKey()
		for {
			tcpConn, err := tcpListener.Accept()
			if err != nil {
				log.DefaultLogger.Warn("could not accept connection", zap.Error(err))
				return
			}

			conn := &Connection{
				Conn:     tcpConn,
				RemoteID: "unknown: handshaking",
			}

			typedSyn, err := Read(conn)
			if err != nil {
				log.DefaultLogger.Warn("waiting for syn failed", zap.Error(err))
				continue
			}

			nonce := typedSyn.(*HandshakeSyn).Nonce
			synAck := &HandshakeSynAck{
				Nonce: nonce,
			}
			if err := Write(synAck, conn, blocks.SignWith(signer)); err != nil {
				log.DefaultLogger.Warn("sending for syn-ack failed", zap.Error(err))
				continue
			}

			typedAck, err := Read(conn)
			if err != nil {
				log.DefaultLogger.Warn("waiting for ack failed", zap.Error(err))
				continue
			}

			if typedAck.(*HandshakeAck).Nonce != nonce {
				log.DefaultLogger.Warn("validating syn to ack nonce failed")
				continue
			}

			conn.RemoteID = typedAck.GetSignature().Key.Thumbprint()
			cconn <- conn
		}
	}()

	return cconn, nil
}

func Write(v blocks.Typed, conn *Connection, opts ...blocks.PackOption) error {
	conn.Conn.SetWriteDeadline(time.Now().Add(time.Second))
	p, err := blocks.Pack(v, opts...)
	if err != nil {
		return err
	}
	b, err := blocks.Encode(p)
	if err != nil {
		return err
	}
	if _, err := conn.Conn.Write(b); err != nil {
		return err
	}
	SendBlockEvent(
		"incoming",
		v.GetType(),
		len(b),
	)
	if os.Getenv("DEBUG_BLOCKS") == "true" {
		b, _ := json.MarshalIndent(p, "", "  ")
		log.DefaultLogger.Info(string(b), zap.String("remoteID", conn.RemoteID), zap.String("direction", "outgoing"))
	}
	return nil
}

func Read(conn *Connection) (blocks.Typed, error) {
	pDecoder := ucodec.NewDecoder(conn.Conn, codec.CborHandler())
	p := &blocks.Block{}
	if err := pDecoder.Decode(&p); err != nil {
		return nil, err
	}
	v, err := blocks.Unpack(p)
	if err != nil {
		return nil, err
	}
	SendBlockEvent(
		"incoming",
		v.GetType(),
		pDecoder.NumBytesRead(),
	)
	if os.Getenv("DEBUG_BLOCKS") != "true" {
		m, _ := blocks.Pack(v)
		b, _ := json.MarshalIndent(m, "", "  ")
		log.DefaultLogger.Info(string(b), zap.String("remoteID", conn.RemoteID), zap.String("direction", "incoming"))
	}
	return v, nil
}
