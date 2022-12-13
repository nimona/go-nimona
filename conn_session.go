package nimona

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"net"

	"github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
	"github.com/oasisprotocol/curve25519-voi/primitives/x25519"
	"golang.org/x/crypto/blake2b"
)

// Session wraps a net.Conn and provides methods for passing around encrypted
// chunks of data.
// Each chunk is prefixed with a uvarint that specifies the length of the
// chunk.
type Session struct {
	conn  net.Conn
	suite cipher.AEAD
	// available after handshake
	remotePublicKey ed25519.PublicKey
	remoteNodeAddr  NodeAddr
}

// NewSession returns a new Session that wraps the given net.Conn.
func NewSession(conn net.Conn) *Session {
	s := &Session{
		conn: conn,
	}
	return s
}

// DoServer performs the server-side of the handshake process.
// It uses the provided static key, writes it to the conn, reads the client's
// static key, and then generates a shared secret via x25519 and blake2b.
// The shared secret is then used to derive the cipher suite for encrypting
// and decrypting data.
func (s *Session) DoServer(
	serverPublicKey ed25519.PublicKey,
	serverPrivateKey ed25519.PrivateKey,
) error {
	// write the server's ephemeral keys to the connection
	_, err := s.conn.Write(serverPublicKey)
	if err != nil {
		return err
	}

	// read the client's ephemeral keys from the connection
	var clientEphemeral [32]byte
	_, err = s.conn.Read(clientEphemeral[:])
	if err != nil {
		return err
	}

	// generate the shared secret using the server's and client's ephemeral keys
	shared, err := s.x25519(serverPrivateKey, clientEphemeral[:])
	if err != nil {
		return err
	}

	// use the shared secret and blake2b to generate the key for the cipher suite
	key := blake2b.Sum256(shared[:])

	// derive the cipher suite using the generated key
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return err
	}
	s.suite, err = cipher.NewGCM(block)
	if err != nil {
		return err
	}

	// store the remote node key and address
	s.remotePublicKey = ed25519.PublicKey(clientEphemeral[:])
	s.remoteNodeAddr = NewNodeAddrWithKey(
		s.conn.RemoteAddr().Network(),
		s.conn.RemoteAddr().String(),
		s.remotePublicKey,
	)

	return nil
}

func (s *Session) DoClient(
	clientPublicKey ed25519.PublicKey,
	clientPrivateKey ed25519.PrivateKey,
) error {
	// Read the server's ephemeral keys from the connection
	var serverEphemeral [32]byte
	_, err := s.conn.Read(serverEphemeral[:])
	if err != nil {
		return err
	}

	// Write the client's ephemeral keys to the connection
	_, err = s.conn.Write(clientPublicKey[:])
	if err != nil {
		return err
	}

	// Generate the shared secret using the server's and client's ephemeral keys
	shared, err := s.x25519(clientPrivateKey, serverEphemeral[:])
	if err != nil {
		return err
	}

	// Use the shared secret and blake2b to generate the key for the cipher suite
	key := blake2b.Sum256(shared[:])

	// Derive the cipher suite using the generated key
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return err
	}
	s.suite, err = cipher.NewGCM(block)
	if err != nil {
		return err
	}

	// store the remote node key and address
	s.remotePublicKey = ed25519.PublicKey(serverEphemeral[:])
	s.remoteNodeAddr = NewNodeAddrWithKey(
		s.conn.RemoteAddr().Network(),
		s.conn.RemoteAddr().String(),
		s.remotePublicKey,
	)

	return nil
}

// Read decrypts a packet of data read from the connection.
// The packet is prefixed with a uvarint that designates the length of the packet.
func (s *Session) Read() ([]byte, error) {
	// TODO Update to use varint
	// Read the length of the packet from the connection
	var length uint64
	err := binary.Read(s.conn, binary.BigEndian, &length)
	if err != nil {
		return nil, err
	}

	// Read the encrypted packet from the connection
	packet := make([]byte, length)
	_, err = s.conn.Read(packet)
	if err != nil {
		return nil, err
	}

	// Decrypt the packet using the cipher suite
	nonce := make([]byte, s.suite.NonceSize())
	_, err = s.conn.Read(nonce)
	if err != nil {
		return nil, err
	}
	data, err := s.suite.Open(nil, nonce, packet, nil)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Write encrypts a packet of data and writes it to the connection.
// The packet is prefixed with a uvarint that designates the length of the packet.
func (s *Session) Write(data []byte) (int, error) {
	// Encrypt the data using the cipher suite
	nonce := make([]byte, s.suite.NonceSize())
	_, err := rand.Read(nonce)
	if err != nil {
		return 0, err
	}
	packet := s.suite.Seal(nil, nonce, data, nil)

	// Write the length of the packet to the connection
	err = binary.Write(s.conn, binary.BigEndian, uint64(len(packet)))
	if err != nil {
		return 0, err
	}

	// Write the encrypted packet to the connection
	_, err = s.conn.Write(packet)
	if err != nil {
		return 0, err
	}

	// Write the nonce to the connection
	n, err := s.conn.Write(nonce)
	if err != nil {
		return n, err
	}

	return n, nil
}

// x25519 returns the result of the scalar multiplication (scalar * point),
// according to RFC 7748, Section 5. scalar, point and the return value are
// slices of 32 bytes.
func (s *Session) x25519(
	privateKey ed25519.PrivateKey,
	publicKey ed25519.PublicKey,
) ([]byte, error) {
	publicKeyX, ok := x25519.EdPublicKeyToX25519(publicKey)
	if !ok {
		return nil, errors.New("unable to derive ed25519 key to x25519 key")
	}

	privateKeyX := x25519.EdPrivateKeyToX25519(privateKey)

	shared, err := x25519.X25519(privateKeyX, publicKeyX[:])
	if err != nil {
		return nil, err
	}

	return shared, nil
}

func (s *Session) NodeAddr() NodeAddr {
	return s.remoteNodeAddr
}

func (s *Session) Close() error {
	return s.conn.Close()
}
