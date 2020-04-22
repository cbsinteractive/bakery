package config

import (
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

// Config holds all the configuration for this service
type Config struct {
	Listen      string `envconfig:"HTTP_PORT" default:":8080"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"debug"`
	OriginHost  string `envconfig:"ORIGIN_HOST"`
	Hostname    string `envconfig:"HOSTNAME"  default:"localhost"`
	OriginToken string `envconfig:"ORIGIN_TOKEN"`
	Tracer
	Client
	Propeller
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

// Authenticate will check the token passed in request
func (c Config) Authenticate(token string) bool {
	if c.OriginToken == token {
		return true
	}

	if c.IsLocalHost() {
		return true
	}

	return false
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
