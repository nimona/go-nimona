package main

import (
	"context"
	"fmt"
	"time"

	"github.com/tylertreat/bench"

	"github.com/nimona/go-nimona/dht"
	"github.com/nimona/go-nimona/net"
)

var bootstrapPeerInfos = []net.PeerInfo{
	// net.PeerInfo{
	// 	ID: "7730b73e34ae2e3ad92235aefc7ee0366736602f96785e6f35e8b710923b4562",
	// 	Addresses: []string{
	// 		"tcp:andromeda.nimona.io:26801",
	// 	},
	// 	PublicKey: [32]byte{
	// 		119, 48, 183, 62, 52, 174, 46, 58, 217, 34, 53, 174, 252, 126,
	// 		224, 54, 103, 54, 96, 47, 150, 120, 94, 111, 53, 232, 183, 16,
	// 		146, 59, 69, 98,
	// 	},
	// },
}

func main() {
	p1, _, r1 := newPeer(31000)
	bootstrapPeerInfos = append(bootstrapPeerInfos, p1.ToPeerInfo())
	bootstrapPeerInfos[0].Addresses = []string{"tcp:127.0.0.1:31000"}
	p2, w2, _ := newPeer(32000)
	p2s := p2.ToPeerInfo()
	p2s.Addresses = []string{"tcp:127.0.0.1:32000"}
	r1.PutPeerInfo(&p2s)

	r := &MessengerRequesterFactory{
		messenger: w2,
		recipient: p1.ToPeerInfo(),
		bytes:     100000000,
	}

	benchmark := bench.NewBenchmark(r, 100, 1, 60*time.Second, 0)
	summary, err := benchmark.Run()
	if err != nil {
		panic(err)
	}

	fmt.Println(summary)
	summary.GenerateLatencyDistribution(nil, "net.txt")
}

func newPeer(port int) (*net.SecretPeerInfo, net.Messenger, net.AddressBook) {
	reg := net.NewAddressBook()
	spi, _ := reg.CreateNewPeer()
	reg.PutLocalPeerInfo(spi)

	for _, peerInfo := range bootstrapPeerInfos {
		reg.PutPeerInfo(&peerInfo)
	}

	wre, _ := net.NewMessenger(reg)
	dht.NewDHT(wre, reg)

	wre.Listen(context.Background(), fmt.Sprintf("0.0.0.0:%d", port))

	wre.Handle("foo", func(event *net.Message) error {
		// fmt.Printf("___ Got message %s\n", string(event.Payload))
		return nil
	})

	return spi, wre, reg
}
