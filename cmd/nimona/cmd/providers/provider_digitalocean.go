package providers

import (
	"context"
	"time"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

const (
	cloudInit = `#cloud-config

  coreos:
    units:
      - name: nimona.service
        command: start
        enable: true
        content: |
          [Unit]
          Description=nimona
          After=docker.service
          Requires=docker.service
          After=docker.redis.service
          
          [Service]
          TimeoutStartSec=0
          Restart=always
          ExecStartPre=-/usr/bin/docker stop nimona
          ExecStartPre=-/usr/bin/docker rm nimona
          ExecStartPre=/usr/bin/docker pull nimona/nimona:latest
          ExecStart=/usr/bin/docker run --name nimona --rm -p 21013:21013 nimona/nimona daemon --port=21013 --api-port=8080
          
          [Install]
          WantedBy=multi-user.target`
)

// DigitalOceanProvider prodides a DO operations
type DigitalOceanProvider struct {
	client *godo.Client
}

type tokenSource struct {
	AccessToken string
}

func (t *tokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

// NewDigitalocean creates a new DigitalOcean Provider
func NewDigitalocean(token string) (Provider, error) {
	if token == "" {
		return nil, ErrNoToken
	}

	tokenSource := &tokenSource{
		AccessToken: token,
	}

	oauthClient := oauth2.NewClient(context.Background(), tokenSource)
	client := godo.NewClient(oauthClient)

	return &DigitalOceanProvider{
		client: client,
	}, nil
}

// NewInstance creates a new DO Droplet
func (dp *DigitalOceanProvider) NewInstance(name, sshFingerprint,
	size, region string) (string, error) {
	if size == "" {
		size = "s-1vcpu-1gb"
	}

	if region == "" {
		region = "lon1"
	}

	ctx := context.Background()
	createRequest := &godo.DropletCreateRequest{
		Name:   name,
		Region: region,
		Size:   size,
		Image: godo.DropletCreateImage{
			Slug: "coreos-stable",
		},
		SSHKeys: []godo.DropletCreateSSHKey{godo.DropletCreateSSHKey{
			Fingerprint: sshFingerprint,
		}},
		UserData: cloudInit,
	}

	// Create server
	drop, _, err := dp.client.Droplets.Create(
		ctx, createRequest)
	if err != nil {
		return "", err
	}

	// Wait for the API to return an IP
	for {
		d, _, err := dp.client.Droplets.Get(ctx, drop.ID)
		if err != nil {
			return "", err
		}

		ip, err := d.PublicIPv4()
		if err != nil {
			return "", err
		}
		if ip != "" {
			return ip, nil
		}

		time.Sleep(2 * time.Second)
	}

}
