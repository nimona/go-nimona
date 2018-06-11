package blx

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"

	"github.com/nimona/go-nimona/wire"
	"github.com/sirupsen/logrus"
)

const (
	wireExtention = "blx"
)

var (
	ErrMissingPeer    = errors.New("Peer ID missing")
	ErrInvalidBlock   = errors.New("Invalid Block")
	ErrInvalidRequest = errors.New("Invalid request")
)

// BlockExchange enables the transfer and storage of
// blocks between peers.
type BlockExchange interface {
	Get(key string, recipient string) (*Block, error)
	Send(recipient string, data []byte,
		meta map[string][]byte) (string, int, error)
	GetLocalBlocks() ([]string, error)
	Subscribe(fn subscriptionCb) (string, error)
	Unsubscribe(id string)
}

type subscriptionCb func(string)

type blockExchange struct {
	wire          wire.Wire
	storage       Storage
	getRequests   sync.Map
	subscriptions sync.Map
}

// NewBlockExchange get Wire and a Storage as parameters and returns a new
// block exchange protocol.
func NewBlockExchange(wr wire.Wire, pr Storage) (BlockExchange, error) {
	blx := &blockExchange{
		wire:        wr,
		storage:     pr,
		getRequests: sync.Map{},
	}

	wr.HandleExtensionEvents(wireExtention, blx.handleMessage)

	return blx, nil
}

func (blx *blockExchange) handleMessage(message *wire.Message) error {
	switch message.PayloadType {
	case PayloadTypeTransferBlock:
		err := blx.handleTransferBlock(message)
		if err != nil {
			return err
		}
	case PayloadTypeRequestBlock:
		err := blx.handleRequestBlock(message)
		if err != nil {
			return err
		}
	default:
		logrus.WithField("message.PayloadType", message.PayloadType).
			Warn("Payload type not known")
		return nil
	}
	return nil
}

func (blx *blockExchange) handleTransferBlock(message *wire.Message) error {
	payload := &payloadTransferBlock{}
	if err := message.DecodePayload(payload); err != nil {
		return err
	}

	if payload.Block != nil {
		err := blx.storage.Store(payload.Block.Key, payload.Block)
		blx.publish(payload.Block.Key)

		if err != nil {
			return err
		}
	}

	// Check if nonce exists in local addressBook
	value, ok := blx.getRequests.Load(payload.Nonce)
	if !ok {
		return nil
	}

	req, ok := value.(*payloadTransferRequestBlock)
	if !ok {
		return ErrInvalidRequest
	}

	req.response <- payload

	return nil
}

func (blx *blockExchange) handleRequestBlock(message *wire.Message) error {
	payload := &payloadTransferRequestBlock{}
	if err := message.DecodePayload(payload); err != nil {
		return err
	}

	status := StatusOK

	// TODO handle block request
	pblock, err := blx.storage.Get(payload.Key)
	if err != nil {
		pblock = nil
		status = StatusNotFound
	}

	resp := payloadTransferBlock{
		Nonce:  payload.Nonce,
		Block:  pblock,
		Status: status,
	}

	ctx := context.Background()
	err = blx.wire.Send(ctx, wireExtention, PayloadTypeTransferBlock, resp,
		[]string{payload.RequestingPeerID})
	if err != nil {
		return err
	}

	return nil
}

func (blx *blockExchange) Get(key string, recipient string) (
	*Block, error) {
	// TODO remember to remove nonce
	nonce := wire.RandStringBytesMaskImprSrc(8)

	req := &payloadTransferRequestBlock{
		RequestingPeerID: "MISSING-REQUEST-PEER-ID", // TODO Missing requesting peer id
		Nonce:            nonce,
		Key:              key,
		response:         make(chan interface{}, 1),
	}

	// Check local storage for block
	block, err := blx.storage.Get(key)
	if err == nil {
		return block, nil
	}
	if err != ErrNotFound {
		return nil, err
	}
	if recipient == "" {
		return nil, ErrMissingPeer
	}

	blx.getRequests.Store(req.Nonce, req)

	// Request block
	ctx := context.Background()
	err = blx.wire.Send(ctx, wireExtention, PayloadTypeRequestBlock,
		req, []string{recipient})
	if err != nil {
		return nil, err
	}

	for {
		select {
		case response := <-req.response:
			resp, ok := response.(*payloadTransferBlock)
			if !ok {
				return nil, ErrInvalidBlock
			}
			if resp.Status != StatusOK {
				return nil, ErrNotFound
			}

			return resp.Block, nil

		case <-ctx.Done():
			return nil, ErrNotFound
		}
	}
	return nil, ErrNotFound
}

func (blx *blockExchange) Send(recipient string, data []byte,
	values map[string][]byte) (string, int, error) {

	hs := blx.hash(data)

	block := Block{
		Key:  hs,
		Meta: Meta{Values: values},
		Data: data,
	}

	nonce := wire.RandStringBytesMaskImprSrc(8)

	resp := payloadTransferBlock{
		Nonce: nonce,
		Block: &block,
	}

	blx.storage.Store(block.Key, &block)
	blx.publish(block.Key)

	ctx := context.Background()
	err := blx.wire.Send(ctx, wireExtention, PayloadTypeTransferBlock, resp,
		[]string{recipient})
	if err != nil {
		return "", 0, err
	}

	return hs, len(data), nil
}

func (blx *blockExchange) GetLocalBlocks() ([]string, error) {
	return blx.storage.List()
}

// Subscribe registers a function to be called when an event happens
// returns the id for the registration
func (blx *blockExchange) Subscribe(fn subscriptionCb) (string, error) {
	id := wire.RandStringBytesMaskImprSrc(8)
	blx.subscriptions.Store(id, fn)
	return id, nil
}

// Unsubscribe unregisters the id
func (blx *blockExchange) Unsubscribe(id string) {
	blx.subscriptions.Delete(id)
}
func (blx *blockExchange) publish(newKey string) {
	blx.subscriptions.Range(func(k, v interface{}) bool {
		f, ok := v.(subscriptionCb)
		if !ok {
			return false
		}

		f(newKey)
		return true
	})
}

func (b *blockExchange) hash(data []byte) string {
	h := sha256.New()
	h.Write(data)

	return string(hex.EncodeToString(h.Sum(nil)))
}
