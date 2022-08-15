package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/meteorae/meteorae-server/helpers"
)

// Client represents an API client.
type Client struct {
	Endpoint string
	APIKey   string
	Client   *http.Client
}

// New returns a new client.
func New() *Client {
	return &Client{
		Endpoint: "https://webservice.fanart.tv/v3",
		APIKey:   "84d310b84b0b62da0cb23f8355271442",
		Client:   http.DefaultClient,
	}
}

func (c *Client) get(ctx context.Context, url string, resType interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", fmt.Sprintf("Meteorae/%s ( https://meteorae.tv )", helpers.Version))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("api-key", c.APIKey)

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Decode the JSON response
	if err = json.NewDecoder(resp.Body).Decode(&resType); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
