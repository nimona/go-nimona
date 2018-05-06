package blx

import (
	"bufio"
	"context"
	"crypto/sha256"
	"fmt"
	"io"

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
		logrus.WithField("message.PayloadType", message.PayloadType).Warn("Payload type not known")
		return nil
	}
	return nil
}

func (blx *blockExchange) handleTransferBlock(message *wire.Message) {
	payload := &payloadTransferBlock{}
	if err := message.DecodePayload(payload); err != nil {
		return
	}

	blx.transferedBlocks <- payload.Block
}

func (blx *blockExchange) handleRequestBlock(message *wire.Message) {
	payload := &payloadTransferRequestBlock{}
	if err := message.DecodePayload(payload); err != nil {
		return
	}

	// TODO handle block request
}

func (blx *blockExchange) Get(key string, recipient string) (*payloadTransferBlock, error) {
	req := &payloadTransferRequestBlock{
		Key: key,
	}

	ctx := context.Background()
	blx.wire.Send(ctx, wireExtention, PayloadTypeRequestBlock, req, []string{recipient})

	// TODO wait to get block and return

	return nil, nil
}

func (blx *blockExchange) Send(recipient string, r io.Reader, meta map[string][]byte) error {

	data := make([]byte, 16000, 16000)
	hs := blx.hash(r)
	fmt.Println("------> ", hs)

	bf := bufio.NewReader(r)
	for {
		// Reads until EOF
		_, err := bf.Read(data)
		if err != nil {
			return err
		}

		block := &Block{
			Key:  hs,
			Meta: meta,
			Data: data,
		}
		resp := payloadTransferBlock{
			Block: block,
		}

		blx.storage.Store(block.Key, block)

		ctx := context.Background()
		blx.wire.Send(ctx, wireExtention, PayloadTypeTransferBlock, resp, []string{recipient})
	}

	return nil
}

func (b *blockExchange) hash(r io.Reader) string {
	br := bufio.NewReader(r)
	h := sha256.New()

	data := make([]byte, 1000, 1000)
	n := 0
	for {
		pd, err := br.Peek(100)
		if err != nil {
			logrus.WithError(err).Error("asdfasdf")
			break
		}

		n++
		fmt.Println(n)
		fmt.Println(string(pd))
		data = append(data, pd...)
	}

	h.Write(data)

	return string(h.Sum(nil))
}
