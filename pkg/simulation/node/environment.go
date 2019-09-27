package node

import (
	"nimona.io/pkg/context"
	"fmt"
	"math/rand"

	"nimona.io/pkg/simulation/containers"
)

type Environment struct {
	net *containers.Network
}

const defaultNetworkName = "NimNet"

func NewEnvironment() (*Environment, error) {
	ctx := context.Background()

	nnet, err := containers.NewNetwork(
		ctx,
		fmt.Sprintf("%s-%d", defaultNetworkName, rand.Intn(100)),
	)
	if err != nil {
		return nil, err
	}

	return &Environment{
		net: nnet,
	}, nil
}

func (env *Environment) Stop() error {
	ctx := context.Background()

	err := env.net.Remove(ctx)
	if err != nil {
		return err
	}

	return nil
}
