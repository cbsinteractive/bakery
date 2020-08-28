package tracing

import (
	"context"
	"net/http"
)

type nopCloser struct{}

func (nopCloser) Close(error) {}

type NoopTracer struct{}

func (NoopTracer) Init() error                        { return nil }
func (NoopTracer) Client(c *http.Client) *http.Client { return c }
func (NoopTracer) BeginSegment(context.Context, string) (context.Context, interface{ Close(error) }) {
	return context.Background(), nopCloser{}
}
func (NoopTracer) BeginSubsegment(context.Context, string) (context.Context, interface{ Close(error) }) {
	return context.Background(), nopCloser{}
}
func (NoopTracer) Handle(_ interface{ Name(host string) string }, h http.Handler) http.Handler {
	return h
}
