package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/cbsinteractive/pkg/tracing"
	propeller "github.com/cbsinteractive/propeller-go/client"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/rs/zerolog"
)

// env map is used for setting env vars for tests
type env map[string]string

// getConfig will return a config to use in tests based on provided values
func getDefaultConfig(auth bool, c Client, t Tracer, p Propeller) Config {
	return Config{
		Listen:      "8080",
		LogLevel:    "panic",
		OriginHost:  "http://localhost:8080",
		Hostname:    "hostname",
		OriginKey:   "x-bakery-origin-token",
		OriginToken: "authenticate-me",
		AuthEnabled: auth,
		Client:      c,
		Tracer:      t,
		Propeller:   p,
	}
}

// getClientConfig will return a Cient config to use in tests based on provided values
func getClientConfig(t time.Duration, trace tracing.Tracer) Client {
	return Client{
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
func getPropellerConfig(enabled bool, scheme, hostname, usr, pw string, t time.Duration, client HTTPClient) Propeller {
	var creds, host string

	if usr != "" && pw != "" {
		creds = fmt.Sprintf("%v:%v", usr, pw)
	}

	if scheme != "" {
		host = fmt.Sprintf("%v://%v", scheme, hostname)
	}

	return Propeller{
		Enabled: enabled,
		Host:    host,
		Creds:   creds,
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
	defaultClientConfig := getClientConfig(defaultTime, noopTracer)

	tests := []struct {
		name         string
		envs         []env
		expectConfig Config
		expectErr    bool
	}{
		{
			name: "When loading Config, if env vars not set, DO NOT throw error for propeller creds for client",
			expectConfig: Config{
				Listen:      ":8080",
				LogLevel:    "debug",
				Hostname:    "localhost",
				OriginKey:   "x-bakery-origin-token",
				OriginToken: "",
				Client:      defaultClientConfig,
				Tracer:      disabledTraceConfig,
				Propeller:   getPropellerConfig(false, "", "", "", "", time.Duration(0*time.Second), nil),
			},
		},
		{
			name: "When loading Config, if env vars are set for propeller, return config with propeller client",
			envs: []env{
				map[string]string{"BAKERY_PROPELLER_ENABLED": "true"},
				map[string]string{"BAKERY_PROPELLER_CREDS": "usr:pw"},
				map[string]string{"BAKERY_PROPELLER_HOST": "http://propeller.dev.com"},
			},
			expectConfig: Config{
				Listen:      ":8080",
				LogLevel:    "debug",
				Hostname:    "localhost",
				OriginKey:   "x-bakery-origin-token",
				OriginToken: "",
				Client:      defaultClientConfig,
				Tracer:      disabledTraceConfig,
				Propeller:   getPropellerConfig(true, "http", "propeller.dev.com", "usr", "pw", defaultTime, noopTracer.Client(&http.Client{})),
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

func TestConfig_ValidateAuthHeader(t *testing.T) {
	tests := []struct {
		name      string
		c         Config
		expectErr bool
	}{
		{
			name: "Don't throw error when authentication properly set",
			c:    Config{AuthEnabled: true, OriginToken: "sometoken", OriginKey: "somekey"},
		},
		{
			name: "Don't throw error when localhost",
			c:    Config{Hostname: "localhost"},
		},
		{
			name:      "Throw error when authenticaion token not set",
			c:         Config{AuthEnabled: true, OriginKey: "somekey"},
			expectErr: true,
		},
		{
			name:      "Throw error when authenticaion key not set",
			c:         Config{AuthEnabled: true, OriginToken: "sometoken"},
			expectErr: true,
		},
		{
			name:      "Throw error when authenticaion not set",
			c:         Config{AuthEnabled: true},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.c.ValidateAuthHeader()

			if err != nil && !tc.expectErr {
				t.Errorf("GetAuthHeader() got error did not expect error thrown")
			} else if err == nil && tc.expectErr {
				t.Errorf("GetAuthHeader() got no error expected error thrown")
			}
		})
	}
}

func TestConfig_Middleware(t *testing.T) {
	noopTracer := tracing.NoopTracer{}
	disabledTraceConfig := getTracerConfig(false, false)

	defaultTime := time.Duration(5 * time.Second)
	defaultClientConfig := getClientConfig(defaultTime, noopTracer)

	tests := []struct {
		name         string
		authtoken    string
		authEnabled  bool
		expectErr    string
		expectStatus int
	}{
		{
			name:         "when request is made with the correct auth token and auth is enabled",
			authtoken:    "authenticate-me",
			authEnabled:  true,
			expectStatus: 200,
		},
		{
			name:         "when request is made with no auth token and auth is disabled",
			expectStatus: 200,
		},
		{
			name:         "when request is made with bad auth token, expect a 403 and error message",
			authtoken:    "bad-auth-token",
			authEnabled:  true,
			expectStatus: 403,
			expectErr:    "you must pass a valid api token as \"x-bakery-origin-token\"\n",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatalf("could not create request got error: %v", err)
			}

			req.Header.Set("x-bakery-origin-token", tc.authtoken)
			rec := httptest.NewRecorder()

			c := getDefaultConfig(tc.authEnabled, defaultClientConfig, disabledTraceConfig, Propeller{})
			handler := c.SetupMiddleware().Then(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			handler.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.expectStatus {
				t.Errorf("Wrong status expected status %v; got %v", tc.expectStatus, res.StatusCode)
			}

			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(string(body), tc.expectErr) {
				t.Errorf("Wrong body returned\ngot %v\nexpected: %v\ndiff: %v",
					string(body), tc.expectErr, cmp.Diff(string(body), tc.expectErr))
			}
		})
	}
}
