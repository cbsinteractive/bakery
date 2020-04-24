package config

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/cbsinteractive/pkg/tracing"
	propeller "github.com/cbsinteractive/propeller-go/client"
)

// Propeller holds associated credentials for propeller api
type Propeller struct {
	Host  string `envconfig:"PROPELLER_HOST"`
	Creds string `envconfig:"PROPELLER_CREDS"`
	Auth  propeller.Auth
	API   *url.URL
	propeller.Client
}

func (p *Propeller) init(trace tracing.Tracer, timeout time.Duration) error {
	if p.Host == "" || p.Creds == "" {
		return fmt.Errorf("your Propeller configs are not set")
	}

	pURL, err := url.Parse(p.Host)
	if err != nil {
		return fmt.Errorf("parsing propeller host url: %w", err)
	}

	auth, err := propeller.NewAuth(p.Creds, pURL.String())
	if err != nil {
		return err
	}

	p.Client = propeller.Client{
		HostURL:    pURL,
		Timeout:    timeout,
		HTTPClient: trace.Client(&http.Client{}),
		Auth:       auth,
	}

	return nil
}

//UpdateContext will update client context
func (p *Propeller) UpdateContext(ctx context.Context) {
	p.Client.Context = ctx
}
