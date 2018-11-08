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

	"github.com/davecgh/go-spew/spew"

	igd "github.com/emersion/go-upnp-igd"
	ucodec "github.com/ugorji/go/codec"
	"go.uber.org/zap"

	"nimona.io/go/codec"
	"nimona.io/go/log"
	"nimona.io/go/peers"
	"nimona.io/go/primitives"
)

// Networker interface for mocking Network
type Networker interface {
	Dial(ctx context.Context, address string) (*Connection, error)
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
func (n *Network) Dial(ctx context.Context, address string) (*Connection, error) {
	logger := log.Logger(ctx)

	addressType := strings.Split(address, ":")[0]
	switch addressType {
	case "peer":
		logger.Debug("dialing peer", zap.String("peer", address))
		peerID := strings.Replace(address, "peer:", "", 1)
		peerInfo, err := n.addressBook.GetPeerInfo(peerID)
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

	case "tcp":
		addr := strings.Replace(address, "tcp:", "", 1)
		dialer := net.Dialer{Timeout: time.Second}
		logger.Debug("dialing", zap.String("address", addr))
		tcpConn, err := dialer.DialContext(ctx, "tcp", addr)
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

		signer := n.addressBook.GetLocalPeerKey()
		nonce := RandStringBytesMaskImprSrc(8)
		syn := &HandshakeSyn{
			Nonce:    nonce,
			PeerInfo: n.addressBook.GetLocalPeerInfo(),
		}
		synBlock := syn.Block()
		if err := primitives.Sign(synBlock, signer); err != nil {
			return nil, err
		}

		if err := Write(synBlock, conn); err != nil {
			return nil, err
		}

		blockSynAck, err := Read(conn)
		if err != nil {
			return nil, err
		}

		synAck := &HandshakeSynAck{}
		synAck.FromBlock(blockSynAck)
		if synAck.Nonce != nonce {
			return nil, errors.New("invalid handhshake.syn-ack")
		}

		// store who is on the other side - peer id
		conn.RemoteID = synAck.Signature.Key.Thumbprint()
		if err := n.addressBook.PutPeerInfo(synAck.PeerInfo); err != nil {
			log.DefaultLogger.Panic("could not add remote peer", zap.Error(err))
		}

		ack := &HandshakeAck{
			Nonce: nonce,
		}
		ackBlock := ack.Block()
		if err := primitives.Sign(ackBlock, signer); err != nil {
			return nil, err
		}
		if err := Write(ackBlock, conn); err != nil {
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
func (n *Network) Listen(ctx context.Context, address string) (chan *Connection, error) {
	logger := log.Logger(ctx).Named("network")
	tcpListener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	port := tcpListener.Addr().(*net.TCPAddr).Port
	logger.Info("Listening and service nimona", zap.Int("port", port))
	addresses := GetAddresses(tcpListener)
	devices := make(chan igd.Device, 10)

	if n.addressBook.LocalHostname != "" {
		addresses = append(addresses, fmtAddress(n.addressBook.LocalHostname, port))
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
	n.addressBook.AddLocalPeerAddress(addresses...)

	cconn := make(chan *Connection, 10)
	go func() {
		signer := n.addressBook.GetLocalPeerKey()
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

			blockSyn, err := Read(conn)
			if err != nil {
				log.DefaultLogger.Warn("waiting for syn failed", zap.Error(err))
				// TODO close conn?
				continue
			}

			// TODO check type

			syn := &HandshakeSyn{}
			syn.FromBlock(blockSyn)
			nonce := syn.Nonce

			// store the peer on the other side
			if err := n.addressBook.PutPeerInfo(syn.PeerInfo); err != nil {
				log.DefaultLogger.Panic("could not add remote peer", zap.Error(err))
			}

			synAck := &HandshakeSynAck{
				Nonce:    nonce,
				PeerInfo: n.addressBook.GetLocalPeerInfo(),
			}
			synAckBlock := synAck.Block()
			if err := primitives.Sign(synAckBlock, signer); err != nil {
				log.DefaultLogger.Warn("could not sigh for syn ack block", zap.Error(err))
				// TODO close conn?
				continue
			}
			if err := Write(synAckBlock, conn); err != nil {
				log.DefaultLogger.Warn("sending for syn-ack failed", zap.Error(err))
				// TODO close conn?
				continue
			}

			blockAck, err := Read(conn)
			if err != nil {
				log.DefaultLogger.Warn("waiting for ack failed", zap.Error(err))
				// TODO close conn?
				continue
			}

			ack := &HandshakeAck{}
			ack.FromBlock(blockAck)
			if ack.Nonce != nonce {
				log.DefaultLogger.Warn("validating syn to ack nonce failed")
				// TODO close conn?
				continue
			}

			conn.RemoteID = ack.Signature.Key.Thumbprint()
			cconn <- conn
		}
	}()

	return cconn, nil
}

func Write(p *primitives.Block, conn *Connection) error {
	conn.Conn.SetWriteDeadline(time.Now().Add(time.Second))
	if p == nil {
		log.DefaultLogger.Error("block for fw cannot be nil")
		return errors.New("missing block")
	}

	b, err := primitives.Marshal(p)
	if err != nil {
		return err
	}

	if _, err := conn.Conn.Write(b); err != nil {
		return err
	}

	SendBlockEvent(
		"incoming",
		p.Type,
		len(b),
	)

	if os.Getenv("DEBUG_BLOCKS") == "true" {
		b, _ := json.MarshalIndent(primitives.BlockToMap(p), "", "  ")
		log.DefaultLogger.Info(string(b), zap.String("remoteID", conn.RemoteID), zap.String("direction", "outgoing"))
	}
	return nil
}

func Read(conn *Connection) (*primitives.Block, error) {
	logger := log.DefaultLogger

	pDecoder := ucodec.NewDecoder(conn.Conn, codec.CborHandler())
	b := &primitives.Block{}
	if err := pDecoder.Decode(&b); err != nil {
		return nil, err
	}

	defer func() {
		if r := recover(); r != nil {
			spew.Dump(b)
			logger.Error("Recovered while processing", zap.Any("r", r))
		}
	}()

	d, err := b.Digest()
	if err != nil {
		return nil, err
	}

	if b.Signature != nil {
		if err := primitives.Verify(b.Signature, d); err != nil {
			return nil, err
		}
	} else {
		fmt.Println("--------------------------------------------------------")
		fmt.Println("----- BLOCK NOT SIGNED ---------------------------------")
		fmt.Println("--------------------------------------------------------")
		fmt.Println("-----", b.Type)
		fmt.Println("-----", b.Payload)
		fmt.Println("--------------------------------------------------------")
	}

	SendBlockEvent(
		"incoming",
		b.Type,
		pDecoder.NumBytesRead(),
	)
	if os.Getenv("DEBUG_BLOCKS") == "true" {
		bs, _ := json.MarshalIndent(primitives.BlockToMap(b), "", "  ")
		logger.Info(string(bs), zap.String("remoteID", conn.RemoteID), zap.String("direction", "incoming"))
	}
	return b, nil
}
