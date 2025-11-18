package supabase

import (
	"github.com/supabase-community/supabase-go"
)

// Client wraps the Supabase client
type Client struct {
	client *supabase.Client
}

// New creates a new Supabase client
func New(url, anonKey string) (*Client, error) {
	client, err := supabase.NewClient(url, anonKey, nil)
	if err != nil {
		return nil, err
	}

	return &Client{client: client}, nil
}

// GetClient returns the underlying Supabase client
func (c *Client) GetClient() *supabase.Client {
	return c.client
}

