package wire

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/nimona/go-nimona/mesh"

	"github.com/stretchr/testify/suite"
)

type wireTestSuite struct {
	suite.Suite
	bootstrapPeerInfos []mesh.PeerInfo
}

func (suite *wireTestSuite) SetupTest() {
	suite.bootstrapPeerInfos = []mesh.PeerInfo{}
}

func (suite *wireTestSuite) TestSendSuccess() {
	p1, w1, r1 := suite.newPeer(31012)
	p2, w2, r2 := suite.newPeer(32012)

	p1s := p1.ToPeerInfo()
	p1s.Addresses = []string{"tcp:127.0.0.1:31012"}
	r2.PutPeerInfo(&p1s)

	p2s := p2.ToPeerInfo()
	p2s.Addresses = []string{"tcp:127.0.0.1:32012"}
	r1.PutPeerInfo(&p2s)

	time.Sleep(time.Second)

	payload := map[string]string{
		"foo": "bar",
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	w1.HandleExtensionEvents("foo", func(message *Message) error {
		decPayload := map[string]string{}
		err := message.DecodePayload(&decPayload)
		suite.NoError(err)
		suite.Equal(payload, decPayload)
		wg.Done()
		return nil
	})

	ctx := context.Background()
	err := w2.Send(ctx, "foo", "bar", payload, []string{p1.ID})
	suite.NoError(err)
	wg.Wait()
}

func (suite *wireTestSuite) newPeer(port int) (*mesh.SecretPeerInfo, Wire, mesh.Registry) {
	reg := mesh.NewRegisty()
	spi, _ := reg.CreateNewPeer()
	reg.PutLocalPeerInfo(spi)

	for _, peerInfo := range suite.bootstrapPeerInfos {
		err := reg.PutPeerInfo(&peerInfo)
		suite.NoError(err)
	}

	wre, _ := NewWire(reg)
	_, _, lErr := wre.Listen(fmt.Sprintf("0.0.0.0:%d", port))
	suite.NoError(lErr)
	return spi, wre, reg
}

func TestWireTestSuite(t *testing.T) {
	suite.Run(t, new(wireTestSuite))
}
