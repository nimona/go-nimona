package node

import (
	"fmt"
	"strconv"

	"nimona.io/pkg/context"

	"nimona.io/pkg/simulation/containers"
)

type Node struct {
	name      string
	container *containers.Container
}

func New(
	image string,
	Environment *Environment,
	opts ...Option,
) ([]*Node, error) {
	ctx := context.Background()
	nodes := []*Node{}

	options := &Options{
		Name:         "NimNode",
		Count:        1,
		Command:      []string{},
		PortMappings: map[int]int{},
	}
	for _, opt := range opts {
		opt(options)
	}

	ports := map[string]string{}

	for i := 0; i < options.Count; i++ {
		for containerPort, nodePort := range options.PortMappings {
			ports[strconv.Itoa(containerPort)] = strconv.Itoa(nodePort + i)
		}

		cnt, err := containers.New(
			ctx,
			image,
			fmt.Sprintf("%s-%d", options.Name, i),
			Environment.net.Name,
			ports,
			options.Command,
			options.Env,
		)
		// if this fails the containers need to be cleaned up
		if err != nil {
			return nodes, err
		}

		nodes = append(
			nodes,
			&Node{
				name:      options.Name,
				container: cnt,
			},
		)
	}
	return nodes, nil
}

func Stop(nodes []*Node) error {
	ctx := context.Background()

	for _, nd := range nodes {
		err := nd.container.Stop(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
