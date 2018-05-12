package blx

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/davecgh/go-spew/spew"

	"github.com/nimona/go-nimona/wire"
	"github.com/sirupsen/logrus"
)

const (
	wireExtention = "blx"
)

type blockExchange struct {
	wire             wire.Wire
	storage          Storage
	transferedBlocks chan *Block
}

func NewBlockExchange(wr wire.Wire) (*blockExchange, error) {
	blx := &blockExchange{
		wire:             wr,
		storage:          newMemoryStore(),
		transferedBlocks: make(chan *Block),
	}

	wr.HandleExtensionEvents("blx", blx.handleMessage)

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

func (blx *blockExchange) handleTransferBlock(message *wire.Message) {
	payload := &payloadTransferBlock{}
	fmt.Println("transfer")
	if err := message.DecodePayload(payload); err != nil {
		fmt.Println(err)
		return
	}
	spew.Println(payload.Block)
	// blx.transferedBlocks <- payload.Block
}

func (blx *blockExchange) handleRequestBlock(message *wire.Message) {
	payload := &payloadTransferRequestBlock{}
	if err := message.DecodePayload(payload); err != nil {
		fmt.Println(err)

		return
	}

	// TODO handle block request
}

func (blx *blockExchange) Get(key string, recipient string) (
	*payloadTransferBlock, error) {
	req := &payloadTransferRequestBlock{
		Key: key,
	}

	ctx := context.Background()
	blx.wire.Send(ctx, wireExtention, PayloadTypeRequestBlock, req,
		[]string{recipient})

	// TODO wait to get block and return

	return nil, nil
}

func (blx *blockExchange) Send(recipient string, data []byte,
	meta map[string][]byte) (string, int, error) {

	hs := blx.hash(data)

	block := Block{
		Key:  hs,
		Meta: meta,
		Data: data,
	}

	resp := payloadTransferBlock{
		Block: &block,
	}

	// blx.storage.Store(block.Key, block)

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
