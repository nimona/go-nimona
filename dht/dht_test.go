package dht

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/nimona/go-nimona/mesh"
	"github.com/nimona/go-nimona/mutation"

	"github.com/stretchr/testify/mock"
	suite "github.com/stretchr/testify/suite"
)

type dhtTestSuite struct {
	suite.Suite
	mockMesh   *mesh.MockMesh
	mockPubSub *mesh.MockPubSub
	messages   chan interface{}
	peers      chan interface{}
	dht        *DHT
}

func (suite *dhtTestSuite) SetupTest() {
	suite.mockMesh = &mesh.MockMesh{}
	suite.mockPubSub = &mesh.MockPubSub{}
	suite.messages = make(chan interface{}, 10)
	suite.peers = make(chan interface{}, 10)
	suite.mockPubSub.On("Subscribe", "dht:.*").Return(suite.messages, nil)
	suite.mockPubSub.On("Subscribe", "peer:.*").Return(suite.peers, nil)
	bootstrapMutation := mutation.PeerProtocolDiscovered{
		PeerID:          "bootstrap",
		ProtocolName:    "messaging",
		ProtocolAddress: "bootstrap-address",
		Pinned:          true,
	}
	suite.mockPubSub.On("Publish", bootstrapMutation, mutation.PeerProtocolDiscoveredTopic).Return(nil)
	suite.dht, _ = NewDHT(suite.mockPubSub, "local-peer", false, "bootstrap-address")
}

func (suite *dhtTestSuite) TestFilterSuccess() {
	ctx := context.Background()
	key := "a"
	value := "b"
	labels := map[string]string{
		"c": "d",
		"e": "f",
	}
	payload := &messageGet{
		OriginPeerID: "local-peer",
		Key:          key,
		Labels:       labels,
	}
	expMessage := mesh.Message{
		Recipient: "bootstrap",
		Sender:    "local-peer",
		Payload:   nil,
		Topic:     MessageTypeGet,
		Codec:     "json",
	}
	suite.mockPubSub.
		On("Publish", mock.AnythingOfType("mesh.Message"), "message:send").
		Return(nil).
		Run(func(args mock.Arguments) {
			// get the parts of the message that are variabls
			reqPublishedMessage := args.Get(0).(mesh.Message)
			expMessage.Nonce = reqPublishedMessage.Nonce
			// get the parts of the payload that are variabls
			reqPayload := &messageGet{}
			json.Unmarshal(reqPublishedMessage.Payload, &reqPayload)
			payload.QueryID = reqPayload.QueryID
			pbs, _ := json.Marshal(payload)
			expMessage.Payload = pbs
			// check message
			suite.Assert().Equal(expMessage, reqPublishedMessage)
			// publish a PUT message with the expected outcome
			retPayload := &messagePut{
				OriginPeerID: "some-peer",
				QueryID:      reqPayload.QueryID,
				Key:          "a",
				Value:        "b",
				Labels:       labels,
			}
			rpbs, _ := json.Marshal(retPayload)
			retPublishMessage := mesh.Message{
				Recipient: "local-peer",
				Sender:    "some-peer",
				Payload:   rpbs,
				Topic:     MessageTypePut,
				Codec:     "json",
			}
			suite.messages <- retPublishMessage
		})

	res, err := suite.dht.Filter(ctx, key, labels)
	time.Sleep(time.Second)
	suite.Nil(err)
	suite.NotEmpty(res)
	retVal := <-res
	suite.Equal(value, retVal.GetValue())
}

func (suite *dhtTestSuite) TestPutSuccess() {
	ctx := context.Background()
	key := "a"
	value := "b"
	labels := map[string]string{
		"c": "d",
		"e": "f",
	}
	payload := &messagePut{
		OriginPeerID: "local-peer",
		Key:          key,
		Value:        value,
		Labels:       labels,
	}
	pbs, _ := json.Marshal(payload)
	expMessage := mesh.Message{
		Recipient: "bootstrap",
		Sender:    "local-peer",
		Payload:   pbs,
		Topic:     MessageTypePut,
		Codec:     "json",
	}
	suite.mockPubSub.
		On("Publish", mock.AnythingOfType("mesh.Message"), "message:send").
		Return(nil).
		Run(func(args mock.Arguments) {
			// get the parts of the message that are variabls
			reqPublishedMessage := args.Get(0).(mesh.Message)
			expMessage.Nonce = reqPublishedMessage.Nonce
			suite.Assert().Equal(expMessage, reqPublishedMessage)
		})

	err := suite.dht.Put(ctx, key, value, labels)
	suite.Nil(err)
}

func TestDHTTestSuite(t *testing.T) {
	suite.Run(t, new(dhtTestSuite))
}
