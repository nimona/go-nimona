package containers

import (
	"io"
	"io/ioutil"
	"time"

	"nimona.io/pkg/context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type Container struct {
	Image   string
	ID      string
	Name    string
	Address string
}

func New(
	ctx context.Context,
	image string,
	name string,
	networkName string,
	portMap map[string]string,
	entrypoint []string,
	command []string,
	env []string,
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
	// swallow error for now to allow for local repos
	if err == nil {
		io.Copy(ioutil.Discard, imgr) // nolint: errcheck
	}

	// Create the container
	cont, err := cli.ContainerCreate(
		ctx,
		&container.Config{
			Image:        image,
			Entrypoint:   entrypoint,
			Cmd:          command,
			ExposedPorts: portsSet,
			Env:          env,
		},
		&container.HostConfig{
			AutoRemove:   false,
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

	ins, err := cli.ContainerInspect(ctx, cont.ID)
	if err != nil {
		return nil, err
	}

	gwAddress := ""
	for _, p := range ins.NetworkSettings.Ports {
		if len(p) > 0 {
			addr := ins.NetworkSettings.DefaultNetworkSettings.IPAddress
			gwAddress = addr + ":" + p[0].HostPort
			break
		}
	}
	return &Container{
		ID:      cont.ID,
		Name:    name,
		Image:   image,
		Address: gwAddress,
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

	if err := cli.ContainerRemove(ctx, cnt.ID, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}); err != nil {
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
		types.ContainerLogsOptions{
			ShowStdout: true,
			Follow:     true,
		},
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
