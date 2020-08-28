package xrayutil

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/aws/aws-xray-sdk-go/awsplugins/ec2"
	"github.com/aws/aws-xray-sdk-go/awsplugins/ecs"
	"github.com/aws/aws-xray-sdk-go/xray"
)

type XrayTracer struct {
	EnableAWSPlugins bool
	InfoLogFn        func(string, ...interface{})
}

func (t XrayTracer) Init() error {
	if t.InfoLogFn == nil {
		t.InfoLogFn = func(s string, i ...interface{}) {}
	}

	var (
		emitter xray.Emitter
		err     error
	)

	emitter, err = xray.NewDefaultEmitter(&net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 2000})
	if err != nil {
		return fmt.Errorf("creating xray emitter: %v", err)
	}

	if t.EnableAWSPlugins {
		ec2.Init()
		ecs.Init()
	}

	err = xray.Configure(xray.Config{
		ContextMissingStrategy: ctxMissingStrategy{logFn: t.InfoLogFn},
		Emitter:                emitter,
	})
	if err != nil {
		return fmt.Errorf("configuring xray: %v", err)
	}

	return nil
}

func (t XrayTracer) Handle(segmentNamer interface{ Name(host string) string }, h http.Handler) http.Handler {
	return xray.Handler(segmentNamer, h)
}

func (t XrayTracer) Client(c *http.Client) *http.Client {
	return xray.Client(c)
}

func (t XrayTracer) BeginSegment(ctx context.Context, name string) (context.Context, interface{ Close(error) }) {
	return xray.BeginSegment(ctx, name)
}

func (t XrayTracer) BeginSubsegment(ctx context.Context, name string) (context.Context, interface{ Close(error) }) {
	return xray.BeginSubsegment(ctx, name)
}

type ctxMissingStrategy struct {
	logFn func(string, ...interface{})
}

func (s ctxMissingStrategy) ContextMissing(v interface{}) {
	s.logFn("request sent without context: %v", v)
}
