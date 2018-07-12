package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"time"

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

	ctx := context.Background()
	p1ID := p1.ID
	size := 100000

	fmt.Println("Size (KB), Time (MS)")

	payload := make([]byte, size)
	if _, err := rand.Read(payload); err != nil {
		log.Fatal(err)
	}

	total := 0
	start := time.Now()
	for i := 0; i < 10000; i++ {
		if err := w2.Send(ctx, "foo", "bar", payload, []string{p1ID}); err != nil {
			fmt.Println(err)
		}
		total = total + size
	}
	elapsed := time.Since(start)
	fmt.Printf("Wrote %d kb in %d ms\n", total/1000, elapsed.Nanoseconds()/int64(time.Millisecond))
}

func newPeer(port int) (*net.SecretPeerInfo, net.Messenger, net.AddressBook) {
	reg := net.NewAddressBook()
	spi, _ := reg.CreateNewPeer()
	reg.PutLocalPeerInfo(spi)

	for _, peerInfo := range bootstrapPeerInfos {
		reg.PutPeerInfo(&peerInfo)
	}

	wre, _ := net.NewWire(reg)
	dht.NewDHT(wre, reg)

	wre.Listen(fmt.Sprintf("0.0.0.0:%d", port))

	wre.HandleExtensionEvents("foo", func(event *net.Message) error {
		// fmt.Printf("___ Got message %s\n", string(event.Payload))
		return nil
	})

	return spi, wre, reg
}
