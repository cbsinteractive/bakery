package main

import (
	"log"
	"net/http"

	"github.com/cbsinteractive/bakery/config"
	"github.com/cbsinteractive/bakery/handlers"
	"github.com/cbsinteractive/pkg/tracing"
)

func main() {
	c, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	middleware := c.SetupMiddleware()
	authMiddleware := c.AuthMiddlewareFrom(middleware)
	handler := authMiddleware.Then(handlers.LoadHandler(c))

	c.Logger.Info().Str("port", c.Listen).Msg("server starting")
	http.Handle("/", c.Client.Tracer.Handle(tracing.FixedNamer("bakery"), handler))
	if err := http.ListenAndServe(c.Listen, nil); err != nil {
		log.Fatal(err)
	}
}
