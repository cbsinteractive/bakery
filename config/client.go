package config

import (
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
	Timeout time.Duration `envconfig:"CLIENT_TIMEOUT" default:"5s"`
	Tracer  tracing.Tracer
	HTTPClient
}

// SetContext will set the context on the incoming requests
func (c *Client) init(t tracing.Tracer) {
	c.Tracer = t
	c.HTTPClient = c.Tracer.Client(&http.Client{
		Timeout: c.Timeout,
	})
}
