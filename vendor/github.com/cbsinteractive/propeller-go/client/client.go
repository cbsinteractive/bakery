package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Client is the server's intermediate client that accesses manifests, playlists
// and transport streams. Public fields are optional and take reasonable defaults
// when nil.
type Client struct {
	HostURL *url.URL
	Timeout time.Duration
	Auth    Auth
	HTTPClient

	err error
}

// HTTPClient hold interface declaration of our http clients
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

//NewClient returns a new propeller client
func NewClient(credentials string, hostURL *url.URL) (*Client, error) {
	auth, err := NewAuth(credentials, hostURL.String())
	if err != nil {
		return &Client{}, err
	}

	return &Client{
		HostURL:    hostURL,
		Timeout:    time.Second * 30,
		HTTPClient: http.DefaultClient,
		Auth:       auth,
	}, nil
}

// CreateChannel creates a propeller Channel
func (c *Client) CreateChannel(ctx context.Context, orgID string, channel *Channel) error {
	path := fmt.Sprintf("/v1/organization/%v/channel", orgID)

	return c.post(ctx, path, channel, channel)
}

// GetChannel returns a propeller Channel
func (c *Client) GetChannel(ctx context.Context, orgID string, channelID string) (Channel, error) {
	resp := Channel{}
	path := fmt.Sprintf("/v1/organization/%v/channel/%v", orgID, channelID)

	err := c.get(ctx, path, &resp)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// DeleteChannel deletes a propeller Channel
func (c *Client) DeleteChannel(ctx context.Context, orgID string, channelID string) (MessageResponse, error) {
	resp := MessageResponse{}
	path := fmt.Sprintf("/v1/organization/%v/channel/%v", orgID, channelID)

	err := c.delete(ctx, path, &resp)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// CreateClip will create a propeller clip
func (c *Client) CreateClip(ctx context.Context, orgID string, clip *Clip) error {
	path := fmt.Sprintf("/v1/organization/%v/clip/%v", orgID, clip.ID)

	return c.put(ctx, path, clip, clip)
}

// GetClip returns a propeller clip based on channel-id if its channel archive was set to true
// If you do not have a clip ID, you can grab the current archive by setting your
// clipID = channelID-archive
func (c *Client) GetClip(ctx context.Context, orgID string, clipID string) (Clip, error) {
	resp := Clip{}
	path := fmt.Sprintf("/v1/organization/%v/clip/%v", orgID, clipID)

	err := c.get(ctx, path, &resp)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// DeleteClip will delete a propeller clip
func (c *Client) DeleteClip(ctx context.Context, orgID string, clipID string) (MessageResponse, error) {
	resp := MessageResponse{}
	path := fmt.Sprintf("/v1/organization/%v/clip/%v", orgID, clipID)

	err := c.delete(ctx, path, &resp)
	if err != nil {
		return resp, err
	}

	return resp, nil
}
