package nimona

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"net"

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
	remotePublicKey PublicKey
	remotePeerAddr  PeerAddr
	rpc             *RPC
	codec           Codec
	// for testing
	skipRPC bool
}

// NewSession returns a new Session that wraps the given net.Conn.
func NewSession(conn net.Conn) *Session {
	s := &Session{
		codec: &CodecCBOR{},
		conn:  conn,
	}
	return s
}

// DoServer performs the server-side of the handshake process.
// It uses the provided static key, writes it to the conn, reads the client's
// static key, and then generates a shared secret via x25519 and blake2b.
// The shared secret is then used to derive the cipher suite for encrypting
// and decrypting data.
func (s *Session) DoServer(
	serverPublicKey PublicKey,
	serverPrivateKey PrivateKey,
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
	key := blake2b.Sum256(shared)

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
	s.remotePublicKey = PublicKey(clientEphemeral[:])
	s.remotePeerAddr = PeerAddr{
		Network:   s.conn.RemoteAddr().Network(),
		Address:   s.conn.RemoteAddr().String(),
		PublicKey: s.remotePublicKey,
	}

	// create the rpc
	if !s.skipRPC {
		s.rpc = NewRPC(s)
	}

	return nil
}

func (s *Session) DoClient(
	clientPublicKey PublicKey,
	clientPrivateKey PrivateKey,
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
	key := blake2b.Sum256(shared)

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
	s.remotePublicKey = PublicKey(serverEphemeral[:])
	s.remotePeerAddr = PeerAddr{
		Network:   s.conn.RemoteAddr().Network(),
		Address:   s.conn.RemoteAddr().String(),
		PublicKey: s.remotePublicKey,
	}

	// create the rpc
	if !s.skipRPC {
		s.rpc = NewRPC(s)
	}

	return nil
}

// read decrypts a packet of data read from the connection.
// The packet is prefixed with a uvarint that designates the length of the packet.
func (s *Session) read() ([]byte, error) {
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

// write encrypts a packet of data and writes it to the connection.
// The packet is prefixed with a uvarint that designates the length of the packet.
func (s *Session) write(data []byte) (int, error) {
	// Encrypt the data using the cipher suite
	nonce := make([]byte, s.suite.NonceSize())
	_, err := rand.Read(nonce)
	if err != nil {
		return 0, err
	}
	packet := s.suite.Seal(nil, nonce, data, nil)

	// Write the length of the packet to the connection
	// TODO Update to use varint
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

func (s *Session) Request(
	ctx context.Context,
	req Cborer,
) (*Response, error) {
	// Encode the request
	b, err := s.codec.Encode(req)
	if err != nil {
		return nil, fmt.Errorf("unable to encode request: %w", err)
	}

	// Send the request
	resBytes, err := s.rpc.Request(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("unable to send request: %w", err)
	}

	// Decode the response
	codec := s.codec
	res := &Response{}
	err = codec.Decode(resBytes, res)
	if err != nil {
		return nil, fmt.Errorf("unable to decode response: %w", err)
	}

	res.Body = resBytes
	res.Codec = codec
	res.Decode = messageDecoder(codec, resBytes)

	return res, nil
}

func messageDecoder(c Codec, b []byte) func(Cborer) error {
	return func(v Cborer) error {
		return c.Decode(b, v)
	}
}

func (s *Session) Read() (*Request, error) {
	// Read the message from the connection
	req, cb, err := s.rpc.Read()
	if err != nil {
		return nil, fmt.Errorf("unable to read message: %w", err)
	}

	codec := s.codec
	msgReq := &Request{}
	err = codec.Decode(req, msgReq)
	if err != nil {
		return nil, fmt.Errorf("unable to decode message: %w", err)
	}

	msgReq.Body = req
	msgReq.Codec = codec
	msgReq.Decode = messageDecoder(codec, req)
	msgReq.Respond = func(res Cborer) error {
		// Encode the response
		b, err := codec.Encode(res)
		if err != nil {
			return fmt.Errorf("unable to encode response: %w", err)
		}

		// Send the response
		err = cb(b)
		if err != nil {
			return fmt.Errorf("unable to send response: %w", err)
		}

		return nil
	}

	return msgReq, nil
}

// x25519 returns the result of the scalar multiplication (scalar * point),
// according to RFC 7748, Section 5. scalar, point and the return value are
// slices of 32 bytes.
func (s *Session) x25519(
	privateKey PrivateKey,
	publicKey PublicKey,
) ([]byte, error) {
	publicKeyX, err := publicKey.X25519()
	if err != nil {
		return nil, errors.New("unable to derive public ed25519 key to x25519 key")
	}

	privateKeyX, err := privateKey.X25519()
	if err != nil {
		return nil, errors.New("unable to derive private ed25519 key to x25519 key")
	}

	shared, err := x25519.X25519(privateKeyX, publicKeyX)
	if err != nil {
		return nil, err
	}

	return shared, nil
}

func (s *Session) PeerAddr() PeerAddr {
	return s.remotePeerAddr
}

func (s *Session) PublicKey() PublicKey {
	return s.remotePublicKey
}

// Close both the connection and the rpc.
func (s *Session) Close() error {
	if s.rpc != nil {
		s.rpc.Close()
	}
	return s.conn.Close()
}
