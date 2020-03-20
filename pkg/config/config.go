package config

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	propeller "github.com/cbsinteractive/propeller-client-go/pkg/client"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

// Config holds all the configuration for this service
type Config struct {
	Listen     string `envconfig:"HTTP_PORT" default:":8080"`
	LogLevel   string `envconfig:"LOG_LEVEL" default:"debug"`
	OriginHost string `envconfig:"ORIGIN_HOST"`
	Hostname   string `envconfig:"HOSTNAME"  default:"localhost"`
	Client     HTTPClient
	Propeller
}

// Propeller holds the client ands its associated credentials
type Propeller struct {
	Host   string `envconfig:"PROPELLER_HOST"`
	Creds  string `envconfig:"PROPELLER_CREDS"`
	Client *propeller.Client
}

// HTTPClient will issue requests to the manifest
type HTTPClient struct {
	Timeout time.Duration `envconfig:"CLIENT_TIMEOUT" default:"5s"`
}

// New creates a new instance of the HTTP Client
func (h HTTPClient) New() *http.Client {
	// https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
	client := &http.Client{
		Timeout: h.Timeout,
	}

	return client
}

// LoadConfig loads the configuration with environment variables injected
func LoadConfig() (Config, error) {
	var c Config
	err := envconfig.Process("bakery", &c)
	if err != nil {
		return c, err
	}

	return c, c.Propeller.init()
}

func (p *Propeller) init() error {
	if p.Host == "" || p.Creds == "" {
		return fmt.Errorf("your Propeller configs are not set")
	}

	pURL, err := url.Parse(p.Host)
	if err != nil {
		return fmt.Errorf("parsing propeller host url: %w", err)
	}

	p.Client, err = propeller.NewClient(p.Creds, pURL)

	return err
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
