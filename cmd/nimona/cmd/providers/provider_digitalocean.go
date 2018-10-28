package providers

import (
	"context"
	"log"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

type DigitalOceanProvider struct {
	client *godo.Client
}

type TokenSource struct {
	AccessToken string
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

func NewDigitalocean(token string) Provider {
	tokenSource := &TokenSource{
		AccessToken: token,
	}

	oauthClient := oauth2.NewClient(context.Background(), tokenSource)
	client := godo.NewClient(oauthClient)

	return &DigitalOceanProvider{
		client: client,
	}
}

func (dp *DigitalOceanProvider) NewInstance(name string) error {

	createRequest := &godo.DropletCreateRequest{
		Name:   name,
		Region: "lon1",
		Size:   "s-1vcpu-1gb",
		Image: godo.DropletCreateImage{
			Slug: "coreos-stable",
		},
		SSHKeys: []godo.DropletCreateSSHKey{godo.DropletCreateSSHKey{}},
	}

	drop, resp, err := dp.client.Droplets.Create(
		context.Background(), createRequest)
	if err != nil {
		log.Println(err)
		// return err
	}

	// list, resp, _ := dp.client.Images.List(context.Background(), &godo.ListOptions{
	// 	PerPage: 1000,
	// })

	log.Printf("%+v\n", drop)
	log.Printf("%+v\n", resp)

	return nil
}
