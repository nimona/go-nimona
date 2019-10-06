package node

import (
	"fmt"

	"nimona.io/pkg/context"

	"nimona.io/pkg/simulation/containers"
)

const defaultContainerPort = 28000

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
		Name:          "NimNode",
		Count:         1,
		Command:       []string{},
		ContainerPort: defaultContainerPort,
	}
	for _, opt := range opts {
		opt(options)
	}

	ports := map[string]string{}

	for i := 0; i < options.Count; i++ {
		if options.NodePort > 0 {
			ports[fmt.Sprintf("%d", options.ContainerPort)] =
				fmt.Sprintf("%d", options.NodePort+i)
		}

		cnt, err := containers.New(
			ctx,
			image,
			fmt.Sprintf("%s-%d", options.Name, i),
			Environment.net.Name,
			ports,
			options.Command,
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
