package node

import (
	"fmt"

	"nimona.io/internal/rand"
	"nimona.io/internal/simulation/containers"
	"nimona.io/pkg/context"
)

type Environment struct {
	net *containers.Network
}

const defaultNetworkName = "nimona-network"

func NewEnvironment() (*Environment, error) {
	ctx := context.Background()

	nnet, err := containers.NewNetwork(
		ctx,
		fmt.Sprintf("%s-%s", defaultNetworkName, rand.String(8)),
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
