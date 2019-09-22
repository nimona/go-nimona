package containers

import (
	"context"
	"io"
	"io/ioutil"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type Container struct {
	Image string
	ID    string
	Name  string
}

func New(
	ctx context.Context,
	image string,
	name string,
	networkName string,
	portMap map[string]string,
	command []string,
) (*Container, error) {
	// Init the env
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	// Create the port mapping
	portBinding := nat.PortMap{}
	portsSet := nat.PortSet{}

	for conPort, hstPort := range portMap {
		hostBinding := nat.PortBinding{
			HostIP:   "0.0.0.0",
			HostPort: hstPort,
		}

		containerPort, err := nat.NewPort("tcp", conPort)
		if err != nil {
			// TODO log error
			continue
		}

		portBinding[containerPort] = []nat.PortBinding{hostBinding}
		portsSet[containerPort] = struct{}{}
	}

	// Pull the image
	// TODO This should not be here, should be a pre-requisite
	imgr, err := cli.ImagePull(
		ctx,
		image,
		types.ImagePullOptions{},
	)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(ioutil.Discard, imgr)
	if err != nil {
		return nil, err
	}

	// Create the container
	cont, err := cli.ContainerCreate(
		ctx,
		&container.Config{
			Image:        image,
			Cmd:          command,
			ExposedPorts: portsSet,
		},
		&container.HostConfig{
			AutoRemove:   true,
			PortBindings: portBinding,
		},
		nil,
		name,
	)
	if err != nil {
		return nil, err
	}

	if err := cli.NetworkConnect(
		ctx,
		networkName,
		cont.ID,
		&network.EndpointSettings{},
	); err != nil {
		return nil, err
	}
	// start the container
	err = cli.ContainerStart(
		ctx,
		cont.ID,
		types.ContainerStartOptions{},
	)
	if err != nil {
		return nil, err
	}

	return &Container{
		ID:    cont.ID,
		Name:  name,
		Image: image,
	}, nil

}

func (cnt *Container) Stop(ctx context.Context) error {
	// Init the env
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	timeout := 0 * time.Second

	if err := cli.ContainerStop(
		ctx,
		cnt.ID,
		&timeout,
	); err != nil {
		return err
	}

	return nil

}

func (cnt *Container) Delete(ctx context.Context) error {
	// Init the env
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	if err := cli.ContainerRemove(
		ctx, cnt.ID,
		types.ContainerRemoveOptions{}); err != nil {
		return err
	}
	return nil
}

func (cnt *Container) Logs(ctx context.Context) (io.ReadCloser, error) {
	// Init the env
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	rc, err := cli.ContainerLogs(
		ctx,
		cnt.ID,
		types.ContainerLogsOptions{ShowStdout: true},
	)
	if err != nil {
		return nil, err
	}

	return rc, nil
}

// Logs combines all the logs from all the containers
func Logs(containers []Container) (io.Reader, error) {
	ctx := context.Background()

	rdrs := []io.Reader{}

	for _, cnt := range containers {
		rd, err := cnt.Logs(ctx)
		if err != nil {
			return nil, err
		}

		rdrs = append(rdrs, rd)
	}

	mr := io.MultiReader(rdrs...)

	return mr, nil
}
