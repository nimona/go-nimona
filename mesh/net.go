package mesh

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/emersion/go-upnp-igd"
)

var (
	ErrAllAddressesFailed = errors.New("all addresses failed to dial")
)

type Net struct {
	registry     Registry
	accepted     chan net.Conn
	reusableLock sync.RWMutex
	reusable     map[string]*reusableConn
	handlers     map[string]Handler
	close        chan bool
}

func New(registry Registry) *Net {
	n := &Net{
		registry: registry,
		close:    make(chan bool),
		accepted: make(chan net.Conn),
		reusable: map[string]*reusableConn{},
		handlers: map[string]Handler{},
	}
	n.RegisterHandler("id", &ID{registry})
	n.RegisterHandler("yamux", &Yamux{})
	n.RegisterHandler("relay", &Relay{n})
	return n
}

func (n *Net) RegisterHandler(protocol string, handler Handler) error {
	n.handlers[protocol] = handler
	return nil
}

func (n *Net) Dial(ctx context.Context, peerID string, commands ...string) (net.Conn, error) {
	// fmt.Println("Dial()ing ", peerID)
	peerInfo, err := n.registry.GetPeerInfo(peerID)
	if err != nil {
		return nil, err
	}
	var conn net.Conn

	n.reusableLock.RLock()
	if reusableConn, ok := n.reusable[peerID]; ok {
		newConn, err := reusableConn.NewConn()
		if err != nil {
			// TODO remove reusable conn
		} else {
			// fmt.Println("dial reusing conn")
			conn = newConn
		}
	}
	n.reusableLock.RUnlock()
	if conn == nil {
		for _, addr := range peerInfo.Addresses {
			if strings.HasPrefix(addr, "relay:") {
				relayID := strings.Replace(addr, "relay:", "", 1)
				// fmt.Println("dialing the relay")
				relayConn, err := n.Dial(ctx, relayID)
				if err != nil {
					// TODO blacklist address for a bit
					// TODO hold error maybe?
					// return nil, err
					continue
				}
				localAddress := peerAddress{
					network: "tcp",
					peerID:  n.registry.GetLocalPeerInfo().ID,
				}
				remoteAddress := peerAddress{
					network: "tcp",
					peerID:  peerID,
				}
				conn = NewAddressableConn(relayConn, localAddress, remoteAddress)
				commands = append([]string{"relay"}, commands...)
			} else {
				addr = strings.Replace(addr, "tcp:", "", 1)
				// fmt.Println("dial dialing new conn to", addr)
				dialer := net.Dialer{Timeout: time.Second * 5}
				newConn, err := dialer.DialContext(ctx, "tcp", addr)
				if err != nil {
					// TODO blacklist address for a bit
					// TODO hold error maybe?
					// return nil, err
					continue
				}
				localAddress := peerAddress{
					network: "tcp",
					peerID:  n.registry.GetLocalPeerInfo().ID,
				}
				remoteAddress := peerAddress{
					network: "tcp",
					peerID:  peerID,
				}
				conn = NewAddressableConn(newConn, localAddress, remoteAddress)
				commands = append([]string{"id", "yamux"}, commands...)
			}
			break
		}
	}
	if conn == nil {
		return nil, ErrAllAddressesFailed
	}

	finalConn, err := n.Select(conn, commands...)
	if err != nil {
		if err := conn.Close(); err != nil {
			fmt.Println("could not close connection after failure to select")
		}
		fmt.Println("error selecting", err)
		return nil, err
	}

	return finalConn, nil
}

func (n *Net) Select(conn net.Conn, commands ...string) (net.Conn, error) {
	for _, command := range commands {
		if err := WriteToken(conn, []byte(command)); err != nil {
			return nil, err
		}
		token, err := ReadToken(conn)
		if err != nil {
			return nil, err
		}
		// fmt.Printf("Dialer got token %s for command %s\n", string(token), command)
		if string(token) != "ok" {
			return nil, errors.New("unexpected token response")
		}
		handler, ok := n.handlers[command]
		if !ok {
			return nil, errors.New("no such handler")
		}
		newConn, err := handler.Initiate(conn)
		if err != nil {
			return nil, err
		}
		// TODO maybe use a switch
		if reusableConn, ok := newConn.(*reusableConn); ok {
			go reusableConn.Accepted(n.accepted)
			// TODO lock
			// fmt.Println("client storing reusable")
			n.reusableLock.Lock()
			n.reusable[conn.RemoteAddr().String()] = reusableConn
			n.reusableLock.Unlock()
		}
		conn = newConn
	}
	return conn, nil
}

// TODO do we need to return a listener?
func (n *Net) Listen(addr string) (Listener, string, error) {
	tcpListener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, "", err
	}
	port := tcpListener.Addr().(*net.TCPAddr).Port

	lock := sync.Mutex{}
	addresses := GetAddresses(tcpListener)

	devices := make(chan igd.Device)
	go func() {
		for device := range devices {
			upnp := true
			upnpFlag := os.Getenv("UPNP")
			if upnpFlag != "" {
				upnp, _ = strconv.ParseBool(upnpFlag)
			}
			if !upnp {
				continue
			}
			externalAddress, err := device.GetExternalIPAddress()
			if err != nil {
				fmt.Println("could not get external ip")
				continue
			}
			desc := "nimona"
			ttl := time.Hour * 24 * 365
			if _, err := device.AddPortMapping(igd.TCP, port, port, desc, ttl); err != nil {
				fmt.Println("could not add port mapping", err)
			} else {
				lock.Lock()
				addresses = append(addresses, fmt.Sprintf("tcp:%s:%d", externalAddress.String(), port))
				lock.Unlock()
			}
		}
	}()

	if err := igd.Discover(devices, 5*time.Second); err != nil {
		log.Println("could not discover devices")
	}

	lock.Lock()
	addresses = append(addresses, "relay:andromeda.nimona.io")
	lock.Unlock()

	// TODO replace this with something like n.registry.GetLocalPeerInfo().SetAddresses()
	n.registry.PutLocalPeerInfo(&PeerInfo{
		ID:        n.registry.GetLocalPeerInfo().ID,
		Addresses: addresses,
		PublicKey: n.registry.GetLocalPeerInfo().PublicKey,
	})

	go func() {
		for {
			conn := <-n.accepted
			go n.HandleSelection(conn)
		}
	}()

	closed := false

	go func() {
		closed = true
		<-n.close
		// fmt.Println("Closing")
		tcpListener.Close()
	}()

	go func() {
		for {
			conn, err := tcpListener.Accept()
			if err != nil {
				if closed {
					return
				}
				fmt.Println("Error accepting: ", err.Error())
				// TODO check conn is still alive and return
				return
			}
			localAddress := peerAddress{
				network: "tcp",
				peerID:  n.registry.GetLocalPeerInfo().ID,
			}
			remoteAddress := peerAddress{
				network: "tcp",
				peerID:  conn.RemoteAddr().String(),
			}
			conn = NewAddressableConn(conn, localAddress, remoteAddress)
			n.accepted <- conn
		}
	}()

	return nil, tcpListener.Addr().String(), nil
}

func (n *Net) Close() error {
	n.close <- true
	return nil
}

func (n *Net) HandleSelection(conn net.Conn) (net.Conn, error) {
	for {
		token, err := ReadToken(conn)
		if err != nil {
			return nil, err
		}
		// fmt.Println("selection handler got token", string(token))
		handler, ok := n.handlers[string(token)]
		if !ok {
			if err := WriteToken(conn, []byte("error")); err != nil {
				return nil, err
			}
			return nil, errors.New("no such handler")
		}
		if err := WriteToken(conn, []byte("ok")); err != nil {
			return nil, err
		}
		newConn, err := handler.Handle(conn)
		if err != nil {
			return nil, err
		}
		// TODO maybe use a switch
		if reusableConn, ok := newConn.(*reusableConn); ok {
			go reusableConn.Accepted(n.accepted)
			// TODO lock
			// TODO check if remote addr is indeed a peer id
			// fmt.Println("server storing reusable")
			n.reusableLock.Lock()
			n.reusable[conn.RemoteAddr().String()] = reusableConn
			n.reusableLock.Unlock()
		}
		conn = newConn
	}
}
