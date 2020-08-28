package tracing

import (
	"context"
	"net/http"
)

type Tracer interface {
	Init() error
	Client(*http.Client) *http.Client
	Handle(interface{ Name(host string) string }, http.Handler) http.Handler
	BeginSegment(ctx context.Context, name string) (context.Context, interface{ Close(error) })
	BeginSubsegment(ctx context.Context, name string) (context.Context, interface{ Close(error) })
}

func FixedNamer(n string) interface{ Name(host string) string } {
	return fixedSegmentNamer{fixedName: n}
}

type fixedSegmentNamer struct {
	fixedName string
}

func (n fixedSegmentNamer) Name(host string) string {
	return n.fixedName
}
