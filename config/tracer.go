package config

import (
	"github.com/cbsinteractive/pkg/tracing"
	"github.com/cbsinteractive/pkg/xrayutil"
	"github.com/rs/zerolog"
)

//Tracer holds configuration for initating the tracing of requests
type Tracer struct {
	EnableXRay        bool `envconfig:"ENABLE_XRAY" default:"false"`
	EnableXRayPlugins bool `envconfig:"ENABLE_XRAY_PLUGINS" default:"false"`
}

// init will set up the tracer to track clients requests
func (t *Tracer) init(logger zerolog.Logger) tracing.Tracer {
	var tracer tracing.Tracer

	if t.EnableXRay {
		tracer = xrayutil.XrayTracer{
			EnableAWSPlugins: t.EnableXRayPlugins,
			InfoLogFn:        logger.Info().Msgf,
		}
	} else {
		tracer = tracing.NoopTracer{}
	}

	err := tracer.Init()
	if err != nil {
		logger.Fatal().Msgf("initializing tracer: %v", err)
	}

	return tracer
}
