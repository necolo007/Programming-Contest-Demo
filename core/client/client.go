package client

import (
	"Programming-Demo/config"
	"github.com/northes/go-moonshot"
)

type Client struct {
	client *moonshot.Client
}

var MoonClient *Client

func InitClient() {
	MoonClient = &Client{}
	MoonClient.client, _ = moonshot.NewClient(config.GetConfig().Apikey)
}

func (c *Client) GetClient() *moonshot.Client {
	return c.client
}
