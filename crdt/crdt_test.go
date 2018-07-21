package crdt

import (
	"testing"

	suite "github.com/stretchr/testify/suite"

	events "github.com/nimona/go-nimona-events"
	peerstore "github.com/nimona/go-nimona-peerstore"
	mock "github.com/stretchr/testify/mock"
)

type HashgraphTestSuite struct {
	suite.Suite
	hashgraph  *Hashgraph
	blockstore *BlockStore
	eventbus   *events.MockEventBus
	handler    func(block *Block) error
	peerID     string
}

func (s *HashgraphTestSuite) SetupTest() {
	s.peerID = "peer-owner"
	s.eventbus = &events.MockEventBus{}
	s.blockstore = &BlockStore{
		blocks: map[string]*Block{},
	}
	s.hashgraph = &Hashgraph{
		peer: &peerstore.BasicPeer{
			ID: peerstore.ID(s.peerID),
		},
		blocks:   s.blockstore,
		eventBus: s.eventbus,
	}
}

func (s *HashgraphTestSuite) TestCreateGraphSucceeds() {
	title := "new-graph"
	recipients := []string{"peer-1"}

	// reset blockstore
	for k := range s.blockstore.blocks {
		delete(s.blockstore.blocks, k)
	}

	// check that the blockstore is empty
	s.Len(s.blockstore.blocks, 0)

	// keep some things from the eventbus for future tests
	var event *events.Event
	var block *Block

	// mock event bus send
	s.eventbus.On("Send", mock.AnythingOfType("*events.Event")).Return(nil).
		Run(func(args mock.Arguments) {
			s.IsType(&events.Event{}, args[0])
			event = args[0].(*events.Event)

			s.Equal(s.peerID, event.OwnerID)
			s.Equal(s.peerID, event.SenderID)
			s.IsType(&Block{}, event.Payload)
			block = event.Payload.(*Block)
			s.Equal(s.peerID, block.Event.Author)
			s.Equal(title, block.Event.Data)
			s.Len(block.Event.Parents, 0)
			s.Equal(EventTypeGraphCreate, block.Event.Type)
		})

	// create the graph
	hash, err := s.hashgraph.CreateGraph(title, recipients)
	s.Nil(err)
	s.Equal(hash, block.Hash)

	// check that the blockstore has at least one block
	s.Len(s.blockstore.blocks, 1)
	// and check something in the block just in case
	s.Equal(title, s.blockstore.blocks[block.Hash].Event.Data)
}

func (s *HashgraphTestSuite) TestSubscribeGraphSucceeds() {
	title := "new-graph"
	recipients := []string{"peer-1"}

	// reset blockstore
	for k := range s.blockstore.blocks {
		delete(s.blockstore.blocks, k)
	}

	// check that the blockstore is empty
	s.Len(s.blockstore.blocks, 0)

	// keep some things from the eventbus for future tests
	var event *events.Event
	var block *Block

	// mock event bus send
	s.eventbus.On("Send", mock.AnythingOfType("*events.Event")).Return(nil).
		Run(func(args mock.Arguments) {
			s.IsType(&events.Event{}, args[0])
			event = args[0].(*events.Event)
			s.Equal(s.peerID, event.OwnerID)
			s.Equal(s.peerID, event.SenderID)
			s.IsType(&Block{}, event.Payload)
			block = event.Payload.(*Block)
			s.Equal(s.peerID, block.Event.Author)
			s.Equal(title, block.Event.Data)
			s.Len(block.Event.Parents, 0)
			s.Equal(EventTypeGraphCreate, block.Event.Type)
		})

	// create the graph
	hash, err := s.hashgraph.CreateGraph(title, recipients)
	s.Nil(err)
	s.Equal(hash, block.Hash)

	// clear previous mocks
	s.eventbus.Mock = mock.Mock{}

	// mock event bus send
	s.eventbus.On("Send", mock.AnythingOfType("*events.Event")).Return(nil).
		Run(func(args mock.Arguments) {
			s.IsType(&events.Event{}, args[0])
			event = args[0].(*events.Event)
			s.Equal(s.peerID, event.OwnerID)
			s.Equal(s.peerID, event.SenderID)
			s.IsType(&Block{}, event.Payload)
			block = event.Payload.(*Block)
			s.Equal(s.peerID, block.Event.Author)
			s.Equal([]string{hash}, block.Event.Parents)
			s.Equal(EventTypeGraphSubscribe, block.Event.Type)
		})

	// create the graph
	hash, err = s.hashgraph.Subscribe(hash)
	s.Nil(err)
	s.Equal(hash, block.Hash)

	// check that the blockstore has at least one block
	s.Len(s.blockstore.blocks, 2)
	// and check something in the block just in case
	s.Equal(s.peerID, s.blockstore.blocks[block.Hash].Event.Author)
}

func TestHashgraphTestSuite(t *testing.T) {
	suite.Run(t, new(HashgraphTestSuite))
}
