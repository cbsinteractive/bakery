package config

import (
	"context"
	"net/http"
	"time"

	"github.com/cbsinteractive/pkg/tracing"
)

// HTTPClient hold interface declaration of our http clients
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client holds configuration for http clients
type Client struct {
	Context context.Context
	Timeout time.Duration `envconfig:"CLIENT_TIMEOUT" default:"5s"`
	Tracer  tracing.Tracer
	HTTPClient
}

// New creates a new instance of the HTTP Client
func (c Client) New() HTTPClient {
	if c.HTTPClient != nil { // only set during testing
		return c.HTTPClient
	}

	// https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
	return c.Tracer.Client(&http.Client{
		Timeout: c.Timeout,
	})

}

// SetContext will set the context on the incoming requests
func (c *Client) SetContext(r *http.Request) {
	c.Context = r.Context()
}
