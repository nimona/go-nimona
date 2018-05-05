package blx

import (
	"testing"

	"github.com/davecgh/go-spew/spew"

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
			reqPublishedMessage := args.Get(0).(mesh.Message)
			spew.Println(reqPublishedMessage)
		})

	err := suite.blx.Send("test02", "test01",
		[]byte("test"), map[string][]byte{})
	suite.NoError(err)
}

func TestRunBlxTestSuite(t *testing.T) {
	suite.Run(t, new(blxTestSuite))
}
