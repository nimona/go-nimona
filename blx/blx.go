package blx

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/nimona/go-nimona/wire"
	uuid "github.com/satori/go.uuid"
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

type blockExchange struct {
	wire        wire.Wire
	storage     Storage
	getRequests sync.Map
}

func NewBlockExchange(wr wire.Wire) (*blockExchange, error) {
	blx := &blockExchange{
		wire:        wr,
		storage:     newMemoryStore(),
		getRequests: sync.Map{},
	}

	wr.HandleExtensionEvents(wireExtention, blx.handleMessage)

	return blx, nil
}

func (blx *blockExchange) handleMessage(message *wire.Message) error {
	switch message.PayloadType {
	case PayloadTypeTransferBlock:
		blx.handleTransferBlock(message)
	case PayloadTypeRequestBlock:
		blx.handleRequestBlock(message)
	default:
		logrus.WithField("message.PayloadType", message.PayloadType).
			Warn("Payload type not known")
		return nil
	}
	return nil
}

func (blx *blockExchange) handleTransferBlock(message *wire.Message) error {
	fmt.Println("Handle transfer")
	payload := &payloadTransferBlock{}
	if err := message.DecodePayload(payload); err != nil {
		fmt.Println(err)
		return err
	}

	if payload.Block != nil {
		err := blx.storage.Store(payload.Block.Key, payload.Block)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	// Check if nonce exists in local registry
	value, ok := blx.getRequests.Load(payload.Nonce)
	if ok == false {
		return nil
	}

	fmt.Println(value)

	req, ok := value.(payloadTransferRequestBlock)
	if ok == false {
		return ErrInvalidRequest
	}

	// TODO is this shit blocking?
	req.response <- payload

	return nil
}

func (blx *blockExchange) handleRequestBlock(message *wire.Message) error {
	fmt.Println("Handle request")
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
		[]string{message.From})
	if err != nil {
		return err
	}

	return nil
}

func (blx *blockExchange) Get(key string, recipient string) (
	*Block, error) {
	nonce := uuid.NewV4().String()

	req := &payloadTransferRequestBlock{
		Nonce:    nonce,
		Key:      key,
		response: make(chan interface{}, 1),
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

	fmt.Println("Request sent")
	// TODO not receiving in the channel
	select {
	case <-req.response:
		fmt.Println("Received")
		// resp, ok := response.(payloadTransferBlock)
		// if !ok {
		// 	return nil, ErrInvalidBlock
		// }
		// if resp.Status != StatusOK {
		// 	return nil, ErrNotFound
		// }

		// return resp.Block, nil
		return nil, nil

	case <-time.After(5 * time.Second):
		return nil, ErrNotFound
	case <-ctx.Done():
		return nil, ErrNotFound
	}
	return nil, ErrNotFound
}

func (blx *blockExchange) Send(recipient string, data []byte,
	meta map[string][]byte) (string, int, error) {

	hs := blx.hash(data)

	block := Block{
		Key:  hs,
		Meta: meta,
		Data: data,
	}

	nonce := uuid.NewV4().String()

	resp := payloadTransferBlock{
		Nonce: nonce,
		Block: &block,
	}

	// blx.storage.Store(block.Key, &block)

	ctx := context.Background()
	err := blx.wire.Send(ctx, wireExtention, PayloadTypeTransferBlock, resp,
		[]string{recipient})
	if err != nil {
		return "", 0, err
	}

	return hs, len(data), nil
}

func (b *blockExchange) hash(data []byte) string {
	h := sha256.New()
	h.Write(data)

	return string(hex.EncodeToString(h.Sum(nil)))
}
