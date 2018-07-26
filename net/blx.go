package net

// import (
// 	"context"
// 	"crypto/sha256"
// 	"encoding/hex"
// 	"errors"
// 	"sync"

// 	"github.com/sirupsen/logrus"
// )

// // const (
// // 	messengerExtention = "blx"
// // )

// var (
// 	ErrMissingPeer    = errors.New("Peer ID missing")
// 	ErrInvalidBlock   = errors.New("Invalid Block")
// 	ErrInvalidRequest = errors.New("Invalid request")
// )

// // BlockExchange enables the transfer and storage of
// // blocks between peers.
// type BlockExchange interface {
// 	Get(key string, recipient string) (*Block, error)
// 	Send( data []byte,
// 		meta map[string][]byte) (string, int, error)
// 	GetLocalBlocks() ([]string, error)
// 	Subscribe(fn subscriptionCb) (string, error)
// 	Unsubscribe(id string)
// }

// type subscriptionCb func(string)

// type blockExchange struct {
// 	net           Messenger
// 	storage       Storage
// 	getRequests   sync.Map
// 	subscriptions sync.Map
// }

// // NewBlockExchange get Messenger and a Storage as parameters and returns a new
// // block exchange protocol.
// func NewBlockExchange(n Messenger, pr Storage) (BlockExchange, error) {
// 	blx := &blockExchange{
// 		net:         n,
// 		storage:     pr,
// 		getRequests: sync.Map{},
// 	}

// 	// n.Handle(messengerExtention, blx.handleBlock)

// 	return blx, nil
// }

// func (blx *blockExchange) handleBlock(block *Block) error {
// 	contentType := block.Metadata.Type
// 	switch contentType {
// 	case PayloadTypeTransferBlock:
// 		err := blx.handleTransferBlock(block)
// 		if err != nil {
// 			return err
// 		}
// 	case PayloadTypeRequestBlock:
// 		err := blx.handleRequestBlock(block)
// 		if err != nil {
// 			return err
// 		}
// 	default:
// 		logrus.WithField("block.PayloadType", contentType).
// 			Warn("Payload type not known")
// 		return nil
// 	}
// 	return nil
// }

// func (blx *blockExchange) handleTransferBlock(block *Block) error {
// 	payload := block.Payload.(PayloadTransferBlock)
// 	if payload.Block != nil {
// 		err := blx.storage.Store(payload.Block.Metadata.ID, payload.Block)
// 		blx.publish(payload.Block.Metadata.ID)

// 		if err != nil {
// 			return err
// 		}
// 	}

// 	// Check if nonce exists in local addressBook
// 	value, ok := blx.getRequests.Load(payload.RequestID)
// 	if !ok {
// 		return nil
// 	}

// 	req, ok := value.(*PayloadRequestBlock)
// 	if !ok {
// 		return ErrInvalidRequest
// 	}

// 	req.response <- payload

// 	return nil
// }

// func (blx *blockExchange) handleRequestBlock(incBlock *Block) error {
// 	payload := incBlock.Payload.(PayloadRequestBlock)
// 	status := StatusOK

// 	// TODO handle block request
// 	pblock, err := blx.storage.Get(payload.ID)
// 	if err != nil {
// 		pblock = nil
// 		status = StatusNotFound
// 	}

// 	resp := PayloadTransferBlock{
// 		RequestID:  payload.RequestID,
// 		Block:  pblock,
// 		Status: status,
// 	}

// 	ctx := context.Background()
// 	block := NewBlock(PayloadTypeTransferBlock, []string{payload.RequestingPeerID}, resp)
// 	if err := blx.Send(ctx, block); err != nil {
// 		logrus.WithError(err).Warnf("blx.handleRequestBlock could not send block")
// 		return err
// 	}

// 	return nil
// }

// func (blx *blockExchange) Get(key string, recipient string) (
// 	*Block, error) {
// 	// TODO remember to remove nonce
// 	nonce := RandStringBytesMaskImprSrc(8)

// 	req := &PayloadRequestBlock{
// 		RequestingPeerID: "MISSING-REQUEST-PEER-ID", // TODO Missing requesting peer id
// 		RequestID:            nonce,
// 		Key:              key,
// 		response:         make(chan interface{}, 1),
// 	}

// 	// Check local storage for block
// 	block, err := blx.storage.Get(key)
// 	if err == nil {
// 		return block, nil
// 	}
// 	if err != ErrNotFound {
// 		return nil, err
// 	}
// 	if recipient == "" {
// 		return nil, ErrMissingPeer
// 	}

// 	blx.getRequests.Store(req.RequestID, req)

// 	// Request block
// 	ctx := context.Background()
// 	block := NewBlock(PayloadTypeRequestBlock, []string{recipient}, req)
// 	if err := blx.Send(ctx, block); err != nil {
// 		logrus.WithError(err).Warnf("blx.Get could not send block")
// 		return nil, err
// 	}

// 	for {
// 		select {
// 		case response := <-req.response:
// 			resp, ok := response.(*PayloadTransferBlock)
// 			if !ok {
// 				return nil, ErrInvalidBlock
// 			}
// 			if resp.Status != StatusOK {
// 				return nil, ErrNotFound
// 			}

// 			return resp.Block, nil

// 		case <-ctx.Done():
// 			return nil, ErrNotFound
// 		}
// 	}
// 	return nil, ErrNotFound
// }

// func (blx *blockExchange) Send(recipient string, data []byte,
// 	values map[string][]byte) (string, int, error) {

// 	hs := blx.hash(data)

// 	block := Block{
// 		Key:  hs,
// 		Meta: Meta{Values: values},
// 		Data: data,
// 	}

// 	nonce := RandStringBytesMaskImprSrc(8)

// 	resp := PayloadTransferBlock{
// 		RequestID: nonce,
// 		Block: &block,
// 	}

// 	blx.storage.Store(block.Metadata.ID, &block)
// 	blx.publish(block.Metadata.ID)

// 	ctx := context.Background()
// 	block := NewBlock(PayloadTypeTransferBlock, []string{recipient}, resp)
// 	if err := blx.Send(ctx, block); err != nil {
// 		logrus.WithError(err).Warnf("blx.Send could not send block")
// 		return "", 0, err
// 	}

// 	return hs, len(data), nil
// }

// func (blx *blockExchange) GetLocalBlocks() ([]string, error) {
// 	return blx.storage.List()
// }

// // Subscribe registers a function to be called when an event happens
// // returns the id for the registration
// func (blx *blockExchange) Subscribe(fn subscriptionCb) (string, error) {
// 	id := RandStringBytesMaskImprSrc(8)
// 	blx.subscriptions.Store(id, fn)
// 	return id, nil
// }

// // Unsubscribe unregisters the id
// func (blx *blockExchange) Unsubscribe(id string) {
// 	blx.subscriptions.Delete(id)
// }
// func (blx *blockExchange) publish(newKey string) {
// 	blx.subscriptions.Range(func(k, v interface{}) bool {
// 		f, ok := v.(subscriptionCb)
// 		if !ok {
// 			return false
// 		}

// 		f(newKey)
// 		return true
// 	})
// }

// func (b *blockExchange) hash(data []byte) string {
// 	h := sha256.New()
// 	h.Write(data)

// 	return string(hex.EncodeToString(h.Sum(nil)))
// }
