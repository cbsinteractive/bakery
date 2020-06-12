package config

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/justinas/alice"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

// Config holds all the configuration for this service
type Config struct {
	Listen      string `envconfig:"HTTP_PORT" default:":8080"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"debug"`
	OriginHost  string `envconfig:"ORIGIN_HOST"`
	Hostname    string `envconfig:"HOSTNAME"  default:"localhost"`
	OriginKey   string `encovnfig:"ORIGIN_KEY" default:"x-bakery-origin-token"`
	OriginToken string `envconfig:"ORIGIN_TOKEN"`
	Logger      zerolog.Logger
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

	c.Logger = c.getLogger()

	tracer := c.Tracer.init(c.Logger)
	c.Client.init(tracer)

	return c, c.Propeller.init(tracer, c.Client.Timeout)
}

// IsLocalHost returns true if env is localhost
func (c Config) IsLocalHost() bool {
	if c.Hostname == "localhost" {
		return true
	}

	return false
}

// GetLogger generates a logger
func (c Config) getLogger() zerolog.Logger {
	level, err := zerolog.ParseLevel(c.LogLevel)
	if err != nil || level == zerolog.NoLevel {
		level = zerolog.DebugLevel
	}

	return zerolog.New(os.Stderr).
		With().
		Timestamp().
		Logger().
		Level(level)
}

//ValidateAuthHeader returns key,value or error if not set
func (c Config) ValidateAuthHeader() error {
	if c.IsLocalHost() {
		return nil
	}

	if c.OriginKey == "" || c.OriginToken == "" {
		return fmt.Errorf("Authentication not set.\nKey: %v,Value: %v", c.OriginKey, c.OriginToken)
	}

	return nil
}

//SetupMiddleware appends request logging context to use in your handler
func (c Config) SetupMiddleware() alice.Chain {
	chain := alice.New()
	chain = chain.Append(hlog.NewHandler(c.Logger))

	chain = chain.Append(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Str("raddr", r.RemoteAddr).
			Str("ref", r.Referer()).
			Str("ua", r.UserAgent()).
			Msg("served request")
	}))
	chain = chain.Append(hlog.RemoteAddrHandler("ip"))
	chain = chain.Append(hlog.RequestIDHandler("req_id", "Request-Id"))

	chain = chain.Append(c.authMiddleware())

	return chain
}

//authMiddlewareFrom appends an authentication middleware to your handler
func (c Config) authMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get(c.OriginKey) != c.OriginToken {
				hlog.FromRequest(r).UpdateContext(func(zctx zerolog.Context) zerolog.Context {
					return zctx.Interface("headers", r.Header).Str("error", "failed authenticating request")
				})

				http.Error(w, fmt.Sprintf("you must pass a valid api token as %q", c.OriginKey),
					http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
