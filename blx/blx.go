package blx

import (
	"fmt"

	"github.com/nimona/go-nimona/mesh"
)

type blockExchange struct {
	pubSub mesh.PubSub
}

func NewBlockExchange(ps mesh.PubSub) (*blockExchange, error) {

	// Subscribe to blx events
	messages, err := ps.Subscribe("blx:.*")
	if err != nil {
		return nil, nil
	}

	go func() {
		// Handle incoming message
		for omsg := range messages {
			msg, ok := omsg.(mesh.Message)
			if !ok {
				continue
			}
			fmt.Println("Message received ===> ", msg)
		}
	}()

	return &blockExchange{
		pubSub: ps,
	}, nil
}
