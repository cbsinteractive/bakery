package config

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/cbsinteractive/pkg/tracing"
	"github.com/cbsinteractive/pkg/xrayutil"
	propeller "github.com/cbsinteractive/propeller-go/client"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

// Config holds all the configuration for this service
type Config struct {
	Listen     string `envconfig:"HTTP_PORT" default:":8080"`
	LogLevel   string `envconfig:"LOG_LEVEL" default:"debug"`
	OriginHost string `envconfig:"ORIGIN_HOST"`
	Hostname   string `envconfig:"HOSTNAME"  default:"localhost"`
	Tracer
	Client
	Propeller
}

//Tracer holds configuration for initating the tracing of requests
type Tracer struct {
	EnableXRay        bool `envconfig:"ENABLE_XRAY" default:"false"`
	EnableXRayPlugins bool `envconfig:"ENABLE_XRAY_PLUGINS" default:"false"`
}

// Propeller holds associated credentials for propeller api
type Propeller struct {
	Host  string `envconfig:"PROPELLER_HOST"`
	Creds string `envconfig:"PROPELLER_CREDS"`
	Auth  propeller.Auth
	API   *url.URL
}

// Client will issue requests to the manifest
type Client struct {
	Context context.Context
	Timeout time.Duration `envconfig:"CLIENT_TIMEOUT" default:"5s"`
	Tracer  tracing.Tracer
}

// New creates a new instance of the HTTP Client
func (c Client) New() *http.Client {
	// https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
	client := c.Tracer.Client(&http.Client{
		Timeout: c.Timeout,
	})

	return client
}

// SetContext will set the context on the incoming requests
func (c *Client) SetContext(r *http.Request) {
	c.Context = r.Context()
}

// LoadConfig loads the configuration with environment variables injected
func LoadConfig() (Config, error) {
	var c Config
	err := envconfig.Process("bakery", &c)
	if err != nil {
		return c, err
	}

	c.Client.Tracer = c.Tracer.init(c.GetLogger())

	return c, c.Propeller.init()
}

// init will set up the tracer to track clients requests
func (t *Tracer) init(logger *logrus.Logger) tracing.Tracer {
	var tracer tracing.Tracer

	if t.EnableXRay {
		tracer = xrayutil.XrayTracer{
			EnableAWSPlugins: t.EnableXRayPlugins,
			InfoLogFn:        logger.Infof,
		}
	} else {
		tracer = tracing.NoopTracer{}
	}

	err := tracer.Init()
	if err != nil {
		logger.Fatalf("initializing tracer: %v", err)
	}

	return tracer
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

// IsLocalHost returns true if env is localhost
func (c Config) IsLocalHost() bool {
	if c.Hostname == "localhost" {
		return true
	}

	return false
}

// GetLogger generates a logger
func (c Config) GetLogger() *logrus.Logger {
	level, err := logrus.ParseLevel(c.LogLevel)
	if err != nil {
		level = logrus.DebugLevel
	}

	logger := logrus.New()
	logger.Out = os.Stdout
	logger.Level = level

	return logger
}
