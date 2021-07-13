package config

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/cbsinteractive/pkg/tracing"
	propeller "github.com/cbsinteractive/propeller-go/client"
)

// Propeller holds associated credentials for propeller api
type Propeller struct {
	Enabled bool   `envconfig:"PROPELLER_ENABLED" default:"false"`
	Host    string `envconfig:"PROPELLER_HOST"`
	Creds   string `envconfig:"PROPELLER_CREDS"`
	propeller.Client
}

func (p *Propeller) IsEnabled() bool {
	return p.Enabled
}

func (p *Propeller) init(trace tracing.Tracer, timeout time.Duration) error {
	if !p.Enabled {
		return nil
	}

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
