package crdt

import (
	"fmt"

	uuid "github.com/google/uuid"
)

type Hashgraph struct {
	blocks   *BlockStore
	handlers []func(block *blx.Block) error
}

func (hg *Hashgraph) HandleBlock(handler func(block *Block) error) error {
	hg.handlers = append(hg.handlers, handler)
	return nil
}

func (hg *Hashgraph) CreateGraph(title, owner string, recipiends []string) (string, error) {
	// create block
	bl := &Block{
		Event: BlockEvent{
			Author: owner,
			Nonce:  RandStringBytesMaskImprSrc(10),
			Data:   title,
			Type:   EventTypeGraphCreate,
		},
	}
	bl.Hash = HashBlock(bl)
	if err := hg.blocks.Add(bl); err != nil {
		return "", err
	}

	return hg.sendBlock(bl, recipiends)
}

func (hg *Hashgraph) Subscribe(graphRootID string) (string, error) {
	// find tip of graph
	parents, err := hg.blocks.FindTip(graphRootID)
	if err != nil {
		return "", err
	}

	// create block
	bl := &Block{
		Event: BlockEvent{
			Author:  string(hg.peer.GetID()),
			Nonce:   RandStringBytesMaskImprSrc(10),
			Parents: parents,
			Type:    EventTypeGraphSubscribe,
		},
	}
	bl.Hash = HashBlock(bl)

	// store block
	if err := hg.blocks.Add(bl); err != nil {
		return "", err
	}

	// find recipients
	recipients, err := hg.blocks.FindSubscribers(graphRootID)
	if err != nil {
		fmt.Println("Could not find subscribers", err)
		return "", err
	}

	return hg.sendBlock(bl, recipients)
}

func (hg *Hashgraph) Append(threadID, eventType, data string) (string, error) {
	// find tip of thread
	parents, err := hg.blocks.FindTip(threadID)
	if err != nil {
		return "", err
	}

	// create block
	bl := &Block{
		Event: BlockEvent{
			Author:  string(hg.peer.GetID()),
			Nonce:   RandStringBytesMaskImprSrc(10),
			Parents: parents,
			Type:    eventType,
			Data:    data,
		},
	}
	bl.Hash = HashBlock(bl)

	// store block
	if err := hg.blocks.Add(bl); err != nil {
		return "", err
	}

	// find recipients
	recipients, err := hg.blocks.FindSubscribers(threadID)
	if err != nil {
		return "", err
	}

	return hg.sendBlock(bl, recipients)
}

func (hg *Hashgraph) sendBlock(bl *Block, recipients []string) (string, error) {
	peerID := string(hg.peer.GetID())

	// create PDU event
	ev := &events.Event{
		Type:       events.EventTypePDU,
		ID:         uuid.New().String(),
		OwnerID:    peerID,
		SenderID:   peerID,
		Recipients: recipients,
		Payload:    bl,
	}

	// sent block to recipiends
	if err := hg.eventBus.Send(ev); err != nil {
		return "", err
	}

	return bl.Hash, nil
}
