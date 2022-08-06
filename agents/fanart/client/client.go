package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/meteorae/meteorae-server/helpers"
)

// DefaultEndpoint represents the default endpoint of the API.
var DefaultEndpoint = "https://webservice.fanart.tv/v3"

// Client represents an API client.
type Client struct {
	Endpoint string
	APIKey   string
	Client   *http.Client
}

// New returns a new client.
func New() *Client {
	return &Client{
		Endpoint: DefaultEndpoint,
		APIKey:   "84d310b84b0b62da0cb23f8355271442",
		Client:   http.DefaultClient,
	}
}

func (c *Client) get(url string, resType interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", fmt.Sprintf("Meteorae/%s ( https://meteorae.tv )", helpers.Version))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("api-key", c.APIKey)

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}

	// Decode the JSON response
	if err = json.NewDecoder(resp.Body).Decode(&resType); err != nil {
		return err
	}

	return nil
}
