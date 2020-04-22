package config

import (
	"fmt"
	"net/http"
	"net/url"

	propeller "github.com/cbsinteractive/propeller-go/client"
)

// Propeller holds associated credentials for propeller api
type Propeller struct {
	Host  string `envconfig:"PROPELLER_HOST"`
	Creds string `envconfig:"PROPELLER_CREDS"`
	Auth  propeller.Auth
	API   *url.URL
}

func (p *Propeller) init() error {
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

	p.Auth = auth
	p.API = pURL

	return nil
}

// NewClient will set up the propeller client to track clients requests
func (p *Propeller) NewClient(c Client) (*propeller.Client, error) {
	return &propeller.Client{
		HostURL: p.API,
		Context: c.Context,
		Timeout: c.Timeout,
		Client:  c.Tracer.Client(&http.Client{}),
		Auth:    p.Auth,
	}, nil
}
