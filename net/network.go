package net

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"

	igd "github.com/emersion/go-upnp-igd"
	ucodec "github.com/ugorji/go/codec"
	"go.uber.org/zap"

	"nimona.io/go/crypto"
	"nimona.io/go/encoding"
	"nimona.io/go/log"
	"nimona.io/go/peers"
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
		signedSyn, err := crypto.Sign(syn.ToObject(), signer)
		if err != nil {
			return nil, err
		}

		if err := Write(signedSyn, conn); err != nil {
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
		conn.RemoteID = synAck.RawObject.GetSignerKey().HashBase58()
		if err := n.addressBook.PutPeerInfo(synAck.PeerInfo); err != nil {
			log.DefaultLogger.Panic("could not add remote peer", zap.Error(err))
		}

		ack := &HandshakeAck{
			Nonce: nonce,
		}
		signedAck, err := crypto.Sign(ack.ToObject(), signer)
		if err != nil {
			return nil, err
		}

		if err := Write(signedAck, conn); err != nil {
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

			synObj, err := Read(conn)
			if err != nil {
				log.DefaultLogger.Warn("waiting for syn failed", zap.Error(err))
				// TODO close conn?
				continue
			}

			syn := &HandshakeSyn{}
			if err := syn.FromObject(synObj); err != nil {
				// TODO close conn?
				continue
			}

			// store the peer on the other side
			if err := n.addressBook.PutPeerInfo(syn.PeerInfo); err != nil {
				log.DefaultLogger.Panic("could not add remote peer", zap.Error(err))
			}

			synAck := &HandshakeSynAck{
				Nonce:    syn.Nonce,
				PeerInfo: n.addressBook.GetLocalPeerInfo(),
			}
			singedSynAck, err := crypto.Sign(synAck.ToObject(), signer)
			if err != nil {
				log.DefaultLogger.Warn("could not sigh for syn ack block", zap.Error(err))
				// TODO close conn?
				continue
			}
			if err := Write(singedSynAck, conn); err != nil {
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
				continue
			}

			if ack.Nonce != syn.Nonce {
				log.DefaultLogger.Warn("validating syn to ack nonce failed")
				// TODO close conn?
				continue
			}

			conn.RemoteID = ack.RawObject.GetSignerKey().HashBase58()
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
		b, _ := json.MarshalIndent(o.Map(), "", "  ")
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
			spew.Dump(r)
			logger.Error("Recovered while processing", zap.Any("r", r))
		}
	}()

	o, err := encoding.NewObjectFromBytes([]byte(*r))
	if err != nil {
		return nil, err
	}

	// TODO(geoah) fix sig
	// if b.Signature != nil {
	// 	if err := crypto.Verify(b.Signature, d); err != nil {
	// 		return nil, err
	// 	}
	// } else {
	// 	fmt.Println("--------------------------------------------------------")
	// 	fmt.Println("----- BLOCK NOT SIGNED ---------------------------------")
	// 	fmt.Println("--------------------------------------------------------")
	// 	fmt.Println("-----", b.Type)
	// 	fmt.Println("-----", b.Payload)
	// 	fmt.Println("--------------------------------------------------------")
	// }

	SendBlockEvent(
		"incoming",
		o.GetType(),
		pDecoder.NumBytesRead(),
	)
	if os.Getenv("DEBUG_BLOCKS") == "true" {
		b, _ := json.MarshalIndent(o.Map(), "", "  ")
		logger.Info(string(b), zap.String("remoteID", conn.RemoteID), zap.String("direction", "incoming"))
	}
	return o, nil
}
