package bochalient

import (
	"Programming-Demo/config"
	"Programming-Demo/pkg/utils/bocha"
)

type Client struct {
	client *bocha.Client
}

var BochaClient *Client

func InitBochaClient() {
	apiKey := config.GetConfig().BOCHA_Apikey
	BochaClient = &Client{
		client: bocha.NewClient(apiKey),
	}

}

func (c *Client) GetClient() *bocha.Client {
	return c.client
}
