package blx

import (
	"github.com/nimona/go-nimona/mesh"
	"github.com/nimona/go-nimona/mutation"
	"github.com/stretchr/testify/mock"
	suite "github.com/stretchr/testify/suite"
)

type blxTestSuite struct {
	suite.Suite
	mockPubSub *mesh.MockPubSub
	messages   chan interface{}
	blx        *blockExchange
}

func (suite *blxTestSuite) SetupTest() {
	suite.mockPubSub = &mesh.MockPubSub{}
	suite.messages = make(chan interface{}, 10)
	suite.mockPubSub.On("Subscribe", "blx:.*").Return(suite.messages, nil)
	bootstrapMutation := mutation.PeerProtocolDiscovered{
		PeerID:          "bootstrap",
		ProtocolName:    "messaging",
		ProtocolAddress: "bootstrap-address",
		Pinned:          true,
	}
	suite.mockPubSub.On("Publish", bootstrapMutation,
		mutation.PeerProtocolDiscoveredTopic).Return(nil)

	suite.blx, _ = NewBlockExchange(suite.mockPubSub)
}

func (suite *blxTestSuite) TestReceiveMessage() {
	suite.mockPubSub.
		On("Publish", mock.AnythingOfType("mesh.Message"), "message:send").
		Return(nil).
		Run(func(args mock.Arguments) {

		})
}
