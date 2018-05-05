package blx

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"

	"github.com/nimona/go-nimona/mesh"
	"github.com/sirupsen/logrus"
)

type blockExchange struct {
	pubSub           mesh.PubSub
	storage          Storage
	transferedBlocks chan *Block
}

func NewBlockExchange(ps mesh.PubSub) (*blockExchange, error) {

	// Subscribe to blx events
	messages, err := ps.Subscribe("blx:.*")
	if err != nil {
		return nil, nil
	}

	be := blockExchange{
		pubSub:           ps,
		storage:          newMemoryStore(),
		transferedBlocks: make(chan *Block),
	}

	go func() {
		// Handle incoming message
		for omsg := range messages {
			msg, ok := omsg.(mesh.Message)
			if !ok {
				continue
			}

			logrus.Infof("Incoming message ====> ", msg.String())

			switch msg.Topic {
			case MessageTypeRequest:
				br := BlockRequest{}
				json.Unmarshal(msg.Payload, &br)
				// bl, err := be.storage.Get(br.Key)
				// if err != nil {

				// }
				// be.Send(msg.Sender, msg.Recipient, bytes.NewReader(bl.Data), bl.Meta)
			case MessageTypeTransfer:
				bl := Block{}
				json.Unmarshal(msg.Payload, &bl)
				be.transferedBlocks <- &bl

			}
		}
	}()

	return &be, nil
}

func (b *blockExchange) Get(key string, recipient,
	sender string) (*Block, error) {
	blr := BlockRequest{
		Key: key,
	}

	blrm, err := json.Marshal(blr)
	if err != nil {
		return nil, err
	}

	msg := mesh.Message{
		Recipient: recipient,
		Sender:    sender,
		Payload:   blrm,
		Topic:     MessageTypeRequest,
		Codec:     "json",
		Nonce:     mesh.RandStringBytesMaskImprSrc(8),
	}

	if err := b.pubSub.Publish(msg, "message:send"); err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *blockExchange) Send(recipient, sender string, r io.Reader,
	meta map[string][]byte) error {

	data := make([]byte, 16000, 16000)
	hs := b.hash(r)
	fmt.Println("------> ", hs)

	bf := bufio.NewReader(r)
	for {
		// Reads until EOF
		_, err := bf.Read(data)
		if err != nil {
			return err
		}

		bl := Block{
			Key:  hs,
			Meta: meta,
			Data: data,
		}

		b.storage.Store(bl.Key, &bl)

		blm, err := json.Marshal(bl)
		if err != nil {
			return err
		}

		msg := mesh.Message{
			Recipient: recipient,
			Sender:    sender,
			Payload:   blm,
			Topic:     MessageTypeTransfer,
			Codec:     "json",
			Nonce:     mesh.RandStringBytesMaskImprSrc(8),
		}

		// logrus.Info("Publishing message", msg.String())

		if err := b.pubSub.Publish(msg, "message:send"); err != nil {
			return err
		}
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
