package config

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/cbsinteractive/pkg/tracing"
	propeller "github.com/cbsinteractive/propeller-go/client"
	"github.com/google/go-cmp/cmp"
)

// env map is used for setting env vars for tests
type env map[string]string

// getConfig will return a config to use in tests based on provided values
func getConfig(listen, log, host, token string, c Client, t Tracer, p Propeller) Config {
	return Config{
		Listen:      ":8080",
		LogLevel:    "debug",
		Hostname:    "localhost",
		OriginToken: "",
		Client:      c,
		Tracer:      t,
		Propeller:   p,
	}
}

// getClientConfig will return a Cient config to use in tests based on provided values
func getClientConfig(c context.Context, t time.Duration, trace tracing.Tracer) Client {
	return Client{
		Context:    c,
		Timeout:    t,
		Tracer:     trace,
		HTTPClient: trace.Client(&http.Client{Timeout: t}),
	}
}

// getTracerConfig will return a Tracer config to use in tests based on provided values
func getTracerConfig(xray, plugin bool) Tracer {
	return Tracer{
		EnableXRay:        xray,
		EnableXRayPlugins: plugin,
	}
}

// getPropellerConfig will return a Propeller config to use in tests based on provided values
func getPropellerConfig(scheme, hostname, usr, pw string, t time.Duration, client HTTPClient) Propeller {
	var creds, host string

	if usr != "" && pw != "" {
		creds = fmt.Sprintf("%v:%v", usr, pw)
	}

	if scheme != "" {
		host = fmt.Sprintf("%v://%v", scheme, hostname)
	}

	return Propeller{
		Host:  host,
		Creds: creds,
		Client: propeller.Client{
			Auth: propeller.Auth{
				User: usr,
				Pass: pw,
				Host: hostname,
			},
			HostURL: &url.URL{
				Scheme: scheme,
				Host:   hostname,
			},
			Timeout:    t,
			HTTPClient: client,
		},
	}
}

func TestConfig_LoadConfig(t *testing.T) {
	noopTracer := tracing.NoopTracer{}
	disabledTraceConfig := getTracerConfig(false, false)

	defaultTime := time.Duration(5 * time.Second)
	defaultClientConfig := getClientConfig(nil, defaultTime, noopTracer)

	tests := []struct {
		name         string
		envs         []env
		expectConfig Config
		expectErr    bool
	}{
		{
			name: "When loading Config, if env vars not set, throw error for propeller creds for client",
			expectConfig: Config{
				Listen:      ":8080",
				LogLevel:    "debug",
				Hostname:    "localhost",
				OriginToken: "",
				Client:      defaultClientConfig,
				Tracer:      disabledTraceConfig,
				Propeller:   getPropellerConfig("", "", "", "", time.Duration(0*time.Second), nil),
			},
			expectErr: true,
		},
		{
			name: "When loading Config, if env vars are set for propeller, return config with propeller client",
			envs: []env{
				map[string]string{"BAKERY_PROPELLER_CREDS": "usr:pw"},
				map[string]string{"BAKERY_PROPELLER_HOST": "http://propeller.dev.com"},
			},
			expectConfig: Config{
				Listen:      ":8080",
				LogLevel:    "debug",
				Hostname:    "localhost",
				OriginToken: "",
				Client:      defaultClientConfig,
				Tracer:      disabledTraceConfig,
				Propeller:   getPropellerConfig("http", "propeller.dev.com", "usr", "pw", defaultTime, noopTracer.Client(&http.Client{})),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, env := range tc.envs {
				for k, v := range env {
					os.Setenv(k, v)
				}
			}

			got, err := LoadConfig()

			if err != nil && !tc.expectErr {
				t.Errorf("LoadConfig() didnt expect an error to be returned, got: %v", err)
				return
			} else if err == nil && tc.expectErr {
				t.Error("LoadConfig() expected an error, got nil")
				return
			}

			// Safe to ignore asthe unexported field is `err` field does not get triggered
			// during manual creation of the propeller Client config.
			ignore := cmpopts.IgnoreUnexported(propeller.Client{}, zerolog.Logger{})
			if !cmp.Equal(got, tc.expectConfig, ignore) {
				t.Errorf("Wrong Tracer config loaded\ngot %v\nexpected %v\ndiff: %v",
					got, tc.expectConfig, cmp.Diff(got, tc.expectConfig, ignore))
			}
		})
	}
}

func TestConfig_GetLogger(t *testing.T) {
	tests := []struct {
		name   string
		c      Config
		expect zerolog.Level
	}{
		{
			name: "if log level not set by env, GetLogger() will return default value",
			c: Config{
				LogLevel: "",
			},
			expect: zerolog.DebugLevel,
		},
		{
			name: "if log level not set by env, GetLogger() will return default value",
			c: Config{
				LogLevel: "panic",
			},
			expect: zerolog.PanicLevel,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.c.getLogger().GetLevel(); got != tc.expect {
				t.Errorf("Wrong log level \ngot %v\nexpected: %v", got, tc.expect)
			}
		})
	}
}

func TestConfig_Authentication(t *testing.T) {
	tests := []struct {
		name   string
		token  string
		expect bool
		c      Config
	}{
		{
			name:   "When localhost, return authentication true",
			token:  "",
			expect: true,
			c:      Config{Hostname: "localhost", OriginToken: ""},
		},
		{
			name:   "When localhost, return authentication true even if token is set",
			token:  "",
			expect: true,
			c:      Config{Hostname: "localhost", OriginToken: "sometoken"},
		},
		{
			name:   "When token is properly set and not localhost, return authentication true",
			token:  "authenticateMeImValid",
			expect: true,
			c:      Config{Hostname: "bakery.com", OriginToken: "authenticateMeImValid"},
		},
		{
			name:   "When token is not properly set and not localhost, return authentication false",
			token:  "",
			expect: false,
			c:      Config{Hostname: "bakery.com", OriginToken: "authenticateMeImValid"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.c.Authenticate(tc.token); got != tc.expect {
				t.Errorf("Wrong authenitcation response\ngot %v\nexpected: %v", got, tc.expect)
			}
		})
	}
}
